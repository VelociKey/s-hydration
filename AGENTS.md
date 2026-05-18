# Agent Registry

This workspace is managed using the **Antigravity Agent Manager** and the **Gemini CLI**.

## The Cognitive Linguistic Foundation of 00flow

Our entire software and agent development platform leads with formal grammar, strict semantic/lexical taxonomies, and Domain Specific Languages (DSLs) to optimize integration with LLMs and SLMs in a language-based world:
- **Deterministic Cognitive Rails:** Restricting system configurations, build harnesses, and promotions to non-recursive, flat-block DSL grammars (`.wag` files) eliminates AI agent hallucination and state-tracking errors. The grammar represents the absolute mathematical boundary of valid intent.
- **Lexical Directory Taxonomy:** Our standardized 5-digit/4-digit directory structure acts as a physical lexical grammar, enabling LLMs to navigate the fleet with $O(1)$ search latency and zero semantic drift.
- **Syntactic Supply-Chain Auditing:** Formal language contracts enable specialized agents to audit system architecture, cryptoseals, and dependencies purely at the syntactic and AST level, guaranteeing air-gapped zero-trust execution.

## Core Architectural Patterns

### 1. Shared Agent Registry
The global `AGENTS.md` file serves as the definitive, shared registry for discovering and documenting agents across all UI tiers (Gemini CLI, Antigravity IDE, and Antigravity Agent Manager). This single source of truth ensures that all platforms maintain synchronized access to agent roles, responsibilities, and invocation patterns.

### 2. Universal SACP Daemon Pattern
All agents in the fleet implement a foundational background listener pattern for the **Sovereign Agent Control Protocol (SACP)**. 
- **Implementation:** Agents implement a `RunDaemon(port)` function that opens a raw UDP/QUIC socket to receive SACP telemetry and commands (e.g., `CALL:<command>;`).
- **Interaction:** This pattern facilitates seamless, zero-copy communication between orchestrators (like `natvs-engine`) and actors (like `jules.exe`), and allows the **Antigravity Agent Manager** to orchestrate the fleet via direct UDP/QUIC frames.

### 3. Universal Build Hydration
The compilation and promotion of internal artifacts (like `natvs-engine.exe`) are governed globally by the `00flow/s-hydration` workspace acting as a **Bazel orchestrator**.
- **Build Harness:** Each workspace contains a `workspace.harness` file defining its build requirements (e.g., `type: NATIVE`).
- **Promotion Registry:** The orchestrator maps these build types to specific output paths in `s-forge` using a central mapping registry (e.g., `00flow/s-hydration/400-registry/sforge_paths_hydrator.webnf`), ensuring that all binaries are promoted to their authoritative sovereign locations.

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

### NATVS Engine (Orchestration Actor)
- **Role:** Fleet-aware execution engine.
- **Responsibility:** Managing the NATVS lifecycle (Negotiation, Assimilation, Transformation, Verification, Synthesis).
- **Function:** Orchestrates multi-silo tasks using the Jules Agent and sovereign toolchains.
- **Interface:** Invoked via Gemini-CLI or Antigravity Agent Manager.
- **Gemini CLI Integration:**
  - **Direct Invocation:** To delegate a goal to the engine, use the `run_shell_command` tool to execute: `C:\aCogSpaceSeed\00flow\s-forge\94000-internal-actors\natvs-engine.exe "<objective>" "<context_path>"`.
  - **Conversational Protocol:** The engine orchestrates a conversation with the Jules agent. Monitor the engine's stdout to track progress and identify when it requires external input or approval.
  - **Sovereign Approval:** When the engine workflow pauses for verification or reaches a critical "Negotiation" or "Synthesis" boundary, use the `ask_user` tool to obtain explicit human approval before proceeding with the next phase of the NATVS lifecycle.

## Platform Structure

### 000all (The Cognitive Silo)
- **s-cognition:** Central repository for architectural plans, blueprints, and agent narratives.

### 00flow (SDLC Platform)
- **s-seed:** Bootstrap workspace for platform initialization and taxonomy management. Contains `init-workspace.ps1`.
- **s-actors:** Dedicated home for active actor source code, cognitive engines, and registrars.
- **s-forge:** Central repository for all external and internal binaries/artifacts.
- **s-hydration:** Specialized workspace for dynamic build synthesis, static asset ingestion, and capability discovery.
- **s-latentlingua:** The "Grammar Studio" for developing DSLs using simplified eBNF/WSN.
### 00aaif (Agent Architecture & Integration Framework)
- Focuses on transforming and modernizing agentic standards.

### 86SREF (Reference Silo)
- Contains clones from the **Velockey** GitHub organization.
- **Mandate:** This silo is strictly **read-only**. Files must never be modified here; they must be migrated to an active workspace for adaptation.

## Technology Stack
- **UI:** Dart, Flutter, Firebase.
- **WASM:** Go (using WASM-GC).
- **Backend:** Go.
- **Build System:** Bazel (Orchestration logic housed in `21000-build-orchestration`).
- **Networking:** QUIC (raw UDP socket using quic-go in Go WASM-GC).
- **Logic:** Domain Specific Languages (DSLs) based on White's evolved Backus-Naur Form (weBNF, inspired by Wirth's styling).
- **Strict Architecture Invariant:** Never create, install, run, or utilize npm, node, react, javascript, typescript, or related web frameworks/packages. The platform is strictly Go and Dart/Flutter; web adapters compile exclusively through Go WASM-GC.


## Workflow Conventions
- **Task Tracking:** Complex tasks should be documented in `/conductor/tasks/`.
- **Taxonomy Initialization:** Use `00flow/s-seed/init-workspace.ps1` to initialize or refine any workspace structure.
- **Sovereign Root:** The project root is defined by the **`.gitroot`** file. This ensures that Gemini-CLI and Antigravity perceive the entire fleet as a single coordinated entity, regardless of nested `.git` repositories in individual workspaces.
- **AAIF Standard:** We strictly follow the AAIF specification. All instructional context is stored in **`AGENTS.md`** files.
 The CLI is configured to bypass `.git` boundaries and stop only at the `.gitroot` to allow global instructions to flow down.
- **Semantic Taxonomy:** All workspaces follow a semantically based directory-tree taxonomy:
    - `nnnnn-<semantic-name>`: "Old school" repository/saved folders (5-digit sequence, strictly lowercase).
    - `cnnnn-<semantic-name>`: Cognitive ephemeral directories (4-digit sequence with 'c' prefix, strictly lowercase), used by agents, Antigravity, and Gemini-CLI.
    - `nnn-<semantic-name>`: Subdirectories within the above (3-digit sequence, strictly lowercase).
- **Context Management:** Use `AGENTS.md` files in subdirectories for module-specific rules.
- **Integration:** All agents should respect the `TERM=xterm-256color` setting for optimal terminal output.

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
- Sovereign Go executable: `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe`
- Sovereign Bazel Orchestration: Managed under `C:\aCogSpaceSeed\00flow\s-hydration\21000-build-orchestration`
- Sovereign Ephemeral Scratch: `C:\aCogSpaceSeed\c0990-ephemeral-scratch`
- Cognitive Layer Silo (000all): `C:\aCogSpaceSeed\000all`
- Sovereign Meta-Grammar (webnf.sn): `C:\aCogSpaceSeed\00flow\s-latentlingua\30100-meta-foundation\webnf.sn` (The authoritative grammar used for all generated DSL grammars, which will have the name form `<dslname>.wag`, and all grammar-conforming programs, which will have the name form `<program_name>_<dsl_name>.webnf`).
- External Artifact Registry Grammar (external_artifact.wag): Authorized DSL grammar for mapping hydrated supply-chain dependencies. The grammar resides at `00flow/s-latentlingua/external_artifact.wag` and governs the conforming program database named `sbom_external_artifact.webnf` under `00flow/s-forge/90100-rehydration-seed/sbom_external_artifact.webnf`.