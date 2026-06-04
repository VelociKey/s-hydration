package hydration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	dependencyMap := make(map[string][]string)
	for wsAbs := range wsModuleMap {
		wsName := filepath.Base(wsAbs)
		imports, err := queryWorkspaceImports(goExe, wsAbs)
		if err != nil {
			dependencyMap[wsName] = []string{}
			continue
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
	sb.WriteString("}\n")

	destPath := filepath.Join(projectRoot, "c0990-ephemeral-scratch", "dependencies.webnf")
	err = os.MkdirAll(filepath.Dir(destPath), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, []byte(sb.String()), 0644)
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
