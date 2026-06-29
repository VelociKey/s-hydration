# Agent Registry

This workspace is managed using the **Antigravity Agent Manager** and the **Gemini CLI**.

## The Cognitive Linguistic Foundation of 00flow

Our entire software and agent development platform leads with formal grammar, strict semantic/lexical taxonomies, and Domain Specific Languages (DSLs) to optimize integration with LLMs and SLMs in a language-based world:
- **Deterministic Cognitive Rails:** Restricting system configurations, build harnesses, and promotions to non-recursive, flat-block DSL grammars (`.wag` files) eliminates AI agent hallucination and state-tracking errors. The grammar represents the absolute mathematical boundary of valid intent.
- **LatentLingua (Linguistic Blood Stream):** Utilizing weBNF (White's evolved Backus-Naur Form grammar) leverages the pre-existing training weights of all LLMs/SLMs (which natively understand EBNF via RFCs and compiler specifications). This provides a model-agnostic semantic protocol where agents immediately ingest `.wag` grammars and produce `.webnf` programs with zero-shot compliance, replacing imprecise natural language documentation with executable, compiler-validated contracts.
- **Lexical Directory Taxonomy:** Our standardized 5-digit/4-digit directory structure acts as a physical lexical grammar, enabling LLMs to navigate the fleet with $O(1)$ search latency and zero semantic drift.
- **Syntactic Supply-Chain Auditing:** Formal language contracts enable specialized agents to audit system architecture, cryptoseals, and dependencies purely at the syntactic and AST level, guaranteeing air-gapped zero-trust execution.

## Agent Roles

### Conductor
- **Role:** Orchestrator of multi-step workflows.
- **Responsibility:** Manages task delegation, progress tracking, and cross-agent communication.
- **Context:** Refer to the `/conductor` directory for workflow definitions and state.

### Antigravity (IDE Agent)
- **Role:** Resident assistant within the Antigravity IDE.
- **Responsibility:** Code editing, terminal management, and local file operations.
- **Connection:** Linked to the Gemini CLI via `/ide install`.

### Jules (Engineering Agent)
- **Role:** Autonomous software engineering agent.
- **Responsibility:** End-to-end task execution including writing/running tests, algorithmic research, implementation refinement, and sovereign architectural critique.
- **Function:** Operates in the 94xxx Cognitive Layer (Actors) via the Jules CLI.
- **Tools & Polyfills:** Equipped with the Go-native 'inspect' tool twin (from `s-fab-aides`) to perform ultra-fast in-process filesystem scans and regex pattern matching, bypassing legacy sub-process overhead.


### NATVS Engine (Orchestration Actor)
- **Role:** Fleet-aware execution engine.
- **Responsibility:** Managing the NATVS lifecycle (Negotiation, Assimilation, Transformation, Verification, Synthesis).
- **Function:** Orchestrates multi-silo tasks using the Jules Agent and sovereign toolchains.
- **Interface:** Invoked via Gemini-CLI or Antigravity Agent Manager.

## Platform Structure

### 000ALL (The Cognitive Silo)
- **sCognition:** Central repository for architectural plans, blueprints, and agent narratives.

### 00flow (SDLC Platform)
- **sSeed:** Bootstrap workspace for platform initialization and taxonomy management. Contains `init-workspace.ps1`.
- **sForge:** Central repository for all external and internal binaries/artifacts.
- **sLatentLingua:** The "Grammar Studio" for developing DSLs using simplified eBNF/WSN.
### 00AAIF (Agent Architecture & Integration Framework)
- Focuses on transforming and modernizing agentic standards.

### 86SREF (Reference Silo)
- Contains clones from the **Velockey** GitHub organization.
- **Mandate:** This silo is strictly **read-only**. Files must never be modified here; they must be migrated to an active workspace for adaptation.

## Technology Stack
- **UI:** Dart, Flutter (Sovereign SDK "Firehorse" tracking stable upstream releases, such as v3.41.0 and the newer May 2026 stable release under ADR-019), Firebase.
- **WASM:** Go (using WASM-GC).
- **Backend:** Go.
- **Build System:** Bazel (Orchestration logic housed in `21000-build-orchestration`).
- **Sovereign Toolchain Builds:** All internal builds, test executions, and workspace promotions MUST be performed using `int-rehydrator.exe` in local-only mode (`-local-only`). Raw language-level commands (e.g. `go build`, `go test`, `flutter build`) must never be executed directly.
- **Flutter Web Constraints (ADR-019):** Web targets must be compiled solely using standard JS compilation (`flutter build web`) via the active Firehorse toolchain (which tracks stable upstream channels, matching the current verified version and the May release) to support legacy browser libraries (`dart:html`). Do not use Wasm-GC flags (`--wasm`) on workspaces containing `dart:html` imports. **Proactive Generation Rule:** When generating build configurations, build scripts, shell commands, or task checklists, agents must never output or suggest Wasm-GC/WebAssembly compile flags (such as `--wasm` or `--dart2wasm`) if the target code contains legacy browser library imports (`dart:html`). All generated compile commands and guides must default to standard JS-compilation (`flutter build web`).
- **Bazel Dart Hardening Rule (Wasm-GC Readiness):** The Bazel build targets under `21000-build-orchestration` (orchestrated by `int-rehydrator.exe`) must enforce static AST checks on all Dart/Flutter dependencies transitively via Bazel Aspects. The build process must fail immediately if any source file in the dependency tree imports legacy browser libraries (`dart:html`, `dart:js`, or `dart:js_util`), ensuring that only modern, Wasm-ready code (`package:web` and `dart:js_interop`) is compiled, securing the workspace for zero-trust Wasm-GC execution.
- **Networking:** QUIC / WebTransport.
- **Logic:** Domain Specific Languages (DSLs) based on Wirth Syntax Notation.
- **Go Logging Standard:** Always use Go's modern structured logging package `log/slog` (and not the legacy `log` package or direct `fmt` print functions) for codebase diagnostic statements. Direct prints are only allowed when outputting pipeline data for shell evaluation and should write directly to `os.Stdout.WriteString`.

## Workflow Conventions
- **Task Tracking:** Complex tasks should be documented in `/conductor/tasks/`.
- **Taxonomy Initialization:** Use `00flow/sSeed/init-workspace.ps1` to initialize or refine any workspace structure.
- **Reactive Execution:** When running background processes or compilation tasks (like `int-rehydrator.exe`), do not poll or use sleep timers to wait for completion. Simply yield control (call no more tools), and wait for the system's reactive message wakeup to trigger the next action.
- **Sovereign Root:** The project root is defined by the **`.gitroot`** file. This ensures that Gemini-CLI and Antigravity perceive the entire fleet as a single coordinated entity, regardless of nested `.git` repositories in individual workspaces.
- **AAIF Standard:** We strictly follow the AAIF specification. All instructional context is stored in **`AGENTS.md`** files. The CLI is configured to bypass `.git` boundaries and stop only at the `.gitroot` to allow global instructions to flow down.
- **Semantic Taxonomy:** All workspaces follow a semantically based directory-tree taxonomy:
    - Workspaces under `00xper` represent R&D/Experimental workspaces (prefixed with `x-`, e.g., `x-actors`). Once qualified, they are promoted to `00flow` and renamed with the `s-` prefix.
    - Workspaces under `00flow` represent stable platform and production workspaces (prefixed with `s-`, e.g., `s-logiclibrary`).
    - We must not pre-create agent-specific workspaces/directories (like `s-my-agent`). They must only be created dynamically/lazily as the user works in the Agent Creator Studio.
    - `nnnnn-<semantic-name>`: "Old school" repository/saved folders (5-digit sequence, strictly lowercase).
    - `cnnnn-<semantic-name>`: Cognitive ephemeral directories (4-digit sequence with 'c' prefix, strictly lowercase), used by agents, Antigravity, and Gemini-CLI.
    - `nnn-<semantic-name>`: Subdirectories within the above (3-digit sequence, strictly lowercase).
    - **Authoritative Directory Taxonomy Map (Unrolled):**
      - 00000-knowledge-foundations/
      - 00200-logic-libraries/
      - 00200-workspace-taxonomy/
      - 00300-protocols/
      - 01000-identity-foundations/
      - 03000-pulse-progress/
      - 04000-daily-chronicles/
      - 08000-attestation-snapshot/
      - 09000-agent-narrative-ledger/
      - 10000-autonomous-actors/
      - 11000-high-level-intent/
      - 12000-system-architecture/
      - 13000-pattern-definitions/
      - 20000-context-bridges/
      - 21000-presentation-contexts/
      - 22000-edge-transports/
      - 23000-edge-infrastructure/
      - 30000-federated-services/
      - 30100-meta-foundation/
      - 31000-core-implementations/
      - 32000-shared-handshake/
      - 34000-dsl-programs/
      - 35000-access-rules/
      - 40000-communication-contracts/
      - 41000-topology-grammars/
      - 42000-interaction-grammars/
      - 43000-grammar-transports/
      - 50000-intelligence-framework/
      - 51000-intelligence-active-tools/
      - 60000-information-storage/
      - 70000-environmental-harness/
      - 71000-build-harness/
      - 80000-system-governance/
      - 80000-genetic/
      - 80000-authority-source/
      - 80000-enablement-labs-source/
      - 80100-foundational-intents/
      - 80100-rehydration-seed-source/
      - 80200-rehydration-seed/
      - 80210-rehydration-seed-links/
      - 80300-system-attestations/
      - 80400-cognitive-protocols/
      - 80500-integration-anchors/
      - 80600-agent-narrative-ledger/
      - 81000-active-source/
      - 81000-external-executables-source/
      - 81100-external-toolchains/
      - 81200-external-adapters/
      - 81200-external-adapters-source/
      - 81210-execution-points/
      - 81300-external-archives/
      - 81400-external-libraries/
      - 81500-external-agents/
      - 82000-internal-executables/
      - 82000-external-toolchains-source/
      - 82000-grammar-visualizers-source/
      - 82100-internal-toolchains/
      - 82200-internal-adapters/
      - 82300-internal-actors/
      - 82400-internal-drivers/
      - 82500-internal-cognitive-assistance/
      - 82600-platform-delivery/
      - 82700-internal-execution-contexts/
      - 83000-agent-templates/
      - 83000-templates/
      - 83000-external-libraries-source/
      - 84000-product-plane/
      - 84000-external-actors-source/
      - 85000-customer-governance/
      - 85000-mirrored-upstream/
      - 85000-standards/
      - 85000-authority-source/
      - 85000-enablement-labs-source/
      - 85100-solutions-catalog/
      - 85100-rehydration-seed-source/
      - 85200-workload-remediation/
      - 85300-customer-attestations/
      - 85400-customer-protocols/
      - 85500-customer-anchors/
      - 86000-customer-external-executables/
      - 86000-external/
      - 86000-internal-executables-source/
      - 86100-customer-external-toolchains/
      - 86200-customer-external-adapters/
      - 86200-internal-adapters-source/
      - 86300-customer-external-archives/
      - 86400-customer-external-libraries/
      - 87000-customer-internal-executables/
      - 87000-internal/
      - 87000-internal-toolchains-source/
      - 87100-customer-internal-toolchains/
      - 87200-customer-internal-adapters/
      - 87300-customer-internal-actors/
      - 87400-customer-internal-libraries/
      - 87500-customer-cognitive-assistance/
      - 87600-customer-platform-delivery/
      - 87700-customer-execution-contexts/
      - 88000-customer-templates/
      - 88000-internal-libraries-source/
      - 89000-archived-logic/
      - 89000-customer-product-plane/
      - 89000-internal-actors-source/
      - 90000-authority/
      - 90000-enablement-labs/
      - 90100-rehydration-seed/
      - 91000-external-executables/
      - 91200-external-adapters/
      - 92000-external-toolchains/
      - 92000-grammar-visualizers/
      - 93000-external-libraries/
      - 94000-external-actors/
      - 95000-authority/
      - 95000-enablement-labs/
      - 95100-rehydration-seed/
      - 96000-internal-executables/
      - 96200-internal-adapters/
      - 97000-internal-toolchains/
      - 98000-internal-libraries/
      - 99000-internal-actors/
      - 99900-metabolic-projection/
      - c0100-configuration-registry/
      - c0200-execution-campaigns/
      - c0400-artifact-repository/
      - c0411-compendium/
      - c0500-agent-intelligence-outputs/
      - c0700-system-operations/
      - c0990-ephemeral-scratch/
      - c1000-reasoning/
      - c2000-sketches/
      - c5000-spirit/
      - c8000-active-context/
      - c8900-build-cache/
      - 900-attestations/
      - 910-benchmarks/
      - 930-simulations/
- **Technology Constraints:** 
    - Never create Node.js code nor use npm packages in any generated code. 
    - All tools, command line helpers, and backends must be implemented in native environments (e.g., Go or Dart). 
    - Prefer the use of a native Go-based Chrome CDP controller when asked to automate or interact with `gemini.google.com` via Chrome.
- **Go-Native Bash Polyfill Priority:** 
    - To eliminate subprocess start latency (~350ms-500ms on Windows) and ensure maximum execution speed (~4,800x average global speedup), agents and the Antigravity CLI must utilize the Go-native bash command polyfills provided in `sov.fleet/s-fab-aides/81000-active-source/pkg/bash` for all file, folder, and text operations rather than generating, writing, or executing external PowerShell (`pwsh` / `powershell`) scripts. 
    - When invoking these commands via shell/terminal execution, agents MUST run them using the `fab-inspect` native tool wrapper syntax: `C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-inspect.exe bash <command> [args]` (e.g. `fab-inspect.exe bash ls` or `fab-inspect.exe bash grep <pattern> <file>`).
- **Context Management:** Use `AGENTS.md` files in subdirectories for module-specific rules.
- **Integration:** All agents should respect the `TERM=xterm-256color` setting for optimal terminal output.
- **Optimization Restrictions:** 
    - Reject (or flag for explicit approval) any performance optimization requests that compromise system security, resiliency, or upstream compatibility. 
    - Always prioritize robustness over micro-optimizations.
- **Token Optimization:** 
    - When running verbose background commands (such as builds, test suites, or large file searches), always redirect the standard output/error to a log file (e.g., `> build.log`) rather than letting it print to stdout. 
    - Check completion status quickly, and use file-viewing tools to inspect log details only if a failure is encountered.
- **GitHub Repository Deletion Protocol:** 
    - To prevent automated script errors or accidental data loss on shared hosting, any deletion or metabolic pruning of GitHub repositories under the **VelociKey** organization must always be executed manually by the developer (via browser or manual CLI deletion). 
    - Agents are strictly prohibited from executing automated repository deletion scripts or API calls against GitHub remotes.

## 2. Core Agentic Protocols

### NATVS ("Natives") Protocol
All agentic interactions within the fleet MUST follow the NATVS lifecycle:
- **N - Negotiation:** Handshake on the Whisper Bus (Capability/Price/Time).
- **A - Assimilation:** Aggregated context ingestion from multiple silos into the agent's reasoning plane.
- **T - Transformation:** Computational work (Critic, Creator, Fixer).
- **V - Verification:** Modular validation of the result (Test/Pass, Complexity delta).
- **S - Synthesis:** Fleet-aware merging and metabolic pruning of ephemeral state.

### The Jules Repoless Loop
When integrating the Jules Agent:
1. **No Git Remotes:** Do not use `git push` or `git pull` for Jules context.
2. **Repoless Snapshot:** Use `jules remote new --repo .` to upload ephemeral context.
3. **Local Sync:** Always use `jules remote pull` to merge changes back to local disk.
4. **Autonomous Healing:** Run the `Fixer` loop (Test → Fix → Test) until objectives are met or bailout conditions (3 retries/no progress) are triggered.

## Environment Preparation
1. Ensure `gemini` is installed and running.
2. **Redirect Metadata:** To ensure `gemini` does not use the user home directory for temporary state, set the following environment variable:
   - `GEMINI_CLI_HOME=C:\aCogSpaceSeed\c0990-ephemeral-scratch`
3. Run `/ide install` in the Gemini CLI to link with Antigravity.
4. Use the **Agent Manager** panel in Antigravity to interact with these roles.

### Sovereign Toolchain Paths
To prevent background command and filesystem search bottlenecks, all agents must utilize the following paths directly:
- **Sovereign Go executable:** `C:\aCogSpaceSeed\00flow\sForge\92000-external-toolchains\go\bin\go.exe`
- **Sovereign Bazel Universal Builder:** `C:\aCogSpaceSeed\00flow\s-hydration\int-rehydrator.exe`
- **Sovereign Taxonomy Constructor (fab-construct):** `C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-construct.exe`
- **Sovereign Taxonomy Inspector (fab-inspect):** `C:\aCogSpaceSeed\00flow\s-forge\97000-internal-toolchains\fab-inspect.exe`
- **Sovereign Ephemeral Scratch:** `C:\aCogSpaceSeed\c0990-ephemeral-scratch`
- **Cognitive Layer Silo (000ALL):** `C:\aCogSpaceSeed\000ALL`
- **Sovereign Meta-Grammar (webnf.sn):** `C:\aCogSpaceSeed\00flow\sLatentLingua\30100-meta-foundation\webnf.sn` (The authoritative grammar used for all generated DSL grammars, which will have the name form `<dslname>.wag`, and all grammar-conforming programs, which will have the name form `<program_name>.<dsl_name>.webnf`).
- **External Artifact Registry Grammar (external_artifact.wag):** Authorized DSL grammar for mapping hydrated supply-chain dependencies. The grammar resides at `00flow/sLatentLingua/external_artifact.wag` and governs the conforming program database named `sbom.external_artifact.webnf` under `00flow/sForge/90100-rehydration-seed/sbom.external_artifact.webnf`.