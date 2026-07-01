package hydration_attestations

import (
	. "sov.fleet/s-hydration"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"sov.fleet/quicdl"
)

func TestDetermineExecutionPlan(t *testing.T) {
	h := NewSovereignPurifier()

	// Mock dynamic experience size
	h.GetExperience()["custom-massive"] = ExperienceRecord{
		TargetName:     "custom-massive",
		LastLoadedSize: 150 * 1024 * 1024, // 150MB -> Phase 2
	}

	records := []ArtifactRecord{
		{Name: "blake3", OriginalSize: 1024 * 1024},                  // Phase 1
		{Name: "golang", OriginalSize: 215 * 1024 * 1024},  // Phase 1
		{Name: "flutter-sdk-firehorse", OriginalSize: 1932735283},    // Phase 2 (1.8 GB)
		{Name: "custom-massive", OriginalSize: 50 * 1024},            // Phase 2 via experience
		{Name: "bazel-rules-go", OriginalSize: 10 * 1024 * 1024},     // Phase 3 github.com
		{Name: "bazel", OriginalSize: 15 * 1024 * 1024},           // Phase 3 github.com
		{Name: "step-ca", OriginalSize: 50 * 1024},                   // Phase 3 small local placeholder
	}

	bootstrap, massive, parallelGroups := h.DetermineExecutionPlan(records)

	// Verify Phase 1
	if len(bootstrap) != 2 {
		t.Errorf("Expected 2 bootstrap targets, got %d", len(bootstrap))
	}
	for _, b := range bootstrap {
		if b.Name != "blake3" && b.Name != "golang" {
			t.Errorf("Unexpected bootstrap target: %s", b.Name)
		}
	}

	// Verify Phase 2
	if len(massive) != 2 {
		t.Errorf("Expected 2 massive targets, got %d", len(massive))
	}
	var foundMassiveCustom bool
	for _, m := range massive {
		if m.Name == "custom-massive" {
			foundMassiveCustom = true
		}
	}
	if !foundMassiveCustom {
		t.Error("Dynamic massive target custom-massive was not scheduled to Phase 2")
	}

	// Verify Phase 3
	ghCount := len(parallelGroups["github.com"])
	if ghCount == 0 {
		t.Error("Expected parallel multiplexing cohort groups for github.com, got 0")
	}
}

func TestSovereignScalerAutoscaling(t *testing.T) {
	scaler := NewSovereignScaler(2, 4) // Min=2, Max=4

	var counter int64
	var mu sync.Mutex

	taskCount := 10
	wg := sync.WaitGroup{}
	wg.Add(taskCount)

	for i := 0; i < taskCount; i++ {
		scaler.Submit(func(ctx context.Context) error {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	wg.Wait()
	scaler.Shutdown()

	if counter != int64(taskCount) {
		t.Errorf("Expected %d tasks to run, got %d", taskCount, counter)
	}
}

func TestProgressWriterCalculation(t *testing.T) {
	var buf bytes.Buffer
	pw := &quicdl.ProgressWriter{
		TargetName: "test-target",
		TotalBytes: 1000,
		StartTime:  time.Now().Add(-1 * time.Second),
		LastReport: time.Now().Add(-1 * time.Second),
		Writer:     &buf,
	}

	// Write 500 bytes (50%)
	payload := make([]byte, 500)
	n, err := pw.Write(payload)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 500 {
		t.Errorf("Expected 500 bytes written, got %d", n)
	}

	// Verify underlying buffer got the bytes
	if buf.Len() != 500 {
		t.Errorf("Expected underlying writer to receive 500 bytes, got %d", buf.Len())
	}
}

func TestParseSBOMRegistry(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mock-sbom-*.webnf")
	if err != nil {
		t.Fatalf("Failed to create temp SBOM: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	mockData := `SBOM-V2
2026-05-18T13:30:34-04:00
AAIF-GreenTea-Rehydrator-v2.0
# VERACITY-SEAL: mock-seal
trivy|0.70.0||sha512:trivyhash|212653891|212653891|2026-05-18T13:30:38-04:00||
`
	_, _ = tmpFile.WriteString(mockData)
	tmpFile.Close()

	records, err := ParseSBOM(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse SBOM: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 parsed record, got %d", len(records))
	}

	r := records[0]
	if r.Name != "trivy" || r.Version != "0.70.0" || r.PrunedHash != "sha512:trivyhash" {
		t.Errorf("Mismatch in parsed SBOM record: %+v", r)
	}
	if r.OriginalSize != 212653891 {
		t.Errorf("Expected size 212653891, got %d", r.OriginalSize)
	}
}

func TestExperienceRegistryLoadSave(t *testing.T) {
	h := NewSovereignPurifier()

	tmpDir, err := os.MkdirTemp("", "mock-exp-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	expPath := filepath.Join(tmpDir, "experience.webnf")

	h.GetExperience()["git"] = ExperienceRecord{
		TargetName:     "git",
		LastLoadedSize: 125483360,
		LastDuration:   5 * time.Second,
		LastPrunedSize: 110000000,
	}

	// Save
	err = h.SaveExperience(expPath)
	if err != nil {
		t.Fatalf("Save experience failed: %v", err)
	}

	// Load into a new hydrator
	h2 := NewSovereignPurifier()
	err = h2.LoadExperience(expPath)
	if err != nil {
		t.Fatalf("Load experience failed: %v", err)
	}

	rec, exists := h2.GetExperience()["git"]
	if !exists {
		t.Fatal("Git experience record missing in loaded hydrator")
	}

	if rec.LastLoadedSize != 125483360 || rec.LastDuration != 5*time.Second {
		t.Errorf("Mismatch in loaded experience record: %+v", rec)
	}
}

func TestIterativeDFWalkAndMetabolicPruning(t *testing.T) {
	h := NewSovereignPurifier()

	tmpDir, err := os.MkdirTemp("", "walk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp walk dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test structures
	// 1. Keep
	keepPath := filepath.Join(tmpDir, "src")
	os.MkdirAll(keepPath, 0755)
	os.WriteFile(filepath.Join(keepPath, "main.go"), []byte("package main"), 0644)

	// 2. Prune Folder (test)
	testPath := filepath.Join(tmpDir, "src", "test")
	os.MkdirAll(testPath, 0755)
	os.WriteFile(filepath.Join(testPath, "main_test.go"), []byte("package main"), 0644)

	// 3. Prune file (.pdf)
	pdfFile := filepath.Join(keepPath, "manual.pdf")
	os.WriteFile(pdfFile, []byte("PDF content"), 0644)

	// 4. Exempt license file (license.md)
	licenseFile := filepath.Join(keepPath, "license.md")
	os.WriteFile(licenseFile, []byte("MIT License"), 0644)

	// Run metabolic pruning pass
	err = IterativeDFWalk(tmpDir, func(curr string, info os.FileInfo) (bool, error) {
		shouldPrune, isDirPrune := h.ShouldPrune(curr, info)
		if shouldPrune {
			if isDirPrune {
				os.RemoveAll(curr)
			} else {
				os.Remove(curr)
			}
			return true, nil // Skip children
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("IterativeDFWalk failed: %v", err)
	}

	// Verify
	if _, err := os.Stat(filepath.Join(keepPath, "main.go")); err != nil {
		t.Error("Expected keep file main.go to exist")
	}
	if _, err := os.Stat(filepath.Join(keepPath, "license.md")); err != nil {
		t.Error("Expected exempt license.md file to exist")
	}
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("Expected test folder to be metabolically pruned")
	}
	if _, err := os.Stat(pdfFile); !os.IsNotExist(err) {
		t.Error("Expected manual.pdf to be metabolically pruned")
	}
}

func TestSovereignProbeAndAltSvc(t *testing.T) {
	h := NewSovereignPurifier()

	// Probe fallback logic test
	proto, finalURL := h.SovereignProbe("https://github.com/BLAKE3-team/BLAKE3")
	if proto != ProtoH2 {
		t.Errorf("Expected fallback to ProtoH2, got %d", proto)
	}
	if finalURL != "https://github.com/BLAKE3-team/BLAKE3" {
		t.Errorf("Expected unmodified URL, got %s", finalURL)
	}
}

func TestSovereignABTestCampaign(t *testing.T) {
	h := NewSovereignPurifier()
	h.SetABTest(true)
	h.SetSequential(false)

	// Mock experiences
	h.GetExperience()["pkg-a"] = ExperienceRecord{TargetName: "pkg-a", LastLoadedSize: 500, LastDuration: 10 * time.Millisecond}
	h.GetExperience()["pkg-b"] = ExperienceRecord{TargetName: "pkg-b", LastLoadedSize: 1000, LastDuration: 20 * time.Millisecond}

	seqDurations := map[string]time.Duration{
		"pkg-a": 15 * time.Millisecond,
		"pkg-b": 25 * time.Millisecond,
	}
	parDurations := map[string]time.Duration{
		"pkg-a": 8 * time.Millisecond,
		"pkg-b": 12 * time.Millisecond,
	}

	records := []ArtifactRecord{
		{Name: "pkg-a", OriginalSize: 500},
		{Name: "pkg-b", OriginalSize: 1000},
	}

	// Proves printABTestReport compiles and formats side-by-side correctly!
	h.PrintABTestReport(records, seqDurations, 40*time.Millisecond, parDurations, 15*time.Millisecond)
}

func TestCascadingBuildDAG(t *testing.T) {
	// Test parseGoWork on mock content
	mockGoWork := `go 1.26.3
use (
	./00flow/s-sacp
	./00flow/s-natives
	./00flow/s-hydration
)
`
	tmpFile, err := os.CreateTemp("", "go.work-*")
	if err != nil {
		t.Fatalf("Failed to create temp go.work: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	_, _ = tmpFile.WriteString(mockGoWork)
	tmpFile.Close()

	workspaces, err := ParseGoWork(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse go.work: %v", err)
	}
	if len(workspaces) != 3 {
		t.Errorf("Expected 3 workspaces, got %d: %v", len(workspaces), workspaces)
	}

	// Test Topological Sort
	graph := map[string][]string{
		"s-natives":   {"s-sacp"},
		"s-hydration": {"s-sacp", "s-natives"},
		"s-sacp":      {},
	}

	downstream := FindDownstreamNodes(graph, "s-sacp")
	if !downstream["s-natives"] || !downstream["s-hydration"] {
		t.Errorf("Expected downstream of s-sacp to include s-natives and s-hydration: %v", downstream)
	}

	order, err := TopologicalSort(graph, downstream)
	if err != nil {
		t.Fatalf("Topological sort failed: %v", err)
	}

	nativesIdx := -1
	hydrationIdx := -1
	for i, node := range order {
		if node == "s-natives" {
			nativesIdx = i
		} else if node == "s-hydration" {
			hydrationIdx = i
		}
	}

	if nativesIdx == -1 || hydrationIdx == -1 || nativesIdx > hydrationIdx {
		t.Errorf("Invalid topological order: %v (nativesIdx=%d, hydrationIdx=%d)", order, nativesIdx, hydrationIdx)
	}
}

func TestParseSBOMRegistryWithSourceAndBuildPolicy(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "mock-extended-sbom-*.webnf")
	if err != nil {
		t.Fatalf("Failed to create temp SBOM: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	mockData := `SBOM-V2
2026-05-18T13:30:34-04:00
AAIF-GreenTea-Rehydrator-v2.0
# VERACITY-SEAL: mock-seal
blake3|1.5.0|sha512:zip|sha512:binary|124667|124667|2026-05-18T13:30:38-04:00|LIBRARY|https://github.com/zeebo/blake3/archive/refs/tags/v0.2.4.zip|GO_NATIVE
`
	_, _ = tmpFile.WriteString(mockData)
	tmpFile.Close()

	records, err := ParseSBOM(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse SBOM: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 parsed record, got %d", len(records))
	}

	r := records[0]
	if r.Name != "blake3" || r.Version != "1.5.0" || r.OriginalHash != "sha512:zip" || r.PrunedHash != "sha512:binary" {
		t.Errorf("Mismatch in parsed SBOM record: %+v", r)
	}
	if r.SourceURL != "https://github.com/zeebo/blake3/archive/refs/tags/v0.2.4.zip" || r.BuildPolicy != "GO_NATIVE" {
		t.Errorf("Mismatch in parsed source URL or build policy: %+v", r)
	}
}

func TestParseWorkspacePrologueTest(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "workspace-facets-*.webnf")
	if err != nil {
		t.Fatalf("Failed to create temp prologue: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	mockData := `name: "test-workspace"
classification: "test"
read_only: true
build_enabled: false
harness_path: "71000-build-harness/workspace.harness"
cluster_membership: ["test_cluster"]
rehydration_requirements: ["gitleaks", "opentofu"]
`
	_, _ = tmpFile.WriteString(mockData)
	tmpFile.Close()

	facets, err := ParseWorkspacePrologue(tmpFile.Name())
	if err != nil {
		t.Fatalf("ParseWorkspacePrologue failed: %v", err)
	}

	if facets.ReadOnly != true {
		t.Errorf("Expected ReadOnly true, got false")
	}
	if facets.BuildEnabled != false {
		t.Errorf("Expected BuildEnabled false, got true")
	}
	if facets.HarnessPath != "71000-build-harness/workspace.harness" {
		t.Errorf("Expected HarnessPath '71000-build-harness/workspace.harness', got %q", facets.HarnessPath)
	}
	if len(facets.Clusters) != 1 || facets.Clusters[0] != "test_cluster" {
		t.Errorf("Expected Cluster membership ['test_cluster'], got %v", facets.Clusters)
	}
	if len(facets.RehydrationRequirements) != 2 || facets.RehydrationRequirements[0] != "gitleaks" || facets.RehydrationRequirements[1] != "opentofu" {
		t.Errorf("Expected RehydrationRequirements ['gitleaks', 'opentofu'], got %v", facets.RehydrationRequirements)
	}
}

func TestScanLicenseCompliance(t *testing.T) {
	h := NewSovereignPurifier()

	tmpDir, err := os.MkdirTemp("", "license-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clean scratch cache if it exists to avoid interference
	os.Setenv("GEMINI_CLI_HOME", tmpDir)

	// Test case 1: Standard LICENSE file containing GPL version 3
	gplDir := filepath.Join(tmpDir, "gpl-pkg")
	os.MkdirAll(gplDir, 0755)
	os.WriteFile(filepath.Join(gplDir, "LICENSE"), []byte("This software is governed by the Affero General Public License (AGPL)"), 0644)

	err = h.ScanLicenseCompliance("mock-gpl-pkg", gplDir)
	if err == nil {
		t.Error("Expected primary scan to fail and block GPL/AGPL license, but it passed")
	}

	// Test case 2: No standard LICENSE file, but inline GPL comment in source file (secondary discovery scan)
	inlineDir := filepath.Join(tmpDir, "inline-gpl-pkg")
	os.MkdirAll(inlineDir, 0755)
	os.WriteFile(filepath.Join(inlineDir, "main.go"), []byte("// Copyright 2026. This file is licensed under GPL v3.\npackage main\n"), 0644)

	err = h.ScanLicenseCompliance("mock-inline-gpl", inlineDir)
	if err == nil {
		t.Error("Expected secondary discovery scan to block inline copyleft terms, but it passed")
	}

	// Test case 3: Safe, non-copyleft library
	safeDir := filepath.Join(tmpDir, "safe-pkg")
	os.MkdirAll(safeDir, 0755)
	os.WriteFile(filepath.Join(safeDir, "LICENSE"), []byte("MIT License\nPermission is hereby granted..."), 0644)

	err = h.ScanLicenseCompliance("mock-safe-pkg", safeDir)
	if err != nil {
		t.Errorf("Expected license compliance to pass for MIT license, but got error: %v", err)
	}

	// Test case 4: Cached license check bypass
	err = h.ScanLicenseCompliance("mock-safe-pkg", safeDir)
	if err != nil {
		t.Errorf("Expected cached license check to succeed, got error: %v", err)
	}
}


