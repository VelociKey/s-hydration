# Agent Registry

This workspace is managed using the **Antigravity Agent Manager** and the **Gemini CLI**.

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

## Platform Structure

### 000ALL (The Cognitive Silo)
- **sCognition:** Central repository for architectural plans, blueprints, and agent narratives.

### 00FLOW (SDLC Platform)
- **sSeed:** Bootstrap workspace for platform initialization and taxonomy management. Contains `init-workspace.ps1`.
- **sForge:** Central repository for all external and internal binaries/artifacts.
- **sBuilder:** Specialized workspace for the build orchestration and cache management.
- **sLatentLingua:** The "Grammar Studio" for developing DSLs using simplified eBNF/WSN.

### 00AAIF (Agent Architecture & Integration Framework)
- Focuses on transforming and modernizing agentic standards.

### 86SREF (Reference Silo)
- Contains clones from the **Velockey** GitHub organization.
- **Mandate:** This silo is strictly **read-only**. Files must never be modified here; they must be migrated to an active workspace for adaptation.

## Technology Stack
- **UI:** Dart, Flutter, Firebase.
- **WASM:** Go (using WASM-GC).
- **Backend:** Go.
- **Build System:** Bazel (Orchestration logic housed in `21000-build-orchestration`).
- **Networking:** QUIC / WebTransport.
- **Logic:** Domain Specific Languages (DSLs) based on Wirth Syntax Notation.

## Workflow Conventions
- **Task Tracking:** Complex tasks should be documented in `/conductor/tasks/`.
- **Taxonomy Initialization:** Use `00FLOW/sSeed/init-workspace.ps1` to initialize or refine any workspace structure.
- **Sovereign Root:** The project root is defined by the **`.gitroot`** file. This ensures that Gemini-CLI and Antigravity perceive the entire fleet as a single coordinated entity, regardless of nested `.git` repositories in individual workspaces.
- **AAIF Standard:** We strictly follow the AAIF specification. All instructional context is stored in **`AGENTS.md`** files. The CLI is configured to bypass `.git` boundaries and stop only at the `.gitroot` to allow global instructions to flow down.
- **Semantic Taxonomy:** All workspaces follow a semantically based directory-tree taxonomy:
    - `nnnnn-<semantic-name>`: "Old school" repository/saved folders (5-digit sequence).
    - `Cnnnn-<semantic-name>`: Cognitive ephemeral directories (4-digit sequence with 'C' prefix), used by agents, Antigravity, and Gemini-CLI.
    - `nnn-<semantic-name>`: Subdirectories within the above (3-digit sequence).
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
   - `GEMINI_CLI_HOME=C:\aCogSpaceSeed\C0990-ephemeral-scratch`
3. Run `/ide install` in the Gemini CLI to link with Antigravity.
4. Use the **Agent Manager** panel in Antigravity to interact with these roles.