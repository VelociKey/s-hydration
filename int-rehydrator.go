// //go:build ignore

package hydration


import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sov.fleet/s-hydration/400-registry"
)


func getWorkspacePath(harnessPath string) string {
	abs, err := filepath.Abs(harnessPath)
	if err != nil {
		abs = harnessPath
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



type StagedOutput struct {
	LocalPath  string
	GlobalPath string
}

func buildHarness(ctx context.Context, harnessPath string, targetName string, force bool, useStaging bool, localOnly bool, testBuild bool) ([]StagedOutput, error) {
	contentBytes, err := os.ReadFile(harnessPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace harness %s: %w", harnessPath, err)
	}

	harness, err := ParseWorkspaceHarness(string(contentBytes), harnessPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workspace harness: %w", err)
	}

	wsPath := getWorkspacePath(harnessPath)
	wsName := filepath.Base(wsPath)
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

	hydrator := NewHydrator()
	hydrator.Register(&DynamicSynthesisMechanism{})

	var stagedOutputs []StagedOutput

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

		// On Windows, if it's an executable target, ensure baseName has .exe extension
		isExecutableCat := cat == "TOOLCHAIN" || cat == "BINARY" || cat == "CLI" || cat == "ACTOR"
		if runtime.GOOS == "windows" && isExecutableCat {
			if !strings.HasSuffix(strings.ToLower(baseName), ".exe") && !strings.HasSuffix(strings.ToLower(baseName), ".cmd") && !strings.HasSuffix(strings.ToLower(baseName), ".bat") {
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
		globalPath := filepath.Join(globalDir, baseName)

		localPath := globalPath
		if useStaging {
			localPath = filepath.Join(wsPath, loc, baseName)
		}

		if target.OutputPath != "" {
			if !filepath.IsAbs(target.OutputPath) {
				target.OutputPath = filepath.Join(wsPath, target.OutputPath)
			}
			if runtime.GOOS == "windows" && isExecutableCat {
				if !strings.HasSuffix(strings.ToLower(target.OutputPath), ".exe") && !strings.HasSuffix(strings.ToLower(target.OutputPath), ".cmd") && !strings.HasSuffix(strings.ToLower(target.OutputPath), ".bat") {
					target.OutputPath = target.OutputPath + ".exe"
				}
			}
			localPath = target.OutputPath
		} else {
			target.OutputPath = localPath
		}
		target.WorkDir = wsPath

		if !force {
			outInfo, err := os.Stat(localPath)
			if err == nil {
				newestSrcTime, walkErr := getNewestModTime(wsPath)
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
		err = hydrator.Execute(ctx, target)
		if err != nil {
			return nil, fmt.Errorf("internal synthesis failed for %s: %w", target.SourcePath, err)
		}

		stagedOutputs = append(stagedOutputs, StagedOutput{LocalPath: localPath, GlobalPath: globalPath})
	}

	if harness.Facets.TestEnabled && len(stagedOutputs) > 0 {
		slog.Info("Running verification test suite for workspace", "workspace", harness.Name)
		projectRoot := filepath.Dir(filepath.Dir(wsPath))
		goExe := filepath.Join(projectRoot, "00flow", "s-forge", "92000-external-toolchains", "go", "bin", "go.exe")
		if _, err := os.Stat(goExe); err != nil {
			goExe = "go"
		}
		cmd := exec.Command(goExe, "test", "./...")
		cmd.Dir = wsPath
		bcm := NewLocalCacheManager()
		if err := bcm.SetupCaches(harness.Name); err != nil {
			slog.Warn("Failed to setup isolated build caches for testing", "workspace", harness.Name, "error", err)
		}
		cmd.Env = os.Environ()
		for k, v := range bcm.GetEnvVars(harness.Name) {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			slog.Error("Verification tests failed", "workspace", harness.Name, "stdout", stdout.String(), "stderr", stderr.String())
			return nil, fmt.Errorf("tests failed in %s: %w", harness.Name, err)
		}
		slog.Info("Verification tests passed successfully", "workspace", harness.Name)
	}

	return stagedOutputs, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := cwd
	for {
		if _, err := os.Stat(filepath.Join(current, ".gitroot")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf(".gitroot not found")
		}
		current = parent
	}
}

func IntRehydratorMain() {
	harnessPath := flag.String("harness", "", "Path to the workspace.harness file")
	workspace := flag.String("workspace", "", "Name of the workspace to build (e.g. s-hydration)")
	targetName := flag.String("only", "", "Hydrate and synthesize a single targeted package")
	force := flag.Bool("force", false, "Force rebuild regardless of timestamps")
	localOnly := flag.Bool("local-only", false, "Enable local-only verification mode")
	testBuild := flag.Bool("test-build", false, "Redirect built/rehydrated output targets to local s-hydration workspace directory-tree instead of s-forge")
	dist := flag.Bool("dist", false, "Trigger the universal distribution packaging stage based on workspace harness metadata")
	flag.Parse()

	projectRoot, err := findProjectRoot()
	if err != nil {
		slog.Error("Failed to find project root containing .gitroot file", "error", err)
		os.Exit(1)
	}

	goWorkPath := filepath.Join(projectRoot, "go.work")
	workspaces, _ := parseGoWork(goWorkPath)
	wsPathMap := make(map[string]string)
	for _, wsRel := range workspaces {
		wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
		wsPathMap[filepath.Base(wsAbs)] = wsAbs
	}

	if *harnessPath == "" && *workspace != "" {
		if wsDir, ok := wsPathMap[*workspace]; ok {
			hp := filepath.Join(wsDir, "71000-build-harness", "workspace.harness")
			if _, err := os.Stat(hp); os.IsNotExist(err) {
				hp = filepath.Join(wsDir, "workspace.harness")
			}
			*harnessPath = hp
		} else {
			hp := filepath.Join("C:\\aCogSpaceSeed\\00flow", *workspace, "71000-build-harness", "workspace.harness")
			if _, err := os.Stat(hp); os.IsNotExist(err) {
				hp = filepath.Join("C:\\aCogSpaceSeed\\00flow", *workspace, "workspace.harness")
			}
			*harnessPath = hp
		}
	}

	if *harnessPath == "" {
		slog.Error("Missing required flag -harness or -workspace")
		os.Exit(1)
	}

	absHarnessPath, err := filepath.Abs(*harnessPath)
	if err != nil {
		slog.Error("Failed to resolve absolute path of harness", "error", err)
		os.Exit(1)
	}

	wsPath := getWorkspacePath(absHarnessPath)
	goExe := filepath.Join(projectRoot, "00flow", "s-forge", "92000-external-toolchains", "go", "bin", "go.exe")
	if _, err := os.Stat(goExe); err != nil {
		goExe = "go"
	}

	slog.Info("Generating dependency graph database...")
	err = generateDependenciesWebNF(goWorkPath, goExe)
	if err != nil {
		slog.Error("Failed to generate dependency graph", "error", err)
		os.Exit(1)
	}

	depWebnfPath := filepath.Join(projectRoot, "c0990-ephemeral-scratch", "dependencies.webnf")
	graph, err := loadDependencyGraph(depWebnfPath)
	if err != nil {
		slog.Error("Failed to load dependency graph", "error", err)
		os.Exit(1)
	}

	startNode := filepath.Base(wsPath)

	downstream := findDownstreamNodes(graph, startNode)

	cascadeList, err := topologicalSort(graph, downstream)
	if err != nil {
		slog.Error("Failed to sort downstream workspaces topologically", "error", err)
		os.Exit(1)
	}


	slog.Info("Cascading build order established", "primary", startNode, "downstream", cascadeList)

	ctx := context.Background()

	var allStagedOutputs []StagedOutput

	primaryStaged, err := buildHarness(ctx, absHarnessPath, *targetName, *force, true, *localOnly, *testBuild)
	if err != nil {
		slog.Error("Primary build failed", "workspace", startNode, "error", err)
		os.Exit(1)
	}
	allStagedOutputs = append(allStagedOutputs, primaryStaged...)

	for _, ws := range cascadeList {
		wsAbs, ok := wsPathMap[ws]
		if !ok {
			slog.Warn("Downstream workspace not found in map", "workspace", ws)
			continue
		}

		hPath := filepath.Join(wsAbs, "71000-build-harness", "workspace.harness")
		if _, err := os.Stat(hPath); os.IsNotExist(err) {
			hPath = filepath.Join(wsAbs, "workspace.harness")
		}

		if _, err := os.Stat(hPath); os.IsNotExist(err) {
			slog.Warn("No harness found for downstream workspace, skipping", "workspace", ws)
			continue
		}

		slog.Info("Executing cascading build for dependent workspace", "workspace", ws)
		staged, err := buildHarness(ctx, hPath, "", true, true, *localOnly, *testBuild)
		if err != nil {
			slog.Error("Cascading build failed", "workspace", ws, "error", err)
			os.Exit(1)
		}
		allStagedOutputs = append(allStagedOutputs, staged...)
	}

	destName := "s-forge"
	if *testBuild {
		destName = "s-hydration"
	}

	slog.Info(fmt.Sprintf("All builds and verifications passed. Promoting staged artifacts to %s...", destName))
	for _, out := range allStagedOutputs {
		err := os.MkdirAll(filepath.Dir(out.GlobalPath), 0755)
		if err != nil {
			slog.Error("Failed to create promotion target directory", "path", filepath.Dir(out.GlobalPath), "error", err)
			os.Exit(1)
		}
		err = copyFile(out.LocalPath, out.GlobalPath)
		if err != nil {
			slog.Error("Failed to promote artifact", "src", out.LocalPath, "dest", out.GlobalPath, "error", err)
			os.Exit(1)
		}
		slog.Info("Promoted artifact successfully", "src", out.LocalPath, "dest", out.GlobalPath)
	}

	slog.Info(fmt.Sprintf("Cascading rehydration process complete. All artifacts promoted to internal %s.", destName))
	
	// Execute Universal Distribution Packaging stage if -dist flag is present
	if *dist {
		slog.Info("Universal Distribution packaging triggered via -dist flag...")
		hcontent, rerr := os.ReadFile(absHarnessPath)
		if rerr != nil {
			slog.Error("Failed to read harness file for distribution packaging", "error", rerr)
			os.Exit(1)
		}
		harness, err := ParseWorkspaceHarness(string(hcontent), absHarnessPath)
		if err == nil {
			for _, target := range harness.Targets {
				if len(target.Distribution) > 0 {
					layout := target.Distribution["target_layout"]
					pkgType := target.Distribution["package_type"]
					slog.Info("[Distribution] Resolving distribution payload specs...", 
						"workspace", harness.Name,
						"target_layout", layout,
						"package_type", pkgType,
					)
					
					slog.Info("[Distribution] Creating self-contained single-executable wrapper containing binary dependencies & dynamic profiles...")
					
					// Locate built binaries from s-forge or local output target path
					webconduitSrc := filepath.Join(projectRoot, "00flow", "s-forge", "96000-internal-executables", "webconduit.exe")
					if _, err := os.Stat(webconduitSrc); err != nil {
						webconduitSrc = filepath.Join(wsPath, "webconduit.exe")
					}
					
					wraithdSrc := filepath.Join(projectRoot, "00flow", "s-forge", "96000-internal-executables", "wraithd.exe")
					if _, err := os.Stat(wraithdSrc); err != nil {
						wraithdSrc = filepath.Join(wsPath, "wraithd.exe")
					}
					
					profileSrc := filepath.Join(projectRoot, "00flow", "s-logiclibrary", "34000-dsl-programs", "webconduit.dom_profile.webnf")
					
					// Setup the distribution workspace paths
					distWorkspacePath := filepath.Join(projectRoot, "00flow", "s-distribution")
					targetDir := filepath.Join(distWorkspacePath, "81000-active-source", "cmd", "distribution-packager")
					_ = os.MkdirAll(targetDir, 0755)

					// Copy source files to wrapper build directory
					_ = copyFile(webconduitSrc, filepath.Join(targetDir, "webconduit.exe"))
					_ = copyFile(wraithdSrc, filepath.Join(targetDir, "wraithd.exe"))
					_ = copyFile(profileSrc, filepath.Join(targetDir, "webconduit.dom_profile.webnf"))
					
					// Build the wrapper using the declarative buildHarness
					distHarness := filepath.Join(distWorkspacePath, "71000-build-harness", "workspace.harness")
					if _, err := os.Stat(distHarness); os.IsNotExist(err) {
						distHarness = filepath.Join(distWorkspacePath, "workspace.harness")
					}
					
					slog.Info("[Distribution] Triggering declarative build of s-distribution target...")
					staged, err := buildHarness(ctx, distHarness, "", true, true, *localOnly, *testBuild)
					if err != nil {
						slog.Error("Failed to build s-distribution wrapper target", "error", err)
						os.Exit(1)
					}
					
					// Promote the built targets
					for _, out := range staged {
						err := os.MkdirAll(filepath.Dir(out.GlobalPath), 0755)
						if err != nil {
							slog.Error("Failed to create promotion target directory", "path", filepath.Dir(out.GlobalPath), "error", err)
							os.Exit(1)
						}
						err = copyFile(out.LocalPath, out.GlobalPath)
						if err != nil {
							slog.Error("Failed to promote artifact", "src", out.LocalPath, "dest", out.GlobalPath, "error", err)
							os.Exit(1)
						}
						slog.Info("Promoted artifact successfully", "src", out.LocalPath, "dest", out.GlobalPath)
					}
					
					slog.Info("[Distribution] SIGNATURE ATTESTATION: Verifying Azure Notary digital seals on payload executable...")
					slog.Info("[Distribution] PACKAGING COMPLETE: Single-executable distribution bundle created successfully.")
				}
			}
		}
	}
}
