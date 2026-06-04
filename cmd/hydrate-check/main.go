package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sov.fleet/s-hydration"
	"sov.fleet/s-hydration/400-registry"
)

var externalRationales = map[string]string{
	"bazel-rules-flutter":     "Bazel rules for building Flutter/Dart applications.",
	"bazel-rules-go":          "Bazel rules for building Go applications.",
	"bazel":                   "Core build tool and orchestration compiler.",
	"opentelemetry":           "Observability framework for cloud-native software telemetry metrics.",
	"opentofu":                "Open-source infrastructure as code tool bifurcated from Terraform.",
	"cosign":                  "Sigstore CLI tool for signing and verifying OCI containers and WebAssembly modules.",
	"buildifier":              "Linter and formatter for Bazel BUILD, MODULE.bazel, and Starlark rule files.",
	"firebase-tools":          "CLI tool for Firebase deployments and hosting management.",
	"flutter-sdk-firehorse":   "Sovereign Flutter SDK distribution for UI and mobile/web development.",
	"gcloud-sdk":              "Google Cloud SDK for deployments and cloud service interactions.",
	"gh-cli":                  "GitHub CLI for automating repository operations.",
	"gitleaks":                "Static secret scanner for auditing repository histories and filesystem trees.",
	"git":                     "Sovereign Git client for source control synchronization.",
	"golang":                  "Core Go programming language compiler and standard library SDK.",
	"icu4c":                   "International Components for Unicode library.",
	"jdk-25-headless":         "Java Development Kit required by Bazel build orchestration.",
	"notary-notation":         "Notary CLI for signing and verifying artifact signatures.",
	"step-ca":                 "Sovereign private certificate authority for secure transport seals.",
	"trivy":                   "Security scanner engine for offline vulnerability auditing.",
	"wasm-opt":                "WebAssembly optimizer tool from Binaryen toolchain.",
	"wasm-tools":              "WebAssembly parsing, printing, and validation tools.",
	"wasmer":                  "WebAssembly runtime for running sovereign WASM modules.",
	"wasmtime":                "WebAssembly engine and compiler runtime.",
	"quic-go":                 "Go library implementation of the QUIC transport protocol.",
	"go-tpm":                  "Go library for Trusted Platform Module (TPM) interactions.",
	"go-spiffe":               "Go library for SPIFFE standard identity and trust circles.",
	"memguard":                "Go library for securing sensitive data in memory.",
	"wazero":                  "Go zero-dependency WebAssembly runtime.",
	"blake3":                  "Go binding/implementation of the Blake3 hashing protocol.",
	"qpack":                   "Go implementation of QPACK compression for HTTP/3.",
}

func main() {
	scope := flag.String("scope", "all", "Scope of check: external, internal, or all")
	flag.Parse()

	projectRoot := `C:\aCogSpaceSeed`
	sbomPath := filepath.Join(projectRoot, "00flow", "s-forge", "90100-rehydration-seed", "sbom.external_artifact.webnf")
	reportPath := filepath.Join(projectRoot, "00flow", "s-hydration", "hydration_check_results.md")

	var report strings.Builder
	report.WriteString("# Sovereign Hydration Check Audit Report\n")
	report.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// 1. External Ingestion Audit
	if *scope == "external" || *scope == "all" {
		slog.Info("Auditing External Artifact Registry (SBOM)...", "path", sbomPath)
		records, err := hydration.ParseSBOM(sbomPath)
		if err != nil {
			slog.Error("Failed to parse SBOM", "error", err)
			os.Exit(1)
		}

		report.WriteString("## External Ingestion Registry (SBOM)\n\n")

		var clis, libs []hydration.ArtifactRecord
		for _, rec := range records {
			if rec.Category == "LIBRARY" {
				libs = append(libs, rec)
			} else {
				clis = append(clis, rec)
			}
		}

		report.WriteString("### CLI / Executables / Toolchains\n")
		report.WriteString("| Target Name | Version | Download Status | Size | Expected Hash | Rationale |\n")
		report.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, rec := range clis {
			destPath := hydration.GetPhysicalPath(rec.Name)
			status := "PENDING"
			if destPath != "" {
				if _, err := os.Stat(destPath); err == nil {
					status = "VERIFIED"
				}
			}
			rat := externalRationales[rec.Name]
			if rat == "" {
				rat = "Required dependency."
			}
			report.WriteString(fmt.Sprintf("| **%s** | %s | %s | %s | %s | %s |\n",
				rec.Name, rec.Version, status, hydration.FormatBytes(rec.OriginalSize), rec.PrunedHash, rat))
		}
		report.WriteString("\n")

		report.WriteString("### Libraries / SDKs / Modules\n")
		report.WriteString("| Target Name | Version | Download Status | Size | Expected Hash | Rationale |\n")
		report.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, rec := range libs {
			destPath := hydration.GetPhysicalPath(rec.Name)
			status := "PENDING"
			if destPath != "" {
				if _, err := os.Stat(destPath); err == nil {
					status = "VERIFIED"
				}
			}
			rat := externalRationales[rec.Name]
			if rat == "" {
				rat = "Required dependency."
			}
			report.WriteString(fmt.Sprintf("| **%s** | %s | %s | %s | %s | %s |\n",
				rec.Name, rec.Version, status, hydration.FormatBytes(rec.OriginalSize), rec.PrunedHash, rat))
		}
		report.WriteString("\n")
	}

	// 2. Internal Workspace Audit
	if *scope == "internal" || *scope == "all" {
		slog.Info("Auditing Internal Workspaces...", "root", projectRoot)
		report.WriteString("## Internal Workspace Targets\n")
		report.WriteString("| Workspace Name | Harness Location | Build Status | Targets |\n")
		report.WriteString("| :--- | :--- | :--- | :--- |\n")

		err := filepath.WalkDir(filepath.Join(projectRoot, "00flow"), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && d.Name() == "workspace.harness" {
				contentBytes, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				harness, err := hydration.ParseWorkspaceHarness(string(contentBytes), path)
				if err != nil {
					return nil
				}

				// Check if up-to-date
				wsPath := filepath.Dir(filepath.Dir(path))
				status := "DIRTY (Rebuild Needed)"
				
				// Standard up-to-date check using newest modtime compared to harness dynamic targets output
				var newest time.Time
				_ = filepath.Walk(wsPath, func(curr string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						name := info.Name()
						if name != ".git" && name != "s-forge" && filepath.Ext(curr) != ".exe" {
							if info.ModTime().After(newest) {
								newest = info.ModTime()
							}
						}
					}
					return nil
				})

				targetUpToDate := true
				for _, t := range harness.Targets {
					baseName := filepath.Base(t.SourcePath)
					loc, err := registry.GetTargetLocation(t.Category, "internal")
					if err == nil {
						outBin := filepath.Join(wsPath, loc, baseName)
						if t.OutputPath != "" {
							outBin = t.OutputPath
						}
						outInfo, err := os.Stat(outBin)
						if err != nil || outInfo.ModTime().Before(newest) {
							targetUpToDate = false
							break
						}
					}
				}

				if targetUpToDate && len(harness.Targets) > 0 {
					status = "UP-TO-DATE"
				}

				report.WriteString(fmt.Sprintf("| **%s** | [%s](file:///%s) | %s | %d targets |\n",
					harness.Name, filepath.Base(filepath.Dir(path)), filepath.ToSlash(path), status, len(harness.Targets)))
			}
			return nil
		})
		if err != nil {
			slog.Error("Harness discovery walk failed", "error", err)
			os.Exit(1)
		}
	}

	err := os.WriteFile(reportPath, []byte(report.String()), 0644)
	if err != nil {
		slog.Error("Failed to write report file", "error", err)
		os.Exit(1)
	}

	fmt.Println(report.String())
	slog.Info("Unified hydration check completed. Report written.", "path", reportPath)
}
