package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
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

func main() {
	harnessPath := flag.String("harness", "", "Path to the workspace_hydrator.webnf file")
	targetName := flag.String("only", "", "Hydrate and synthesize a single targeted package")
	force := flag.Bool("force", false, "Force rebuild regardless of timestamps")
	flag.Parse()

	if *harnessPath == "" {
		slog.Error("Missing required flag -harness")
		os.Exit(1)
	}

	contentBytes, err := os.ReadFile(*harnessPath)
	if err != nil {
		slog.Error("Failed to read workspace harness", "path", *harnessPath, "error", err)
		os.Exit(1)
	}

	harness, err := ParseWorkspaceHarness(string(contentBytes), *harnessPath)
	if err != nil {
		slog.Error("Failed to parse and validate workspace harness", "error", err)
		os.Exit(1)
	}

	slog.Info("Workspace harness verified", "name", harness.Name, "read_only", harness.Facets.ReadOnly, "build_enabled", harness.Facets.BuildEnabled, "classification", harness.Facets.FunctionalClassification)

	hydrator := NewHydrator()
	hydrator.Register(&DynamicSynthesisMechanism{})

	ctx := context.Background()
	for _, target := range harness.Targets {
		if target.Mode != ModeDynamic {
			continue // int-rehydrator only processes internal DYNAMIC compilation targets
		}
		
		if *targetName != "" && !strings.Contains(target.SourcePath, *targetName) && !strings.Contains(target.OutputPath, *targetName) {
			continue
		}

		// Route the artifact to the correct internal s-forge directory based on category.
		// Internal range: 96000-internal-binaries, 97000-internal-toolchains,
		//                 98000-internal-libraries, 99000-internal-actors.
		baseName := target.SourcePath
		if idx := strings.LastIndex(baseName, "/"); idx != -1 {
			baseName = baseName[idx+1:]
		}
		if idx := strings.LastIndex(baseName, ":"); idx != -1 {
			baseName = baseName[idx+1:]
		}

		cat := target.Category
		if cat == "" {
			cat = "BINARY"
		}

		dir, err := registry.GetTargetPhysicalPath("", cat, "internal")
		if err != nil {
			slog.Warn("Failed to resolve physical path for category, falling back to BINARY", "category", cat, "error", err)
			dir, _ = registry.GetTargetPhysicalPath("", "BINARY", "internal")
		}
		target.OutputPath = filepath.Join(dir, baseName)

		// Optimize: Check timestamps if not forced
		if !*force {
			outInfo, err := os.Stat(target.OutputPath)
			if err == nil {
				wsPath := getWorkspacePath(*harnessPath)
				newestSrcTime, walkErr := getNewestModTime(wsPath)
				if walkErr == nil && !newestSrcTime.IsZero() {
					if outInfo.ModTime().After(newestSrcTime) {
						slog.Info("SKIPPING target compilation (up to date, no changes since last build)", "target", target.SourcePath, "out", target.OutputPath)
						continue
					}
				}
			}
		}

		slog.Info("Orchestrating internal rehydration for target", "src", target.SourcePath, "category", target.Category, "assumed_out", target.OutputPath)
		err = hydrator.Execute(ctx, target)
		if err != nil {
			slog.Error("Internal synthesis failed", "target", target.SourcePath, "error", err)
			os.Exit(1)
		}
	}

	slog.Info("Internal rehydration process complete. All artifacts promoted to internal s-forge.")
}
