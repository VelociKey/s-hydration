package hydration_attestations

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceSynthesis(t *testing.T) {
	// Synthesize a test workspace named 's-test-synthesis'
	projectRoot := "C:\\aCogSpaceSeed"
	fabConstructPath := filepath.Join(projectRoot, "00flow", "s-forge", "97000-internal-toolchains", "fab-construct.exe")
	if _, err := os.Stat(fabConstructPath); os.IsNotExist(err) {
		fabConstructPath = filepath.Join(projectRoot, "00flow", "s-hydration", "97000-internal-toolchains", "fab-construct.exe")
	}
	if _, err := os.Stat(fabConstructPath); os.IsNotExist(err) {
		t.Skip("fab-construct.exe not found, skipping synthesis test")
	}

	wsName := "s-test-synthesis"
	targetPath := filepath.Join(projectRoot, "00flow", wsName)
	
	// Clean up any existing stale test workspace
	_ = os.RemoveAll(targetPath)
	defer os.RemoveAll(targetPath)

	cmd := exec.Command(fabConstructPath, "workspace", "-name", wsName, "-silo", "00flow")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fab-construct failed: %v, output: %s", err, string(output))
	}

	// Verify paths created
	expectedPaths := []string{
		filepath.Join(targetPath, "go.mod"),
		filepath.Join(targetPath, "71000-build-harness", "workspace.harness"),
		filepath.Join(targetPath, "81000-active-source", "cmd", wsName, "main.go"),
		filepath.Join(targetPath, "81000-active-source", "cmd", wsName, "main_test.go"),
	}

	for _, p := range expectedPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("Expected path does not exist: %s", p)
		}
	}

	// Clean up go.work registration to keep the workspace clean
	goWorkPath := filepath.Join(projectRoot, "go.work")
	if data, err := os.ReadFile(goWorkPath); err == nil {
		lines := strings.Split(string(data), "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, wsName) {
				newLines = append(newLines, line)
			}
		}
		_ = os.WriteFile(goWorkPath, []byte(strings.Join(newLines, "\n")), 0644)
	}
}

func TestRollbackOnFailure(t *testing.T) {
	// Synthesize a test workspace named 's-test-rollback'
	projectRoot := "C:\\aCogSpaceSeed"
	fabConstructPath := filepath.Join(projectRoot, "00flow", "s-forge", "97000-internal-toolchains", "fab-construct.exe")
	if _, err := os.Stat(fabConstructPath); os.IsNotExist(err) {
		fabConstructPath = filepath.Join(projectRoot, "00flow", "s-hydration", "97000-internal-toolchains", "fab-construct.exe")
	}
	if _, err := os.Stat(fabConstructPath); os.IsNotExist(err) {
		t.Skip("fab-construct.exe not found, skipping rollback test")
	}

	wsName := "s-test-rollback"
	targetPath := filepath.Join(projectRoot, "00flow", wsName)
	
	_ = os.RemoveAll(targetPath)
	defer os.RemoveAll(targetPath)

	// Construct it
	cmd := exec.Command(fabConstructPath, "workspace", "-name", wsName, "-silo", "00flow")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fab-construct failed: %v, output: %s", err, string(output))
	}

	// Read original main.go
	mainGoPath := filepath.Join(targetPath, "81000-active-source", "cmd", wsName, "main.go")
	originalContent, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	// Invoke int-rehydrator to build successfully first and generate cache
	intRehydratorPath := filepath.Join(projectRoot, "00flow", "s-hydration", "int-rehydrator.exe")
	if _, err := os.Stat(intRehydratorPath); os.IsNotExist(err) {
		intRehydratorPath = filepath.Join(projectRoot, "00flow", "s-forge", "97000-internal-toolchains", "int-rehydrator.exe")
	}

	cmdSuccess := exec.Command(intRehydratorPath, "-local-only", "-rollback-on-failure", "-workspace", wsName)
	cmdSuccess.Env = os.Environ()
	if out, err := cmdSuccess.CombinedOutput(); err != nil {
		t.Fatalf("Initial successful build failed: %v, output: %s", err, string(out))
	}

	// Make a compiler error modification
	brokenContent := "package main\n\nfunc main() {\n\tcompiler_error_here(\n}\n"
	err = os.WriteFile(mainGoPath, []byte(brokenContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write broken main.go: %v", err)
	}

	cmdRehydrate := exec.Command(intRehydratorPath, "-local-only", "-rollback-on-failure", "-workspace", wsName)
	cmdRehydrate.Env = os.Environ()
	// Let's run and it MUST fail because of the compiler error
	errRun := cmdRehydrate.Run()
	if errRun == nil {
		t.Fatal("Expected int-rehydrator build to fail with compilation error, but it passed")
	}

	// Assert that main.go was rolled back/restored to the original content
	restoredContent, err := os.ReadFile(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to read restored main.go: %v", err)
	}

	if string(restoredContent) != string(originalContent) {
		t.Errorf("Expected main.go to be restored to original content, but got:\n%s", string(restoredContent))
	}

	// Clean up go.work registration
	goWorkPath := filepath.Join(projectRoot, "go.work")
	if data, err := os.ReadFile(goWorkPath); err == nil {
		lines := strings.Split(string(data), "\n")
		var newLines []string
		for _, line := range lines {
			if !strings.Contains(line, wsName) {
				newLines = append(newLines, line)
			}
		}
		_ = os.WriteFile(goWorkPath, []byte(strings.Join(newLines, "\n")), 0644)
	}
}
