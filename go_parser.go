package hydration

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
)

func parseGoWork(goWorkPath string) ([]string, error) {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}
	f, err := modfile.ParseWork(goWorkPath, content, nil)
	if err != nil {
		return nil, err
	}
	var workspaces []string
	for _, u := range f.Use {
		workspaces = append(workspaces, u.Path)
	}
	return workspaces, nil
}

func parseGoModModuleName(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	f, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return "", err
	}
	if f.Module == nil {
		return "", fmt.Errorf("module name not found in %s", goModPath)
	}
	return f.Module.Mod.Path, nil
}

func queryWorkspaceImports(goExe string, wsPath string) ([]string, error) {
	uniqueImports := make(map[string]bool)

	err := filepath.WalkDir(wsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" || name == "s-forge" || name == "86sref" || name == "s-hydrationcache" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
			fset := token.NewFileSet()
			fileAST, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err == nil {
				for _, spec := range fileAST.Imports {
					if spec.Path != nil {
						val, err := strconv.Unquote(spec.Path.Value)
						if err == nil {
							uniqueImports[val] = true
						}
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
	for imp := range uniqueImports {
		result = append(result, imp)
	}
	return result, nil
}
