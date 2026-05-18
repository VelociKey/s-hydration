# Actionable Hand-off Track: Sovereign Aether Workspace Migration
**Track ID:** `00flow-TRACK-AETHER-MIGRATION`  
**Execution Context:** Sovereign Workspace Silos (`00flow/s-hermes`, `00flow/s-aether`)  
**OS/Shell Target:** Windows PowerShell / Gemini CLI / NATVS Engine  

---

## 📋 Global Execution Context & Variables

The NATVS Engine utilizes these sovereign coordinates directly during execution:
*   **Sovereign Go compiler:** `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe`
*   **Source Workspace:** `C:\aCogSpaceSeed\00flow\s-hermes`
*   **Target Workspace:** `C:\aCogSpaceSeed\00flow\s-aether`
*   **Offline Registry Catalog:** `C:\aCogSpaceSeed\00flow\s-forge\90100-rehydration-seed\sbom_external_artifact.webnf`

---

## 🚀 Step-by-Step Actionable Execution Track

### Step 1: Initialize the Sovereign `s-aether` Taxonomy
*   **Target Folder:** `C:\aCogSpaceSeed\00flow\s-aether`
*   **Actionable Task Description:**  
    Setup the standard semantic directory taxonomy inside the new workspace using the platform bootstrap taxonomy matrix:
    1.  Create core subdirectories: `03000-pulse-progress`, `04000-daily-chronicles`, `08000-attestation-snapshot`, `10000-autonomous-actors`, `c0990-ephemeral-scratch`, and `c1000-reasoning`.
    2.  Write the base `AGENTS.md` and add `.gitkeep` to all directories.
*   **Execution Commands:**
    ```powershell
    # Execute workspace initialization natively via Conductor bootstrap rules
    & "PowerShell.exe" -ExecutionPolicy Bypass -File "C:\aCogSpaceSeed\00flow\s-seed\init-workspace.ps1" -WorkspacePath "C:\aCogSpaceSeed\00flow\s-aether"
    ```

---

### Step 2: Migrate Core Source, Test, and Harness Assets
*   **Target Location:** `C:\aCogSpaceSeed\00flow\s-aether`
*   **Actionable Task Description:**  
    Move the re-animated, proprietary Aether code assets from the legacy s-hermes workspace to s-aether:
    1.  Copy all active Go source files: `hermes_core.go`, `aether_evolver.go`, `hermes_conformance_test.go`.
    2.  Copy modular definitions: `go.mod`, `go.sum`, `workspace.harness`.
    3.  Create the clean-room legal attribution file `LICENSE-HERMES.txt` in the root of the workspace.
*   **Execution Commands:**
    ```powershell
    # 1. Copy Go files, harnesses, and mod configs
    Copy-Item "C:\aCogSpaceSeed\00flow\s-hermes\*.go" "C:\aCogSpaceSeed\00flow\s-aether\" -Force
    Copy-Item "C:\aCogSpaceSeed\00flow\s-hermes\go.*" "C:\aCogSpaceSeed\00flow\s-aether\" -Force
    Copy-Item "C:\aCogSpaceSeed\00flow\s-hermes\workspace.harness" "C:\aCogSpaceSeed\00flow\s-aether\" -Force

    # 2. Configure legal attribution
    Set-Content -Path "C:\aCogSpaceSeed\00flow\s-aether\LICENSE-HERMES.txt" -Value @"
    ========================================================================
    AETHER SOVEREIGN ENGINE - PROPRIETARY CLEAN-ROOM ATTRIBUTION & LICENSE
    ========================================================================

    This workspace houses Project Aether, a proprietary clean-room derivative
    implementing SACP zero-copy memory topography and native Go genetic
    self-evolutionary algorithms.

    Portions of turn-protection algorithms are derived from Nous Research's 
    original 'Hermes' implementation under standard MIT copyright terms.
    All proprietary additions (including in-memory prompt/skill mutation,
    raw QUIC UDP multi-stream telemetry dialers, and SACP symmetric proxy
    layouts) remain the intellectual property of the 00flow SDLC Platform.
    "@
    ```

---

### Step 3: Conformance & Verification Checks
*   **Target Location:** `C:\aCogSpaceSeed\00flow\s-aether`
*   **Actionable Task Description:**  
    Execute the entire conformance test suite natively inside the new workspace using the offline Go toolchain to assert 100% successful compilation and execution of genetic loops:
    1.  Resolve local package dependencies inside `go.work`.
    2.  Run the tests and assert that all 7 active conformance checks pass in under 150ms.
*   **Execution Commands:**
    ```powershell
    # 1. Update go.work to register the new s-aether workspace
    # 2. Run the tests in s-aether
    & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" test -v "sov.fleet/s-aether"
    ```

---

### Step 4: Re-Compile WASM Binary and Perform Metabolic Pruning
*   **Target Location:** `C:\aCogSpaceSeed\00flow\s-forge\bin`
*   **Actionable Task Description:**  
    Conclude the NATVS migration cycle by compiling the new production binary and pruning the old duplicate structures in `s-hermes`:
    1.  Compile the Go package inside `s-aether` to target `GOOS=wasip1 GOARCH=wasm`.
    2.  Assert that output size is exactly **962 KB** and calculate its BLAKE3 hash.
    3.  Metabolically prune the redundant Go, harness, and test files inside `s-hermes` to ensure zero code duplication.
*   **Execution Commands:**
    ```powershell
    # 1. Re-compile WASM-GC from s-aether
    cmd /c "set GOOS=wasip1&&set GOARCH=wasm&&C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe build -o C:\aCogSpaceSeed\00flow\s-forge\bin\hermes_wasm_gc.wasm ./00flow/s-aether"

    # 2. Perform metabolic pruning of s-hermes
    Remove-Item "C:\aCogSpaceSeed\00flow\s-hermes\*.go" -Force
    Remove-Item "C:\aCogSpaceSeed\00flow\s-hermes\go.*" -Force
    Remove-Item "C:\aCogSpaceSeed\00flow\s-hermes\workspace.harness" -Force
    ```
