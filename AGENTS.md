# Workspace: sHydrator

This workspace houses the NATVS Hydration Engine, which orchestrates Dynamic build synthesis and Static asset ingestion for the 00FLOW platform and licensees.

## Mandates
- **Orchestration:** Executes the centralized Hydration parser and compiler runners across the global fleet.
- **Dual Hydration:** Manages both dynamic compilers (compilation rules in `100-synthesis-engine/`) and static retrievals (in `200-ingestion-engine/`).
- **Integration:** Promotes verified compiled/acquired artifacts symmetrically to sForge (Platform) and sCauldron (Licensees) using the central `300-registry/hydrator.wag` mappings.
