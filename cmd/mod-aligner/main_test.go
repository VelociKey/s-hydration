package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGoWork(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gowork-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	goWorkContent := `go 1.26.3

use (
	./00flow/s-latentlingua
	"./00flow/s-adk"
	./00flow/s-hydration
)
`
	goWorkPath := filepath.Join(tempDir, "go.work")
	if err := os.WriteFile(goWorkPath, []byte(goWorkContent), 0644); err != nil {
		t.Fatalf("failed to write go.work: %v", err)
	}

	workspaces, err := parseGoWork(goWorkPath)
	if err != nil {
		t.Fatalf("parseGoWork failed: %v", err)
	}

	expected := []string{"./00flow/s-latentlingua", "./00flow/s-adk", "./00flow/s-hydration"}
	if len(workspaces) != len(expected) {
		t.Fatalf("expected %d workspaces, got %d", len(expected), len(workspaces))
	}
	for i, v := range expected {
		if workspaces[i] != v {
			t.Errorf("expected workspace %d to be %q, got %q", i, v, workspaces[i])
		}
	}
}

func TestTopoSort(t *testing.T) {
	wsA := &WorkspaceInfo{Name: "wsA", ModuleName: "sov.fleet/a", Imports: []string{"sov.fleet/b"}}
	wsB := &WorkspaceInfo{Name: "wsB", ModuleName: "sov.fleet/b", Imports: []string{"sov.fleet/c"}}
	wsC := &WorkspaceInfo{Name: "wsC", ModuleName: "sov.fleet/c", Imports: []string{}}

	localModules := map[string]*WorkspaceInfo{
		"sov.fleet/a": wsA,
		"sov.fleet/b": wsB,
		"sov.fleet/c": wsC,
	}

	wsList := []*WorkspaceInfo{wsA, wsB, wsC}
	sorted := topoSort(wsList, localModules)

	if len(sorted) != 3 {
		t.Fatalf("expected sorted length 3, got %d", len(sorted))
	}

	// Order should be C, B, A
	if sorted[0].Name != "wsC" || sorted[1].Name != "wsB" || sorted[2].Name != "wsA" {
		t.Errorf("unexpected sorting order: %v", getWsNames(sorted))
	}
}

func TestAlignGoMod(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gomod-align-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	wsAPath := filepath.Join(tempDir, "wsA")
	wsBPath := filepath.Join(tempDir, "wsB")
	_ = os.MkdirAll(wsAPath, 0755)
	_ = os.MkdirAll(wsBPath, 0755)

	initialGoMod := `module sov.fleet/a

go 1.26.3

require (
	github.com/google/uuid v1.3.0
	sov.fleet/b v0.0.0
)

replace sov.fleet/b => ../old-wsB
`
	if err := os.WriteFile(filepath.Join(wsAPath, "go.mod"), []byte(initialGoMod), 0644); err != nil {
		t.Fatalf("failed to write initial go.mod: %v", err)
	}

	wsA := &WorkspaceInfo{
		Name:       "wsA",
		Path:       wsAPath,
		ModuleName: "sov.fleet/a",
		Imports:    []string{"sov.fleet/b"},
	}
	wsB := &WorkspaceInfo{
		Name:       "wsB",
		Path:       wsBPath,
		ModuleName: "sov.fleet/b",
	}

	localModules := map[string]*WorkspaceInfo{
		"sov.fleet/a": wsA,
		"sov.fleet/b": wsB,
	}

	err = alignGoMod(wsA, localModules)
	if err != nil {
		t.Fatalf("alignGoMod failed: %v", err)
	}

	contentBytes, err := os.ReadFile(filepath.Join(wsAPath, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read aligned go.mod: %v", err)
	}

	content := string(contentBytes)
	if !strings.Contains(content, "github.com/google/uuid v1.3.0") {
		t.Errorf("external requirement was lost: %s", content)
	}
	if strings.Contains(content, "../old-wsB") {
		t.Errorf("old replacement path was not pruned: %s", content)
	}
	if !strings.Contains(content, "sov.fleet/b => ../wsB") {
		t.Errorf("new correct replace path was not injected: %s", content)
	}
}
