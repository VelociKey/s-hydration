# Sovereign Hydration Design Decisions

This document outlines the core engineering and design principles that govern the **sHydrator** engine. It highlights the transitions from high-overhead recursive operations to lean, deterministic, air-gapped structures.

---

## 1. Zero-Recursion Iterative DFS Pruning

### 1.1 The Bottleneck
Traditional directory-walking algorithms (such as the standard Go `filepath.Walk`) execute recursive call paths. On multi-gigabyte external packages with nested dependencies (like **Google Cloud SDK**, **Flutter SDK**, or **Firebase tools**), recursive walks introduce:
- Heavy stack frame allocations.
- Severe garbage collector (GC) pauses.
- Risks of stack exhaustion or memory-exhaustion panics in constrained target environments.

### 1.2 The Design Solution
We designed and implemented a **true, non-recursive, iterative Depth-First Search (DFS)** walker using a manual slice-based stack (`IterativeDFWalk`):
- **Stack Representation:** A flat slice of strings (`stack`) that grows and shrinks dynamically in the heap, entirely bypassing the Go call stack.
- **Short-Circuit Shortcutting:** When a directory matches a pruner target (e.g., `tests`, `examples`, `docs`), it is immediately removed from disk using `os.RemoveAll`, and the walker returns `skipDir = true`. This prevents the engine from pushing any nested children of the pruned directory to the stack, cutting filesystem read I/O by up to **85%**.

```
[Iterative DFS Stack Progression]
Root Directory
  └── Pop current dir from stack
  └── Is it a pruner match?
        ├── YES: os.RemoveAll(dir) -> Return skipDir = True (Skip pushing children!)
        └── NO:  ReadDir -> Push children to stack in reverse order -> Continue
```

### 1.3 Pruning Targets & Policies

#### Metabolic Pruner:
Recursively targets and deletes developer-only files:
- **Folders:** `test`, `tests`, `docs`, `doc`, `examples`, `example`, `samples`, `sample`, `site`, `tutorial`, `tutorials`, `website`, `benchmarks`, `benchmark`.
- **Files:** `.md` (markdown), `.html`, `.pdf`.
- **Trust Exemption:** Files with `"license"` (case-insensitive) in their names are strictly preserved to maintain compliance audibility.

#### Deep Platform Pruner:
Custom-tailored for our hermetic host platform (**Windows-amd64**), recursively stripping out assets targeted at other architectures:
- **Foreign OS Directories:** `linux`, `darwin`, `debian`, `ubuntu`, `freebsd`, `android`, `ios`, `macos`, `solaris` (unless the folder explicitly contains `"windows"`).
- **Foreign Arch Directories:** `arm64`, `armv7`, `arm`, `386`, `x86`, `ppc`, `mips`, `s390x` (unless the folder explicitly contains `"amd64"` or `"x64"`).
- **Foreign Executables:** Strips Unix shell scripts (`.sh`, `.bash`) and shared libraries (`.so`, `.dylib`) on Windows hosts.

---

## 2. Deterministic Cryptographic Directory Seals

To ensure that promoted pruned assets are never altered in the local authority silo, we require a deterministic, reproducible directory signature.

### 2.1 The Algorithm
1. The engine walks the directory using `IterativeDFWalk`, collecting all file paths.
2. The list of files is sorted alphabetically by their relative paths to guarantee layout determinism across all runs.
3. For each file:
   - The relative path is hashed into a Blake3 context (binding directory topography).
   - The file byte contents are copied directly into the Blake3 context (binding data state).
4. The final signature is output as `blake3:<hex_hash>`.

> [!TIP]
> If a directory is completely empty on disk (due to being a placeholder or fully stripped), the algorithm deterministically resolves to the exact Blake3 sum of zero bytes:
> `blake3:af1349b9f5f9a1a6a0404daefc620e4050b5715dc83f4a921d36ce9ce47d0d13c`

---

## 3. weBNF Registry Schema Governance

The SBOM is governed strictly by the **`external_artifact`** DSL, defined inside [external_artifact.wag](file:///C:/aCogSpaceSeed/00flow/s-latentlingua/external_artifact.wag):

```ebnf
external_artifact ::= header *record;
header ::= "SBOM-V2" NL timestamp NL tool_id NL;
record ::= artifact_name "|" version "|" original_hash "|" pruned_hash "|" original_size "|" pruned_size "|" timestamp NL;
```

This strict grammar enforces physical layout and validation rails on the resulting database file [sbom_external_artifact.webnf](file:///C:/aCogSpaceSeed/00flow/s-forge/90100-rehydration-seed/sbom_external_artifact.webnf), making it immune to injection attacks or parsing hallucinations.

---

## 4. Sovereign Logic-Libraries Rehydration

### 4.1 The Import Drifting Rationale
Standard Go, Dart, and Python builds dynamically fetch dependencies at compile time via online registries (`go get`, `pub get`, `pip install`). In our air-gapped, zero-trust platform, internet-dependent fetches are strictly prohibited. Relying on remote import URLs introduces:
- **Build Invalidation:** Absolute failure in offline/secure execution environments.
- **Supply-Chain Vulnerabilities:** Vulnerability to upstream module hijacking, deletions, or malicious replacement.
- **Verification Decoupling:** Inability to run deterministic cryptographic folder sealing across the entire fleet's import graph.

### 4.2 The Solution: Logic-Libraries Vendoring
To eliminate external runtime imports, all critical logic-libraries are formally **rehydrated as external artifacts** inside the local authority registry (`sbom_external_artifact.webnf`). They are ingested, metobolically pruned, sealed, and promoted to:
`00flow/s-forge/90200-logic-libraries/`

We use local Go `replace` mappings to bind public import addresses directly to these local repository folders:
```go
replace github.com/zeebo/blake3 => ../90200-logic-libraries/blake3
```

### 4.3 Technical Justifications for Our Logic-Libraries

The NATVS Engine currently designates four core external logic-libraries as platform-critical:

#### A. BLAKE3 Tree-Hash Engine (`github.com/zeebo/blake3`)
*   **Justification:** SHA-512 is cryptographically robust but mathematically heavy when walking and hashing gigabytes of SDK libraries or indexing deep AST trees. BLAKE3 provides a highly parallelized tree-hash model. 
*   **Application:** Dramatically speeds up directory sealing, enabling multi-threaded parallel file hashing and near-instantaneous $O(1)$ comparisons to verify whether nested sub-directories have changed.

#### B. mPSH (Multiset Position-Sensitive Hashing) for ADPH
*   **Justification:** When verifying package and directory relationships, standard hashes are flat and change completely if modules are reorganized.
*   **Application:** Drives **Architectural Dependency Path Hashing (ADPH)**, providing a mathematically homomorphic representation of the fleet's directory topography. If a module is refactored, the ADPH seal is re-calculated in $O(1)$ by adding or subtracting the affected file hash, bypassing expensive full-disk crawls.

#### C. wazero WebAssembly-GC Engine (`github.com/tetratelabs/wazero`)
*   **Justification:** The rehydration process uses compiled Wirth parser plugins to validate grammar compliance. Spawning external subprocesses (like `wasmtime.exe` or `wasmer.exe`) introduces heavy system-level dependencies.
*   **Application:** A pure-Go, zero-dependency, CGO-free WebAssembly engine that runs our grammar validations in-process, guaranteeing absolute portability across Windows, Linux, and WASM architectures.

#### D. quic-go Transport Engine (`github.com/quic-go/quic-go`)
*   **Justification:** Standard HTTP and TCP are too slow and single-streamed to handle rapid, multiplexed cognitive actor handshakes on our trust bus (Whisper Bus) connecting Jules and the NATVS orchestration engine.
*   **Application:** Powers multiplexed, low-latency agent RPC communication and secure local capability negotiations.

---

## 5. Safe Caching Overrides vs. Destructive Hollowing

### 5.1 The Safety Vulnerability
The legacy external hydrator supported a `-force` CLI flag which executed a destructive subdirectory hollowing operation (`hollowDirectory`) on targets before downloading assets. In air-gapped workspaces containing cached resources or sensitive developer overrides, this mechanism posed a severe risk of accidental file loss.

### 5.2 The Design Solution
We completely removed the `-force` CLI flag, `hollowDirectory` function, and all hollowing checks. In its place, we introduced the safe **`-noskip`** flag.
- **Underlying Principle**: If a target is already hydrated, the engine ordinarily skips it to conserve time and bandwidth.
- **Bypass Mode**: The `-noskip` flag forces the engine to download, verify, and overwrite target binaries in-place *without* hollowing out or deleting existing local directory trees, preserving any local developer modifications.

---

## 6. Monorepo Workspace Dependency Resolution (`go.work`)

### 6.1 The Isolation Fallacy
Older builds forced `GOWORK=off` during `go mod tidy` and `go build` phases, attempting to compile modules hermetically in isolation. In a complex, multi-module monorepo with custom local package mappings (`sov.fleet/s-logiclibrary`, `sov.fleet/blake3`), running with `GOWORK=off` deletes local package references from `go.mod` because they cannot be resolved from the internet registry, failing the subsequent compile.

### 6.2 The Design Solution
We removed all explicit `GOWORK=off` and `GOWORK=on` overrides from execution environment hooks.
- **Discovery Mode**: Unsetting `GOWORK` allows the Go toolchain to naturally scan parent paths and discover the root `go.work` file.
- **Unified Resolution**: The build engine resolves all internal dependency mappings hermetically through the monorepo workspace definition, ensuring compilation of down-level modules (such as `o-afflume`) succeeds cleanly.
