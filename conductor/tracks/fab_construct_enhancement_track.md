# Actionable Hand-off Track: Sovereign Workspace Construction & Semantic Stage/Publish Engine

**Track ID:** `00flow-TRACK-FAB-CONSTRUCT-ENHANCEMENT`  
**Execution Context:** Sovereign Workspace Silos (`00flow/s-fab-aides`)  
**Target:** NATVS Engine / Gemini-CLI / Jules Agent

This track provides the exact, actionable steps for the NATVS Engine to enhance the `fab-construct` helper tool with modular workspace/GitHub twin construction and semantic, Trivy-audited git actions (`add` and `publish`).

---

## 📋 Sovereign Execution Context & Constants

The executing agent must utilize these locked system coordinates directly:
*   **Sovereign Go executable:** `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe`
*   **Target Tool Code Directory:** `C:\aCogSpaceSeed\00flow\s-fab-aides\81000-active-source`
*   **Trivy Binary (if available in path):** `trivy` (assumed pre-hydrated or installed globally in execution context)
*   **Sovereign Root Tracker:** `C:\aCogSpaceSeed\.gitroot`
*   **Unified Workspace Ledger:** `C:\aCogSpaceSeed\go.work`

---

## 🚀 Step-by-Step Production Plan

### Step 1: Implement Modular Subcommands in `fab-construct`
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-fab-aides\81000-active-source\cmd\fab-construct\main.go`
*   **Actionable Task:**  
    Expand `main.go` to cleanly decouple local and remote operations:
    1.  **`fab-construct workspace -name <name> [-silo <silo>]`**: Instantiates only local directory paths via `init-workspace.ps1` and updates `go.work`.
    2.  **`fab-construct github -name <name> [-org <org>]`**: Verifies name uniqueness under the GitHub organization and creates the private repository.
    3.  **`fab-construct both -name <name> [-silo <silo> -org <org>]`**: Performs the full unified local-and-remote bootstrap.
*   **Verification:** Verify compiling `fab-construct.exe` succeeds and running `fab-construct -h` outputs the correct help options.

---

### Step 2: Implement Semantic Staging (`add` command)
*   **Target Path:** `C:\aCogSpaceSeed\00flow\s-fab-aides\81000-active-source\pkg/git` and `cmd/fab-construct`
*   **Actionable Task:**  
    Develop the **`add`** subcommand for semantic staging:
    1.  **Viral License Scan:** Execute Trivy to scan codebase dependencies (e.g., `trivy fs --scanners license --license-full --severity HIGH,CRITICAL --format json .`).
        *   Analyze JSON output. If any unapproved copyleft license (like GPLv3, AGPL) is detected, halt execution and print the non-compliant dependency name.
    2.  **SBOM Generation:** Create a structured Software Bill of Materials (SBOM) metadata block detailing dependency packages and their license mappings.
    3.  **Local Staging & Automated Remarking:** Stage all modified files, analyze differences, compile a structured commit description template (`[type]: brief description`), verify user message formatting, and execute the local `git commit`.
*   **Verification:** Attempt to stage a file in a workspace with a mock GPLv3 package. Verify the tool blocks the commit with a license violation notice.

---

### Step 3: Implement Semantic Publishing (`publish` command)
*   **Target Path:** `C:\aCogSpaceSeed\00flow\s-fab-aides\81000-active-source\pkg/git` and `cmd/fab-construct`
*   **Actionable Task:**  
    Develop the **`publish`** subcommand for secure remote promotion:
    1.  **Vulnerability & Secret Scanning:** Run Trivy to scan for CVEs and hardcoded secrets (tokens, keys) across the project directory:
        ```powershell
        trivy fs --scanners vuln,secret --severity HIGH,CRITICAL --format json .
        ```
        *   Inspect results. If vulnerabilities or secrets are detected, block the push and report the findings.
    2.  **Software Maturity Assessment:** Audit compliance markers:
        *   Verify the existence of the module-specific `AGENTS.md` instructions.
        *   Scan for basic Go unit test suites (`*_test.go`).
        *   Compute a compliance score (fail if critical documentation is missing).
    3.  **Push Gate:** If and only if all scans return zero High/Critical violations and the maturity score passes the threshold, execute `git push origin zenith`.
*   **Verification:** Run `fab-construct publish` on a folder containing a dummy AWS API secret key. Verify the tool blocks the push and reports the secret leak.
