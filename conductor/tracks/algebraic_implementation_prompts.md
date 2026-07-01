# Conductor Track Prompts: Algebraic Platform Implementation

This document contains high-fidelity, unambiguous implementation prompts designed for the **Conductor** to spin up execution tracks (delegating to Jules, go_builder, or Antigravity subagents) for each of the four roadmap phases.

---

## Track Prompt 1: Phase 1 — Cryptographic & Path Security Foundations

```markdown
Role: Sovereign Security Engineer (Jules / Go-Builder)
Goal: Implement the s-hardware-bridge, integrate Lattice-Based Path Confinement inside s-agentbox, and build s-taxonomy-guard.

Instructions:

1. Create the s-hardware-bridge workspace:
   - Location: C:\aCogSpaceSeed\00flow\s-hardware-bridge
   - Package Name: sov.fleet/s-logiclibrary/00200-logic-libraries/hardware
   - Requirements:
     * Implement Go dynamic syscall wrappers using standard CGO or syscall.NewLazyDLL to bind with the local Windows TPM 2.0 PKCS#11 API.
     * Expose 'SignPayloadWithTPM(keyID string, payload []byte) ([]byte, error)'.
     * Expose 'GetHardwareSVID() (*x509.Certificate, error)' to extract the hardware-bound identity certificate.
     * Verify logic against a software-mock fallback if hardware TPM is missing.
     * Implement test assertions in 900-attestations/ verifying payload signature verification cycles.

2. Integrate Lattice-Based Path Confinement:
   - Locations: C:\aCogSpaceSeed\00flow\s-logiclibrary\00200-logic-libraries\sandbox\ and C:\aCogSpaceSeed\00flow\s-agentbox\
   - Requirements:
     * Map workspace paths to dense Bitset array positions using FNV-1a perfect hashes.
     * Implement 'func EvaluateLatticePath(targetPath string, allowedMask *Bitset) bool' checking the bit representation in sub-nanosecond lookups.
     * Hook this function inside the Custom File Open/Read wrappers of s-agentbox, intercepting operations and blocking actions that fail the mask check.
     * Write tests verifying symlink and path traversal ('../../') attempts are blocked.

3. Build s-taxonomy-guard:
   - Location: C:\aCogSpaceSeed\00flow\s-taxonomy-guard
   - Package Name: sov.fleet/s-natives/cmd/taxonomy-guard
   - Requirements:
     * Set up a filesystem watcher (using fsnotify) on the workspace root directory.
     * On file modify/rename events, retrieve the calling PID.
     * Retrieve the calling process SVID certificate via s-hardware-bridge, evaluate its path lattice mask, and instantly revert directory changes if the credentials fail validation.
     * Compile to 'taxonomy-guard.exe'.
```

---

## Track Prompt 2: Phase 2 — Speed & Concurrency Exploitation

```markdown
Role: Performance & Build Systems Engineer (Jules / Go-Builder)
Goal: Implement Algebraic Workspace Synchronization and deploy Algebraic Topological Work Scheduling.

Instructions:

1. Deploy Algebraic Workspace Synchronization:
   - Location: C:\aCogSpaceSeed\00flow\s-hydration/pkg/sync
   - Requirements:
     * Map a mock workspace containing 1,000,000 files to index positions in a dense Bitset array.
     * Implement the bitwise XOR delta function: 'CalculateQuotientDelta(base, local *Bitset) *Bitset' representing state divergence modulo unchanged directories.
     * Implement 'ResolveMerge(main, branchA, branchB *Bitset) (*Bitset, error)' performing intersection checks to evaluate merge compatibility.
     * Wire this package into the primary staging and distribution pipelines of int-rehydrator.go and wraithd.
     * Write tests simulating 100 concurrent Actor merges to confirm conflict checks finish in microsecond timescales.

2. Deploy Algebraic Topological Work Scheduling:
   - Location: C:\aCogSpaceSeed\00flow\s-hydration/pkg/scheduler
   - Requirements:
     * Model the global package dependency graph as a sparse adjacency matrix (MD) using Compressed Sparse Row (CSR) formats.
     * Implement AVX-512 vectorized matrix-vector multiplication (MD * v) over the binary field F_2 to compute target task in-degrees.
     * Loop through in-degrees to isolate task batch vectors (v_0) containing zero unresolved dependencies.
     * Schedule tasks in batch v_0 in parallel Go goroutines, verifying that their write-ideals are disjoint.
     * Wire this scheduler into the build cascader of int-rehydrator.go to replace the procedural topological sort.
```

---

## Track Prompt 3: Phase 3 — Communication Safety & Message Conformity

```markdown
Role: Language Parser & Assurance Engineer (Jules / Go-Builder)
Goal: Build s-schema-broker and deploy AST Code-Gen Linting inside s-latentlingua.

Instructions:

1. Build s-schema-broker:
   - Location: C:\aCogSpaceSeed\00flow\s-sacp/cmd/schema-broker
   - Requirements:
     * Create an IPC interceptor that sits on Unix/TCP connection listener loops.
     * For each received packet, extract the topic string and look up its matching Wirth Syntax Notation EBNF schema compiled via s-latentlingua.
     * Parse the raw byte payload against the EBNF schema. If it fails, discard the message, return a SchemaValidationError frame, and close the socket.
     * Compile to 'schema-broker.exe' and wire it as a middleware daemon on SACP buses.
     * Write tests verifying that malformed payloads are successfully blocked.

2. Deploy AST Code-Gen Linting:
   - Location: C:\aCogSpaceSeed\00flow\s-latentlingua/pkg/linter
   - Requirements:
     * Extend sn-lint.exe to parse .wag grammar files.
     * Add compiler-validation checks that block grammars containing dynamic string evaluation patterns that could run unchecked shell executions.
     * Analyze generated Go/Dart templates to prove they contain zero raw system calls.
```

---

## Track Prompt 4: Phase 4 — Experimental Isolation & Promotion

```markdown
Role: Systems Virtualization & Lifecycle Engineer (Jules / Go-Builder)
Goal: Build s-actor-attestator, create x-sandbox-lab using Firecracker wrappers, and automate promotion in x-transform-antigravity.

Instructions:

1. Build s-actor-attestator:
   - Location: C:\aCogSpaceSeed\00flow\s-actors/pkg/attestation
   - Requirements:
     * Create a post-compile utility that computes a Blake3 hash of Actor binaries.
     * Sign the hash using the local TPM private key via s-hardware-bridge and append the signature block to the binary footer.
     * Add startup verification checks inside the native execution engine, parsing the binary footer and verifying its signature against the trust circle anchors before allowing execution.
     * Verify that launching an unsigned or modified binary triggers an attestation failure.

2. Build x-sandbox-lab:
   - Location: C:\aCogSpaceSeed\00xper/x-sandbox-lab
   - Requirements:
     * Implement Go wrappers that launch and configure Firecracker micro-VM containers.
     * Spin up a micro-VM containing a minimal Linux kernel and mount the test workspace read-only.
     * Execute untrusted Actors from x-actors inside the container and log resource usage parameters.
     * Verify path containment boundaries.

3. Automate Grammar-Guided Workspace Promotion:
   - Location: C:\aCogSpaceSeed\00xper/x-transform-antigravity/pkg/promotion
   - Requirements:
     * Write EBNF transformation rules in a .wag mapping schema.
     * Parse experimental Go code files inside x-actors, rewrite import paths, and promote them directly into production s-actors paths under standard SWDT casing formats.
     * Verify promoted code conforms to conformance.exe checks.
```
