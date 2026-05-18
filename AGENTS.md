# Workspace: s-hydration

This workspace houses the NATVS Hydration Engine, which orchestrates Dynamic build synthesis, Static asset ingestion, and Capability Discovery for the 00flow platform and licensees.

## Mandates
- **Orchestration:** Executes the centralized Hydration parser and compiler runners across the global fleet.
- **Triple Hydration:** Manages dynamic compilers (compilation rules in `100-synthesis-engine/`), static retrievals (in `200-ingestion-engine/`), and capability discovery (in `300-discovery-engine/`).
- **Integration:** Promotes verified compiled/acquired/discovered artifacts symmetrically to s-forge (Platform) and sCauldron (Licensees) using the central `400-registry/hydrator.wag` mappings.

## SACP / qAPC raw QUIC Core Integration
This workspace holds the fully implemented, compiled, and verified SACP / qAPC raw QUIC networking core:
*   **Architecture & Design Specs:** Read **[SACP_ARCHITECTURE.md](file:///C:/aCogSpaceSeed/00flow/s-hydration/SACP_ARCHITECTURE.md)** for rationales and layered system designs.
*   **Grammar Schema:** **[sacp.wag](file:///C:/aCogSpaceSeed/00flow/s-hydration/400-registry/sacp.wag)** conforming to AAIF standard.
*   **Symmetric Memory-Slot Proxies:** **[sacp_proxy.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_proxy.go)** (64-byte zero-copy serialization).
*   **Host-Side SACP Broker:** **[sacp_broker.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_broker.go)** (raw UDP/QUIC listener and monotonic checker).
*   **Conformance Test Suite:** **[sacp_conformance_test.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_conformance_test.go)** (asserts and validates all security invariants).

