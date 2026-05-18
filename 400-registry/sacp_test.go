package registry

import (
	"os"
	"testing"
	"sov.fleet/s-latentlingua/02000-logic-libraries/snparser"
)

func TestSACPGrammarSyntax(t *testing.T) {
	// 1. Read the grammar file
	content, err := os.ReadFile("sacp.wag")
	if err != nil {
		t.Fatalf("Failed to read sacp.wag: %v", err)
	}

	// 2. Parse the WAG grammar
	lexer := snparser.NewLexer(string(content))
	parser := snparser.NewParser(lexer)
	grammarAST, err := parser.Parse()

	if err != nil {
		t.Fatalf("Syntax error in sacp.wag: %v", err)
	}

	// 3. Verify key rules are present
	expectedRules := []string{"SACPMessage", "QAPCHeaderBlock", "AuthorityLevel", "AttestationBlock", "CapabilityFrameBlock"}
	for _, rule := range expectedRules {
		if _, ok := grammarAST.Rules[rule]; !ok {
			t.Errorf("Expected rule %q not found in AST of sacp.wag", rule)
		}
	}

	t.Log("Successfully verified sacp.wag structural grammar syntax!")
}
