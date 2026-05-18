# Sovereign Hydration Engine Architecture

The **sHydrator** engine acts as the primary gatekeeper for the air-gapped VelociKey/00flow platform. It provides a dual-zone rehydration structure to ingest, verify, prune, and promote all external software binaries and internal Bazel build layers. This guarantees absolute offline reliability and zero-internet builds.

```mermaid
graph TD
    subgraph External Hydration (Green Tea Engine)
        Manifests[Authority Manifests .webnf.manifest] -->|Load| Engine[rehydrate.go Engine]
        Engine -->|Download/Ingest| Shadow[Shadow Sandbox Dir]
        Shadow -->|Iterative DFS Metabolic Pruning| Pruner[DFS Metabolic & Deep Pruner]
        Pruner -->|Deterministic Seal| Hash[SignPrunedProduct sha512]
        Hash -->|Promote & Register| SBOM[sbom_external_artifact.webnf]
        SBOM -->|Verify Paths| Paths[FLEET-PATHS.webnf]
    end

    subgraph Internal Hydration (Bazel Orchestration)
        BazelRules[Bazel Rulesets & Rules_Go] -->|Offline Decoupled Promotion| LocalForge[s-forge Repository Silo]
        LocalForge -->|Relative Local Mapping| MODULE[MODULE.bazel / WORKSPACE]
        MODULE -->|Bazel Orchestrator| Compilation[Local AMD64 Hermetic Compilation]
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
