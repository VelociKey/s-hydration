package registry_attestations

import (
	. "sov.fleet/s-hydration/400-registry"
"path/filepath"
	"testing"
)

func TestGetTargetLocation(t *testing.T) {
	tests := []struct {
		category string
		origin   string
		expected string
		expectErr bool
	}{
		// External authority range
		{"binary", "external", "91000-external-executables", false},
		{"BINARY", "external", "91000-external-executables", false},
		{"executable", "EXTERNAL", "91000-external-executables", false},
		{"toolchain", "external", "92000-external-toolchains", false},
		{"library", "external", "93000-external-libraries", false},
		{"actor", "external", "94000-external-actors", false},

		// Internal authority range
		{"binary", "internal", "96000-internal-executables", false},
		{"toolchain", "internal", "97000-internal-toolchains", false},
		{"library", "internal", "98000-internal-libraries", false},
		{"actor", "internal", "99000-internal-actors", false},

		// External source/genetic range
		{"binary", "source", "81000-external-executables-source", false},
		{"toolchain", "source-external", "82000-external-toolchains-source", false},
		{"library", "source", "83000-external-libraries-source", false},
		{"actor", "source-external", "84000-external-actors-source", false},

		// Internal source/genetic range
		{"binary", "source-internal", "86000-internal-executables-source", false},
		{"toolchain", "source-internal", "87000-internal-toolchains-source", false},
		{"library", "source-internal", "88000-internal-libraries-source", false},
		{"actor", "source-internal", "89000-internal-actors-source", false},

		// Trim and casing variations
		{"  library  ", "  INTERNAL  ", "98000-internal-libraries", false},

		// Error cases
		{"invalid_category", "external", "", true},
		{"binary", "invalid_origin", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.category+"_"+tc.origin, func(t *testing.T) {
			got, err := GetTargetLocation(tc.category, tc.origin)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error for category=%q, origin=%q, but got nil", tc.category, tc.origin)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, got)
				}
			}
		})
	}
}

func TestGetTargetPhysicalPath(t *testing.T) {
	t.Run("default sforge base", func(t *testing.T) {
		got, err := GetTargetPhysicalPath("", "library", "internal")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(`C:\aCogSpaceSeed\00flow\s-forge`, "98000-internal-libraries")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("custom base path", func(t *testing.T) {
		got, err := GetTargetPhysicalPath(`C:\my\custom\base`, "actor", "external")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := filepath.Join(`C:\my\custom\base`, "94000-external-actors")
		if got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("error propagation", func(t *testing.T) {
		_, err := GetTargetPhysicalPath("", "bad", "bad")
		if err == nil {
			t.Errorf("expected error on invalid input but got nil")
		}
	})
}
