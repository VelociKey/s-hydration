package asthash

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// writeTempGoFile writes a Go source string to a temporary file and returns its path.
func writeTempGoFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "temp.go")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write temp go file: %v", err)
	}
	return filePath
}

func TestComputeASTHash_Basic(t *testing.T) {
	content1 := `package main
// Comment here
import "fmt"

func Main() {
	fmt.Println("Hello")
}
`
	content2 := `package main

import "fmt"

func Main() {
	fmt.Println("Hello")
}
`

	fp1 := writeTempGoFile(t, content1)
	fp2 := writeTempGoFile(t, content2)

	hash1, err := ComputeASTHash(fp1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hash2, err := ComputeASTHash(fp2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash1 == "" {
		t.Error("hash should not be empty")
	}

	if hash1 != hash2 {
		t.Errorf("expected hashes to match for semantically identical content, got: %s vs %s", hash1, hash2)
	}
}

func TestComputeASTHash_StressConcurrency(t *testing.T) {
	// Create several files with varying content
	var files []string
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf(`package stress%d
type Struct%d struct {
	FieldA int
	FieldB string
}
`, i, i)
		files = append(files, writeTempGoFile(t, content))
	}

	const numGoroutines = 200
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			// Pick a file at random and compute its hash
			fp := files[rand.Intn(len(files))]
			_, err := ComputeASTHash(fp)
			if err != nil {
				t.Errorf("Goroutine %d: error computing AST hash: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}

func TestComputeASTHash_StressDeepNesting(t *testing.T) {
	// Generate a deeply nested/large Go source file structures
	var nestedStructs string
	for i := 0; i < 1000; i++ {
		nestedStructs += fmt.Sprintf("type DeepStruct%d struct {\n\tChild *DeepStruct%d\n\tVal int\n}\n\n", i, i+1)
	}

	content := fmt.Sprintf(`package deepnesting
%s
func Process() {
	_ = &DeepStruct0{}
}
`, nestedStructs)

	fp := writeTempGoFile(t, content)

	hash, err := ComputeASTHash(fp)
	if err != nil {
		t.Fatalf("failed to hash deeply nested AST structure: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash for nested structure")
	}
}

func BenchmarkComputeASTHash(b *testing.B) {
	content := `package benchmark
type Record struct {
	ID        int
	Name      string
	Value     float64
	IsActive  bool
}

func ProcessRecord(r *Record) string {
	if r.IsActive {
		return r.Name
	}
	return ""
}
`
	tmpDir := b.TempDir()
	filePath := filepath.Join(tmpDir, "benchmark.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write temp file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ComputeASTHash(filePath)
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
	}
}
