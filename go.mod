module sov.fleet/s-hydration

go 1.26.3

require (
	sov.fleet/blake3 v0.2.4
	sov.fleet/quic-go v0.59.1
	sov.fleet/quicdl v0.0.0
	sov.fleet/s-agentbox v0.0.0
)

replace (
	sov.fleet/blake3 => ../s-forge/93000-external-libraries/blake3
	sov.fleet/qpack => ../s-forge/93000-external-libraries/qpack
	sov.fleet/quic-go => ../s-forge/93000-external-libraries/quic-go
	sov.fleet/quicdl => ../s-forge/98000-internal-libraries/quicdl
	sov.fleet/s-agentbox => ../s-agentbox
)

require (
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	sov.fleet/qpack v0.6.0 // indirect
)
