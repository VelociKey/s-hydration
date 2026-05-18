# Actionable Hand-off Track: SACP natvs-engine Integration
**Track ID:** `00flow-TRACK-SACP-NATVS`  
**Execution Context:** Sovereign Workspace Silos (`00flow/s-hydration`, `00aaif/p-aaif`, `00flow/s-latentlingua`)  
**OS/Shell Target:** Windows PowerShell / Gemini CLI  

---

## 📋 Global Execution Context & Variables

To ensure zero-trust compliance, the executing agent must utilize these sovereign coordinates directly:
*   **Sovereign Go compiler:** `C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe`
*   **Sovereign Meta-Grammar:** `C:\aCogSpaceSeed\00flow\s-latentlingua\30100-meta-foundation\webnf.sn`
*   **AAIF Master Grammar:** `C:\aCogSpaceSeed\00aaif\AAIF.webnf`
*   **Symmetric Proxy Generator:** `C:\aCogSpaceSeed\00flow\s-latentlingua\02000-logic-libraries/gogen/generator.go`
*   **Sovereign Ephemeral Scratch:** `C:\aCogSpaceSeed\c0990-ephemeral-scratch`

---

## 🚀 Step-by-Step Actionable Execution Track

### Step 1: Define `sacp.wag` Grammar conforming to `AAIF.webnf`
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-hydration\400-registry\sacp.wag`
*   **Actionable Task Description:**  
    Create the SACP WAG grammar defining the generic capability frames and the qAPC attestation headers. The grammar must strictly implement:
    1.  `qapc_header`: Maps `OriginalUUID` (string), `OriginalAuthority` (enum), `CurrentAuthority` (enum), and an array of `AttestationBlocks`.
    2.  `capability_frame`: Maps generalized `domain` (`resource`, `tool`, `prompt`), `action` (`register`, `discover`, `call`), `id`, and parameters.
*   **Execution Commands:**
    ```powershell
    # 1. Create file and populate with grammar rules
    # 2. Run WAG parser compilation to verify grammar conformance
    & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" test -v sov.fleet/s-hydration/400-registry/...
    ```

---

### Step 2: Compile Symmetric proxies using `gogen`
*   **Target Output File:** `C:\aCogSpaceSeed\00flow\s-hydration\100-synthesis-engine\sacp_proxy.go`
*   **Actionable Task Description:**  
    Utilize our logic library's symmetric proxy compiler to generate caller and callee SACP serialization layers.
    1.  Call the `GenerateSymmetricProxy` routine for SACP capability frames.
    2.  Configure direct memory offsets based on `slotSize := 64` blocks.
    3.  Generate the following direct setters/getters in `sacp_proxy.go`:
        *   `SetOriginalUUID` / `GetOriginalUUID`
        *   `SetOriginalAuthority` / `GetOriginalAuthority`
        *   `SetCurrentAuthority` / `GetCurrentAuthority`
        *   `SetAttestationBlock` / `GetAttestationBlock`
*   **Code Structure Outline for `sacp_proxy.go`:**
    ```go
    package synthesis

    type SACPCallerProxy struct {
        buffer []byte
    }

    func (p *SACPCallerProxy) SetOriginalUUID(val []byte) {
        copy(p.buffer[0:64], val)
    }
    func (p *SACPCallerProxy) GetOriginalUUID() []byte {
        return p.buffer[0:64]
    }
    // Repeat for all Monotonic Authority & Attestation slots...
    ```

---

### Step 3: Implement Host-Side SACP UDP/QUIC Loopback Broker
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-hydration\100-synthesis-engine\sacp_broker.go`
*   **Actionable Task Description:**  
    Build the SACP loopback broker handling UDP port `9099` QUIC connections.
    1.  **Monotonic Authority Guardrail:** Implement the structural verification check enforcing:
        $$\text{CurrentAuthority} \le \text{OriginalAuthority}$$
        If a packet fails this check, drop it at the socket wire layer and log a security alert.
    2.  **Metabolic Handshake:** Implement the listener on legacy WebSocket TCP port `8080`. When a boot negotiation is received, issue a single-use transitions token, start the UDP QUIC listener, and safely transition the session live.
    3.  **VM Tool Hydration Link:** Implement the virtio-fs symlink mounting utility, mapping guest on-demand tools (like `tinygo`) under `tasks/<task_id>/92000-external-toolchains`.
*   **Execution Commands:**
    ```powershell
    # Verify compilation of the broker
    & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" build -v -o C:\aCogSpaceSeed\c0990-ephemeral-scratch\bin\sacp_broker.exe C:\aCogSpaceSeed\00flow\s-hydration\100-synthesis-engine\sacp_broker.go
    ```

---

### Step 4: Firecracker Guest `socat` Bridge Configuration
*   **Target Path (Guest Context):** `/etc/init.d/S99qapc-bridge` (Inside Guest rootfs image)
*   **Actionable Task Description:**  
    Configure the Firecracker VM guest startup scripts to bridge virtual socket traffic directly to loopback UDP streams.
    1.  Map vsock CID 3 (port 1000) directly to guest loopback UDP port `9099`.
    2.  Configure guest-side Jules startup rules to direct SACP telemetry packets to local `socat` targets.
*   **Guest Startup Script Snippet:**
    ```bash
    #!/bin/sh
    # S99qapc-bridge - Startup bridge daemon inside Guest MicroVM
    socat vsock-listen:1000,fork udp:127.0.0.1:9099 &
    ```

---

### Step 5: Conformance and Integration Verification tests
*   **Target File Location:** `C:\aCogSpaceSeed\00flow\s-hydration\100-synthesis-engine\sacp_conformance_test.go`
*   **Actionable Task Description:**  
    Write comprehensive unit and integration tests asserting SACP security invariants and zero-copy performance metrics.
    1.  **TestMonotonicAuthority:** Validate that the broker blocks any attempts to escalate privileges.
    2.  **TestMetabolicHotSwap:** Assert that the active session hot-swaps live from TCP WebSockets to UDP QUIC without state loss.
    3.  **TestZeroCopyLatency:** Validate that proxy serialization and brace-balanced parsing takes under **500 microseconds** per frame.
*   **Execution Commands:**
    ```powershell
    # Run the comprehensive SACP conformance test suite
    & "C:\aCogSpaceSeed\00flow\s-forge\92000-external-toolchains\go\bin\go.exe" test -v -run TestSACPConformance sov.fleet/s-hydration/100-synthesis-engine/...
    ```

---
> [!IMPORTANT]
> The executing agent must ensure that all Go files are mapped in `go.work` and build cleanly under the target toolchains without external internet access. Every compilation must succeed with 0 errors before proceeding to the next step.
