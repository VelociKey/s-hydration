package main

import (
	"bytes"
	"crypto/sha512"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ArtifactRecord represents a parsed line in the sbom_external_artifact.webnf database.
type ArtifactRecord struct {
	Name         string
	Version      string
	OriginalHash string
	PrunedHash   string
	OriginalSize int64
	ExpandedSize int64
	PrunedSize   int64
	Timestamp    string
}

// RehydrateV2 represents the V2 external hydration and metabolic pruning engine.
type RehydrateV2 struct{}

func main() {
	checkFlag := flag.Bool("check", false, "Execute full dry-run checklist")
	forceFlag := flag.Bool("force", false, "Force rehydrate and prune all artifacts")
	onlyFlag := flag.String("only", "", "Hydrate and prune a single targeted package")

	flag.Parse()

	// Default to check if no flags provided
	if !*checkFlag && !*forceFlag && *onlyFlag == "" {
		*checkFlag = true
	}

	r := &RehydrateV2{}
	err := r.Run(*checkFlag, *forceFlag, *onlyFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run executes the hydration check or pruning loop based on input flags.
func (r *RehydrateV2) Run(check, force bool, only string) error {
	sbomPath := `C:\aCogSpaceSeed\00flow\s-forge\90100-rehydration-seed\sbom_external_artifact.webnf`
	records, err := r.parseSBOM(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to parse SBOM: %w", err)
	}

	fmt.Printf("Loaded %d external artifact records from SBOM.\n", len(records))

	// Filter down selection based on cli arguments
	var selectedRecords []ArtifactRecord
	for _, rec := range records {
		if only != "" && rec.Name != only {
			continue
		}
		selectedRecords = append(selectedRecords, rec)
	}

	// Determine dynamic execution plan: bootstrap sorting & domain-authority QUIC groupings
	r.determineExecutionPlan(selectedRecords)

	fmt.Println(strings.Repeat("-", 80))

	actionableCount := 0
	okCount := 0
	missingCount := 0
	errorCount := 0

	for _, rec := range selectedRecords {
		physPath := getPhysicalPath(rec.Name)
		if physPath == "" {
			fmt.Printf("[SKIPPED] No physical path mapping for %s\n", rec.Name)
			continue
		}

		_, err := os.Lstat(physPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("[MISSING] %s at path: %s\n", rec.Name, physPath)
				missingCount++
				continue
			}
			fmt.Printf("[ERROR] Failed to stat %s: %v\n", rec.Name, err)
			errorCount++
			continue
		}

		isBlake3 := strings.HasPrefix(rec.PrunedHash, "blake3:")

		// Run dry-run checklist / actual prune
		hash, size, err := r.PruneAndHash(physPath, force, isBlake3)
		if err != nil {
			fmt.Printf("[ERROR] Failed to process %s: %v\n", rec.Name, err)
			errorCount++
			continue
		}

		if hash == rec.PrunedHash && size == rec.PrunedSize {
			fmt.Printf("[OK] %s is up-to-date\n", rec.Name)
			fmt.Printf("     Path: %s\n", physPath)
			fmt.Printf("     Hash: %s, Size: %d bytes\n", hash, size)
			okCount++
		} else {
			actionableCount++
			if force {
				// We forced the prune, let's re-verify
				newHash, newSize, err := r.PruneAndHash(physPath, false, isBlake3)
				if err != nil {
					fmt.Printf("[ERROR] Verification failed after force pruning %s: %v\n", rec.Name, err)
					errorCount++
					continue
				}
				if newHash == rec.PrunedHash && newSize == rec.PrunedSize {
					fmt.Printf("[PRUNED] %s successfully pruned and verified!\n", rec.Name)
					fmt.Printf("         Hash: %s, Size: %d bytes\n", newHash, newSize)
				} else {
					fmt.Printf("[MISMATCH] %s pruning completed but hash still mismatches!\n", rec.Name)
					fmt.Printf("           Expected Hash: %s, Size: %d bytes\n", rec.PrunedHash, rec.PrunedSize)
					fmt.Printf("           Actual Hash:   %s, Size: %d bytes\n", newHash, newSize)
					errorCount++
				}
			} else {
				fmt.Printf("[ACTIONABLE] %s has mismatch\n", rec.Name)
				fmt.Printf("             Path:          %s\n", physPath)
				fmt.Printf("             SBOM Hash:     %s, Size: %d bytes\n", rec.PrunedHash, rec.PrunedSize)
				fmt.Printf("             Current Hash:  %s, Size: %d bytes\n", hash, size)
			}
		}
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Summary:\n")
	fmt.Printf("  OK:         %d\n", okCount)
	fmt.Printf("  ACTIONABLE: %d\n", actionableCount)
	fmt.Printf("  MISSING:    %d\n", missingCount)
	fmt.Printf("  ERRORS:     %d\n", errorCount)

	return nil
}

// determineExecutionPlan performs selection, priority ordering, and QUIC cohort grouping.
func (r *RehydrateV2) determineExecutionPlan(records []ArtifactRecord) {
	var bootstrapCohort []ArtifactRecord
	var sequentialLargeCohort []ArtifactRecord
	parallelCohortGroups := make(map[string][]ArtifactRecord)

	for _, rec := range records {
		// 1. Separation by priority and size
		if rec.Name == "go-sdk-green-tea" || rec.Name == "blake3" {
			bootstrapCohort = append(bootstrapCohort, rec)
			continue
		}

		if rec.OriginalSize > 100*1024*1024 { // >100MB
			sequentialLargeCohort = append(sequentialLargeCohort, rec)
			continue
		}

		// Parallel Grouping by Domain Authority
		dlURL := getDownloadURL(rec.Name)
		if dlURL == "" {
			parallelCohortGroups["local-placeholder"] = append(parallelCohortGroups["local-placeholder"], rec)
			continue
		}

		parsed, err := url.Parse(dlURL)
		if err != nil {
			parallelCohortGroups["unknown"] = append(parallelCohortGroups["unknown"], rec)
			continue
		}
		host := parsed.Host
		parallelCohortGroups[host] = append(parallelCohortGroups[host], rec)
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("             DYNAMIC ASSIMILATION & REHYDRATION EXECUTION PLAN")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Println("PHASE 1: Mandatory Front-End Bootstrap (Sequential Priority)")
	for _, rec := range bootstrapCohort {
		fmt.Printf("  -> Ingest and seal: %-25s [Size: %s]\n", rec.Name, formatBytes(rec.OriginalSize))
	}
	if len(bootstrapCohort) == 0 {
		fmt.Println("  (No bootstrap records in current scope)")
	}

	fmt.Println("\nPHASE 2: Massive Payload Ingestions (Sequential Network Isolation)")
	for _, rec := range sequentialLargeCohort {
		fmt.Printf("  -> Ingest and seal: %-25s [Size: %s]\n", rec.Name, formatBytes(rec.OriginalSize))
	}
	if len(sequentialLargeCohort) == 0 {
		fmt.Println("  (No massive sequential payloads in current scope)")
	}

	fmt.Println("\nPHASE 3: Cohort Connection Multiplexing (HTTP/3 QUIC Streams)")
	// Sort hosts for deterministic plan printout
	var hosts []string
	for host := range parallelCohortGroups {
		hosts = append(hosts, host)
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		fmt.Printf("  Host Authority Context: [%s] (Persistent UDP/QUIC Connection Established)\n", host)
		for _, rec := range parallelCohortGroups[host] {
			fmt.Printf("    * Multiplexed HTTP/3 stream: %-20s [Size: %s]\n", rec.Name, formatBytes(rec.OriginalSize))
		}
	}
	if len(parallelCohortGroups) == 0 {
		fmt.Println("  (No parallel cohorts in current scope)")
	}
	fmt.Println(strings.Repeat("=", 80))
}

// formatBytes returns a human-readable representation of sizes.
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseSBOM parses the SBOM Wirth-style WebNF format database file.
func (r *RehydrateV2) parseSBOM(path string) ([]ArtifactRecord, error) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(contentBytes), "\n")
	var records []ArtifactRecord

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if line == "SBOM-V2" || strings.Contains(line, "AAIF-GreenTea") {
			continue
		}
		if !strings.Contains(line, "|") {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}

		origSize, _ := strconv.ParseInt(parts[4], 10, 64)

		var expandedSize int64
		var prunedSize int64
		var timestamp string

		if len(parts) >= 8 {
			// Backwards compatibility with 8 columns (Expanded size included):
			// Name|Version|OrigHash|PrunedHash|OrigSize|ExpandedSize|PrunedSize|Timestamp
			expandedSize, _ = strconv.ParseInt(parts[5], 10, 64)
			prunedSize, _ = strconv.ParseInt(parts[6], 10, 64)
			timestamp = parts[7]
		} else {
			// Legacy 7 columns:
			// Name|Version|OrigHash|PrunedHash|OrigSize|PrunedSize|Timestamp
			prunedSize, _ = strconv.ParseInt(parts[5], 10, 64)
			timestamp = parts[6]
		}

		rec := ArtifactRecord{
			Name:         parts[0],
			Version:      parts[1],
			OriginalHash: parts[2],
			PrunedHash:   parts[3],
			OriginalSize: origSize,
			ExpandedSize: expandedSize,
			PrunedSize:   prunedSize,
			Timestamp:    timestamp,
		}
		records = append(records, rec)
	}

	return records, nil
}

// IterativeDFWalk executes a true non-recursive iterative DFS walk over path using a heap stack.
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

// PruneAndHash prunes un-needed items and calculates sizes and hashes.
func (r *RehydrateV2) PruneAndHash(path string, force bool, isBlake3 bool) (string, int64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "sha512:missing", 0, err
	}

	if !info.IsDir() {
		// Single file
		if isBlake3 {
			h, err := blake3Hash(path)
			if err != nil {
				return "", 0, err
			}
			return h, info.Size(), nil
		} else {
			h := sha512.New()
			f, err := os.Open(path)
			if err != nil {
				return "", 0, err
			}
			defer f.Close()
			_, err = io.Copy(h, f)
			if err != nil {
				return "", 0, err
			}
			return fmt.Sprintf("sha512:%x", h.Sum(nil)), info.Size(), nil
		}
	}

	// For directories
	var remainingFiles []string
	var totalSize int64

	err = r.IterativeDFWalk(path, func(curr string, currInfo os.FileInfo) (bool, error) {
		shouldPrune, isDirPrune := r.shouldPrune(curr, currInfo)
		if shouldPrune {
			if force {
				if isDirPrune {
					_ = os.RemoveAll(curr)
				} else {
					_ = os.Remove(curr)
				}
			}
			// Skip pushing children
			return true, nil
		}

		if !currInfo.IsDir() {
			remainingFiles = append(remainingFiles, curr)
			totalSize += currInfo.Size()
		}
		return false, nil
	})

	if err != nil {
		return "", 0, err
	}

	// Deterministic alphabetical sorting of relative paths
	sort.Strings(remainingFiles)

	h := sha512.New()
	for _, fpath := range remainingFiles {
		rel, err := filepath.Rel(path, fpath)
		if err != nil {
			continue
		}
		// Standardize relative paths on forward slashes to ensure hash determinism across OSes
		rel = filepath.ToSlash(rel)
		h.Write([]byte(rel))

		f, err := os.Open(fpath)
		if err == nil {
			_, _ = io.Copy(h, f)
			f.Close()
		}
	}

	return fmt.Sprintf("sha512:%x", h.Sum(nil)), totalSize, nil
}

// shouldPrune defines the metabolic and platform-specific pruning criteria.
func (r *RehydrateV2) shouldPrune(path string, info os.FileInfo) (bool, bool) {
	name := strings.ToLower(info.Name())

	if info.IsDir() {
		// 1. Metabolic Folder Pruner
		metabolicFolders := []string{
			"test", "tests", "docs", "doc", "examples", "example",
			"samples", "sample", "site", "tutorial", "tutorials",
			"website", "benchmarks", "benchmark",
		}
		for _, f := range metabolicFolders {
			if name == f {
				return true, true
			}
		}

		// 2. Deep Platform Foreign OS Directories
		foreignOS := []string{
			"linux", "darwin", "debian", "ubuntu", "freebsd",
			"android", "ios", "macos", "solaris",
		}
		for _, osName := range foreignOS {
			if name == osName {
				// unless the folder explicitly contains "windows"
				if !strings.Contains(name, "windows") {
					return true, true
				}
			}
		}

		// 3. Deep Platform Foreign Arch Directories
		foreignArch := []string{
			"arm64", "armv7", "arm", "386", "x86", "ppc", "mips", "s390x",
		}
		for _, archName := range foreignArch {
			if name == archName {
				// unless the folder explicitly contains "amd64" or "x64"
				if !strings.Contains(name, "amd64") && !strings.Contains(name, "x64") {
					return true, true
				}
			}
		}

		return false, false
	} else {
		// Trust Exemption: preserve files with "license" (case-insensitive) in their names
		if strings.Contains(name, "license") {
			return false, false
		}

		// 1. Metabolic File Pruner
		if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".pdf") {
			return true, false
		}

		// 2. Deep Platform Foreign Executables & Shared Libraries on Windows hosts
		if strings.HasSuffix(name, ".sh") || strings.HasSuffix(name, ".bash") ||
			strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".dylib") {
			return true, false
		}

		return false, false
	}
}

// blake3Hash calculates the BLAKE3 hash of a file using b3sum.exe.
func blake3Hash(filePath string) (string, error) {
	cmd := exec.Command(`C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\blake3\b3sum.exe`, filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run b3sum.exe: %w", err)
	}
	fields := strings.Fields(out.String())
	if len(fields) < 1 {
		return "", fmt.Errorf("invalid b3sum output: %s", out.String())
	}
	return "blake3:" + fields[0], nil
}

// getPhysicalPath resolves the s-forge authority target paths.
func getPhysicalPath(name string) string {
	switch name {
	case "bazel-rules-flutter":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\bazel-rules\rules_flutter`
	case "bazel-rules-go":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\bazel-rules\rules_go`
	case "bazelisk":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\bazel\bazelisk.exe`
	case "blake3":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\blake3`
	case "buildifier":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\bazel\buildifier.exe`
	case "firebase-tools":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\firebase`
	case "flutter-sdk-firehorse":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\flutter`
	case "gcloud-sdk":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\gcloud`
	case "gh-cli":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\gh`
	case "git":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\git`
	case "go-sdk-green-tea":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go`
	case "hermes-wasm-gc":
		return `C:\aCogSpaceSeed\00flow\s-forge\bin\hermes_wasm_gc.wasm`
	case "icu4c":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\icu`
	case "jdk-25-headless":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\jdk`
	case "notary-notation":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\notation`
	case "step-ca":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\step-ca`
	case "trivy":
		return `C:\aCogSpaceSeed\00flow\s-forge\91000-external-binaries\trivy`
	case "wasm-opt":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\wasm\binaryen`
	case "wasm-tools":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\wasm\wasm-tools`
	case "wasmer":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\wasm\wasmer`
	case "wasmtime":
		return `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\wasm\wasmtime`
	default:
		return ""
	}
}

// getDownloadURL resolves the download URL for targets.
func getDownloadURL(name string) string {
	switch name {
	case "bazel-rules-flutter":
		return "https://github.com/flutter/rules_flutter/archive/refs/tags/v2.1.0.zip"
	case "bazel-rules-go":
		return "https://github.com/bazelbuild/rules_go/releases/download/v0.51.0/rules_go-v0.51.0.zip"
	case "bazelisk":
		return "https://github.com/bazelbuild/bazelisk/releases/download/v1.20.0/bazelisk-windows-amd64.exe"
	case "blake3":
		return "https://github.com/BLAKE3-team/BLAKE3/releases/download/1.5.0/b3sum_windows_x64_bin.zip"
	case "buildifier":
		return "https://github.com/bazelbuild/buildtools/releases/download/v7.1.2/buildifier-windows-amd64.exe"
	case "firebase-tools":
		return "https://github.com/firebase/firebase-tools/releases/download/v13.11.2/firebase-tools-win.exe"
	case "flutter-sdk-firehorse":
		return "https://storage.googleapis.com/flutter_infra_release/releases/stable/windows/flutter_windows_3.41.0-stable.zip"
	case "gcloud-sdk":
		return "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-476.0.0-windows-x86_64.zip"
	case "gh-cli":
		return "https://github.com/cli/cli/releases/download/v2.49.0/gh_2.49.0_windows_amd64.zip"
	case "git":
		return "https://github.com/git-for-windows/git/releases/download/v2.45.0.windows.1/MinGit-2.45.0-64-bit.zip"
	case "go-sdk-green-tea":
		return "https://storage.googleapis.com/golang/go1.26.0.windows-amd64.zip"
	case "hermes-wasm-gc":
		return "https://github.com/facebook/hermes/releases/download/v1.0.0/hermes-cli-windows.zip"
	case "icu4c":
		return "https://github.com/unicode-org/icu/releases/download/release-75-1/icu4c-75_1-Win64-MSVC2019.zip"
	case "jdk-25-headless":
		return "https://storage.googleapis.com/jdk-releases/jdk-25-headless_windows-x64_bin.zip"
	case "notary-notation":
		return "https://github.com/notaryproject/notation/releases/download/v1.1.0/notation_1.1.0_windows_amd64.zip"
	case "step-ca":
		return "https://github.com/smallstep/certificates/releases/download/v0.26.2/step-ca_windows_0.26.2_amd64.zip"
	case "trivy":
		return "https://github.com/aquasecurity/trivy/releases/download/v0.70.0/trivy_0.70.0_windows-64bit.zip"
	case "wasm-opt":
		return "https://github.com/WebAssembly/binaryen/releases/download/version_117/binaryen-version_117-x86_64-windows.tar.gz"
	case "wasm-tools":
		return "https://github.com/bytecodealliance/wasm-tools/releases/download/wasm-tools-1.210.0/wasm-tools-1.210.0-x86_64-windows.zip"
	case "wasmer":
		return "https://github.com/wasmerio/wasmer/releases/download/v4.3.0/wasmer-windows-amd64.zip"
	case "wasmtime":
		return "https://github.com/bytecodealliance/wasmtime/releases/download/v21.0.0/wasmtime-v21.0.0-x86_64-windows.zip"
	default:
		return ""
	}
}
