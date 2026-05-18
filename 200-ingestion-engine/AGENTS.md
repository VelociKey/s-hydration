# Static Ingestion Engine (sHydrator/200-ingestion-engine)

This directory houses the static materialization and declarative retrieval adapters of the NATVS Hydration Engine.

## Key Mandates:
1.  **Declarative Hydration:** Resolves static mirror download targets, external packages, and pre-compiled developer toolchains.
2.  **Veracity Validation:** Cryptographically checks the SHA-256 signatures and `VERACITY_SEAL` marks of every fetched asset against the trusted authority keys inside `90000`/`95000` before extracting or launching them.
3.  **Encapsulation Boundaries:** Restores directories and registers symlinks hermetically inside `EXTERNAL_TOOLCHAIN` or `EXTERNAL_BLUEPRINT` paths without contaminating platform source code spaces.
