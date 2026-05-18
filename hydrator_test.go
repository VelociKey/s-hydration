package shydration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHydratorPluginRegistration(t *testing.T) {
	hydrator := NewHydrator()
	
	// Register mechanisms
	hydrator.Register(&DynamicSynthesisMechanism{})
	hydrator.Register(&StaticIngestionMechanism{})
	hydrator.Register(&MCPDiscoveryMechanism{})
	
	// Verify count
	if len(hydrator.plugins) != 3 {
		t.Errorf("Expected 3 registered plugins, got %d", len(hydrator.plugins))
	}
	
	// Verify modes
	if _, ok := hydrator.plugins[ModeDynamic]; !ok {
		t.Error("Dynamic synthesis plugin missing")
	}
	if _, ok := hydrator.plugins[ModeStatic]; !ok {
		t.Error("Static ingestion plugin missing")
	}
	if _, ok := hydrator.plugins[ModeDiscover]; !ok {
		t.Error("MCP Capability Discovery plugin missing")
	}
}

func TestDynamicSynthesisPlugin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dynamic-synthesis-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	outputPath := filepath.Join(tmpDir, "internal_actor_main")
	target := HydrationTarget{
		Mode:       ModeDynamic,
		Engine:     "go-wasm-compiler",
		SourcePath: "00flow/s-forge/94000-internal-actors/source",
		OutputPath: outputPath,
	}
	
	hydrator := NewHydrator()
	hydrator.Register(&DynamicSynthesisMechanism{})
	
	err = hydrator.Execute(context.Background(), target)
	if err != nil {
		t.Fatalf("Hydration execute failed: %v", err)
	}
	
	// Verify output file exists
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	if !strings.Contains(string(content), "Mode: DYNAMIC") {
		t.Errorf("Expected output to contain 'Mode: DYNAMIC', got %q", string(content))
	}
	if !strings.Contains(string(content), "Engine: go-wasm-compiler") {
		t.Errorf("Expected output to contain 'Engine: go-wasm-compiler', got %q", string(content))
	}
}

func TestStaticIngestionPlugin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "static-ingestion-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	outputPath := filepath.Join(tmpDir, "external_toolchain_trivy")
	target := HydrationTarget{
		Mode:       ModeStatic,
		Engine:     "github-release-retriever",
		SourcePath: "https://github.com/aquasecurity/trivy/releases/v0.51.0",
		OutputPath: outputPath,
	}
	
	hydrator := NewHydrator()
	hydrator.Register(&StaticIngestionMechanism{})
	
	err = hydrator.Execute(context.Background(), target)
	if err != nil {
		t.Fatalf("Hydration execute failed: %v", err)
	}
	
	// Verify output file exists
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	if !strings.Contains(string(content), "Mode: STATIC") {
		t.Errorf("Expected output to contain 'Mode: STATIC', got %q", string(content))
	}
}

func TestMCPDiscoveryPlugin(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mcp-discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	outputPath := filepath.Join(tmpDir, "capability_catalog.webnf")
	target := HydrationTarget{
		Mode:       ModeDiscover,
		Engine:     "mcp-client-discovery",
		SourcePath: "127.0.0.1:8080/mcp",
		OutputPath: outputPath,
	}
	
	hydrator := NewHydrator()
	mcpPlugin := &MCPDiscoveryMechanism{
		MockCapabilities: []string{"git_commit", "trivy_audit", "jules_self_heal"},
	}
	hydrator.Register(mcpPlugin)
	
	err = hydrator.Execute(context.Background(), target)
	if err != nil {
		t.Fatalf("Hydration execute failed: %v", err)
	}
	
	// Verify output file exists
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	
	fileStr := string(content)
	if !strings.Contains(fileStr, "Mode: DISCOVER") {
		t.Errorf("Expected output to contain 'Mode: DISCOVER', got %q", fileStr)
	}
	if !strings.Contains(fileStr, "capability_catalog {") {
		t.Errorf("Expected output to open capability_catalog block, got %q", fileStr)
	}
	if !strings.Contains(fileStr, "MCP_TOOL_jules_self_heal = \"enabled\" ;") {
		t.Errorf("Expected output to enable discovered tool, got %q", fileStr)
	}
}

func TestParseWebNFTargets(t *testing.T) {
	webnfContent := `; Sample targets file
workspace_harness {
    name : "test_workspace" ;
    targets {
        target {
            mode : DYNAMIC ;
            engine : "go-wasm-compiler" ;
            src : "00flow/s-forge/94000-internal-actors/source" ;
            out : "C:/aCogSpaceSeed/00flow/s-forge/94000-internal-actors/bin/actor_test" ;
        }
        target {
            mode : STATIC ;
            engine : "git-lfs-mirror" ;
            src : "https://github.com/aquasecurity/trivy/releases" ;
            out : "C:/aCogSpaceSeed/00flow/s-forge/92000-external-toolchains/trivy" ;
        }
    }
}
`
	targets, err := ParseWebNFTargets(webnfContent)
	if err != nil {
		t.Fatalf("ParseWebNFTargets failed: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("Expected 2 parsed targets, got %d", len(targets))
	}

	if targets[0].Mode != ModeDynamic || targets[0].Engine != "go-wasm-compiler" {
		t.Errorf("Mismatch in first target: %+v", targets[0])
	}
	if targets[1].Mode != ModeStatic || targets[1].Engine != "git-lfs-mirror" {
		t.Errorf("Mismatch in second target: %+v", targets[1])
	}
}

func TestInvokeToAddPermanent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "invoke-to-add-permanent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	webnfPath := filepath.Join(tmpDir, "workspace_hydrator.webnf")
	outputPath := filepath.Join(tmpDir, "sealed_actor_main")

	hydrator := NewHydrator()
	hydrator.Register(&DynamicSynthesisMechanism{})

	input := InvokeInput{
		Mode:       ModeDynamic,
		Engine:     "licensee-go-compiler",
		SourcePath: "licensee/custom-actors/src",
		OutputPath: outputPath,
	}

	// Invoke permanent inclusion
	err = hydrator.InvokeToAdd(context.Background(), input, true, webnfPath)
	if err != nil {
		t.Fatalf("InvokeToAdd failed: %v", err)
	}

	// 1. Verify target was run successfully
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read dynamic actor output: %v", err)
	}
	if !strings.Contains(string(content), "Engine: licensee-go-compiler") {
		t.Errorf("Expected output to contain engine name, got %q", string(content))
	}

	// 2. Verify targets database file exists and contains the syntactically correct block
	webnfBytes, err := os.ReadFile(webnfPath)
	if err != nil {
		t.Fatalf("Failed to read persisted WebNF database: %v", err)
	}
	webnfStr := string(webnfBytes)
	if !strings.Contains(webnfStr, "mode : DYNAMIC ;") {
		t.Errorf("Persisted WebNF missing mode string, got %q", webnfStr)
	}
	if !strings.Contains(webnfStr, `engine : "licensee-go-compiler" ;`) {
		t.Errorf("Persisted WebNF missing engine string, got %q", webnfStr)
	}

	// 3. Verify we can parse it back using our ParseWebNFTargets
	parsedTargets, err := ParseWebNFTargets(webnfStr)
	if err != nil {
		t.Fatalf("Failed to re-parse generated WebNF config: %v", err)
	}
	if len(parsedTargets) != 1 {
		t.Fatalf("Expected 1 re-parsed target, got %d", len(parsedTargets))
	}
	if parsedTargets[0].Engine != "licensee-go-compiler" || parsedTargets[0].Mode != ModeDynamic {
		t.Errorf("Parsed target mismatch: %+v", parsedTargets[0])
	}
}

func TestInvokeToAddEphemeral(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "invoke-to-add-ephemeral-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	webnfPath := filepath.Join(tmpDir, "workspace_hydrator.webnf")
	outputPath := filepath.Join(tmpDir, "sealed_actor_main")

	hydrator := NewHydrator()
	hydrator.Register(&DynamicSynthesisMechanism{})

	input := InvokeInput{
		Mode:       ModeDynamic,
		Engine:     "short-term-engine",
		SourcePath: "ephemeral/src",
		OutputPath: outputPath,
	}

	// Invoke ephemeral inclusion
	err = hydrator.InvokeToAdd(context.Background(), input, false, webnfPath)
	if err != nil {
		t.Fatalf("InvokeToAdd failed: %v", err)
	}

	// 1. Verify target was run successfully
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read dynamic actor output: %v", err)
	}
	if !strings.Contains(string(content), "Engine: short-term-engine") {
		t.Errorf("Expected output to contain ephemeral engine, got %q", string(content))
	}

	// 2. Verify targets database file does NOT exist (short-term inclusion!)
	if _, err := os.Stat(webnfPath); !os.IsNotExist(err) {
		t.Error("Expected WebNF file to NOT be created for ephemeral inclusion")
	}
}
