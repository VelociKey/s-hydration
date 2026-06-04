# Implementation Analysis: Pluggable Ingestion Spectrum in s-hydration

**Document Class:** System Architecture Design & Integration Plan  
**Classification:** Technical Architecture Specification  
**System Target:** `s-hydration` Platform Workspace & Unified Hydrator Core  
**Applicability:** Unifying External Ingestion (`hydrator.go`) and Internal Rehydration (`int-rehydrator.go`)

---

## 1. Executive Analysis: Current State vs. Target Ingestion Spectrum

Currently, the `s-hydration` workspace maintains a structural division between **external dependency hydration** (managed by `hydrator.go` / `sov-hydrator.exe`) and **internal workspace rehydration** (managed by `int-rehydrator.go` / `int-rehydrator.exe`). 

While both engines implement robust performance features (such as QUIC-multiplexing, dynamic concurrency scaling, and sandboxed directory copies), running two separate codebases introduces compilation drift, inconsistent validation metrics, and redundant file-moving pipelines.

Adopting the **Pluggable Ingestion Spectrum** replaces this division with a unified 4-Stage state machine that maps all assets into four distinct Ingestion Zones:

```
[EXTERNAL SYSTEM BOUNDARY] ────────────────────────────────────────────────────────┐
 │                                                                                 │
 ├── Zone 1: Public Untrusted Pre-builts (Tier 1)                                  │
 │    └── Sourced: GitHub Releases / Registries (e.g., trivy, step-ca)             │
 │                                                                                 │
 ├── Zone 2: Public Untrusted Source Text (Tier 2)                                 │
 │    └── Sourced: Public Open-Source repos (e.g., bazel-rules-flutter)            │
 │                                                                                 │
 ├── Zone 3: Private Sovereign Source Text (Tier 3)                                │
 │    └── Sourced: Internal Go workspace repositories (e.g., s-logiclibrary)       │
 │                                                                                 │
[LOCAL SYSTEM BOUNDARY] ───────────────────────────────────────────────────────────┘
 │                                                                                 │
 └── Zone 4: Internal Sovereign Pre-builts (Tier 4)                                │
      └── Sourced: Locally compiled Go binaries / WASM modules in s-forge          │
───────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. The 4-Stage Unified Pipeline Pattern

To align the codebase with the specification, all ingestion flows must be refactored to execute sequentially through the same Go-native pipeline interface:

```
[ STAGE 1: SELECTION DISCOVERY ] ──► [ STAGE 2: MANIFEST SYNTHESIS ] ──► [ STAGE 3: SHADOW CAPTURE ] ──► [ STAGE 4: TYPE-MAPPED PROMOTION ]
  - Pluggable list discovery           - Attach canonical type/location   - Quarantine in shadow bay        - Route to targeted folder
  - Output uniform index format        - Enforce structural path gates     - Context-aware deep pruning      - Topological hashing & seal
```

### Stage 1: Selection Discovery
* **Current Gap:** `hydrator.go` uses hardcoded SBOM parsing (`ParseSBOM`), while `int-rehydrator.go` uses hardcoded `go.work` and module parsing.
* **Proposed Alignment:** Abstract list retrieval. Introduce a `ListDiscoveryDriver` interface. Implement two drivers:
  1. `ExternalRegistryDriver`: Fetches target lists from the JSON/weBNF manifests.
  2. `InternalWorkspaceDriver`: Traverses local directories via `go.work` to discover local targets.
* **Output Invariant:** Both output a standardized `SelectionList` containing unique identifiers, tags, and source locations.

### Stage 2: Manifest Synthesis
* **Current Gap:** Target mapping is handled implicitly by metadata categories or manual path configuration.
* **Proposed Alignment:** Convert selection records into schema-conforming declarative manifests. Append two explicit tags to every target:
  1. `ArtifactLocation`: An immutable address (e.g., URL or local relative path offset).
  2. `TypeDesignation`: One of `VENDOR_BIN`, `VENDOR_SRC`, `SOV_PRIV_SRC`, or `SOV_PRE_BUILT`.
* **Guardrail:** Prior to execution, validate that all output targets reside strictly within designated prefixes (rejecting `..` parent lookbacks or absolute outside paths).

### Stage 3: Shadow Isolation Buffer (Quarantine)
* **Current Gap:** `hydrator.go` downloads files to `C:\aCogSpaceSeed\c0990-ephemeral-scratch\C0990-verify-download` before copying. `int-rehydrator.go` compiles directly to intermediate local paths.
* **Proposed Alignment:** Unify all staging to a single volatile shadow scratch space (`c0990-ephemeral-scratch/shadow-staging/`).
  * **Ingest:** Download external packages, or compile internal workspaces (via microVM/sandboxed builders) strictly targeting this shadow workspace.
  * **Sanitation (Pruning):** Run the double-pruner suite:
    * **Metabolic Pruner:** Removes testing directories, docs, examples, and markdown clutter while preserving compliance licenses.
    * **Deep Platform Pruner:** Scans and discards binaries, layouts, and libraries (e.g., `.so` or `.dylib`) that do not match the current host operating system (Windows x64).

### Stage 4: Type-Mapped Promotion Matrix
* **Current Gap:** `int-rehydrator.go` maps directories manually, while `hydrator.go` has hardcoded s-forge target folders.
* **Proposed Alignment:** Read the `TypeDesignation` from Stage 2 and route the sanitized directory to its corresponding target registry folder:
  * **Production Mode:** Routes directly to target segments in the global `s-forge` registry (e.g., `s-forge/97000-internal-toolchains/`, `s-forge/93000-external-libraries/`).
  * **Workspace-Local Test Build Phase:** Routes compile/build artifacts directly to matching folder segments inside the **`s-hydration` local directory-tree** (e.g., `s-hydration/97000-internal-toolchains/`, `s-hydration/93000-external-libraries/`). This mirrors the production storage topology locally, allowing Bazel to execute test builds and verify pipeline integration before committing output files globally.
* **Cryptographic Sealing:** Run a lexicographically sorted topographical directory scan. Blend file paths and sizes into a SHA-512 `Deterministic Directory Seal`. Log this signature into the unified SBOM database (`sbom.external_artifact.webnf`).

---

## 3. Recommended Improvements & Implementation Plan

We propose structuring the unification of `s-hydration` into the following key implementation steps:

### Phase 1: Unifying Core Engine Runtimes
* **Action:** Merge `hydrator.go` and `int-rehydrator.go` into a unified `rehydrator_core.go`.
* **Details:** Re-use the high-performance components (SovereignScaler, HTTP3/QUIC multiplexer, and DFS directory copy engines) as a shared utility library. Replace separate main loops with a single CLI entry point that routes compilation based on manifest properties.
* **Value:** Eliminates compilation drift and duplicates. Reduces the codebase by ~500 lines of boilerplate.

### Phase 2: Implementing the Pluggable Ingestion Drivers
* **Action:** Implement the `IngestionDriver` interface for Stage 1.
* **Details:** Define a stateless conduit API to pull bytes into the Shadow Buffer. 
  * `NetworkTransportConduit`: Pulls bytes from H2/H3 URL streams.
  * `LocalBuildConduit`: Executes local compilers and pipes output directly to the shadow directory.
* **Value:** Clean separation of concerns. Makes the engine highly modular and extensible for new transport protocols.

### Phase 3: Implementing Platform-Specific Deep Pruning
* **Action:** Implement the `DeepPlatformPruner` inside the shadow isolation stage.
* **Details:** Write a non-recursive traversal function that checks binary extensions (`.dll`, `.so`, `.dylib`, `.exe`) and prunes foreign operating system configurations.
* **Value:** Bypasses filesystem clutter. Decreases workspace storage footprint by up to 40% on localized developer machines.

### Phase 4: Standardizing Topographical Directory Sealing
* **Action:** Implement directory tree hashing in Stage 4.
* **Details:** Recursively traverse files, sort names alphabetically to ensure cross-platform reproducibility, hash file bytes, and generate a final composite SHA-512 seal. Commit this seal to `sbom.external_artifact.webnf`.
* **Value:** Guarantees absolute, environment-agnostic rehydration integrity. Enables secure, lightning-fast $O(1)$ local caching across different workstations.

### Phase 5: Workspace-Local Test Build & A/B Benchmarking Phase
* **Action:** Implement local path re-routing and execution timer switches for A/B testing.
* **Details:**
  * Add a `-test-build` command-line switch to `rehydrator_core.go`. When active, it bypasses the `s-forge` destination mapping in Stage 4, routing all target promotions directly to their matching folder counterparts in the `s-hydration` workspace directory (e.g., `s-hydration/97000-internal-toolchains/`).
  * Integrate an A/B benchmark mode (leveraging the existing `-abtest` logic from `hydrator.go`) that measures the execution times of sequential compilation runs vs. parallel Bazel compilation loops directly within these local segments.
* **Value:** Allows full verification of build pipelines and dependency topologies without polluting the shared production registry. Provides immediate performance comparison metrics for cache hits and thread-scaling optimizations.

---

## 4. Value and Impact Matrix

| Refactoring Task | Engineering Value | Risk Level | Metric Offset |
| :--- | :--- | :--- | :--- |
| **Unified Core Engine** | High (Eliminates redundant binaries & duplicate compilation bugs) | Medium | -500 lines boilerplate |
| **Pluggable Drivers** | High (Stateless code abstraction, enables easy MCP/QUIC scaling) | Low | $O(1)$ transport abstraction |
| **Deep Platform Pruner** | Medium (Keeps local environments lean and fast to parse) | Low | Up to 40% local storage savings |
| **Topographical Sealing** | Critical (Ensures absolute reproducibility across workstations) | Medium | $O(1)$ caching validation speed |
| **Volatile Shadow Bay** | High (Guarantees zero residual junk accumulates on developer SSDs) | Low | $O(1)$ storage footprint overhead |
| **Workspace-Local Test & A/B** | High (Safe build sandbox, measures parallel throughput deltas) | Low | Platform-safe sandboxed validation |

---

## 5. Next Steps

1. Create a workspace-specific roadmap task file (`tasks/hydration_unification.md`).
2. Implement the `IngestionDriver` interface and consolidate `hydrator.go` and `int-rehydrator.go` under the unified state machine.
3. Add the `-test-build` switch to route outputs to matching local segments inside the `s-hydration` workspace.
4. Integrate the A/B testing timing metrics for the local test build runs.
5. Migrate existing SBOM mappings into the normalized manifest format.

