package registry_attestations

import (
	"os"
	"path/filepath"
	"testing"

	. "sov.fleet/s-hydration/400-registry"
)

func TestNewPerfectHash(t *testing.T) {
	keys := []string{"blake3", "quic-go", "qpack", "go-tpm", "wazero", "quicdl"}
	ph, err := NewPerfectHash(keys)
	if err != nil {
		t.Fatalf("failed to construct perfect hash: %v", err)
	}
	if ph.Size != len(keys) {
		t.Errorf("expected size %d, got %d", len(keys), ph.Size)
	}

	// Verify all keys are found
	for _, key := range keys {
		idx := ph.LookupIndex(key)
		if idx == -1 {
			t.Errorf("key %q not found in perfect hash", key)
		}
		if ph.Keys[idx] != key {
			t.Errorf("slotted key at index %d is %q, expected %q", idx, ph.Keys[idx], key)
		}
	}

	// Verify non-existent key returns -1
	if idx := ph.LookupIndex("non-existent"); idx != -1 {
		t.Errorf("expected -1 for non-existent key, got %d", idx)
	}
}

func TestLookupRegistryIndex(t *testing.T) {
	// Look up an existing entry
	idx := LookupRegistryIndex("blake3")
	if idx == -1 {
		t.Errorf("expected to find blake3 in registry")
	}

	entry := GetRegistryEntry(idx)
	if entry == nil {
		t.Fatalf("expected registry entry for index %d", idx)
	}
	if entry.Name != "blake3" {
		t.Errorf("expected entry name 'blake3', got %q", entry.Name)
	}

	// Look up internal entry
	idx2 := LookupRegistryIndex("quicdl")
	if idx2 == -1 {
		t.Errorf("expected to find quicdl in registry")
	}
	entry2 := GetRegistryEntry(idx2)
	if entry2 == nil || entry2.Name != "quicdl" || !entry2.IsInternal {
		t.Errorf("expected internal entry for quicdl, got %v", entry2)
	}

	// Unregistered package should return -1
	if idxMissing := LookupRegistryIndex("unknown-package"); idxMissing != -1 {
		t.Errorf("expected LookupRegistryIndex to return -1 for unknown package, got %d", idxMissing)
	}
}

func TestSerializeRegistry(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "registry-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	targetFile := filepath.Join(tempDir, "hydratedregistry.go")

	// Read current registry entries count
	idx := LookupRegistryIndex("blake3")
	if idx == -1 {
		t.Fatal("blake3 not found before serialization test")
	}
	originalEntry := GetRegistryEntry(idx)

	// Modify entry and add a new entry to verify serialization writing
	newEntry := RegistryEntry{
		Name:            "test-serial-pkg",
		Location:        "98000-internal-libraries/test-serial-pkg",
		SourceWorkspace: "s-forge",
		LastChangedUTC:  "2026-05-29T00:00:00Z",
		Category:        "LIBRARY",
		IsInternal:      true,
	}
	SetRegistryEntry(newEntry)
	RebuildRegistry()
	defer func() {
		DeleteRegistryEntry("test-serial-pkg")
		RebuildRegistry()
	}()

	// Serialize
	err = SerializeRegistry(targetFile)
	if err != nil {
		t.Fatalf("SerializeRegistry failed: %v", err)
	}

	// Verify the file was written
	content, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	contentStr := string(content)
	if !stringsContains(contentStr, `"test-serial-pkg"`) {
		t.Errorf("expected serialized file to contain 'test-serial-pkg'")
	}
	if !stringsContains(contentStr, originalEntry.Name) {
		t.Errorf("expected serialized file to contain original entry %q", originalEntry.Name)
	}
}

func TestRegistryConsistency(t *testing.T) {
	sforgeBase := `C:\aCogSpaceSeed\00flow\s-forge`
	sfabaidesBase := `C:\aCogSpaceSeed\00flow\s-fab-aides`

	entries := GetRegistryEntries()
	if len(entries) == 0 {
		t.Fatalf("registry is empty, expected conformed entries")
	}

	for _, entry := range entries {
		t.Run(entry.Name, func(t *testing.T) {
			// 1. Verify filesystem existence
			var fullPath string
			if entry.SourceWorkspace == "s-forge" {
				fullPath = filepath.Join(sforgeBase, entry.Location)
			} else if entry.SourceWorkspace == "s-fab-aides" {
				fullPath = filepath.Join(sfabaidesBase, entry.Location)
			} else {
				t.Fatalf("unknown SourceWorkspace %q for entry %q", entry.SourceWorkspace, entry.Name)
			}

			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("registry entry %q references location that does not exist: %s", entry.Name, fullPath)
			}

			// 2. Verify O(1) Perfect Hash correctness
			idx := LookupRegistryIndex(entry.Name)
			if idx == -1 {
				t.Fatalf("failed O(1) registry lookup for entry: %q", entry.Name)
			}

			retrieved := GetRegistryEntry(idx)
			if retrieved == nil {
				t.Fatalf("GetRegistryEntry returned nil for index %d (entry %q)", idx, entry.Name)
			}

			if retrieved.Name != entry.Name {
				t.Errorf("O(1) lookup mismatch: slotted index %d resolved to %q, expected %q", idx, retrieved.Name, entry.Name)
			}
		})
	}
}

func stringsContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
