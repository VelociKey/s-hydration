package hydration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"sov.fleet/s-hydration/400-registry"
	"sov.fleet/s-latentlingua/02000-logic-libraries/ir"
	qdag "sov.fleet/s-qdag/81000-active-source/pkg/qdag"
)

type GoPackage struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
	Deps       []string `json:"Deps"`
}

type WsCacheEntry struct {
	LastModTime time.Time
	Imports     []string
}

func getWorkspaceNewestModTime(dir string) (time.Time, error) {
	// For Go workspaces, the dependency graph only changes if go.mod is modified.
	// Check the mod time of go.mod directly to avoid recursively walking the directory.
	info, err := os.Stat(filepath.Join(dir, "go.mod"))
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func generateDependenciesWebNF(goWorkPath string, goExe string) error {
	workspaces, err := parseGoWork(goWorkPath)
	if err != nil {
		return fmt.Errorf("failed to parse go.work: %w", err)
	}

	projectRoot := filepath.Dir(goWorkPath)

	wsModuleMap := make(map[string]string)
	moduleToWsName := make(map[string]string)

	for _, wsRel := range workspaces {
		wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
		goModPath := filepath.Join(wsAbs, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			modName, err := parseGoModModuleName(goModPath)
			if err == nil {
				wsModuleMap[wsAbs] = modName
				moduleToWsName[modName] = filepath.Base(wsAbs)
			}
		}
	}

	cachePath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "dependencies.cache.webnf")
	cache := make(map[string]WsCacheEntry)
	
	// Load WebNF cache
	if data, err := os.ReadFile(cachePath); err == nil {
		irRoot, err := registry.ParseConfigFile(string(data))
		if err == nil {
			paths := registry.ExtractLogicalPaths(irRoot)
			for key, val := range paths {
				if strings.HasPrefix(key, "ws_") {
					if strings.HasSuffix(key, "_last_mod_time") {
						wsNormal := strings.TrimPrefix(strings.TrimSuffix(key, "_last_mod_time"), "ws_")
						t, err := time.Parse(time.RFC3339Nano, val)
						if err == nil {
							entry := cache[wsNormal]
							entry.LastModTime = t
							cache[wsNormal] = entry
						}
					} else if strings.HasSuffix(key, "_imports") {
						wsNormal := strings.TrimPrefix(strings.TrimSuffix(key, "_imports"), "ws_")
						var imps []string
						if val != "" {
							for _, imp := range strings.Split(val, ",") {
								imp = strings.TrimSpace(imp)
								if imp != "" {
									imps = append(imps, imp)
								}
							}
						}
						entry := cache[wsNormal]
						entry.Imports = imps
						cache[wsNormal] = entry
					}
				}
			}
		}
	}

	type queryJob struct {
		wsAbs    string
		wsName   string
		wsNormal string
		newest   time.Time
	}

	var jobs []queryJob
	dependencyMap := make(map[string][]string)
	packageDependencyMap := make(map[string]map[string][]string)
	wsProgramsMap := make(map[string]string)

	// Pre-check cache to collect hits and list miss jobs
	for wsAbs := range wsModuleMap {
		wsName := filepath.Base(wsAbs)
		wsNormal := strings.ReplaceAll(wsName, "-", "_")

		newestTime, walkErr := getWorkspaceNewestModTime(wsAbs)
		
		var imports []string
		var cacheHit bool
		if walkErr == nil {
			if entry, ok := cache[wsNormal]; ok {
				if !newestTime.After(entry.LastModTime) {
					imports = entry.Imports
					cacheHit = true
				}
			}
		}

		if cacheHit {
			// Compute dependencies immediately for hits using prefix lookup
			var deps []string
			seenDeps := make(map[string]bool)
			pkgMap := make(map[string][]string)
			for _, imp := range imports {
				imp = strings.ReplaceAll(imp, "\\", "/")
				parts := strings.Split(imp, "/")
				for i := len(parts); i > 0; i-- {
					prefix := strings.Join(parts[:i], "/")
					if otherWsName, ok := moduleToWsName[prefix]; ok {
						if otherWsName != wsName {
							if !seenDeps[otherWsName] {
								seenDeps[otherWsName] = true
								deps = append(deps, otherWsName)
							}
							pkgMap[otherWsName] = append(pkgMap[otherWsName], imp)
						}
						break
					}
				}
			}
			dependencyMap[wsName] = deps
			packageDependencyMap[wsName] = pkgMap

			// Read programs immediately for hits
			harnessPath, _, err := resolveWorkspaceHarness(wsAbs)
			if err == nil {
				if hBytes, err := os.ReadFile(harnessPath); err == nil {
					if harness, err := ParseWorkspaceHarness(string(hBytes), harnessPath); err == nil {
						var targetsList []string
						for _, target := range harness.Targets {
							isGoProgram := target.Engine == "go-compiler" || target.Category == "TOOLCHAIN" || target.Category == "BINARY" || target.Category == "CLI" || target.Category == "ACTOR"
							if isGoProgram {
								outName := target.OutputPath
								if outName == "" {
									outName = filepath.Base(target.SourcePath)
								} else {
									outName = filepath.Base(outName)
								}
								outName = strings.TrimSuffix(outName, ".exe")
								targetsList = append(targetsList, fmt.Sprintf("%s:%s:%s", outName, target.SourcePath, target.Category))
							}
						}
						if len(targetsList) > 0 {
							wsProgramsMap[wsNormal] = strings.Join(targetsList, ",")
						}
					}
				}
			}
		} else {
			jobs = append(jobs, queryJob{
				wsAbs:    wsAbs,
				wsName:   wsName,
				wsNormal: wsNormal,
				newest:   newestTime,
			})
		}
	}

	// Execute cache-miss jobs in parallel
	if len(jobs) > 0 {
		type queryResult struct {
			job     queryJob
			imports []string
			err     error
		}

		resultsChan := make(chan queryResult, len(jobs))
		var g errgroup.Group
		g.SetLimit(8) // Bound concurrency to 8 worker goroutines

		for _, job := range jobs {
			j := job
			g.Go(func() error {
				imps, qErr := queryWorkspaceImports(goExe, j.wsAbs)
				resultsChan <- queryResult{job: j, imports: imps, err: qErr}
				return nil
			})
		}

		_ = g.Wait()
		close(resultsChan)

		for res := range resultsChan {
			if res.err != nil {
				dependencyMap[res.job.wsName] = []string{}
				continue
			}

			// Save to cache
			cache[res.job.wsNormal] = WsCacheEntry{
				LastModTime: res.job.newest,
				Imports:     res.imports,
			}

			// Compute dependencies
			var deps []string
			seenDeps := make(map[string]bool)
			pkgMap := make(map[string][]string)
			for _, imp := range res.imports {
				imp = strings.ReplaceAll(imp, "\\", "/")
				parts := strings.Split(imp, "/")
				for i := len(parts); i > 0; i-- {
					prefix := strings.Join(parts[:i], "/")
					if otherWsName, ok := moduleToWsName[prefix]; ok {
						if otherWsName != res.job.wsName {
							if !seenDeps[otherWsName] {
								seenDeps[otherWsName] = true
								deps = append(deps, otherWsName)
							}
							pkgMap[otherWsName] = append(pkgMap[otherWsName], imp)
						}
						break
					}
				}
			}
			dependencyMap[res.job.wsName] = deps
			packageDependencyMap[res.job.wsName] = pkgMap

			// Read programs
			harnessPath, _, err := resolveWorkspaceHarness(res.job.wsAbs)
			if err == nil {
				if hBytes, err := os.ReadFile(harnessPath); err == nil {
					if harness, err := ParseWorkspaceHarness(string(hBytes), harnessPath); err == nil {
						var targetsList []string
						for _, target := range harness.Targets {
							isGoProgram := target.Engine == "go-compiler" || target.Category == "TOOLCHAIN" || target.Category == "BINARY" || target.Category == "CLI" || target.Category == "ACTOR"
							if isGoProgram {
								outName := target.OutputPath
								if outName == "" {
									outName = filepath.Base(target.SourcePath)
								} else {
									outName = filepath.Base(outName)
								}
								outName = strings.TrimSuffix(outName, ".exe")
								targetsList = append(targetsList, fmt.Sprintf("%s:%s:%s", outName, target.SourcePath, target.Category))
							}
						}
						if len(targetsList) > 0 {
							wsProgramsMap[res.job.wsNormal] = strings.Join(targetsList, ",")
						}
					}
				}
			}
		}
	}

	// Save WebNF cache
	var cb strings.Builder
	cb.WriteString("; Hydration Cache Database\n")
	cb.WriteString("cache {\n")
	for wsNormal, entry := range cache {
		cb.WriteString(fmt.Sprintf("    ws_%s_last_mod_time = %q ;\n", wsNormal, entry.LastModTime.Format(time.RFC3339Nano)))
		cb.WriteString(fmt.Sprintf("    ws_%s_imports = %q ;\n", wsNormal, strings.Join(entry.Imports, ",")))
	}
	cb.WriteString("}\n")
	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	_ = os.WriteFile(cachePath, []byte(cb.String()), 0644)

	// Save Package Dependencies Graph
	pkgDepsPath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "package_dependencies.webnf")
	var psb strings.Builder
	psb.WriteString("; Authoritative Package Dependency Database\n")
	psb.WriteString("package_dependencies {\n")
	for ws, pkgMap := range packageDependencyMap {
		wsKey := strings.ReplaceAll(ws, "-", "_")
		if len(wsKey) > 0 && (wsKey[0] >= '0' && wsKey[0] <= '9') {
			wsKey = "ws_" + wsKey
		}
		for depWs, pkgs := range pkgMap {
			depKey := strings.ReplaceAll(depWs, "-", "_")
			if len(depKey) > 0 && (depKey[0] >= '0' && depKey[0] <= '9') {
				depKey = "ws_" + depKey
			}
			psb.WriteString(fmt.Sprintf("    %s__%s = %q ;\n", wsKey, depKey, strings.Join(pkgs, ",")))
		}
	}
	psb.WriteString("}\n\n")
	_ = os.WriteFile(pkgDepsPath, []byte(psb.String()), 0644)

	// Build dependencies.webnf
	var sb strings.Builder
	sb.WriteString("; Authoritative Dependency Graph Database\n")
	sb.WriteString("dependencies {\n")
	for ws, deps := range dependencyMap {
		key := strings.ReplaceAll(ws, "-", "_")
		if len(key) > 0 && (key[0] >= '0' && key[0] <= '9') {
			key = "ws_" + key
		}
		var formattedDeps []string
		for _, dep := range deps {
			d := strings.ReplaceAll(dep, "-", "_")
			if len(d) > 0 && d[0] >= '0' && d[0] <= '9' {
				d = "ws_" + d
			}
			formattedDeps = append(formattedDeps, d)
		}
		sb.WriteString(fmt.Sprintf("    %s = %q ;\n", key, strings.Join(formattedDeps, ",")))
	}
	sb.WriteString("}\n\n")

	// Append go_programs block
	sb.WriteString("go_programs {\n")
	for wsNormal, targetsVal := range wsProgramsMap {
		sb.WriteString(fmt.Sprintf("    ws_%s = %q ;\n", wsNormal, targetsVal))
	}
	sb.WriteString("}\n\n")

	// Discover and append Dart package dependencies
	dartDeps, err := discoverDartPackages(projectRoot, workspaces)
	if err == nil {
		sb.WriteString("dart_packages {\n")
		for ws, deps := range dartDeps {
			key := strings.ReplaceAll(ws, "-", "_")
			if len(key) > 0 && (key[0] >= '0' && key[0] <= '9') {
				key = "ws_" + key
			}
			var formattedDeps []string
			for _, dep := range deps {
				d := strings.ReplaceAll(dep, "-", "_")
				if len(d) > 0 && d[0] >= '0' && d[0] <= '9' {
					d = "ws_" + d
				}
				formattedDeps = append(formattedDeps, d)
			}
			sb.WriteString(fmt.Sprintf("    %s = %q ;\n", key, strings.Join(formattedDeps, ",")))
		}
		sb.WriteString("}\n\n")
	}

	// Discover and append Grammar bindings
	grammarBindings, err := discoverGrammarBindings(projectRoot, workspaces)
	if err == nil {
		sb.WriteString("grammar_bindings {\n")
		for file, grammars := range grammarBindings {
			key := strings.ReplaceAll(strings.ReplaceAll(file, "-", "_"), ".", "_")
			if len(key) > 0 && (key[0] >= '0' && key[0] <= '9') {
				key = "file_" + key
			}
			var formattedGrammars []string
			for _, g := range grammars {
				formattedGrammars = append(formattedGrammars, strings.ReplaceAll(g, "-", "_"))
			}
			sb.WriteString(fmt.Sprintf("    %s = %q ;\n", key, strings.Join(formattedGrammars, ",")))
		}
		sb.WriteString("}\n")
	}

	// Output logical clusters mapping
	sb.WriteString("\nclusters {\n")
	for ws := range dependencyMap {
		cluster := getClusterForWorkspace(ws)
		key := strings.ReplaceAll(ws, "-", "_")
		if len(key) > 0 && (key[0] >= '0' && key[0] <= '9') {
			key = "ws_" + key
		}
		sb.WriteString(fmt.Sprintf("    %s = %q ;\n", key, cluster))
	}
	sb.WriteString("}\n")

	destPath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "dependencies.webnf")
	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, []byte(sb.String()), 0644)
}

func getClusterForWorkspace(ws string) string {
	switch ws {
	case "s-latentlingua", "s-adk", "s-a2a", "s-sacp":
		return "grammar_languages"
	case "s-parallizer", "s-agentbox", "s-scorecard", "s-scoreboard":
		return "agent_sandbox_orchestration"
	case "s-actors", "s-mcp", "s-natives", "s-animus":
		return "active_actors"
	case "s-hydration", "s-builder", "s-forge", "s-seed", "s-fab-aides", "s-distribution", "s-hydrationcache":
		return "build_toolchain"
	case "s-introspection", "s-taxonomy-guard", "s-assurance", "s-trust-circle", "s-authorize", "s-hardware-bridge", "s-webconduit":
		return "governance_observability"
	default:
		return "other"
	}
}

func loadDependencyGraph(depWebnfPath string) (map[string][]string, error) {
	content, err := os.ReadFile(depWebnfPath)
	if err != nil {
		return nil, err
	}

	irRoot, err := registry.ParseConfigFile(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies.webnf: %w", err)
	}

	var depBlock *ir.Node
	for _, child := range irRoot.Children {
		if child.GetAttributeString("name") == "dependencies" {
			depBlock = child
			break
		}
	}

	graph := make(map[string][]string)
	if depBlock != nil {
		for _, child := range depBlock.Children {
			if child.Type == ir.NodeAttribute {
				wsName := child.GetAttributeString("name")
				wsName = strings.ReplaceAll(wsName, "_", "-")
				if strings.HasPrefix(wsName, "ws-") {
					wsName = strings.TrimPrefix(wsName, "ws-")
				}

				var val string
				if len(child.Children) > 0 && child.Children[0].Type == ir.NodeValue {
					val = child.Children[0].GetAttributeString("value")
				}

				var deps []string
				if val != "" {
					parts := strings.Split(val, ",")
					for _, part := range parts {
						part = strings.TrimSpace(part)
						if part != "" {
							d := strings.ReplaceAll(part, "_", "-")
							if strings.HasPrefix(d, "ws-") {
								d = strings.TrimPrefix(d, "ws-")
							}
							deps = append(deps, d)
						}
					}
				}
				graph[wsName] = deps
			}
		}
	}
	return graph, nil
}

func findDownstreamNodes(graph map[string][]string, startNode string) map[string]bool {
	d := qdag.NewDAG[struct{}]()
	for n := range graph {
		d.AddNode(n, struct{}{})
	}
	for n, deps := range graph {
		for _, dep := range deps {
			if _, ok := d.Nodes[dep]; !ok {
				d.AddNode(dep, struct{}{})
			}
			_ = d.AddEdge(dep, n, nil)
		}
	}
	return d.BlastRadius(startNode)
}

func topologicalSort(graph map[string][]string, nodes map[string]bool) ([]string, error) {
	d := qdag.NewDAGWithStrategy[struct{}](&qdag.DFSSortStrategy{})
	for n := range nodes {
		d.AddNode(n, struct{}{})
	}
	for n, deps := range graph {
		if !nodes[n] {
			continue
		}
		for _, dep := range deps {
			if nodes[dep] && dep != n {
				if _, ok := d.Nodes[dep]; !ok {
					d.AddNode(dep, struct{}{})
				}
				if err := d.AddEdge(dep, n, nil); err != nil {
					return nil, err
				}
			}
		}
	}
	return d.TopologicalSort()
}
