# Capability Discovery Engine (s-hydration/300-discovery-engine)

This directory houses the dynamic capability discovery mechanisms of the NATVS Hydration Engine, enabling federated tool and context ingestion from compliant external systems.

## Key Mandates:
1.  **Federated Interface Ingestion:** Dynamically queries and discovers capabilities (tools, dynamic context resources, and structured prompts) exposed by external or peer agents using standard protocol bindings, primarily the **Model Context Protocol (MCP)**.
2.  **Semantic Capability Mapping:** Synthesizes identified capabilities and tool schemas into formal WAG-conforming definitions (`capability_catalog.webnf`) within `s-forge/91000-ext-artifact-cognition-intelligence/` to validate the interface syntactic boundary before execution.
3.  **Zero-Trust Client Bindings:** Spawns and configures lightweight, secure client bindings (JSON-RPC over standard input/output, WebTransport, or QUIC) to orchestrate external capabilities safely within the NATVS reasoning plane without running un-vetted third-party code in privileged platform spaces.
