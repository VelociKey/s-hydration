package registry

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// RegistryEntry represents a conformed artifact details.
type RegistryEntry struct {
	Name               string
	Location           string
	SourceWorkspace    string
	ProgramLocation    string
	LastChangedUTC     string
	Category           string
	IsInternal         bool
	UpstreamImportPath string // Used to automatically rewrite imports from upstream source namespaces
	Executable         string // Relative path to executable from workspace
}

var (
	registryLock    sync.RWMutex
	perfectHash     *PerfectHash
	registryEntries = []RegistryEntry{
		{Name: "bazel-rules-flutter", Location: "92000-external-toolchains\\bazel-rules\\rules_flutter", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-05-18T13:30:38Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "bazel-rules-go", Location: "92000-external-toolchains\\bazel-rules\\rules_go", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:17Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "bazel", Location: "92000-external-toolchains\\bazel", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:58Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "opentelemetry", Location: "92000-external-toolchains\\opentelemetry", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:57Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "opentofu", Location: "92000-external-toolchains\\opentofu", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:58Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "cosign", Location: "92000-external-toolchains\\cosign", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:38:19Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "buildifier", Location: "92000-external-toolchains\\buildifier", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:56Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "firebase-tools", Location: "92000-external-toolchains\\firebase", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:38:26Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "flutter-sdk-firehorse", Location: "92000-external-toolchains\\flutter", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:03Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "gcloud-sdk", Location: "92000-external-toolchains\\gcloud", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:52Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "gh-cli", Location: "92000-external-toolchains\\gh", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:58Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "gitleaks", Location: "92000-external-toolchains\\gitleaks", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:58Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "git", Location: "92000-external-toolchains\\git", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:02Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "golang", Location: "92000-external-toolchains\\go", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:38:12Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "icu4c", Location: "92000-external-toolchains\\icu", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-05-18T13:30:38Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "jdk-25-headless", Location: "92000-external-toolchains\\jdk", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:13Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "notary-notation", Location: "92000-external-toolchains\\notation", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:40:59Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "step-ca", Location: "92000-external-toolchains\\step-ca", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:00Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "trivy", Location: "92000-external-toolchains\\trivy", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:04Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "wasm-opt", Location: "92000-external-toolchains\\wasm\\binaryen", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:06Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "wasm-tools", Location: "92000-external-toolchains\\wasm\\wasm-tools", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:02Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "wasmer", Location: "92000-external-toolchains\\wasm\\wasmer", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:12Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "wasmtime", Location: "92000-external-toolchains\\wasm\\wasmtime", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:06Z", Category: "CLI", IsInternal: false, UpstreamImportPath: "", Executable: ""},
		{Name: "quic-go", Location: "93000-external-libraries\\go-lib-quic-go", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:09Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/quic-go/quic-go", Executable: ""},
		{Name: "qpack", Location: "93000-external-libraries\\go-lib-qpack", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:12Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/quic-go/qpack", Executable: ""},
		{Name: "go-tpm", Location: "93000-external-libraries\\go-lib-go-tpm", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:13Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/google/go-tpm", Executable: ""},
		{Name: "go-spiffe", Location: "93000-external-libraries\\go-lib-go-spiffe", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:12Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/spiffe/go-spiffe", Executable: ""},
		{Name: "memguard", Location: "93000-external-libraries\\go-lib-memguard", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:13Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/awnumar/memguard", Executable: ""},
		{Name: "wazero", Location: "93000-external-libraries\\go-lib-wazero", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:41:25Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/tetratelabs/wazero", Executable: ""},
		{Name: "blake3", Location: "93000-external-libraries\\go-lib-blake3", SourceWorkspace: "s-forge", ProgramLocation: "", LastChangedUTC: "2026-06-25T00:38:14Z", Category: "LIBRARY", IsInternal: false, UpstreamImportPath: "github.com/zeebo/blake3", Executable: ""},
		{Name: "quicdl", Location: "98000-internal-libraries\\quicdl", SourceWorkspace: "s-forge", ProgramLocation: "quicdl.go", LastChangedUTC: "2026-05-29T18:29:51Z", Category: "LIBRARY", IsInternal: true, UpstreamImportPath: "", Executable: ""},
		{Name: "construct", Location: "81000-active-source\\cmd\\fab-construct", SourceWorkspace: "s-fab-aides", ProgramLocation: "main.go", LastChangedUTC: "2026-05-29T19:15:00Z", Category: "CLI", IsInternal: true, UpstreamImportPath: "", Executable: ""},
		{Name: "inspect", Location: "81000-active-source\\cmd\\fab-inspect", SourceWorkspace: "s-fab-aides", ProgramLocation: "main.go", LastChangedUTC: "2026-05-29T19:15:00Z", Category: "CLI", IsInternal: true, UpstreamImportPath: "", Executable: ""},
	}
)

func init() {
	RebuildRegistry()
}

// RebuildRegistry initializes or regenerates the CHD perfect hash mapping from registry entries.
func RebuildRegistry() {
	registryLock.Lock()
	defer registryLock.Unlock()

	keys := make([]string, len(registryEntries))
	for i, entry := range registryEntries {
		keys[i] = entry.Name
	}
	ph, err := NewPerfectHash(keys)
	if err == nil {
		perfectHash = ph
	}
}

// LookupRegistryIndex maps key to index in HydratedRegistry, returning -1 if not found.
func LookupRegistryIndex(symbol string) int {
	registryLock.RLock()
	defer registryLock.RUnlock()

	if perfectHash == nil {
		return -1
	}
	return perfectHash.LookupIndex(symbol)
}

// GetRegistryEntry returns the RegistryEntry record for a given slot index.
func GetRegistryEntry(index int) *RegistryEntry {
	registryLock.RLock()
	defer registryLock.RUnlock()

	if perfectHash == nil || index < 0 || index >= len(perfectHash.Keys) {
		return nil
	}
	name := perfectHash.Keys[index]
	for i := range registryEntries {
		if registryEntries[i].Name == name {
			return &registryEntries[i]
		}
	}
	return nil
}

// GetRegistryEntries returns a copy of all current RegistryEntry records.
func GetRegistryEntries() []RegistryEntry {
	registryLock.RLock()
	defer registryLock.RUnlock()

	entries := make([]RegistryEntry, len(registryEntries))
	copy(entries, registryEntries)
	return entries
}

// SetRegistryEntry updates or appends a RegistryEntry dynamically.
func SetRegistryEntry(entry RegistryEntry) {
	registryLock.Lock()
	defer registryLock.Unlock()

	found := false
	for i := range registryEntries {
		if registryEntries[i].Name == entry.Name {
			registryEntries[i] = entry
			found = true
			break
		}
	}
	if !found {
		registryEntries = append(registryEntries, entry)
	}
}

// DeleteRegistryEntry removes a RegistryEntry by name dynamically.
func DeleteRegistryEntry(name string) {
	registryLock.Lock()
	defer registryLock.Unlock()

	for i := range registryEntries {
		if registryEntries[i].Name == name {
			registryEntries = append(registryEntries[:i], registryEntries[i+1:]...)
			break
		}
	}
}

// SerializeRegistry writes the current state of registry entries to the target file.
func SerializeRegistry(filePath string) error {
	registryLock.RLock()
	defer registryLock.RUnlock()

	content, err := os.ReadFile(filePath)
	var prefix, suffix string
	if err == nil {
		cStr := string(content)
		startIdx := strings.Index(cStr, "registryEntries = []RegistryEntry{")
		if startIdx != -1 {
			prefix = cStr[:startIdx+len("registryEntries = []RegistryEntry{\n")]
			endIdx := strings.Index(cStr[startIdx:], "func init() {")
			if endIdx != -1 {
				suffix = "\t}\n)\n\n" + cStr[startIdx+endIdx:]
			}
		}
	}

	// Fallbacks in case file reading fails or structure changed
	if prefix == "" {
		prefix = `package registry

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// RegistryEntry represents a conformed artifact details.
type RegistryEntry struct {
	Name               string
	Location           string
	SourceWorkspace    string
	ProgramLocation    string
	LastChangedUTC     string
	Category           string
	IsInternal         bool
	UpstreamImportPath string
	Executable         string
}

var (
	registryLock    sync.RWMutex
	perfectHash     *PerfectHash
	registryEntries = []RegistryEntry{
`
	}
	if suffix == "" {
		suffix = `	}
)

func init() {
	RebuildRegistry()
}

// RebuildRegistry initializes or regenerates the CHD perfect hash mapping from registry entries.
func RebuildRegistry() {
	registryLock.Lock()
	defer registryLock.Unlock()

	keys := make([]string, len(registryEntries))
	for i, entry := range registryEntries {
		keys[i] = entry.Name
	}
	ph, err := NewPerfectHash(keys)
	if err == nil {
		perfectHash = ph
	}
}

// LookupRegistryIndex maps key to index in HydratedRegistry, returning -1 if not found.
func LookupRegistryIndex(symbol string) int {
	registryLock.RLock()
	defer registryLock.RUnlock()

	if perfectHash == nil {
		return -1
	}
	return perfectHash.LookupIndex(symbol)
}

// GetRegistryEntry returns the RegistryEntry record for a given slot index.
func GetRegistryEntry(index int) *RegistryEntry {
	registryLock.RLock()
	defer registryLock.RUnlock()

	if perfectHash == nil || index < 0 || index >= len(perfectHash.Keys) {
		return nil
	}
	name := perfectHash.Keys[index]
	for i := range registryEntries {
		if registryEntries[i].Name == name {
			return &registryEntries[i]
		}
	}
	return nil
}

// GetRegistryEntries returns a copy of all current RegistryEntry records.
func GetRegistryEntries() []RegistryEntry {
	registryLock.RLock()
	defer registryLock.RUnlock()

	entries := make([]RegistryEntry, len(registryEntries))
	copy(entries, registryEntries)
	return entries
}

// SetRegistryEntry updates or appends a RegistryEntry dynamically.
func SetRegistryEntry(entry RegistryEntry) {
	registryLock.Lock()
	defer registryLock.Unlock()

	found := false
	for i := range registryEntries {
		if registryEntries[i].Name == entry.Name {
			registryEntries[i] = entry
			found = true
			break
		}
	}
	if !found {
		registryEntries = append(registryEntries, entry)
	}
}

// DeleteRegistryEntry removes a RegistryEntry by name dynamically.
func DeleteRegistryEntry(name string) {
	registryLock.Lock()
	defer registryLock.Unlock()

	for i := range registryEntries {
		if registryEntries[i].Name == name {
			registryEntries = append(registryEntries[:i], registryEntries[i+1:]...)
			break
		}
	}
}

// SerializeRegistry writes the current state of registry entries to the target file.
func SerializeRegistry(filePath string) error {
	return SerializeRegistry(filePath)
}
`
	}

	var sb strings.Builder
	sb.WriteString(prefix)
	for _, entry := range registryEntries {
		sb.WriteString(fmt.Sprintf("\t\t{Name: %q, Location: %q, SourceWorkspace: %q, ProgramLocation: %q, LastChangedUTC: %q, Category: %q, IsInternal: %t, UpstreamImportPath: %q, Executable: %q},\n",
			entry.Name, entry.Location, entry.SourceWorkspace, entry.ProgramLocation, entry.LastChangedUTC, entry.Category, entry.IsInternal, entry.UpstreamImportPath, entry.Executable))
	}
	if !strings.HasSuffix(prefix, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(suffix)

	return os.WriteFile(filePath, []byte(sb.String()), 0644)
}
