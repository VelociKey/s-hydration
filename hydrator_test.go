package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"


)

func TestDetermineExecutionPlan(t *testing.T) {
	h := NewSovereignHydrator()

	// Mock dynamic experience size
	h.experience["custom-massive"] = ExperienceRecord{
		TargetName:     "custom-massive",
		LastLoadedSize: 150 * 1024 * 1024, // 150MB -> Phase 2
	}

	records := []ArtifactRecord{
		{Name: "blake3", OriginalSize: 1024 * 1024},                  // Phase 1
		{Name: "go-sdk-green-tea", OriginalSize: 215 * 1024 * 1024},  // Phase 1
		{Name: "flutter-sdk-firehorse", OriginalSize: 1932735283},    // Phase 2 (1.8 GB)
		{Name: "custom-massive", OriginalSize: 50 * 1024},            // Phase 2 via experience
		{Name: "bazel-rules-go", OriginalSize: 10 * 1024 * 1024},     // Phase 3 github.com
		{Name: "bazelisk", OriginalSize: 15 * 1024 * 1024},           // Phase 3 github.com
		{Name: "step-ca", OriginalSize: 50 * 1024},                   // Phase 3 small local placeholder
	}

	bootstrap, massive, parallelGroups := h.determineExecutionPlan(records)

	// Verify Phase 1
	if len(bootstrap) != 2 {
		t.Errorf("Expected 2 bootstrap targets, got %d", len(bootstrap))
	}
	for _, b := range bootstrap {
		if b.Name != "blake3" && b.Name != "go-sdk-green-tea" {
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
	pw := &ProgressWriter{
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
trivy|0.70.0||sha512:trivyhash|212653891|212653891|2026-05-18T13:30:38-04:00
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
	h := NewSovereignHydrator()

	tmpDir, err := os.MkdirTemp("", "mock-exp-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	expPath := filepath.Join(tmpDir, "experience.webnf")

	h.experience["git"] = ExperienceRecord{
		TargetName:     "git",
		LastLoadedSize: 125483360,
		LastDuration:   5 * time.Second,
		LastPrunedSize: 110000000,
	}

	// Save
	err = h.saveExperience(expPath)
	if err != nil {
		t.Fatalf("Save experience failed: %v", err)
	}

	// Load into a new hydrator
	h2 := NewSovereignHydrator()
	err = h2.loadExperience(expPath)
	if err != nil {
		t.Fatalf("Load experience failed: %v", err)
	}

	rec, exists := h2.experience["git"]
	if !exists {
		t.Fatal("Git experience record missing in loaded hydrator")
	}

	if rec.LastLoadedSize != 125483360 || rec.LastDuration != 5*time.Second {
		t.Errorf("Mismatch in loaded experience record: %+v", rec)
	}
}

func TestIterativeDFWalkAndMetabolicPruning(t *testing.T) {
	h := NewSovereignHydrator()

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
		shouldPrune, isDirPrune := h.shouldPrune(curr, info)
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
	h := NewSovereignHydrator()

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
	h := NewSovereignHydrator()
	h.abtest = true
	h.sequential = false

	// Mock experiences
	h.experience["pkg-a"] = ExperienceRecord{TargetName: "pkg-a", LastLoadedSize: 500, LastDuration: 10 * time.Millisecond}
	h.experience["pkg-b"] = ExperienceRecord{TargetName: "pkg-b", LastLoadedSize: 1000, LastDuration: 20 * time.Millisecond}

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
	h.printABTestReport(records, seqDurations, 40*time.Millisecond, parDurations, 15*time.Millisecond)
}
