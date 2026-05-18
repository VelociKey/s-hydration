# Active Task Goal: Sovereign Hermes Re-Animation

**Task ID:** `00flow-TASK-HERMES-REANIMATION`  
**Sponsor:** Conductor / NATVS Engine  
**Status:** ACTIVE (Negotation & Assimilation Complete)  
**Target Tracks:** **[hermes_reanimation_production_track.md](file:///C:/aCogSpaceSeed/conductor/tracks/hermes_reanimation_production_track.md)**

---

## 🎯 Global Goal Mandate

The **NATVS Engine** must execute the five sequential phases detailed in the **[Sovereign Hermes Re-Animation Production Track](file:///C:/aCogSpaceSeed/conductor/tracks/hermes_reanimation_production_track.md)** to fully deploy, network, and compile the Go-reanimated **Hermes Agent** natively for production-grade raw QUIC execution and MV4 dashboard visualization.

---

## 📋 Directives & Invariant Constraints

1.  **Zero-Network Invariant:** The compiler must resolve all external dependencies—specifically `quic-go`—locally against the offline cache pooled inside the **[s-forge/go.mod](file:///C:/aCogSpaceSeed/00flow/s-forge/go.mod)** module. Do **NOT** attempt any internet-facing `go get` or online module fetch commands.
2.  **Sovereign Environment Variables:** During Step 2 compilation, the compiler must set:
    *   `GOOS=wasip3`
    *   `GOARCH=wasm`
    And compile natively inside the `00flow/s-hermes` workspace to produce `00flow/s-forge/bin/hermes_wasm_gc.wasm`.
3.  **Monotonic Authority Guard:** Assert that SACP capability registrations in the newly compiled binary strictly validate that current execution authority decays monotonically ($\text{CurrentAuthority} \le \text{OriginalAuthority}$).
4.  **Verification Success Requirement:** Transition to Step 5 (Metabolic Synthesis) only after all unit and integration tests in `sov.fleet/s-hermes` compile and pass with **0 errors**.

---

## 🚀 Active Roadmap to Execute

*   [x] **Phase 1:** Implement raw `quic-go` UDP network loop inside `s-aether/hermes_core.go` mapping stream channels concurrently.
*   *Verification:* Running local mock server compilation tests.
*   [x] **Phase 2:** Compile the production binary for WASM-GC under `wasip1` and verify binary payload size stays under 4MB.
*   [ ] **Phase 3:** Setup the sandboxed Firecracker guest vsock-to-UDP bridge in VM using native Go agent.
*   [ ] **Phase 4:** Calculate the SHA-256 hash, generate the **Veracity Seal**, and seal the re-hydrated software in `sbom_external_artifact.webnf`.
*   [ ] **Phase 5:** Write the Dart JS-Interop & heap-sharing visualizer bindings inside the Flutter frontend (MV4).
