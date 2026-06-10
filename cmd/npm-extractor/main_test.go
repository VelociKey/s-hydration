package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarballBinaries(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "npm-extractor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockTarball := filepath.Join(tempDir, "mock_package.tgz")
	file, err := os.Create(mockTarball)
	if err != nil {
		t.Fatalf("failed to create mock tarball file: %v", err)
	}

	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)

	// Write a mock binary file inside the tarball
	binaryName := "package/bin/jules.exe"
	binaryContent := []byte("mock binary executable payload")

	header := &tar.Header{
		Name:     binaryName,
		Mode:     0755,
		Size:     int64(len(binaryContent)),
		Typeflag: tar.TypeReg,
	}

	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("failed to write header to tar: %v", err)
	}
	if _, err := tw.Write(binaryContent); err != nil {
		t.Fatalf("failed to write body to tar: %v", err)
	}

	// Write an arbitrary JS file to verify it is ignored by extractor
	jsHeader := &tar.Header{
		Name:     "package/index.js",
		Mode:     0644,
		Size:     13,
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(jsHeader); err != nil {
		t.Fatalf("failed to write js header: %v", err)
	}
	if _, err := tw.Write([]byte("console.log()")); err != nil {
		t.Fatalf("failed to write js content: %v", err)
	}

	tw.Close()
	gzw.Close()
	file.Close()

	// Run extraction
	extractedBinary, err := extractTarballBinaries(mockTarball, tempDir)
	if err != nil {
		t.Fatalf("failed to extract mock tarball: %v", err)
	}

	if extractedBinary == "" {
		t.Fatal("expected to extract binary, got none")
	}

	if filepath.Base(extractedBinary) != "jules.exe" {
		t.Errorf("expected extracted binary filename to be jules.exe, got %s", filepath.Base(extractedBinary))
	}

	// Verify JS file was not extracted
	jsPath := filepath.Join(tempDir, "index.js")
	if _, err := os.Stat(jsPath); !os.IsNotExist(err) {
		t.Error("index.js was extracted but should have been ignored")
	}
}

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		pkg      string
		expected string
	}{
		{"@google/jules", "jules"},
		{"jules-cli", "jules"},
		{"@google/jules-tools", "jules-tools"},
	}

	for _, tc := range tests {
		res := sanitizePackageName(tc.pkg)
		if res != tc.expected {
			t.Errorf("sanitizePackageName(%q) = %q; expected %q", tc.pkg, res, tc.expected)
		}
	}
}
