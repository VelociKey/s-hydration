module sov.fleet/s-hydration

go 1.26.4

require (
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	sov.fleet/quic-go v0.0.0 // indirect
	sov.fleet/s-parallizer v0.0.0 // indirect
	sov.fleet/s-scorecard v0.0.0-00010101000000-000000000000 // indirect
)

require (
	golang.org/x/mod v0.36.0
	golang.org/x/sync v0.21.0
	gopkg.in/yaml.v3 v3.0.1
	sov.fleet/blake3 v0.0.0
	sov.fleet/quicdl v0.0.0
	sov.fleet/s-agentbox v0.0.0
	sov.fleet/s-latentlingua v0.0.0
	sov.fleet/s-logiclibrary v0.0.0
	sov.fleet/s-qdag v0.0.0
)

replace (
	c0100-test-cli => ../../99cogt/c0100-test-cli
	c0200-test-library => ../../99cogt/c0200-test-library
	c0300-test-ui-backend/backend => ../../99cogt/c0300-test-ui-backend/backend
	github.com/modelcontextprotocol/go-sdk => ../s-agentframework/83000-external-libraries-source/go-sdk
	sov.fleet/00flow/00flow-info => ../00flow-info
	sov.fleet/00xper/00xper-info => ../../00xper/00xper-info
	sov.fleet/99cbox-root => ../../99cbox/99cbox-root
	sov.fleet/99cbox/99cbox-info => ../../99cbox/99cbox-info
	sov.fleet/99cogt/99cogt-info => ../../99cogt/99cogt-info
	sov.fleet/blake3 => ../s-forge/93000-external-libraries/blake3
	sov.fleet/n-root => ../../00flon/n-root
	sov.fleet/o-afflume => ../../00floo/o-afflume
	sov.fleet/o-afflume-realization => ../../00floo/o-afflume-realization
	sov.fleet/o-bankanchor => ../../00floo/o-bankanchor
	sov.fleet/o-banking => ../../00floo/o-banking
	sov.fleet/o-bankmachine => ../../00floo/o-bankmachine
	sov.fleet/o-billing => ../../00floo/o-billing
	sov.fleet/o-ingestion => ../../00floo/o-ingestion
	sov.fleet/o-invoicing => ../../00floo/o-invoicing
	sov.fleet/o-ledger => ../../00floo/o-ledger
	sov.fleet/o-offering => ../../00floo/o-offering
	sov.fleet/o-paymentprocessing => ../../00floo/o-paymentprocessing
	sov.fleet/o-taxation => ../../00floo/o-taxation
	sov.fleet/o-telemetry => ../../00floo/o-telemetry
	sov.fleet/qpack => ../s-forge/93000-external-libraries/qpack
	sov.fleet/quic-go => ../s-forge/93000-external-libraries/quic-go
	sov.fleet/quicdl => ../s-forge/98000-internal-libraries/quicdl
	sov.fleet/s-a2a => ../s-a2a
	sov.fleet/s-actors => ../s-actors
	sov.fleet/s-adk => ../s-adk
	sov.fleet/s-agentbox => ../s-agentbox
	sov.fleet/s-agentframework => ../s-agentframework
	sov.fleet/s-animus => ../s-animus
	sov.fleet/s-assurance => ../s-assurance
	sov.fleet/s-authorize => ../s-authorize
	sov.fleet/s-cognition => ../../000all/s-cognition
	sov.fleet/s-distribution => ../s-distribution
	sov.fleet/s-fab-aides => ../s-fab-aides
	sov.fleet/s-forge => ../s-forge
	sov.fleet/s-githubmcp => ../s-githubmcp
	sov.fleet/s-hardware-bridge => ../s-hardware-bridge
	sov.fleet/s-hydrationcache => ../s-hydrationcache
	sov.fleet/s-ingestion => ../s-ingestion
	sov.fleet/s-introspection => ../s-introspection
	sov.fleet/s-latentlingua => ../s-latentlingua
	sov.fleet/s-logiclibrary => ../s-logiclibrary
	sov.fleet/s-mcp => ../s-mcp
	sov.fleet/s-mcp-githubtrimmed => ../s-mcp-githubtrimmed
	sov.fleet/s-mcp-greeter => ../s-mcp-greeter
	sov.fleet/s-mcpstudio => ../s-mcpstudio
	sov.fleet/s-natives => ../s-natives
	sov.fleet/s-parallizer => ../s-parallizer
	sov.fleet/s-qdag => ../s-qdag
	sov.fleet/s-riskpapers => ../../51slam/s-riskpapers
	sov.fleet/s-sacp => ../s-sacp
	sov.fleet/s-scoreboard => ../s-scoreboard
	sov.fleet/s-scorecard => ../s-scorecard
	sov.fleet/s-seed => ../s-seed
	sov.fleet/s-taxonomy-guard => ../s-taxonomy-guard
	sov.fleet/s-trust-circle => ../s-trust-circle
	sov.fleet/s-webconduit => ../s-webconduit
	sov.fleet/stripe-go => ../s-forge/93000-external-libraries/stripe-go
	sov.fleet/x-actors => ../../00xper/x-actors
	sov.fleet/x-actorstudio => ../../00xper/x-actorstudio
	sov.nvelwraith/wraithclient => ../../.nvelwraith/src/sov.nvelwraith/wraithclient
	x-transform-antigravity => ../../00xper/x-transform-antigravity
)
