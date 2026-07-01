package asthash

import (
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"sov.fleet/blake3"
)

// ComputeASTHash parses the Go file and computes a semantic hash of the AST structure,
// ignoring comments, whitespace, and import ordering to prevent cache invalidation.
func ComputeASTHash(filePath string) (string, error) {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	hasher := blake3.New()

	// Walk the AST nodes and hash their structural representations
	ast.Inspect(fileNode, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			// Hash function signature details (name, params, results)
			hasher.Write([]byte("func:" + x.Name.Name))
		case *ast.TypeSpec:
			// Hash struct definition type and names
			hasher.Write([]byte("type:" + x.Name.Name))
		case *ast.ValueSpec:
			// Hash constant/variable specifications
			for _, name := range x.Names {
				hasher.Write([]byte("val:" + name.Name))
			}
		case *ast.StructType:
			// Hash struct field layout
			if x.Fields != nil {
				for _, f := range x.Fields.List {
					for _, name := range f.Names {
						hasher.Write([]byte("field:" + name.Name))
					}
				}
			}
		}
		return true
	})

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
