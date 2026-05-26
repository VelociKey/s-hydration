package synthesis

import (
	"fmt"

	"sov.fleet/s-logiclibrary/00200-logic-libraries/bicodec"
)

// AuthorityLevel coordinates monotonic permissions for agent capabilities.
type AuthorityLevel int

const (
	AuthGuest AuthorityLevel = iota
	AuthWorker
	AuthManager
	AuthSovereign
)

func (a AuthorityLevel) String() string {
	switch a {
	case AuthGuest:
		return "AuthGuest"
	case AuthWorker:
		return "AuthWorker"
	case AuthManager:
		return "AuthManager"
	case AuthSovereign:
		return "AuthSovereign"
	default:
		return "Unknown"
	}
}

func mapAuthLevelToString(auth AuthorityLevel) string {
	switch auth {
	case AuthGuest:
		return bicodec.AuthGuest
	case AuthWorker:
		return bicodec.AuthWorker
	case AuthManager:
		return bicodec.AuthManager
	case AuthSovereign:
		return bicodec.AuthSovereign
	default:
		return bicodec.AuthGuest
	}
}

func mapStringToAuthLevel(s string) AuthorityLevel {
	switch s {
	case bicodec.AuthGuest:
		return AuthGuest
	case bicodec.AuthWorker:
		return AuthWorker
	case bicodec.AuthManager:
		return AuthManager
	case bicodec.AuthSovereign:
		return AuthSovereign
	default:
		return AuthGuest
	}
}

// QAPCHeader coordinates monotonic delegation and TPM/Veracity attestations.
type QAPCHeader struct {
	OriginalUUID      string
	OriginalAuthority AuthorityLevel
	CurrentAuthority  AuthorityLevel
}

func NewQAPCHeader(uuid string, auth AuthorityLevel) *QAPCHeader {
	return &QAPCHeader{
		OriginalUUID:      uuid,
		OriginalAuthority: auth,
		CurrentAuthority:  auth,
	}
}

// Delegate checks the Monotonic Invariant: reducedAuth <= CurrentAuthority
func (h *QAPCHeader) Delegate(reducedAuth AuthorityLevel) (*QAPCHeader, error) {
	if reducedAuth > h.CurrentAuthority {
		return nil, fmt.Errorf("authority escalation denied: cannot delegate %s from %s", reducedAuth, h.CurrentAuthority)
	}
	return &QAPCHeader{
		OriginalUUID:      h.OriginalUUID,
		OriginalAuthority: h.OriginalAuthority,
		CurrentAuthority:  reducedAuth,
	}, nil
}

// Symmetric Memory Topography Constants (kept for compatibility)
const (
	SlotSize       = 64
	UUIDOffset     = 0
	OrigAuthOffset = SlotSize
	CurrAuthOffset = SlotSize * 2
	PayloadOffset  = SlotSize * 3
)

// SACPProxyBase embeds the shared header structure and provides common getters.
type SACPProxyBase struct {
	header bicodec.SACPHeader
}

func (p *SACPProxyBase) GetOriginalUUID() string {
	return p.header.UUID
}

func (p *SACPProxyBase) GetOriginalAuthority() AuthorityLevel {
	return mapStringToAuthLevel(p.header.OriginalAuthority)
}

func (p *SACPProxyBase) GetCurrentAuthority() AuthorityLevel {
	return mapStringToAuthLevel(p.header.CurrentAuthority)
}

// SACPCallerProxy is the caller serialization helper.
type SACPCallerProxy struct {
	SACPProxyBase
	payload string
}

func NewSACPCallerProxy(size int) *SACPCallerProxy {
	return &SACPCallerProxy{}
}

func (p *SACPCallerProxy) Buffer() []byte {
	c := &bicodec.SACPCapability{
		Domain: "s-hydration",
		Action: "call",
		Parameters: map[string]string{
			"payload": p.payload,
		},
	}
	buf, err := bicodec.EncodeSACPMessage(&p.header, c)
	if err != nil {
		panic(err)
	}
	return buf
}

func (p *SACPCallerProxy) SetOriginalUUID(uuid string) {
	p.header.UUID = uuid
}

func (p *SACPCallerProxy) SetOriginalAuthority(auth AuthorityLevel) {
	p.header.OriginalAuthority = mapAuthLevelToString(auth)
}

func (p *SACPCallerProxy) SetCurrentAuthority(auth AuthorityLevel) {
	p.header.CurrentAuthority = mapAuthLevelToString(auth)
}

func (p *SACPCallerProxy) SetPayload(payload string) {
	p.payload = payload
}

func (p *SACPCallerProxy) GetPayload() string {
	return p.payload
}

// SACPCalleeProxy is the callee deserialization helper.
type SACPCalleeProxy struct {
	SACPProxyBase
	cap    bicodec.SACPCapability
}

func NewSACPCalleeProxy(buf []byte) *SACPCalleeProxy {
	hasHeader, h, hasCap, c, err := bicodec.DecodeSACPMessage(buf)
	if err != nil {
		return &SACPCalleeProxy{}
	}
	_ = hasHeader
	_ = hasCap
	return &SACPCalleeProxy{
		SACPProxyBase: SACPProxyBase{header: h},
		cap:           c,
	}
}

func (p *SACPCalleeProxy) GetPayload() string {
	if p.cap.Parameters == nil {
		return ""
	}
	return p.cap.Parameters["payload"]
}
