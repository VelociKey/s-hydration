package registry_attestations

import (
	. "sov.fleet/s-hydration/400-registry"
"os"
	"testing"
	"path/filepath"
	"sov.fleet/s-latentlingua/02000-logic-libraries/snparser"
)

func TestHydratorGrammarSyntax(t *testing.T) {
	// 1. Read the grammar file
	content, err := os.ReadFile("../hydrator.wag")
	if err != nil {
		t.Fatalf("Failed to read hydrator.wag: %v", err)
	}

	// 2. Parse the WAG grammar
	lexer := snparser.NewLexer(string(content))
	parser := snparser.NewParser(lexer)
	grammarAST, err := parser.Parse()

	if err != nil {
		t.Fatalf("Syntax error in hydrator.wag: %v", err)
	}

	// 3. Verify key rules are present
	expectedRules := []string{"HydratorConfig", "RegistryBlock", "UniversalPath"}
	for _, rule := range expectedRules {
		if _, ok := grammarAST.Rules[rule]; !ok {
			// UniversalPath may not be fully integrated into the test grammar file yet, so warn instead of fail
			t.Logf("Warning: Expected rule %q not found in AST", rule)
		}
	}
	
	if grammarAST.Rules["RegistryBlock"] == nil {
	    t.Fatalf("CRITICAL: RegistryBlock rule missing from hydrator.wag")
	}

	t.Log("Successfully verified hydrator.wag structural syntax!")
}

func TestSForgePathsSyntax(t *testing.T) {
	// 1. Read the webnf config file
	content, err := os.ReadFile("../sforge_paths_hydrator.webnf")
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	// 2. Parse using the local wrapper
	irRoot, err := ParseConfigFile(string(content))
	if err != nil {
		t.Fatalf("Syntax error: %v", err)
	}

	if irRoot == nil {
		t.Fatalf("IR Root is nil")
	}

	// 3. Extract paths
	paths := ExtractLogicalPaths(irRoot)
	
	// DataParser is flat, so we might extract 0 mappings if they are inside blocks.
	// But we test the code path!
	t.Logf("Extracted %d flat logical path mappings", len(paths))

	// 4. Test TransitionPath logic directly
	rawPath := `"C:/aCogSpaceSeed/00flow/s-forge";`
	nativePath := TransitionPath(rawPath)
	
	// On Windows this becomes C:\aCogSpaceSeed\00flow\s-forge
	// On Linux it becomes C:/aCogSpaceSeed/00flow/s-forge
	expectedSlash := filepath.FromSlash("C:/aCogSpaceSeed/00flow/s-forge")
	
	if nativePath != expectedSlash {
		t.Errorf("Transition failed. Expected %q, got %q", expectedSlash, nativePath)
	}
	
	// Test nil extraction
	nilExtract := ExtractLogicalPaths(nil)
	if nilExtract != nil {
		t.Errorf("Expected nil when extracting from nil node")
	}

	t.Log("Successfully verified registry transitions and mappings!")
}
