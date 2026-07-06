package asthash

import (
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"unsafe"

	"sov.fleet/blake3"
)

var (
	prefixFunc  = []byte("func:")
	prefixType  = []byte("type:")
	prefixVal   = []byte("val:")
	prefixField = []byte("field:")
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
	sw, hasStringWriter := any(hasher).(io.StringWriter)

	// Walk the AST nodes and hash their structural representations
	ast.Inspect(fileNode, func(n ast.Node) bool {
		if n == nil {
			return true
		}

		switch x := n.(type) {
		case *ast.FuncDecl:
			// Hash function signature details (name, params, results)
			if hasStringWriter {
				_, _ = sw.WriteString("func:")
				_, _ = sw.WriteString(x.Name.Name)
			} else {
				_, _ = hasher.Write(prefixFunc)
				b := unsafe.Slice(unsafe.StringData(x.Name.Name), len(x.Name.Name))
				_, _ = hasher.Write(b)
			}
		case *ast.TypeSpec:
			// Hash struct definition type and names
			if hasStringWriter {
				_, _ = sw.WriteString("type:")
				_, _ = sw.WriteString(x.Name.Name)
			} else {
				_, _ = hasher.Write(prefixType)
				b := unsafe.Slice(unsafe.StringData(x.Name.Name), len(x.Name.Name))
				_, _ = hasher.Write(b)
			}
		case *ast.ValueSpec:
			// Hash constant/variable specifications
			for _, name := range x.Names {
				if hasStringWriter {
					_, _ = sw.WriteString("val:")
					_, _ = sw.WriteString(name.Name)
				} else {
					_, _ = hasher.Write(prefixVal)
					b := unsafe.Slice(unsafe.StringData(name.Name), len(name.Name))
					_, _ = hasher.Write(b)
				}
			}
		case *ast.StructType:
			// Hash struct field layout
			if x.Fields != nil {
				for _, f := range x.Fields.List {
					for _, name := range f.Names {
						if hasStringWriter {
							_, _ = sw.WriteString("field:")
							_, _ = sw.WriteString(name.Name)
						} else {
							_, _ = hasher.Write(prefixField)
							b := unsafe.Slice(unsafe.StringData(name.Name), len(name.Name))
							_, _ = hasher.Write(b)
						}
					}
				}
			}
		}
		return true
	})

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
