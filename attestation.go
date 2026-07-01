package hydration

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"sov.fleet/s-hydration/pkg/asthash"
)

// GenerateAttestationRegistry walks all workspaces and computes Blake3 AST hashes of Go files,
// writing them to c0990-ephemeral-scratch/build_hashes.webnf.
func GenerateAttestationRegistry(projectRoot string) error {
	destPath := filepath.Join(projectRoot, "00flow", "s-hydrationcache", "c0990-ephemeral-scratch", "build_hashes.webnf")

	var sb strings.Builder
	sb.WriteString("; Authoritative Blake3 Build Invariant Attestation Registry\n")
	sb.WriteString("build_invariants {\n")

	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "c0990-ephemeral-scratch" || name == "build-caches" || name == "s-forge" || name == "86sref" || name == "s-hydrationcache" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
			// Compute AST-Symbol Hash using Blake3
			hash, err := asthash.ComputeASTHash(path)
			if err == nil {
				// Convert absolute path to a relative logical key format for WebNF format
				rel, err := filepath.Rel(projectRoot, path)
				if err == nil {
					key := strings.ReplaceAll(strings.ReplaceAll(rel, "\\", "_"), "/", "_")
					key = strings.ReplaceAll(strings.ReplaceAll(key, ".", "_"), "-", "_")
					// Guard key starting with a digit
					if len(key) > 0 && (key[0] >= '0' && key[0] <= '9') {
						key = "file_" + key
					}
					sb.WriteString(fmt.Sprintf("    %s = %q ;\n", key, hash))
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	sb.WriteString("}\n")

	_ = os.MkdirAll(filepath.Dir(destPath), 0755)
	err = os.WriteFile(destPath, []byte(sb.String()), 0644)
	if err != nil {
		return err
	}

	slog.Info("Blake3 Build Invariant Attestation Registry successfully updated", "path", destPath)
	return nil
}
