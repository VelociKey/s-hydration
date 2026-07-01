package hydration

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sov.fleet/s-hydration/400-registry"
)

type StagedOutput struct {
	LocalPath  string
	GlobalPath string
}

func getWorkspacePath(harnessPath string) string {
	abs, err := filepath.Abs(harnessPath)
	if err != nil {
		abs = harnessPath
	}
	current := filepath.Dir(abs)
	for {
		if _, err := os.Stat(filepath.Join(current, "00001-workspace-prologue")); err == nil {
			return current
		}
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			if _, gitrootErr := os.Stat(filepath.Join(current, ".gitroot")); gitrootErr != nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	dir := filepath.Dir(abs)
	if filepath.Base(dir) == "71000-build-harness" {
		return filepath.Dir(dir)
	}
	return dir
}

func getNewestModTime(dir string) (time.Time, error) {
	var newest time.Time
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "71000-build-harness" || name == "c0990-ephemeral-scratch" || name == "s-forge" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".exe" {
			return nil
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest, err
}

func buildHarness(ctx context.Context, harnessPath string, targetName string, force bool, useStaging bool, localOnly bool, testBuild bool, distBuild bool, graph map[string][]string, wsPathMap map[string]string, rollbackOnFailure bool, useBazelTest bool, pkgDepsMap map[string]map[string][]string) ([]StagedOutput, bool, error) {
	contentBytes, err := os.ReadFile(harnessPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read workspace harness %s: %w", harnessPath, err)
	}

	harness, err := ParseWorkspaceHarness(string(contentBytes), harnessPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse workspace harness: %w", err)
	}

	wsPath := getWorkspacePath(harnessPath)
	wsName := filepath.Base(wsPath)

	if !harness.Facets.BuildEnabled {
		slog.Info("Workspace build disabled via prologue, skipping target compilation", "workspace", wsName)
		return nil, false, nil
	}

	slog.Info("Workspace harness verified", "workspace", wsName, "read_only", harness.Facets.ReadOnly, "classification", harness.Facets.FunctionalClassification)

	if localOnly {
		webnfPath := filepath.Join(wsPath, ".webnf")
		if _, err := os.Stat(webnfPath); err == nil {
			slog.Info("Reading flat metadata configuration file", "path", webnfPath)
			if content, err := os.ReadFile(webnfPath); err == nil {
				lines := strings.Split(string(content), "\n")
				var mockPath string
				var limits string
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
						continue
					}
					if strings.Contains(line, ":") {
						parts := strings.SplitN(line, ":", 2)
						key := strings.TrimSpace(parts[0])
						val := strings.Trim(strings.TrimSpace(parts[1]), `";`)
						if key == "mock_path" {
							mockPath = val
						} else if key == "isolation_limit" {
							limits = val
						}
					}
				}
				slog.Info("Local-only layout isolation dictates", "mock_path", mockPath, "isolation_limit", limits)
				if mockPath != "" {
					os.Setenv("MOCK_PATH", mockPath)
				}
			}
		}
	}

	purifier := NewPurifier()
	purifier.Register(&DynamicSynthesisMechanism{})

	var stagedOutputs []StagedOutput
	compiledAny := false

	for _, target := range harness.Targets {
		if target.Mode != ModeDynamic {
			continue
		}

		if targetName != "" && !strings.Contains(target.SourcePath, targetName) && !strings.Contains(target.OutputPath, targetName) {
			continue
		}

		baseName := target.OutputPath
		if baseName == "" {
			baseName = target.SourcePath
		}
		if idx := strings.LastIndex(baseName, "/"); idx != -1 {
			baseName = baseName[idx+1:]
		}
		if idx := strings.LastIndex(baseName, "\\"); idx != -1 {
			baseName = baseName[idx+1:]
		}
		if idx := strings.LastIndex(baseName, ":"); idx != -1 {
			baseName = baseName[idx+1:]
		}

		cat := target.Category
		if cat == "" {
			cat = "BINARY"
		}

		isExecutableCat := cat == "TOOLCHAIN" || cat == "BINARY" || cat == "CLI" || cat == "ACTOR"
		if runtime.GOOS == "windows" && isExecutableCat {
			if !strings.HasSuffix(strings.ToLower(baseName), ".wasm") && !strings.HasSuffix(strings.ToLower(baseName), ".exe") && !strings.HasSuffix(strings.ToLower(baseName), ".cmd") && !strings.HasSuffix(strings.ToLower(baseName), ".bat") {
				baseName = baseName + ".exe"
			}
		}

		loc, err := registry.GetTargetLocation(cat, "internal")
		if err != nil {
			slog.Warn("Failed to resolve location for category, falling back to 96000-internal-executables", "category", cat, "error", err)
			loc = "96000-internal-executables"
		}

		sforgeBase := `C:\aCogSpaceSeed\00flow\s-forge`
		if testBuild {
			sforgeBase = `C:\aCogSpaceSeed\00flow\s-hydration`
		}
		globalDir, err := registry.GetTargetPhysicalPath(sforgeBase, cat, "internal")
		if err != nil {
			globalDir, _ = registry.GetTargetPhysicalPath(sforgeBase, "BINARY", "internal")
		}
		
		if distBuild {
			target.Hardening = append(target.Hardening, "trimpath", "strip-symbols")
			globalDir = `C:\aCogSpaceSeed\00flow\s-distribution\81000-active-source\cmd\distribution-packager\src`
		}
		
		globalPath := filepath.Join(globalDir, baseName)
		localPath := globalPath
		if useStaging && !distBuild {
			localPath = filepath.Join(wsPath, loc, baseName)
		}

		if target.OutputPath != "" {
			if !filepath.IsAbs(target.OutputPath) {
				target.OutputPath = filepath.Join(wsPath, target.OutputPath)
			}
			if runtime.GOOS == "windows" && isExecutableCat {
				if !strings.HasSuffix(strings.ToLower(target.OutputPath), ".wasm") && !strings.HasSuffix(strings.ToLower(target.OutputPath), ".exe") && !strings.HasSuffix(strings.ToLower(target.OutputPath), ".cmd") && !strings.HasSuffix(strings.ToLower(target.OutputPath), ".bat") {
					target.OutputPath = target.OutputPath + ".exe"
				}
			}
			if distBuild {
				target.OutputPath = filepath.Join(globalDir, filepath.Base(target.OutputPath))
			}
			localPath = target.OutputPath
		} else {
			target.OutputPath = localPath
		}
		target.WorkDir = wsPath

		if !force {
			outInfo, err := os.Stat(localPath)
			if err == nil {
				newestSrcTime, walkErr := getNewestModTimeWithDeps(wsName, wsPath, graph, wsPathMap, pkgDepsMap)
				if walkErr == nil && !newestSrcTime.IsZero() {
					if outInfo.ModTime().After(newestSrcTime) {
						slog.Info("SKIPPING target compilation (up to date)", "target", target.SourcePath, "out", localPath)
						stagedOutputs = append(stagedOutputs, StagedOutput{LocalPath: localPath, GlobalPath: globalPath})
						continue
					}
				}
			}
		}

		slog.Info("Orchestrating internal rehydration for target", "src", target.SourcePath, "category", target.Category, "out", target.OutputPath)
		err = purifier.Execute(ctx, target)
		if err != nil {
			return nil, false, fmt.Errorf("internal synthesis failed for %s: %w", target.SourcePath, err)
		}
		
		var isDistributionWorkspace bool
		var targetShortName string
		if wsName == "s-webconduit" || wsName == "s-webconnect" {
			isDistributionWorkspace = true
			targetShortName = strings.TrimPrefix(wsName, "s-")
		}

		if distBuild && isDistributionWorkspace {
			distSrcDir := filepath.Join(`C:\aCogSpaceSeed\00flow\s-distribution\81000-active-source\cmd\distribution-packager\src`, targetShortName)
			_ = os.MkdirAll(distSrcDir, 0755)
			
			goModContent := fmt.Sprintf("module sov.nvelwraith/%s\n\ngo 1.26.3\n\nrequire (\n\tsov.fleet/quic-go v0.59.1\n)\n\nreplace (\n\tsov.fleet/quic-go => ../../../../00flow/s-forge/93000-external-libraries/quic-go\n)\n", targetShortName)
			_ = os.WriteFile(filepath.Join(distSrcDir, "go.mod.tmpl"), []byte(goModContent), 0644)
			
			hollowedGoContent := getHollowedGoContent(targetShortName)
			_ = os.WriteFile(filepath.Join(distSrcDir, fmt.Sprintf("%s.go", targetShortName)), []byte(hollowedGoContent), 0644)
			slog.Info("[Dist-Build] Generated hollowed Go source library inside s-distribution src directory.")
		}

		compiledAny = true
		stagedOutputs = append(stagedOutputs, StagedOutput{LocalPath: localPath, GlobalPath: globalPath})
	}

	if harness.Facets.TestEnabled && len(stagedOutputs) > 0 && !testBuild && !rollbackOnFailure {
		slog.Info("Running verification test suite for workspace", "workspace", harness.Name)
		projectRoot := filepath.Dir(filepath.Dir(wsPath))
		
		var cmd *exec.Cmd
		bcm := NewLocalCacheManager()
		if err := bcm.SetupCaches(harness.Name); err != nil {
			slog.Warn("Failed to setup isolated build caches for testing", "workspace", harness.Name, "error", err)
		}

		if useBazelTest {
			bazelExe := filepath.Join(projectRoot, "00flow", "s-forge", "92000-external-toolchains", "bazel", "bazel.exe")
			if _, err := os.Stat(bazelExe); err != nil {
				return nil, false, fmt.Errorf("bazel test compiler not found at %s", bazelExe)
			}
			bazelOut := bcm.GetEnvVars(harness.Name)["BAZEL_OUTPUT_BASE"]
			cmd = exec.Command(bazelExe, "--output_user_root="+bazelOut, "test", "--symlink_prefix=/", "--color=no", "//...")
			cmd.Dir = wsPath
			cmd.Env = os.Environ()
			for k, v := range bcm.GetEnvVars(harness.Name) {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
		} else {
			goExe := filepath.Join(projectRoot, "00flow", "s-forge", "92000-external-toolchains", "go", "bin", "go.exe")
			if _, err := os.Stat(goExe); err != nil {
				goExe = "go"
			}
			cmd = exec.Command(goExe, "test", "./...")
			cmd.Dir = wsPath
			cmd.Env = os.Environ()
			for k, v := range bcm.GetEnvVars(harness.Name) {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			slog.Error("Verification tests failed", "workspace", harness.Name, "stdout", stdout.String(), "stderr", stderr.String())
			return nil, false, fmt.Errorf("tests failed in %s: %w", harness.Name, err)
		}
		slog.Info("Verification tests passed successfully", "workspace", harness.Name)
	}

	return stagedOutputs, compiledAny, nil
}

func resolveWorkspaceHarness(wsDir string) (string, *WorkspaceFacets, error) {
	harnessPath := filepath.Join(wsDir, "71000-build-harness", "workspace.harness")
	if _, err := os.Stat(harnessPath); os.IsNotExist(err) {
		harnessPath = filepath.Join(wsDir, "workspace.harness")
	}
	if _, err := os.Stat(harnessPath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("workspace harness not found in %s", wsDir)
	}
	content, err := os.ReadFile(harnessPath)
	if err != nil {
		return "", nil, err
	}
	harness, err := ParseWorkspaceHarness(string(content), harnessPath)
	if err != nil {
		return "", nil, err
	}
	return harnessPath, &harness.Facets, nil
}

func getTransitiveDependencies(graph map[string][]string, startNode string) []string {
	visited := make(map[string]bool)
	var deps []string
	var traverse func(node string)
	traverse = func(node string) {
		for _, dep := range graph[node] {
			if !visited[dep] {
				visited[dep] = true
				deps = append(deps, dep)
				traverse(dep)
			}
		}
	}
	traverse(startNode)
	return deps
}

func getNewestModTimeWithDeps(wsName string, wsPath string, graph map[string][]string, wsPathMap map[string]string, pkgDepsMap map[string]map[string][]string) (time.Time, error) {
	newest, err := getNewestModTime(wsPath)
	if err != nil {
		return time.Time{}, err
	}

	transDeps := getTransitiveDependencies(graph, wsName)
	for _, dep := range transDeps {
		if depPath, ok := wsPathMap[dep]; ok {
			var depTime time.Time
			var scanErr error

			if specificPkgs, hasSpecific := pkgDepsMap[wsName][dep]; hasSpecific && len(specificPkgs) > 0 {
				goModTime, _ := getNewestModTime(filepath.Join(depPath, "go.mod"))
				depTime = goModTime

				modName, errMod := parseGoModModuleName(filepath.Join(depPath, "go.mod"))
				for _, pkg := range specificPkgs {
					if errMod == nil && strings.HasPrefix(pkg, modName) {
						relPkg := strings.TrimPrefix(pkg, modName)
						relPkg = strings.TrimPrefix(relPkg, "/")
						pkgFullDir := filepath.Join(depPath, filepath.FromSlash(relPkg))
						
						pkgTime, errPkg := getNewestModTime(pkgFullDir)
						if errPkg == nil && pkgTime.After(depTime) {
							depTime = pkgTime
						}
					} else {
						fullTime, _ := getNewestModTime(depPath)
						if fullTime.After(depTime) {
							depTime = fullTime
						}
					}
				}
			} else {
				depTime, scanErr = getNewestModTime(depPath)
			}
			
			if scanErr == nil && depTime.After(newest) {
				newest = depTime
			}
		}
	}
	return newest, nil
}
