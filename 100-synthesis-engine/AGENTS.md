# Dynamic Synthesis Engine (sHydrator/100-synthesis-engine)

This directory houses the dynamic compiler runners and translation adapters (currently backed by Bazel Starlark rules) for the 00flow platform.

## Key Mandates:
1.  **Compiler Orchestration:** Hermetically compiles Go WASM-GC, Dart/Flutter, and backend compiled actors from raw local source code.
2.  **IP Separation:** Handles compilation for both platform core repositories and licensee custom proprietary modules.
3.  **Target Promotion:** Symmetrically routes finished compiled binaries to the correct `INTERNAL_ACTOR` or `INTERNAL_TOOLCHAIN` folders inside `s-forge` or `sCauldron` by querying the central path registry.
