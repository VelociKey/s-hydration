# Sovereign Hydration Check Audit Report
Generated on: 2026-06-01 13:53:32

## External Ingestion Registry (SBOM)

### CLI / Executables / Toolchains
| Target Name | Version | Download Status | Size | Expected Hash | Rationale |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **bazel** | 9.1.0 | PENDING | 11.2 MB | blake3:f99ef339a2c894fbd04c4311f62f416aee47e298ccf7691bb8ec777e7b2d49e8 | Core build tool and orchestration compiler. |
| **opentelemetry** | 0.110.0 | PENDING | 22.9 MB | blake3:4057691574fadba2e7906f71abbd939a78fe2f9bc967bbb4934a258c3eca032b | Observability framework for cloud-native software telemetry metrics. |
| **opentofu** | 1.8.2 | PENDING | 36.5 MB | blake3:6a24ce8a385113bea31978819fc4360604c8cf7f6e5ba2310635301bcf6a0986 | Open-source infrastructure as code tool bifurcated from Terraform. |
| **cosign** | 2.4.1 | PENDING | 15.2 MB | blake3:5e77391b54aacaa86b9e01cea9c1223ce45cc8988bad0bcd66b8007e3ab1605c | Sigstore CLI tool for signing and verifying OCI containers and WebAssembly modules. |
| **buildifier** | 7.1.2 | PENDING | 7.5 MB | blake3:49778cb4ab1c07b522d77a45b855351f026bbf7523604ddf7a2a3af9b865f7cb | Linter and formatter for Bazel BUILD, MODULE.bazel, and Starlark rule files. |
| **firebase-tools** | 13.11.2 | PENDING | 133.0 MB | blake3:7fa1489b4f4f1103f5741c53227381f4f7d01494ef8c8496c09427e5202aa25d | CLI tool for Firebase deployments and hosting management. |
| **flutter-sdk-firehorse** | 3.44.0 | VERIFIED | 0 B | blake3:f74e337653b894f733b8f675f83b3813ac954b56835fb7d3878b30e20b3592fa | Sovereign Flutter SDK distribution for UI and mobile/web development. |
| **gcloud-sdk** | 476.0.0 | VERIFIED | 366.2 MB | blake3:190085c7c083e36a4415d62bcb764025256efece912707077e6e96929cae28b2 | Google Cloud SDK for deployments and cloud service interactions. |
| **gh-cli** | 2.49.0 | PENDING | 60.3 MB | blake3:65dba6fc8322df8b9615f38c67eacd1311af8502d55de55cdc1b2c84213a50da | GitHub CLI for automating repository operations. |
| **gitleaks** | 8.18.2 | PENDING | 13.5 MB | blake3:498649fc27f366770a36802bdb36796656339db2aede0ef19064d5bcbe031a45 | Static secret scanner for auditing repository histories and filesystem trees. |
| **git** | 2.45.0 | PENDING | 119.7 MB | blake3:57905603f25b17c18775b807a7af8ed0d56c109e39f3c05cecc6996323ed0c28 | Sovereign Git client for source control synchronization. |
| **golang** | 1.26.3 | PENDING | 0 B | blake3:f6c3c59497ee553883bfcedf4737c4af90d679e1ce433a5d68a8fdb4040b404c | Required dependency. |
| **jdk-25-headless** | 25.0.0 | VERIFIED | 521.7 MB | blake3:943e6345db31de192046248f257004face3ff22c0a9337aaa7cc32eb7206d3d2 | Java Development Kit required by Bazel build orchestration. |
| **notary-notation** | 1.1.0 | PENDING | 12.5 MB | blake3:5d7c5ed720c197f160d139b9477cc2d1d4d9899f8efedc100502dd6b60a42845 | Notary CLI for signing and verifying artifact signatures. |
| **step-ca** | 0.26.2 | PENDING | 0 B | blake3:e82e9f96254cca0c9fda285b9b3ecbe3679815194a6149931c78b475f5f21ca2 | Sovereign private certificate authority for secure transport seals. |
| **trivy** | 0.70.0 | PENDING | 202.8 MB | blake3:f9544e4565b5ea5a4a9e36260d1478d8a7b0730d4ede116f83261bfdf61410a1 | Security scanner engine for offline vulnerability auditing. |
| **wasm-opt** | 117 | VERIFIED | 0 B | blake3:d38217ef1fddf575b757f576c4b5836376baadcab658195affa2cee048624e7c | WebAssembly optimizer tool from Binaryen toolchain. |
| **wasm-tools** | 1.210.0 | VERIFIED | 0 B | sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e | WebAssembly parsing, printing, and validation tools. |
| **wasmer** | 4.3.0 | VERIFIED | 0 B | sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e | WebAssembly runtime for running sovereign WASM modules. |
| **wasmtime** | 21.0.0 | VERIFIED | 0 B | blake3:72d09915fa4832566cd8fc78fd59afa96201cbcfeb6ecdd0c5f5adfce0004a28 | WebAssembly engine and compiler runtime. |

### Libraries / SDKs / Modules
| Target Name | Version | Download Status | Size | Expected Hash | Rationale |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **bazel-rules-flutter** | 2.1.0 | PENDING | 0 B | blake3:13c3fad23753c58d480c832c741521f7aed1256e1bdd514e3ff2fd9da5cc3c56 | Bazel rules for building Flutter/Dart applications. |
| **bazel-rules-go** | 0.51.0 | PENDING | 10.2 MB | blake3:1c26b60bd67b2c7fd45b63a1f695eccceb9061a0ffc3d54d1e69d8ffac04ec1f | Bazel rules for building Go applications. |
| **quic-go** | 0.59.1 | PENDING | 0 B | blake3:f3dc4bc1731a62a72aa97e15506828fbc17585503d4b44b2ff33e342b16c0aad | Go library implementation of the QUIC transport protocol. |
| **qpack** | 0.6.0 | PENDING | 0 B | blake3:b8d86f78e65e605e02fb4c9440c189b101876074dfea42e12711a07b739f8925 | Go implementation of QPACK compression for HTTP/3. |
| **go-tpm** | 0.9.0 | PENDING | 0 B | blake3:01b953029756b676967940aef905a76c533e049100691f57cf5f8a9858677b62 | Go library for Trusted Platform Module (TPM) interactions. |
| **go-spiffe** | 2.0.0 | PENDING | 0 B | blake3:74a596b408aa8427ca6b9a7228f02b8a4966e9c9d6cc96b89c3bdb22c8fb7d7c | Go library for SPIFFE standard identity and trust circles. |
| **memguard** | 0.22.0 | PENDING | 0 B | blake3:7b7819a2cfeeef30041761a7a9a1e5dc6052e3862a803539dfec7c9c2a78c453 | Go library for securing sensitive data in memory. |
| **wazero** | 1.7.0 | PENDING | 0 B | blake3:59c5f2d87255f0524c3f8fdb879b3d35fcdfb02963ddf8aefb0bb7afcde724c2 | Go zero-dependency WebAssembly runtime. |
| **blake3** | 0.2.4 | PENDING | 0 B | blake3:2cd948e40cf3ca7f6fe86d5980e10fe9d353ea5f4196e4076bc6d1b4c3bf4694 | Go binding/implementation of the Blake3 hashing protocol. |

## Internal Workspace Targets
| Workspace Name | Harness Location | Build Status | Targets |
| :--- | :--- | :--- | :--- |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-a2a/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-actors/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 20 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-adk/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 2 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-animus/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-assurance/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-fab-aides/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 2 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-hydration/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 2 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-introspection/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 12 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-latentlingua/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 5 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-logiclibrary/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-mcp/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 3 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-natives/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **s-natives** | [s-natives](file:///C:/aCogSpaceSeed/00flow/s-natives/workspace.harness) | DIRTY (Rebuild Needed) | 0 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-sacp/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-seed/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **71000-build-harness** | [71000-build-harness](file:///C:/aCogSpaceSeed/00flow/s-trust-circle/71000-build-harness/workspace.harness) | DIRTY (Rebuild Needed) | 1 targets |
| **s-trust-circle** | [s-trust-circle](file:///C:/aCogSpaceSeed/00flow/s-trust-circle/workspace.harness) | UP-TO-DATE | 2 targets |
