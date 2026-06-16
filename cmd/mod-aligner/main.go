package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type WorkspaceInfo struct {
	Name       string
	Path       string // Absolute path
	ModuleName string
	Imports    []string // Imported local module names
}

func main() {
	slog.Info("Initializing mod_aligner...")

	// 1. Locate go.work
	projectRoot, err := findProjectRoot()
	if err != nil {
		slog.Error("Failed to find project root", "error", err)
		os.Exit(1)
	}
	slog.Info("Project root located", "path", projectRoot)

	goWorkPath := filepath.Join(projectRoot, "go.work")
	workspaces, err := parseGoWork(goWorkPath)
	if err != nil {
		slog.Error("Failed to parse go.work", "path", goWorkPath, "error", err)
		os.Exit(1)
	}
	slog.Info("Workspaces parsed from go.work", "count", len(workspaces))

	// 2. Resolve module names and paths
	moduleToWs := make(map[string]*WorkspaceInfo)
	pathToWs := make(map[string]*WorkspaceInfo)
	var wsList []*WorkspaceInfo

	for _, wsRel := range workspaces {
		wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
		goModPath := filepath.Join(wsAbs, "go.mod")
		if _, err := os.Stat(goModPath); err != nil {
			// Skip workspaces without go.mod
			continue
		}

		modName, err := parseGoModModuleName(goModPath)
		if err != nil {
			slog.Warn("Skipping workspace (unable to parse module name)", "path", wsAbs, "error", err)
			continue
		}

		wsInfo := &WorkspaceInfo{
			Name:       filepath.Base(wsAbs),
			Path:       wsAbs,
			ModuleName: modName,
		}
		moduleToWs[modName] = wsInfo
		pathToWs[wsAbs] = wsInfo
		wsList = append(wsList, wsInfo)
	}

	// 3. Scan imports for each workspace using AST
	for _, ws := range wsList {
		imports, err := scanWorkspaceImports(ws.Path, moduleToWs)
		if err != nil {
			slog.Error("Failed to scan workspace imports", "workspace", ws.Name, "error", err)
			os.Exit(1)
		}
		ws.Imports = imports
		slog.Debug("Workspace scan complete", "workspace", ws.Name, "local_imports", imports)
	}

	// 4. Topological Sort of workspaces based on dependencies
	sortedWs := topoSort(wsList, moduleToWs)

	slog.Info("Topological sorting of workspaces complete (best-effort)", "order", getWsNames(sortedWs))

	// 5. Align go.mod and run go mod tidy for each workspace
	goExe := filepath.Join(projectRoot, "00flow", "s-forge", "92000-external-toolchains", "go", "bin", "go.exe")
	if _, err := os.Stat(goExe); err != nil {
		goExe = "go" // fallback
	}

	for _, ws := range sortedWs {
		slog.Info("Aligning workspace modules", "workspace", ws.Name)
		err := alignGoMod(ws, moduleToWs)
		if err != nil {
			slog.Error("Failed to align go.mod", "workspace", ws.Name, "error", err)
			os.Exit(1)
		}

		if ws.Name == "s-forge" || ws.Name == "s-hydrationcache" || strings.HasSuffix(ws.Name, "-info") || strings.Contains(ws.Path, "93000-external-libraries") {
			slog.Info("Skipping tidy for registry/cache/metadata/external workspace", "workspace", ws.Name)
			continue
		}

		slog.Info("Tidying module", "workspace", ws.Name)
		err = tidyModule(goExe, ws.Path)
		if err != nil {
			slog.Error("Failed to tidy module", "workspace", ws.Name, "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Workspace module alignment completed successfully!")
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.work")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", fmt.Errorf("go.work not found")
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

func isDigitString(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func scanWorkspaceImports(wsPath string, localModules map[string]*WorkspaceInfo) ([]string, error) {
	importsMap := make(map[string]bool)
	err := filepath.WalkDir(wsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" {
				return filepath.SkipDir
			}
			// Skip 5-digit taxonomy directories (like 92000-external-toolchains) to avoid mock tests/issues
			if len(name) > 5 && name[5] == '-' && isDigitString(name[:5]) {
				return filepath.SkipDir
			}
			if path != wsPath {
				if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				return nil
			}
			for _, imp := range f.Imports {
				impPath := strings.Trim(imp.Path.Value, `"`)
				// Find matching local module prefix
				for modName := range localModules {
					if impPath == modName || strings.HasPrefix(impPath, modName+"/") {
						importsMap[modName] = true
						break
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var result []string
	for imp := range importsMap {
		result = append(result, imp)
	}
	return result, nil
}

func topoSort(wsList []*WorkspaceInfo, localModules map[string]*WorkspaceInfo) []*WorkspaceInfo {
	visited := make(map[string]int) // 0 = unvisited, 1 = visiting, 2 = visited
	var order []*WorkspaceInfo

	var visit func(ws *WorkspaceInfo)
	visit = func(ws *WorkspaceInfo) {
		if visited[ws.ModuleName] == 1 {
			return
		}
		if visited[ws.ModuleName] == 2 {
			return
		}
		visited[ws.ModuleName] = 1
		for _, depName := range ws.Imports {
			if depWs, ok := localModules[depName]; ok {
				visit(depWs)
			}
		}
		visited[ws.ModuleName] = 2
		order = append(order, ws)
	}

	for _, ws := range wsList {
		if visited[ws.ModuleName] == 0 {
			visit(ws)
		}
	}

	return order
}

func getWsNames(wsList []*WorkspaceInfo) []string {
	var names []string
	for _, ws := range wsList {
		names = append(names, ws.Name)
	}
	return names
}

func alignGoMod(ws *WorkspaceInfo, localModules map[string]*WorkspaceInfo) error {
	goModPath := filepath.Join(ws.Path, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	var outLines []string
	inRequire := false
	inReplace := false

	// Filter out existing local module declarations
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "require (" {
			inRequire = true
			outLines = append(outLines, line)
			continue
		}
		if trimmed == "replace (" {
			inReplace = true
			outLines = append(outLines, line)
			continue
		}
		if trimmed == ")" {
			inRequire = false
			inReplace = false
			outLines = append(outLines, line)
			continue
		}

		if inRequire {
			isLocal := false
			for modName := range localModules {
				if strings.HasPrefix(trimmed, modName+" ") || strings.HasPrefix(trimmed, modName+"\t") {
					isLocal = true
					break
				}
			}
			if isLocal {
				continue
			}
		} else if inReplace {
			isLocal := false
			for modName := range localModules {
				if strings.HasPrefix(trimmed, modName+" ") || strings.HasPrefix(trimmed, modName+"\t") || strings.HasPrefix(trimmed, modName+"=>") || strings.Contains(trimmed, modName+" =>") {
					isLocal = true
					break
				}
			}
			if isLocal {
				continue
			}
		} else {
			// Single-line require or replace
			if strings.HasPrefix(trimmed, "require ") {
				isLocal := false
				for modName := range localModules {
					if strings.Contains(trimmed, "require "+modName) {
						isLocal = true
						break
					}
				}
				if isLocal {
					continue
				}
			}
			if strings.HasPrefix(trimmed, "replace ") {
				isLocal := false
				for modName := range localModules {
					if strings.Contains(trimmed, "replace "+modName) {
						isLocal = true
						break
					}
				}
				if isLocal {
					continue
				}
			}
		}

		outLines = append(outLines, line)
	}

	// Insert correct aligned require and replace blocks
	var requireBlock []string
	var replaceBlock []string

	// Direct imports are required
	for _, depName := range ws.Imports {
		if depName == ws.ModuleName {
			continue
		}
		requireBlock = append(requireBlock, fmt.Sprintf("\t%s %s", depName, getModuleVersion(depName)))
	}

	// EVERY local workspace is replaced to support transitive offline resolution
	for depName, depWs := range localModules {
		if depName == ws.ModuleName {
			continue
		}

		// Calculate relative path
		relPath, err := filepath.Rel(ws.Path, depWs.Path)
		if err != nil {
			return err
		}
		relPathSlash := filepath.ToSlash(relPath)
		if !strings.HasPrefix(relPathSlash, ".") && !strings.HasPrefix(relPathSlash, "/") {
			relPathSlash = "./" + relPathSlash
		}

		replaceBlock = append(replaceBlock, fmt.Sprintf("\t%s => %s", depName, relPathSlash))
	}

	newContent := strings.Join(outLines, "\n")
	if len(requireBlock) > 0 {
		newContent += "\nrequire (\n" + strings.Join(requireBlock, "\n") + "\n)\n"
	}
	if len(replaceBlock) > 0 {
		newContent += "\nreplace (\n" + strings.Join(replaceBlock, "\n") + "\n)\n"
	}

	return os.WriteFile(goModPath, []byte(newContent), 0644)
}

func tidyModule(goExe string, wsPath string) error {
	cmd := exec.Command(goExe, "mod", "tidy")
	cmd.Dir = wsPath
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	
	// Set Windows creation flags to hide console window flashing
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("go mod tidy failed:\nstdout: %s\nstderr: %s\nerror: %w", stdout.String(), stderr.String(), err)
	}
	return nil
}

func getModuleVersion(moduleName string) string {
	idx := strings.LastIndex(moduleName, "/")
	if idx != -1 {
		suffix := moduleName[idx+1:]
		if len(suffix) > 1 && suffix[0] == 'v' {
			isNumber := true
			for _, r := range suffix[1:] {
				if r < '0' || r > '9' {
					isNumber = false
					break
				}
			}
			if isNumber {
				return suffix + ".0.0"
			}
		}
	}
	return "v0.0.0"
}
