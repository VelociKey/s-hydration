package hydration

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateWorkspaceHarness validates workspace.harness rules against physical directories on disk.
func ValidateWorkspaceHarness(harnessPath string) error {
	contentBytes, err := os.ReadFile(harnessPath)
	if err != nil {
		return fmt.Errorf("failed to read workspace harness %s: %w", harnessPath, err)
	}

	harness, err := ParseWorkspaceHarness(string(contentBytes), harnessPath)
	if err != nil {
		return fmt.Errorf("harness syntax error: %w", err)
	}

	wsPath := getWorkspacePath(harnessPath)

	for _, target := range harness.Targets {
		// Rule 1: Ensure engine is valid
		if target.Engine != "go-compiler" && target.Engine != "bazel" && target.Engine != "none" && target.Engine != "" {
			return fmt.Errorf("invalid compiler engine %q on target %q", target.Engine, target.SourcePath)
		}

		// Rule 2: Ensure source target path exists on disk
		srcPath := filepath.Join(wsPath, filepath.FromSlash(target.SourcePath))
		if _, err := os.Stat(srcPath); err != nil {
			return fmt.Errorf("target source path %q does not exist", srcPath)
		}
	}

	return nil
}
