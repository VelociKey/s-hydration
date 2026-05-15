# Workspace: sBuilder

This workspace specializes in build orchestration and cache management for the 00FLOW platform. 

## Mandates
- **Orchestration:** Responsible for the centralized Bazel build process across the global fleet.
- **Operation:** Executes Bazel and other required toolchain executables (e.g., compilers) directly from the `sforge` repository.
- **Integration:** Acts as the primary builder of internal artifacts, promoting verified assets to the `9xxxx` authority ranges.
