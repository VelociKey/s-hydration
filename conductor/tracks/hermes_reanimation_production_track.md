# Actionable Hand-off Track: Sovereign Hermes Re-Animation Production

**Track ID:** `00flow-TRACK-HERMES-REANIMATION`  
**Execution Context:** Sovereign Workspace Silos (`00flow/s-hermes`, `00flow/s-hydration`, `00flow/s-forge`)  
**Target:** NATVS Engine / Go WASM-GC Compiler (`wasip3`)

This track provides the exact, actionable steps for the NATVS Engine to fully deploy, network, and compile the Go-reanimated **Hermes Agent** for production-grade raw QUIC execution and MV4 dashboard visualization.

---

## 📋 Sovereign Execution Context & Constants

The executing agent must utilize these locked system coordinates directly:
*   **Sovereign Go executable:** `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe`
*   **Sovereign Scratch Space:** `C:\aCogSpaceSeed\c0990-ephemeral-scratch`
*   **Production Binary Output:** `C:\aCogSpaceSeed\00flow\s-forge\bin\hermes_wasm_gc.wasm`
*   **Unified Workspace Ledger:** `C:\aCogSpaceSeed\go.work`

---

## 🚀 Step-by-Step Production Plan

### Step 1: Implement the raw QUIC/UDP Dial Loop in `s-hermes`
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-hermes\hermes_core.go`
*   **Actionable Task (100% Offline-Hydrated Invariant):**  
    Expand `hermes_core.go` to include a background raw QUIC UDP network client:
    1.  Import `github.com/quic-go/quic-go` inside the Go core. 
        > [!NOTE]
        > The `quic-go` package is already fully hydrated inside `sov.fleet/s-forge`'s go.mod! The Go workspace compiler resolves this package locally via `go.work` pool mapping; **do NOT attempt any internet-facing go get commands**.
    2.  Implement `DialTelemetryServer(addr string)`: Establishes a permanent, high-priority bi-directional control stream (Stream ID: 0) to host port `9099`.
    3.  Create an automated telemetry dispatcher: Writes SACP zero-copy capability frames directly to the QUIC stream buffer on every skill execution.
*   **Verification:** Run local mock server compilation test:
    ```powershell
    & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" test -v sov.fleet/s-hermes
    ```

---

### Step 2: Compile the Production WASM-GC Binary
*   **Target Path:** `C:\aCogSpaceSeed\00flow\s-forge\bin\hermes_wasm_gc.wasm`
*   **Actionable Task:**  
    Compile the re-animated Go-based core targeting the shared-heap WebAssembly Garbage Collection runtime (`wasip3`):
    1.  Configure env flags: `GOOS=wasip3`, `GOARCH=wasm`.
    2.  Compile using the locked toolchain without internet access:
        ```powershell
        $env:GOOS="wasip3"
        $env:GOARCH="wasm"
        & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" build -o C:\aCogSpaceSeed\00flow\s-forge\bin\hermes_wasm_gc.wasm C:\aCogSpaceSeed\00flow\s-hermes\hermes_core.go
        ```
*   **Verification:** Assert that the compiled WASM binary is successfully saved and its size remains under **4MB** (bypassing Pyodide payload bloats).

---

### Step 3: Implement Dart JS-Interop & heap-sharing visualizers (MV4)
*   **Target Workspace (Frontend):** `00flow/s-hydration/web`
*   **Actionable Task:**  
    Set up the frontend bindings inside Flutter's MV4 surface to read the compiled Hermes WASM heap:
    1.  Write Dart JS-Interop bindings importing `hermes_wasm_gc.wasm`.
    2.  Coordinate memory offsets directly from the 64-byte symmetric slot proxy outputs (`UUIDOffset`, `OrigAuthOffset`, `PayloadOffset`).
    3.  Bind progress coordinates directly to the reactive telemetry visualizer console.
*   **Verification:** Assert that SACP data streams project live UI updates without state copy wrappers.

---

### Step 4: Firecracker Guest Sandboxed Bridge Integration
*   **Target Path (VM Context):** `/etc/init.d/S99hermes-bridge`
*   **Actionable Task:**  
    Configure the transitional Firecracker guest microVM startup to test the Go-based guest agent:
    1.  Spin up the compiled Go-based guest Hermes agent inside a locked VM (strictly avoiding Python to meet zero-overhead guidelines).
    2.  Start the local `socat` bridge mapping vsock CID 3 port 1000 directly to local guest loopback UDP port `9099`.
    3.  Route all standard MCP calls through the native Go SACP frame parser.
*   **Verification:** Ensure guest-to-host UDP loopback packets reach the host `SACPBroker` cleanly.

---

### Step 5: Metabolic Synthesis Sealing
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-forge\90100-rehydration-seed\sbom_external_artifact.webnf`
*   **Actionable Task:**  
    Perform zero-trust compliance sealing on the newly re-animated core:
    1.  Calculate the SHA-256 hash of `hermes_wasm_gc.wasm`.
    2.  Generate a cryptographic **Veracity Seal** using the central Conductor keys.
    3.  Append and format the seal record into the authoritative supply-chain catalog `sbom_external_artifact.webnf`.
*   **Verification:** Assert that the re-hydrated software conforms strictly to WAG grammar validators.
