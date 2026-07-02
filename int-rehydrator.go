package hydration

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"sov.fleet/s-hydration/400-registry"
	qdag "sov.fleet/s-qdag/81000-active-source/pkg/qdag"
)

func IntRehydratorMain() {
	workspaceFlag := flag.String("workspace", "", "Relative or absolute path to target workspace")
	targetName := flag.String("target", "", "Compile specific target binary instead of all dynamic assets in harness")
	force := flag.Bool("force", false, "Force rebuild of targeted workspaces bypassing ModTime skip logic")
	rollbackOnFailure := flag.Bool("rollback-on-failure", true, "Rollback workspace files to last successful build state on compilation or test failures")
	localOnly := flag.Bool("local-only", false, "Compile locally using workspace-bound isolated configurations (bypassing cloud environment variables)")
	testBuild := flag.Bool("test-build", false, "Test build output mode: promotions copy binaries to s-hydration target instead of s-forge")
	distBuild := flag.Bool("dist-build", false, "Distribution build mode: bundles binaries into packaging folder with symbols stripped")
	bazelTest := flag.Bool("bazel-test", false, "Run verification testing suite using Bazel test runner instead of native Go test compiler")
	wasm := flag.Bool("wasm", false, "Compile to WebAssembly target (forcing GOOS=js/wasip1 and GOARCH=wasm)")
	deploy := flag.Bool("deploy", false, "Deploy WASM targets to GCP Artifact Registry and stage via GCS")

	flag.Parse()

	if *wasm {
		os.Setenv("REHYDRATOR_WASM", "true")
	}

	if *workspaceFlag == "" {
		slog.Error("Target workspace flag (-workspace) is required")
		os.Exit(1)
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		slog.Error("Failed to identify project root", "error", err)
		os.Exit(1)
	}

	wsPath := *workspaceFlag
	goWorkPath := filepath.Join(projectRoot, "go.work")

	if !filepath.IsAbs(wsPath) {
		resolved := false
		if workspaces, err := parseGoWork(goWorkPath); err == nil {
			for _, wsRel := range workspaces {
				wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
				if filepath.Base(wsAbs) == wsPath || wsRel == wsPath {
					wsPath = wsAbs
					resolved = true
					break
				}
			}
		}
		if !resolved {
			wsPath = filepath.Clean(filepath.Join(projectRoot, wsPath))
		}
	}

	// Mark aCogSpaceSeed build=no (skip compilation for root seed workspace)
	if wsPath == projectRoot || filepath.Base(wsPath) == "aCogSpaceSeed" {
		slog.Info("Workspace build disabled for aCogSpaceSeed, skipping target compilation", "workspace", "aCogSpaceSeed")
		os.Exit(0)
	}

	absHarnessPath, facets, err := resolveWorkspaceHarness(wsPath)
	if err != nil {
		slog.Error("Failed to locate or parse workspace harness", "workspace", wsPath, "error", err)
		os.Exit(1)
	}

	if !facets.BuildEnabled {
		slog.Info("Workspace build disabled via prologue, skipping target compilation", "workspace", filepath.Base(wsPath))
		os.Exit(0)
	}

	goWorkPath = filepath.Join(projectRoot, "go.work")
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

	slog.Info("Generating Blake3 Build Invariant Attestation Registry...")
	if err = GenerateAttestationRegistry(projectRoot); err != nil {
		slog.Error("Failed to generate invariant attestation registry", "error", err)
		os.Exit(1)
	}

	depWebnfPath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "dependencies.webnf")
	graph, err := loadDependencyGraph(depWebnfPath)
	if err != nil {
		slog.Error("Failed to load dependency graph", "error", err)
		os.Exit(1)
	}

	wsPathMap := make(map[string]string)
	workspaces, err := parseGoWork(goWorkPath)
	if err == nil {
		for _, wsRel := range workspaces {
			wsAbs := filepath.Join(projectRoot, filepath.FromSlash(wsRel))
			wsName := filepath.Base(wsAbs)
			wsPathMap[wsName] = wsAbs
		}
	}

	wsNormalToActual := make(map[string]string)
	for ws := range wsPathMap {
		wsNormal := strings.ReplaceAll(ws, "-", "_")
		if len(wsNormal) > 0 && (wsNormal[0] >= '0' && wsNormal[0] <= '9') {
			wsNormal = "ws_" + wsNormal
		}
		wsNormalToActual[wsNormal] = ws
	}

	pkgDepsPath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "package_dependencies.webnf")
	pkgDepsMap := make(map[string]map[string][]string)
	if pkgData, err := os.ReadFile(pkgDepsPath); err == nil {
		irRoot, err := registry.ParseConfigFile(string(pkgData))
		if err == nil {
			paths := registry.ExtractLogicalPaths(irRoot)
			for key, val := range paths {
				parts := strings.SplitN(key, "__", 2)
				if len(parts) == 2 {
					wsActual := wsNormalToActual[parts[0]]
					depActual := wsNormalToActual[parts[1]]
					if wsActual != "" && depActual != "" {
						var pkgs []string
						if val != "" {
							for _, p := range strings.Split(val, ",") {
								p = strings.TrimSpace(p)
								if p != "" {
									pkgs = append(pkgs, p)
								}
							}
						}
						if pkgDepsMap[wsActual] == nil {
							pkgDepsMap[wsActual] = make(map[string][]string)
						}
						pkgDepsMap[wsActual][depActual] = pkgs
					}
				}
			}
		}
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

	var backupDir string
	if *rollbackOnFailure {
		var berr error
		backupDir, berr = backupWorkspace(wsPath)
		if berr != nil {
			slog.Warn("Failed to create workspace backup", "workspace", startNode, "error", berr)
		}
	}

	primaryStaged, _, err := buildHarness(ctx, absHarnessPath, *targetName, *force, true, *localOnly, *testBuild, *distBuild, graph, wsPathMap, *rollbackOnFailure, *bazelTest, pkgDepsMap)
	if err != nil {
		if *rollbackOnFailure && backupDir != "" {
			_ = restoreWorkspace(wsPath, backupDir)
		}
		slog.Error("Primary build failed", "workspace", startNode, "error", err)
		os.Exit(1)
	}
	if *rollbackOnFailure && backupDir != "" {
		_ = os.RemoveAll(backupDir)
	}
	allStagedOutputs = append(allStagedOutputs, primaryStaged...)

	// 1. Build s-qdag DAG for the cascadeList
	d := qdag.NewDAG[struct{}]()
	for _, ws := range cascadeList {
		d.AddNode(ws, struct{}{})
	}
	for _, ws := range cascadeList {
		if deps, ok := graph[ws]; ok {
			for _, dep := range deps {
				// Only add edges between nodes in cascadeList
				if _, ok := d.Nodes.Load(dep); ok && dep != ws {
					_ = d.AddEdge(dep, ws, nil)
				}
			}
		}
	}

	// 2. Perform concurrent topologically aligned rebuilds
	wsWinnow := qdag.NewWinnowState()
	completed := make(map[string]bool)
	var muCompleted sync.Mutex
	var muOutputs sync.Mutex

	concurrencyLimit := qdag.AuditWorkstation()
	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	for {
		muCompleted.Lock()
		readyNodes := d.GetReadyNodes(wsWinnow, completed)
		muCompleted.Unlock()

		if len(readyNodes) == 0 {
			muCompleted.Lock()
			doneCount := len(completed)
			muCompleted.Unlock()
			if doneCount == len(cascadeList) {
				break
			}
			slog.Error("Circular dependency detected or rehydration deadlock in cascade list")
			os.Exit(1)
		}

		for _, nodeID := range readyNodes {
			muCompleted.Lock()
			completed[nodeID] = false // mark active
			muCompleted.Unlock()

			wg.Add(1)
			go func(ws string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				wsAbs, ok := wsPathMap[ws]
				if !ok {
					slog.Warn("Downstream workspace not found in map", "workspace", ws)
					muCompleted.Lock()
					completed[ws] = true
					muCompleted.Unlock()
					return
				}

				hPath, _, _ := resolveWorkspaceHarness(wsAbs)
				if _, err := os.Stat(hPath); os.IsNotExist(err) {
					slog.Warn("No harness found for downstream workspace, skipping", "workspace", ws)
					muCompleted.Lock()
					completed[ws] = true
					muCompleted.Unlock()
					return
				}

				var bDir string
				if *rollbackOnFailure {
					var berr error
					bDir, berr = backupWorkspace(wsAbs)
					if berr != nil {
						slog.Warn("Failed to create workspace backup", "workspace", ws, "error", berr)
					}
				}

				slog.Info("Executing concurrent cascading build for dependent workspace", "workspace", ws)
				staged, _, err := buildHarness(ctx, hPath, "", false, true, *localOnly, *testBuild, *distBuild, graph, wsPathMap, *rollbackOnFailure, *bazelTest, pkgDepsMap)
				if err != nil {
					if *rollbackOnFailure && bDir != "" {
						_ = restoreWorkspace(wsAbs, bDir)
					}
					slog.Error("Cascading build failed", "workspace", ws, "error", err)
					os.Exit(1)
				}
				if *rollbackOnFailure && bDir != "" {
					_ = os.RemoveAll(bDir)
				}
				if *rollbackOnFailure {
					successDest := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "last_successful_build", ws)
					_ = os.RemoveAll(successDest)
					_ = os.MkdirAll(successDest, 0755)
					_ = copyDir(wsAbs, successDest)
				}

				muOutputs.Lock()
				allStagedOutputs = append(allStagedOutputs, staged...)
				muOutputs.Unlock()

				muCompleted.Lock()
				completed[ws] = true
				muCompleted.Unlock()
			}(nodeID)
		}
		wg.Wait()
	}

	if *rollbackOnFailure {
		successDest := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "last_successful_build", startNode)
		_ = os.RemoveAll(successDest)
		_ = os.MkdirAll(successDest, 0755)
		_ = copyDir(wsPath, successDest)
	}

	slog.Info("All builds and verifications passed. Promoting staged artifacts to s-forge...")
	for _, out := range allStagedOutputs {
		sClean := filepath.Clean(out.LocalPath)
		dClean := filepath.Clean(out.GlobalPath)
		if sClean == dClean {
			continue
		}
		_ = os.Remove(out.GlobalPath)
		_ = os.MkdirAll(filepath.Dir(out.GlobalPath), 0755)
		err := copyFile(out.LocalPath, out.GlobalPath)
		if err != nil {
			slog.Error("Failed to promote artifact", "src", out.LocalPath, "dest", out.GlobalPath, "error", err)
			os.Exit(1)
		}
		slog.Info("Promoted artifact successfully", "src", out.LocalPath, "dest", out.GlobalPath)
	}

	if *deploy {
		slog.Info("Executing GCP Deployment capability run...")
		err := DeployWasmArtifacts(context.Background(), projectRoot, allStagedOutputs)
		if err != nil {
			slog.Error("GCP Deployment failed", "error", err)
			os.Exit(1)
		}
		slog.Info("GCP Deployment completed successfully")
	}

	cleanLocalWorkstationExecutables(projectRoot)
}
