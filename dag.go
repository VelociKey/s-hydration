package hydration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"sov.fleet/s-hydration/400-registry"
)

type GoPackage struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

func parseGoWork(goWorkPath string) ([]string, error) {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var workspaces []string
	inUse := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "use (" {
			inUse = true
			continue
		}
		if line == ")" && inUse {
			inUse = false
			continue
		}
		var path string
		if inUse {
			path = line
		} else if strings.HasPrefix(line, "use ") {
			path = strings.TrimSpace(strings.TrimPrefix(line, "use "))
		} else {
			continue
		}

		path = strings.Trim(path, `"' `)
		if idx := strings.Index(path, "//"); idx != -1 {
			path = strings.TrimSpace(path[:idx])
		}
		if path != "" {
			workspaces = append(workspaces, path)
		}
	}
	return workspaces, nil
}

func parseGoModModuleName(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module name not found in %s", goModPath)
}

func queryWorkspaceImports(goExe string, wsPath string) ([]string, error) {
	cmd := exec.Command(goExe, "list", "-json", "./...")
	cmd.Dir = wsPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var imports []string
	decoder := json.NewDecoder(bytes.NewReader(output))
	for decoder.More() {
		var pkg GoPackage
		if err := decoder.Decode(&pkg); err != nil {
			return nil, err
		}
		imports = append(imports, pkg.Imports...)
	}

	uniqueImports := make(map[string]bool)
	for _, imp := range imports {
		uniqueImports[imp] = true
	}
	var result []string
	for imp := range uniqueImports {
		result = append(result, imp)
	}
	return result, nil
}

type WsCacheEntry struct {
	LastModTime time.Time `json:"last_mod_time"`
	Imports     []string  `json:"imports"`
}

func getWorkspaceNewestModTime(dir string) (time.Time, error) {
	var newest time.Time
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "71000-build-harness" || name == "c0990-ephemeral-scratch" || name == "s-forge" || name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".go" || ext == ".mod" || ext == ".work" {
			info, err := d.Info()
			if err == nil {
				if info.ModTime().After(newest) {
					newest = info.ModTime()
				}
			}
		}
		return nil
	})
	return newest, err
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

	cachePath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "dependencies.cache.json")
	cache := make(map[string]WsCacheEntry)
	if data, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(data, &cache)
	}

	dependencyMap := make(map[string][]string)
	for wsAbs := range wsModuleMap {
		wsName := filepath.Base(wsAbs)

		newestTime, walkErr := getWorkspaceNewestModTime(wsAbs)
		
		var imports []string
		var cacheHit bool
		if walkErr == nil {
			if entry, ok := cache[wsName]; ok {
				if !newestTime.After(entry.LastModTime) {
					imports = entry.Imports
					cacheHit = true
				}
			}
		}

		if !cacheHit {
			var queryErr error
			imports, queryErr = queryWorkspaceImports(goExe, wsAbs)
			if queryErr != nil {
				dependencyMap[wsName] = []string{}
				continue
			}
			cache[wsName] = WsCacheEntry{
				LastModTime: newestTime,
				Imports:     imports,
			}
		}

		var deps []string
		seenDeps := make(map[string]bool)
		for _, imp := range imports {
			for otherMod, otherWsName := range moduleToWsName {
				if otherWsName == wsName {
					continue
				}
				if imp == otherMod || strings.HasPrefix(imp, otherMod+"/") {
					if !seenDeps[otherWsName] {
						seenDeps[otherWsName] = true
						deps = append(deps, otherWsName)
					}
				}
			}
		}
		dependencyMap[wsName] = deps
	}

	_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
	if cacheData, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, cacheData, 0644)
	}

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

	// Discover and append Dart package dependencies
	dartDeps, err := discoverDartPackages(projectRoot)
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
	grammarBindings, err := discoverGrammarBindings(projectRoot)
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

	paths := registry.ExtractLogicalPaths(irRoot)
	graph := make(map[string][]string)

	for key, val := range paths {
		wsName := strings.ReplaceAll(key, "_", "-")
		if strings.HasPrefix(wsName, "ws-") {
			wsName = strings.TrimPrefix(wsName, "ws-")
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
	return graph, nil
}

func findDownstreamNodes(graph map[string][]string, startNode string) map[string]bool {
	revGraph := make(map[string][]string)
	for node, deps := range graph {
		for _, dep := range deps {
			revGraph[dep] = append(revGraph[dep], node)
		}
	}

	downstream := make(map[string]bool)
	var visit func(string)
	visit = func(n string) {
		for _, next := range revGraph[n] {
			if !downstream[next] {
				downstream[next] = true
				visit(next)
			}
		}
	}
	visit(startNode)
	return downstream
}

func topologicalSort(graph map[string][]string, nodes map[string]bool) ([]string, error) {
	visited := make(map[string]int)
	var order []string
	var hasCycle bool

	var visit func(string)
	visit = func(n string) {
		if visited[n] == 1 {
			hasCycle = true
			return
		}
		if visited[n] == 2 {
			return
		}
		visited[n] = 1
		for _, dep := range graph[n] {
			if nodes[dep] || dep == n {
				visit(dep)
			}
		}
		visited[n] = 2
		order = append(order, n)
	}

	for n := range nodes {
		if visited[n] == 0 {
			visit(n)
		}
	}

	if hasCycle {
		return nil, fmt.Errorf("dependency cycle detected")
	}

	return order, nil
}

func discoverDartPackages(projectRoot string) (map[string][]string, error) {
	dartDeps := make(map[string][]string)
	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "pubspec.yaml" {
			wsName := filepath.Base(filepath.Dir(path))
			file, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer file.Close()

			var deps []string
			scanner := bufio.NewScanner(file)
			inDeps := false
			for scanner.Scan() {
				line := scanner.Text()
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}

				if strings.HasPrefix(line, "dependencies:") || strings.HasPrefix(line, "dev_dependencies:") {
					inDeps = true
					continue
				} else if inDeps && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
					inDeps = false
				}

				if inDeps {
					if strings.Contains(trimmed, ":") {
						parts := strings.SplitN(trimmed, ":", 2)
						depName := strings.TrimSpace(parts[0])
						if depName != "" {
							deps = append(deps, depName)
						}
					}
				}
			}
			var uniqueDeps []string
			keys := make(map[string]bool)
			for _, entry := range deps {
				if _, value := keys[entry]; !value {
					keys[entry] = true
					uniqueDeps = append(uniqueDeps, entry)
				}
			}
			if len(uniqueDeps) > 0 {
				dartDeps[wsName] = uniqueDeps
			}
		}
		return nil
	})
	return dartDeps, err
}

func discoverGrammarBindings(projectRoot string) (map[string][]string, error) {
	bindings := make(map[string][]string)
	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" {
				return filepath.SkipDir
			}
			return nil
		}

		name := d.Name()
		if strings.HasSuffix(name, ".webnf") {
			parts := strings.Split(name, ".")
			if len(parts) >= 3 {
				dslName := parts[len(parts)-2]
				bindings[name] = []string{dslName}
			}
		}
		return nil
	})
	return bindings, err
}
