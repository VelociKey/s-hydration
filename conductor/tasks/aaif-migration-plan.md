# Implementation Plan: AAIF Migration to QUIC+CODED & Material View 4

This plan outlines the migration of the Agent Architecture & Integration Framework (AAIF) from standard A2A (Protobuf/gRPC) to a sovereign, grammar-based stack using QUIC and weBNF-defined CODED protocols, with a specialized Flutter/Go hybrid UI architecture.

## 1. Architectural Shift: "The Sovereign Stack"

| Component | Legacy (A2A) | Sovereign (AAIF 2.0) |
| :--- | :--- | :--- |
| **Networking** | TCP / gRPC (A2A v1.0) | **QUIC / WebTransport (WASI 0.3)** |
| **Serialization** | Protobuf / Structs (A2A v1.0) | **CODED / A2UI Blueprints** (weBNF) |
| **Logic Layer** | Dart / Go (Mixed) | **Go 1.26 ("Green Tea")** (`wasip3`) |
| **View Layer** | Flutter (React-style) | **Flutter 3.41 ("Fire Horse")** (MV4) |
| **Orchestration** | Service Mesh | NATVS Lifecycle |
| **Specification** | English Narratives | **weBNF Grammars** |

## 1.1 Grammar-First Specification
AAIF 2.0 is defined as a machine-executable grammar first. English descriptions in `sCognition` serve as the narrative frame for the underlying weBNF logic. This logic is bridged with the **Agentic AI Foundation (Linux Foundation)** standards, specifically the **Model Context Protocol (MCP)** and the **AGENTS.md** convention, to ensure global federation conformance.

## 2. Phase 1: Cognitive Foundation (sCognition)

Add foundational specifications to `000ALL/sCognition`.

### 2.1. CODED Protocol Definition
- **Location**: `43000-grammar-transports/CODED-spec.md`
- **Objective**: Define the binary format derived directly from weBNF grammars.
- **Key Feature**: Zero-overhead parsing where the grammar *is* the schema, eliminating the need for separate code generation steps like `protoc`.

### 2.2. QUIC Handshake & Stream Topology
- **Location**: `22000-edge-transports/QUIC-topology.md`
- **Objective**: Define how NATVS "Negotiation" happens over QUIC streams.

### 2.3. Material View 4 (MV4) Standards
- **Location**: `21000-presentation-contexts/MV4-standards.md`
- **Objective**: Document the transition to MV4 in Flutter, focusing on dynamic tokens and Go-driven state projections.

## 3. Phase 2: Toolchain Implementation (sLatentLingua & sForge)

### 3.1. Foundational CODED Grammar
- Create `CODED.webnf` in `sLatentLingua`.
- This grammar will define the "primitive types" and "message structures" for all agentic communication.

### 3.2. Go-to-Dart "Surface" Bridge
- Develop a Go package that projects internal state machines directly into a format consumable by Material View 4 components.
- Dart becomes a "thin surface" that renders these projections.

## 4. Phase 3: Integration & Verification (NATVS)

### 4.1. Negotiation (N)
- Handshake via QUIC, exchanging weBNF capabilities.

### 4.2. Assimilation (A)
- Ingesting context via CODED-serialized streams.

### 4.3. Transformation (T)
- Go-based processing of the assimilated context.

### 4.4. Verification (V)
- Flutter/MV4 visual validation of the result.

### 4.5. Conformance Verification
All migration steps must be validated against the formal grammars. Conformance status is tracked in the **[Conformance Ledger](file:///c:/aCogSpaceSeed/000ALL/sCognition/80000-system-governance/conformance-ledger.md)**.

## 5. Next Steps

1. [ ] Create the `CODED` specification draft in `sCognition`.
2. [ ] Initialize `00AAIF/standards` with the new protocol markers.
3. [ ] Prototype a simple "Counter" or "Hello" app using Go-Logic -> CODED -> Flutter MV4 Surface.

---
> [!IMPORTANT]
> This migration marks the departure from the "React world" of state management. We are moving logic "back" to the engineering-grade language (Go) and keeping Dart strictly for what it does best: UI Surface rendering.
