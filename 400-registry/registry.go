package registry

import (
	"fmt"
	"path/filepath"
	"strings"
	"sov.fleet/s-latentlingua/02000-logic-libraries/snparser"
	"sov.fleet/s-latentlingua/02000-logic-libraries/ir"
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
	stack := []*ir.Node{node}
	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if curr.Type == ir.NodeAttribute {
			name := curr.GetAttributeString("name")
			if len(curr.Children) > 0 && curr.Children[0].Type == ir.NodeValue {
				val := curr.Children[0].GetAttributeString("value")
				// Automatically apply the Transition Rule to all mapped values
				paths[name] = TransitionPath(val)
			}
		}

		for i := len(curr.Children) - 1; i >= 0; i-- {
			stack = append(stack, curr.Children[i])
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

// GetTargetLocation maps a category (case-insensitive e.g. "binary", "toolchain", "library", "actor")
// and origin (case-insensitive e.g. "external", "internal", "source"/"source-external", "source-internal")
// to its corresponding directory name in the 5-digit/4-digit decade taxonomy structure under s-forge.
func GetTargetLocation(category string, origin string) (string, error) {
	cat := strings.ToUpper(strings.TrimSpace(category))
	ori := strings.ToLower(strings.TrimSpace(origin))

	// Normalize category synonyms
	switch cat {
	case "BINARY", "EXECUTABLE":
		cat = "BINARY"
	case "TOOLCHAIN":
		cat = "TOOLCHAIN"
	case "LIBRARY":
		cat = "LIBRARY"
	case "ACTOR":
		cat = "ACTOR"
	default:
		return "", fmt.Errorf("unrecognized category: %q", category)
	}

	switch ori {
	case "external":
		switch cat {
		case "BINARY":
			return "91000-external-executables", nil
		case "TOOLCHAIN":
			return "92000-external-toolchains", nil
		case "LIBRARY":
			return "93000-external-libraries", nil
		case "ACTOR":
			return "94000-external-actors", nil
		}
	case "internal":
		switch cat {
		case "BINARY":
			return "96000-internal-executables", nil
		case "TOOLCHAIN":
			return "97000-internal-toolchains", nil
		case "LIBRARY":
			return "98000-internal-libraries", nil
		case "ACTOR":
			return "99000-internal-actors", nil
		}
	case "source", "source-external":
		switch cat {
		case "BINARY":
			return "81000-external-executables-source", nil
		case "TOOLCHAIN":
			return "82000-external-toolchains-source", nil
		case "LIBRARY":
			return "83000-external-libraries-source", nil
		case "ACTOR":
			return "84000-external-actors-source", nil
		}
	case "source-internal":
		switch cat {
		case "BINARY":
			return "86000-internal-executables-source", nil
		case "TOOLCHAIN":
			return "87000-internal-toolchains-source", nil
		case "LIBRARY":
			return "88000-internal-libraries-source", nil
		case "ACTOR":
			return "89000-internal-actors-source", nil
		}
	default:
		return "", fmt.Errorf("unrecognized origin: %q", origin)
	}

	return "", fmt.Errorf("unhandled category/origin mapping: category=%q, origin=%q", category, origin)
}

// GetTargetPhysicalPath resolves the target location directory name and joins it with the provided base directory.
// If sforgeBase is empty, it defaults to the standard location `C:\aCogSpaceSeed\00flow\s-forge`.
func GetTargetPhysicalPath(sforgeBase string, category string, origin string) (string, error) {
	if sforgeBase == "" {
		sforgeBase = `C:\aCogSpaceSeed\00flow\s-forge`
	}
	loc, err := GetTargetLocation(category, origin)
	if err != nil {
		return "", err
	}
	return filepath.Join(sforgeBase, loc), nil
}
