# Sovereign Agentic Context Protocol (SACP) Architecture Specification

**Document ID:** `00flow-SACP-ARCH-2026-05-18`  
**Classification:** Technical Architecture / R&D Silo  
**Target:** `s-hydration` Platform Workspace & NATVS Engine

This document details the architectural blueprint, multi-layer designs, high-performance implementations, and technical rationales governing the upgraded **SACP raw QUIC (quic-go) Core** inside the **NATVS Engine**.

---

## 1. Multi-Layer System Architecture & Rationales

The SACP engine decouples communication concern into four strict, independent layers to enforce a zero-trust network boundary and ultra-low latency execution:

```
                            [SACP MULTI-LAYER SYSTEM]
                            
      ┌──────────────────────────────────────────────────────────────────┐
      │  LAYER 4: SOVEREIGN MICROVM CONTEXT CONFINEMENT (Jules BCP)      │
      │  - Firecracker guest-host virtio-fs mapping                      │
      │  - socat vsock-to-UDP socket loops                               │
      └─────────────────────────────────┬────────────────────────────────┘
                                        │
      ┌─────────────────────────────────▼────────────────────────────────┐
      │  LAYER 3: MULTIPLEXED raw QUIC UDP NETWORK CONDUIT (quic-go)     │
      │  - Permanent Control Stream (0) vs. Ephemeral Data Stream (4+)   │
      │  - Unreliable Datagrams for 120Hz MV4 Telemetry                 │
      └─────────────────────────────────┬────────────────────────────────┘
                                        │
      ┌─────────────────────────────────▼────────────────────────────────┐
      │  LAYER 2: ZERO-COPY SYMMETRIC PROXIES (s-latentlingua gogen)      │
      │  - 64-byte aligned symmetric caller/callee offset slotting       │
      │  - Zero-reflection, zero-parsing in-memory serialization          │
      └─────────────────────────────────┬────────────────────────────────┘
                                        │
      ┌─────────────────────────────────▼────────────────────────────────┐
      │  LAYER 1: FORMAL GRAMMAR SCHEMA CORE (sacp.wag)                  │
      │  - eBNF-conforming deterministic parsing                         │
      │  - Generalized domain-agnostic capability frames                  │
      └──────────────────────────────────────────────────────────────────┘
```

---

### 1.1 Layer 1: Formal Grammar Schema Core (`sacp.wag`)
*   **Implementation Location:** **[sacp.wag](file:///C:/aCogSpaceSeed/00flow/s-hydration/400-registry/sacp.wag)**
*   **Design:** A non-recursive, flat-block Wirth Syntax Notation/eBNF grammar defining `qapc_header` and generic `capability_frame` topologies. Rather than hardcoding specific Model Context Protocol (MCP) parameters, the grammar defines abstract frames mapping generic domains (`resource`, `tool`, `prompt`), actions (`register`, `discover`, `call`), and arbitrary parameter dictionaries.
*   **Rationale:** 
    *   ** Hallucination-Free Boundaries:** Restricting schema inputs to strict, formal eBNF grammars ensures that AI agents (Jules, Conductor) operate within a mathematically bounded range of valid intent, completely eliminating malformed payloads.
    *   **Generalized Extension (The "Zero-Change" Invariant):** Decoder routines only parse abstract capability frames. When the MCP standard is re-souled or upgraded, **we write zero Go code changes** to the SACP core; the new features are integrated purely by updating dynamic WebNF dictionary databases.

---

### 1.2 Layer 2: Zero-Copy Symmetric Proxies (`sacp_proxy.go`)
*   **Implementation Location:** **[sacp_proxy.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_proxy.go)**
*   **Design:** Compiled caller and callee symmetric proxy bindings using uniform 64-byte slots (`SlotSize = 64`). The structure contains dedicated setters and getters mapping to fixed, pre-calculated byte offsets within a raw, shared byte slice.
*   **Rationale:**
    *   **Sub-Microsecond Latency:** Traditional JSON serialization relies on reflection, causing CPU and memory bottlenecks. Symmetric proxies perform direct memory operations (`copy(buf[offset:offset+SlotSize], val)`), allowing serialization and deserialization in **under 100 microseconds**.
    *   **WASM-GC Surface Heap Sharing:** Go WASM-GC and Dart/Flutter share a single, unified memory heap in the browser visualizer (**MV4**). Symmetric proxies let the UI project states directly from raw UDP stream offsets in memory, bypassing standard state management wrappers (BLoC/Redux) and delivering liquid-smooth 120Hz visuals.

---

### 1.3 Layer 3: Multiplexed raw QUIC UDP Network Conduit (`quic-go`)
*   **Implementation Location:** **[sacp_broker.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_broker.go)**
*   **Design:** Transition from flat UDP packets to a multiplexed **raw QUIC network topology** using the **`quic-go`** library directly compiled in Go WASM-GC. Assigns specific concerns to concurrent bi-directional streams:
    *   *Stream ID: 0 (Permanent Control Stream):* Dedicated to security handshakes, monotonic authority validations, and real-time cancels/interrupts.
    *   *Stream ID: 4+ (Ephemeral Data Stream):* Dynamically opened to transfer high-bandwidth tool binaries and compiler stubs.
    *   *Unreliable Datagrams:* Feeds real-time progress and stdout console logs.
*   **Rationale:**
    *   **HOL-Blocking Prevention:** Under TCP/WebSockets, high-priority control frames are blocked behind large compilation transfers. QUIC streams are processed independently, ensuring a user cancel or state interrupt reaches the engine instantly.
    *   **No WebTransport Overhead:** WebTransport introduces heavy browser constraints and strict TLS certificate verification rules. Shifting to **raw QUIC (quic-go)** allows us to run pure Go UDP socket code inside the WASM-GC client, removing WebTransport dependencies completely.

---

### 1.4 Layer 4: Sovereign MicroVM Context Confinement (Firecracker)
*   **Implementation Location:** **[sacp_conformance_test.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_conformance_test.go#L94)** (and Guest `/etc/init.d/S99qapc-bridge`)
*   **Design:** Quarantined guest VM boundaries routing virtual socket (`vsock`) connections (Guest CID 3, Port 1000) directly to local host UDP loopback sockets (`127.0.0.1:9099`) via a background `socat` bridge daemon inside the Guest rootfs image.
*   **Rationale:**
    *   **Zero-Trust Isolation:** Compilation and synthesis of dynamic, untrusted actor stubs occur entirely inside independent, sandboxed AWS Firecracker microVMs. The host filesystem is protected, and compiler tools are dynamically mounted on-demand via `virtio-fs`.

---

## 2. Core Security & Transition Invariants

### 2.1 The Monotonic Authority Invariant

To prevent privilege escalation by compromised Guest VM compiler processes, SACP enforces the **Monotonic Authority Invariant** at the physical socket layer:

$$\text{CurrentAuthority} \le \text{OriginalAuthority}$$

If a quarantined guest VM attempts to escalate its privilege level (e.g., executing a command reserved for `AuthManager` while running at `AuthWorker`), the host-side `SACPBroker` validates the header signatures and immediately drops the connection over the wire.

### 2.2 The Metabolic Hot-Swap Handshake

To maintain 100% system operational uptime and avoid dropping active sessions during live upgrades, SACP utilizes a 4-phase metabolic transition:

```
                            [METABOLIC HOT-SWAP PHASE]
                            
   [Phase 1: Boot] ──────► [Phase 2: Token] ──────► [Phase 3: Swap] ──────► [Phase 4: Teardown]
   - WS TCP:8080 active     - Single-use token      - Connects UDP:9099     - Tears down legacy
   - Client requests upgrade  issued to client       - Streams transition     WS TCP connection
```

Because the UDP/QUIC server is initialized before the legacy WebSocket is terminated, sessions are elevated *live* mid-stream with **zero state loss**.

---

## 3. Reference Links

*   **AAIF Master Grammar Specification:** **[AAIF.webnf](file:///C:/aCogSpaceSeed/00AAIF/AAIF.webnf)**
*   **SACP Grammar Schema:** **[sacp.wag](file:///C:/aCogSpaceSeed/00flow/s-hydration/400-registry/sacp.wag)**
*   **Symmetric Proxy Bindings:** **[sacp_proxy.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_proxy.go)**
*   **Host Loopback Broker:** **[sacp_broker.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_broker.go)**
*   **Conformance Verification Tests:** **[sacp_conformance_test.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_conformance_test.go)**
