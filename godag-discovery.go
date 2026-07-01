package hydration

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func discoverDartPackages(projectRoot string, workspaces []string) (map[string][]string, error) {
	dartDeps := make(map[string][]string)
	for _, wsRel := range workspaces {
		wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
		if _, err := os.Stat(wsAbs); err != nil {
			continue
		}
		_ = filepath.WalkDir(wsAbs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" || name == "s-forge" || name == "86sref" || name == "s-hydrationcache" {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() == "pubspec.yaml" {
				wsName := filepath.Base(filepath.Dir(path))
				content, err := os.ReadFile(path)
				if err != nil {
					return nil
				}

				var pubspec struct {
					Dependencies    map[string]interface{} `yaml:"dependencies"`
					DevDependencies map[string]interface{} `yaml:"dev_dependencies"`
				}
				if err := yaml.Unmarshal(content, &pubspec); err != nil {
					return nil
				}

				var uniqueDeps []string
				keys := make(map[string]bool)
				for dep := range pubspec.Dependencies {
					if !keys[dep] {
						keys[dep] = true
						uniqueDeps = append(uniqueDeps, dep)
					}
				}
				for dep := range pubspec.DevDependencies {
					if !keys[dep] {
						keys[dep] = true
						uniqueDeps = append(uniqueDeps, dep)
					}
				}
				if len(uniqueDeps) > 0 {
					dartDeps[wsName] = uniqueDeps
				}
			}
			return nil
		})
	}
	return dartDeps, nil
}

func discoverGrammarBindings(projectRoot string, workspaces []string) (map[string][]string, error) {
	bindings := make(map[string][]string)
	for _, wsRel := range workspaces {
		wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
		if _, err := os.Stat(wsAbs); err != nil {
			continue
		}
		_ = filepath.WalkDir(wsAbs, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" || name == "s-forge" || name == "86sref" || name == "s-hydrationcache" {
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
	}
	return bindings, nil
}
