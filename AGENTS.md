# Workspace: s-hydration

This workspace houses the NATVS Hydration Engine, which orchestrates Dynamic build synthesis, Static asset ingestion, and Capability Discovery for the 00flow platform and licensees.

## Mandates
- **Orchestration:** Executes the centralized Hydration parser and compiler runners across the global fleet.
- **Triple Hydration:** Manages dynamic compilers (compilation rules in `100-synthesis-engine/`), static retrievals (in `200-ingestion-engine/`), and capability discovery (in `300-discovery-engine/`).
- **Integration:** Promotes verified compiled/acquired/discovered artifacts symmetrically to s-forge (Platform) and sCauldron (Licensees) using the central `400-registry/hydrator.wag` mappings.
- **Registry & Manifest Pattern:** All hydration processes (external artifacts, internal builds, and any other source we use) must follow the Routing vs. Audit separation:
  - The registry (`hydratedregistry.go`) functions as a routing table holding stable, version-agnostic names mapped to locations.
  - The SBOM/Manifest (`sbom.external_artifact.webnf`) serves as the audit log containing versions, dates, and verification hash history.
  - Stand-alone synchronization/repair (`hydrator -sync`) keeps the registry aligned using a two-phase filesystem veracity scanner (Phase A: Scan & Add; Phase B: Verify & Delete).

## SACP / qAPC raw QUIC Core Integration
This workspace holds the fully implemented, compiled, and verified SACP / qAPC raw QUIC networking core:
*   **Architecture & Design Specs:** Read **[SACP_ARCHITECTURE.md](file:///C:/aCogSpaceSeed/00flow/s-hydration/SACP_ARCHITECTURE.md)** for rationales and layered system designs.
*   **Grammar Schema:** **[sacp.wag](file:///C:/aCogSpaceSeed/00flow/s-hydration/400-registry/sacp.wag)** conforming to AAIF standard.
*   **Symmetric Memory-Slot Proxies:** **[sacp_proxy.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_proxy.go)** (64-byte zero-copy serialization).
*   **Host-Side SACP Broker:** **[sacp_broker.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_broker.go)** (raw UDP/QUIC listener and monotonic checker).
*   **Conformance Test Suite:** **[sacp_conformance_test.go](file:///C:/aCogSpaceSeed/00flow/s-hydration/100-synthesis-engine/sacp_conformance_test.go)** (asserts and validates all security invariants).

