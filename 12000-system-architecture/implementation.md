# Sovereign Hydration Implementation Details

This document provides concrete source-code descriptions and command-line execution interfaces for both external toolchain hydration and internal Bazel build orchestration.

---

## 1. External Rehydration Implementation (`hydrator.go`)

The external rehydration engine is written in standard, dependency-free Go to guarantee compilation in a clean bootstrap environment.

### 1.1 The DFS Walker Engine
`IterativeDFWalk` drives all metabolic transformations without recursive stack frames:
```go
func (r *RehydrateV2) IterativeDFWalk(root string, walkFn func(path string, info os.FileInfo) (bool, error)) error {
	stack := []string{root}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		info, err := os.Lstat(curr)
		if err != nil {
			continue // Skip missing files/directories
		}

		skipDir, err := walkFn(curr, info)
		if err != nil {
			return err
		}
		if skipDir {
			continue
		}

		if info.IsDir() {
			files, err := os.ReadDir(curr)
			if err != nil {
				continue
			}
			for i := len(files) - 1; i >= 0; i-- {
				stack = append(stack, filepath.Join(curr, files[i].Name()))
			}
		}
	}
	return nil
}
```

### 1.2 Deterministic Topography Signer
`SignPrunedProduct` alphabetizes the file index of the pruned artifact and generates a cryptographic seal:
```go
func (r *RehydrateV2) SignPrunedProduct(path string) (string, error) {
	// If path points to a file, just compute its Blake3 directly
	info, err := os.Lstat(path)
	if err != nil { return "blake3:missing", err }
	
	h := blake3.New()
	if !info.IsDir() {
		f, err := os.Open(path)
		if err != nil { return "blake3:error", err }
		defer f.Close()
		io.Copy(h, f)
		return fmt.Sprintf("blake3:%x", h.Sum(nil)), nil
	}

	// For a directory, collect all files in sorted order of relative paths
	var files []string
	r.IterativeDFWalk(path, func(fpath string, info os.FileInfo) (bool, error) {
		if !info.IsDir() { files = append(files, fpath) }
		return false, nil
	})
	
	sort.Strings(files)

	for _, fpath := range files {
		rel, _ := filepath.Rel(path, fpath)
		h.Write([]byte(rel)) // Bind directory structure
		
		f, err := os.Open(fpath)
		if err == nil {
			io.Copy(h, f) // Bind byte content
			f.Close()
		}
	}
	return fmt.Sprintf("blake3:%x", h.Sum(nil)), nil
}
```

---

## 2. Command-Line Execution Interface

The rehydration engine is invoked via the sovereign Go toolchain and provides key diagnostic and execution controls:

```powershell
# 1. Execute full dry-run checklist (identify actionable/up-to-date targets without modifying disk)
go run hydrator.go -check

# 2. Force rehydrate and prune all artifacts, overwriting existing directories
go run hydrator.go -force

# 3. Hydrate and prune a single targeted package (e.g. 'trivy')
go run hydrator.go -only trivy
```

---

## 3. Internal (Bazel-Based) Rehydration Integration

The internal hydration layer allows Bazel to build compile pipelines without communicating with remote servers.

### 3.1 Target Type Classifications & Harness Designations

To structure air-gapped build redirection, we classify and designate Bazel hermetic targets under distinct **Target Types**. Each target type maps to a specific, cryptographically sealed directory within the `s-forge` authority silo:

| Target Type | Harness Designation (MODULE.bazel) | s-forge Target Location | Purpose & Mapping Guarantee |
| :--- | :--- | :--- | :--- |
| **`toolchain_sdk`** | `register_toolchains` / `go_download_sdk` | `00flow/s-forge/92000-external-toolchains/go` | Hermetic Go SDK runtime environment for target compilation. |
| **`bazel_rules`** | `local_path_override(module_name = "rules_go")` | `00flow/s-forge/92000-external-toolchains/bazel-rules/rules_go` | Offline package management rulesets for local compiler orchestration. |
| **`bazel_rules`** | `local_path_override(module_name = "rules_flutter")` | `00flow/s-forge/92000-external-toolchains/bazel-rules/rules_flutter` | Compilation rules and libraries for cross-compiling local surfaces. |
| **`external_binary`**| `binary_path` / `genrule` tool hooks | `00flow/s-forge/91000-external-binaries/bazel/buildifier.exe` | Formatters, linters, and helper tools run as localized subprocesses. |
| **`wasm_platform`** | `register_execution_platforms` | `00flow/s-forge/92000-external-toolchains/wasm/wasm-tools` | Hermetic platforms for WASM-GC runtime compilation. |

### 3.2 Decoupled Module Mappings & Harness Examples

Inside the workspace's root file (`MODULE.bazel`), external remote references are overridden with relative local hooks mapping back to these designated directories:

```bazel
# 1. Ruleset Designation Redirection
local_path_override(
    module_name = "rules_go",
    path = "00flow/s-forge/92000-external-toolchains/bazel-rules/rules_go",
)

local_path_override(
    module_name = "rules_flutter",
    path = "00flow/s-forge/92000-external-toolchains/bazel-rules/rules_flutter",
)

# 2. Local SDK Designation Redirection
local_path_override(
    module_name = "go_sdk",
    path = "00flow/s-forge/92000-external-toolchains/go",
)
```

This structural architecture guarantees that every external rule invocation in Bazel translates into a local-first AST evaluation, blocking all internet leaks during platform compilation.

---

## 4. Concurrent DAG and Targeted Test Slicing

### 4.1 Targeted Bazel Test Slicing (`computeBazelTestTargets`)
The targeted testing suite identifies and executes only verification test targets impacted by files modified in the active workspace.

```go
func computeBazelTestTargets(wsPath string, bazelExe string, bazelOut string) []string {
	if os.Getenv("DISABLE_TARGET_SLICING") == "true" {
		return []string{"//..."}
	}

	cmdGit := exec.Command("git", "status", "--porcelain")
	cmdGit.Dir = wsPath
	var gitOut bytes.Buffer
	cmdGit.Stdout = &gitOut
	if err := cmdGit.Run(); err != nil {
		return []string{"//..."}
	}

	modifiedPkgs := make(map[string]bool)
	lines := strings.Split(gitOut.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) < 4 { continue }
		
		filePath := line[3:]
		dir := filepath.Dir(filePath)
		dir = strings.ReplaceAll(dir, "\\", "/")
		if dir == "." { dir = "" }
		
		curr := dir
		for {
			if _, err := os.Stat(filepath.Join(wsPath, curr, "BUILD")); err == nil {
				modifiedPkgs["//"+curr] = true
				break
			}
			if _, err := os.Stat(filepath.Join(wsPath, curr, "BUILD.bazel")); err == nil {
				modifiedPkgs["//"+curr] = true
				break
			}
			if curr == "" || curr == "." { break }
			
			idx := strings.LastIndex(curr, "/")
			if idx == -1 { curr = "" } else { curr = curr[:idx] }
		}
	}

	if len(modifiedPkgs) == 0 {
		return []string{"//..."}
	}

	var pkgList []string
	for pkg := range modifiedPkgs {
		pkgList = append(pkgList, fmt.Sprintf("rdeps(//..., %s)", pkg))
	}
	queryExpr := fmt.Sprintf("kind(test, %s)", strings.Join(pkgList, " + "))

	cmdQuery := exec.Command(bazelExe, "--output_user_root="+bazelOut, "query", queryExpr)
	cmdQuery.Dir = wsPath
	var queryOut bytes.Buffer
	cmdQuery.Stdout = &queryOut
	if err := cmdQuery.Run(); err != nil {
		return []string{"//..."}
	}

	var targets []string
	scanner := bufio.NewScanner(&queryOut)
	for scanner.Scan() {
		target := strings.TrimSpace(scanner.Text())
		if target != "" {
			targets = append(targets, target)
		}
	}
	return targets
}
```

### 4.2 Parallel Rehydration Scheduling (`s-qdag` Orchestration)
The build cascade maps dependent modules to nodes and evaluates them concurrently:

```go
// Parallel DAG scheduler utilizing worker goroutines
graph := qdag.NewDAG()
// Populate graph...
for {
	ready := graph.GetReadyNodes()
	if len(ready) == 0 { break }
	
	var wg sync.WaitGroup
	for _, node := range ready {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			err := compileAndVerify(n)
			if err == nil {
				graph.CompleteNode(n)
			}
		}(node)
	}
	wg.Wait()
}
```
