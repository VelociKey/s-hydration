# Sovereign Hydration Architecture

## 1. Executive Summary
The Sovereign Hydration engine (`sov-hydrator.go`) is a highly concurrent, fully air-gapped dependency ingestion orchestrator for the NATVS-engine pipeline. Designed to be a single, unified, 100% portable Go binary, it compiles flawlessly across Windows 11 Pro, Linux, and future macOS platforms. The engine establishes Bazel sovereignty by retrieving external dependencies, metabolically pruning them against explicit capability lists, verifying cryptographic seals, and securely promoting them to the `s-forge` internal registry.

## 2. Core Architectural Components

### 2.1 Single-File Cross-Platform Portability
- **Native Implementation:** All OS-dependent shell executions (such as PowerShell `Copy-Item`) have been replaced with a native, pure Go non-recursive DFS copy mechanism (`CopyDirectory`).
- **Unified Build:** The system is governed by a single Go source file without complex multi-file build tags or OS-specific C/C++ dependencies. 
- **Storage Metrics:** Uses a simplified, cross-platform agnostic stub for storage metric aggregation to guarantee compilation anywhere without NT-kernel or POSIX `statfs` collisions.

### 2.2 SovereignScaler: Dynamic Concurrency
- **HTTP/3 QUIC Multiplexing:** Leverages QUIC streams over a single UDP socket connection to bypass HTTP/2 TCP head-of-line blocking and maximize throughput against the same host (e.g., github.com).
- **Elastic Throttling:** Concurrency scales dynamically from N=2 up to N=8 based on target payload sizes and historic network throughput metrics recorded in the `Experience Registry`.
- **Topological Scheduling:** Massive artifacts (>100MB) are relegated to a dedicated sequential queue, while smaller bootstrap dependencies are processed in massive parallel bursts to prevent disk thrashing.

### 2.3 The NATVS Lifecycle (Verification & Pruning)
- **Ingestion Sandbox:** Assets are downloaded into an ephemeral scratch space (`c0990-ephemeral-scratch`).
- **Offline Security Scan:** Trivy integration prevents CVEs from entering the `s-forge` perimeter.
- **Metabolic Pruning:** Exhaustive DFS graph traversal targets and securely deletes unnecessary bloat (e.g., `*.md`, test cases, extraneous OS distributions), maintaining critical compliance declarations (`license`, `copying`, `patents`).
- **Attestation Sealing:** Employs Blake3 / SHA-512 cryptographic checks against a hardcoded manifest registry before permitting promotion.

## 3. Experience Registry
The engine maintains a historical database (`hydration_experience.webnf`) that chronicles prior ingestion durations, original payload sizes, expanded volume footprints, and pruned footprints. This persistent "memory" enables the SovereignScaler to intelligently partition work across hosts based on predictive timing rather than static rules.

## 4. A/B Test Validation & Optimization

### 4.1 Real-World Metrics Verification
An integrated live A/B Acid Test campaign (`-abtest`) isolates the performance delta between forced-sequential scheduling (N=1) and dynamic QUIC-multiplexed parallel scheduling. 

**Acid Test Results Example:**
```text
=====================================================================================
                         SOVEREIGN HYDRATOR A/B TEST REPORT
=====================================================================================
  TARGET ARTIFACT                | SEQUENTIAL DURATION  | PARALLEL DURATION    | SPEEDUP   
-------------------------------------------------------------------------------------
  bazel-rules-flutter            | 220ms                | 68ms                 | 3.24x     
  bazel-rules-go                 | 1.94s                | 2.201s               | 0.88x     
  bazelisk                       | 706ms                | 2.039s               | 0.35x     
  buildifier                     | 611ms                | 2.039s               | 0.30x     
  step-ca                        | 1.342s               | 2.521s               | 0.53x     
-------------------------------------------------------------------------------------
  GLOBAL PROCESS TOTALS          | 4.819s               | 2.589s               | 1.86x Faster
=====================================================================================
```

### 4.2 Engineering Advancements & Corrections
- **Queue Wait Time Cancellation:** Extracted target download durations are precision-corrected by capturing queue lock wait times. This ensures parallel targets waiting inside the extraction bottleneck FIFO queue do not artificially inflate their parallel execution durations.
- **Target Bay Sanitation:** `s-forge` promotion paths are wiped cleanly before `CopyDirectory` executes, providing a residual-free release state.
- **Minimal Disk Footprint ($O(1)$ Storage Overhead):** Ephemeral shadow verification paths are eradicated instantaneously inside the lock's critical section, guaranteeing that gigabytes of parallel unzipped payload do not stack simultaneously on workstation solid-state drives.

## 5. Deployment Commands
- **Standard Pipeline:** `go run sov-hydrator.go`
- **A/B Acid Test Campaign:** `go run sov-hydrator.go -abtest -only "bazel-rules-flutter,bazelisk,buildifier"`
- **Strict Sequential Mode:** `go run sov-hydrator.go -sequential`
