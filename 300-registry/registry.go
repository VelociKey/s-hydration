package registry

import (
	"path/filepath"
	"strings"
	"sov.fleet/sLatentLingua/02000-logic-libraries/snparser"
	"sov.fleet/sLatentLingua/02000-logic-libraries/ir"
)

// TransitionPath implements the declarative "UniversalPath -> HostPlatformPath" grammar rule.
// It mathematically converts the LLM-optimized universal forward slashes into the guaranteed
// native host path separators, preventing compiler resolution bugs.
func TransitionPath(universalPath string) string {
	// Strip any trailing semicolons or whitespace that might leak from raw tokens
	cleanStr := strings.TrimSpace(universalPath)
	cleanStr = strings.TrimSuffix(cleanStr, ";")
	cleanStr = strings.Trim(cleanStr, `"'`)
	
	// Apply the OS translation mechanism
	native := filepath.FromSlash(cleanStr)
	return filepath.Clean(native)
}

// ExtractLogicalPaths reads a parsed IR root and attempts to extract flat 
// coordinate mappings by leveraging the DataParser.
func ExtractLogicalPaths(node *ir.Node) map[string]string {
	if node == nil {
		return nil
	}

	paths := make(map[string]string)
	for _, child := range node.Children {
		if child.Type == ir.NodeAttribute {
			name := child.GetAttributeString("name")
			if len(child.Children) > 0 && child.Children[0].Type == ir.NodeValue {
				val := child.Children[0].GetAttributeString("value")
				// Automatically apply the Transition Rule to all mapped values
				paths[name] = TransitionPath(val)
			}
		}
	}
	return paths
}

// ParseConfigFile is a helper to read WEBNF configuration using the core lexer/parser
func ParseConfigFile(content string) (*ir.Node, error) {
	lexer := snparser.NewLexer(content)
	dataParser := snparser.NewDataParser(lexer)
	return dataParser.Parse()
}
