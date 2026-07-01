# Sovereign Hydration Engine Architecture

The **sHydrator** engine acts as the primary gatekeeper for the air-gapped VelociKey/00flow platform. It provides a dual-zone rehydration structure to ingest, verify, prune, and promote all external software binaries and internal Bazel build layers. This guarantees absolute offline reliability and zero-internet builds.

```mermaid
graph TD
    subgraph External Hydration (Green Tea Engine)
        Manifests[Authority Manifests] -->|Load| Engine[hydrator.go Engine]
        Engine -->|DAG Cohort Scheduling| QDAG[s-qdag DAG / LPT Sorter]
        QDAG -->|Parallel Ingestion| Sandbox[Shadow Sandbox]
        Sandbox -->|Iterative DFS Pruning| Seal[Deterministic Seal]
        Seal -->|Register| SBOM[sbom_external_artifact.webnf]
    end

    subgraph Internal Hydration (int-rehydrator)
        Workspaces[go.work / workspaces] -->|Dependency Graph| IntQDAG[s-qdag Cascade Scheduler]
        IntQDAG -->|Parallel Workers| Compile[Hermetic Compiles]
        Compile -->|Targeted Bazel Verification| Bazel[rdeps Test Target Slicer]
    end
```

---

## 1. Dual-Zone Rehydration Topology

### 1.1 Zone A: External Ingestion (The Green Tea Engine)
Zone A handles all binary-level external toolchain dependencies (Go compiler, Java JDK, step-ca, notary, trivy, git, gh-cli, gcloud-sdk, etc.) required to bootstrap and manage the workspace.
- **Authority Ground-Truth:** Controlled strictly by non-recursive `.webnf.manifest` files inside [90000-authority](file:///C:/aCogSpaceSeed/00flow/s-forge/90000-authority).
- **Isolation Sandbox:** Artifacts are pulled and unpacked into a sandboxed `c0990-ephemeral-scratch/shadow` area to prevent active environment contamination.
- **Sovereign Promotion:** Verified and pruned packages are promoted to active genetic zones (`91xxx` for binaries, `92xxx` for toolchains) under [00flow/s-forge](file:///C:/aCogSpaceSeed/00flow/s-forge).

### 1.2 Zone B: Internal Build Orchestration (Bazel Sovereignty)
Zone B manages the internal build layers, rulesets, and WASM-GC backends.
- **Rule Decoupling:** Replaces external HTTP imports (`http_archive`, `git_repository`) with hermetic, offline `local_repository` links pointing directly to [s-forge](file:///C:/aCogSpaceSeed/00flow/s-forge).
- **Workspace Decoupling:** Binds compiler rulesets (e.g., `rules_go`, `rules_flutter`, `rules_foreign_cc`) under [21000-build-orchestration](file:///C:/aCogSpaceSeed/00flow/sHydrator/21000-build-orchestration) to enable zero-network builds on developer desktops and sovereign server nodes.

---

## 2. Directory Layout & Governance

The hydration structures strictly adhere to the semantically based, sequence-prefix directory taxonomy defined in **AGENTS.md**:

| Path | Purpose | Prefix / Directory Rules |
| :--- | :--- | :--- |
| `00flow/s-forge/90000-authority` | Ingestion Manifests | `90xxx` Sequence, Strict Manifests |
| `00flow/s-forge/90100-rehydration-seed` | Registry Program Database | Conforming `sbom_external_artifact.webnf` file |
| `00flow/s-forge/91000-external-binaries` | Promoted Standalone Tools | Standalone Windows executable packages |
| `00flow/s-forge/92000-external-toolchains` | Promoted Runtimes & Rulesets | Go compilers, Headless JDK, Bazel rulesets |
| `00flow/s-latentlingua` | Wirth WAG Grammars | Conforming registry definition (`external_artifact.wag`) |
| `00flow/sHydrator/21000-build-orchestration` | Hermetic Build Harness | Core Bazel offline compiler configurations |

> [!NOTE]
> All paths mapped inside this directory layout are strictly validated against **FLEET-PATHS.webnf** during every promotion phase to eliminate workspace path hijacking or environment configuration drift.

---

## 3. Topological Dependency Scheduling (`s-qdag` Engine)

To scale compilation throughput while guaranteeing strict dependency ordering, both internal and external rehydration flows are orchestrated as Directed Acyclic Graphs (DAGs) using `s-qdag`.

### 3.1 Cross-Language Internal Dependency Parsing
The internal cascade scheduler resolves the union of Go and Dart workspace dependency trees:
1. **Dependency Ingestion**: Parses both `"dependencies"` (Go) and `"dart_packages"` (Dart) blocks from `dependencies.webnf`.
2. **Topological Evaluation**: Computes the exact build ordering for compilation targets, preventing downstream targets (like `o-afflume`) from executing before upstream platforms (like `s-qdag` or `s-natives`).
3. **Targeted Verification Tests**: Employs git delta slicing (`git status`) to trace packages with modified files, querying Bazel's AST using target `rdeps` to run verification checks strictly on affected test trees.

### 3.2 Longest Processing Time First (LPT) Ingestion
External toolchain ingestion leverages LPT scheduling constraints to optimize network and I/O concurrency:
1. **Bootstrap Phase**: Core runtimes (`golang` and the JDK) are isolated at index 0 of the dependency queue to resolve first, unblocking down-level compiler dependencies.
2. **Cohort Parallelization**: Standard downloads are grouped into independent cohorts.
3. **LPT Sorting**: Streams are sorted descending by their historical durations (read from `hydration_experience.webnf`), ensuring the slowest downloads start earliest to minimize pipeline tail latencies.
