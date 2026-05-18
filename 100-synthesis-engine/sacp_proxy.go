package synthesis

import (
	"bytes"
	"fmt"
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

// Symmetric Memory Topography Constants
const (
	SlotSize       = 64
	UUIDOffset     = 0
	OrigAuthOffset = SlotSize
	CurrAuthOffset = SlotSize * 2
	PayloadOffset  = SlotSize * 3
)

// SACPCallerProxy is the zero-copy caller serialization helper.
type SACPCallerProxy struct {
	buffer []byte
}

func NewSACPCallerProxy(size int) *SACPCallerProxy {
	if size < PayloadOffset+SlotSize {
		size = PayloadOffset + SlotSize
	}
	return &SACPCallerProxy{buffer: make([]byte, size)}
}

func (p *SACPCallerProxy) Buffer() []byte {
	return p.buffer
}

func (p *SACPCallerProxy) SetOriginalUUID(uuid string) {
	b := []byte(uuid)
	copy(p.buffer[UUIDOffset:UUIDOffset+SlotSize], bytes.Repeat([]byte{0}, SlotSize))
	copy(p.buffer[UUIDOffset:UUIDOffset+SlotSize], b)
}

func (p *SACPCallerProxy) GetOriginalUUID() string {
	return string(bytes.Trim(p.buffer[UUIDOffset:UUIDOffset+SlotSize], "\x00"))
}

func (p *SACPCallerProxy) SetOriginalAuthority(auth AuthorityLevel) {
	p.buffer[OrigAuthOffset] = byte(auth)
}

func (p *SACPCallerProxy) GetOriginalAuthority() AuthorityLevel {
	return AuthorityLevel(p.buffer[OrigAuthOffset])
}

func (p *SACPCallerProxy) SetCurrentAuthority(auth AuthorityLevel) {
	p.buffer[CurrAuthOffset] = byte(auth)
}

func (p *SACPCallerProxy) GetCurrentAuthority() AuthorityLevel {
	return AuthorityLevel(p.buffer[CurrAuthOffset])
}

func (p *SACPCallerProxy) SetPayload(payload string) {
	b := []byte(payload)
	limit := len(p.buffer) - PayloadOffset
	copy(p.buffer[PayloadOffset:], bytes.Repeat([]byte{0}, limit))
	copy(p.buffer[PayloadOffset:], b)
}

func (p *SACPCallerProxy) GetPayload() string {
	return string(bytes.Trim(p.buffer[PayloadOffset:], "\x00"))
}

// SACPCalleeProxy is the zero-copy callee deserialization helper.
type SACPCalleeProxy struct {
	buffer []byte
}

func NewSACPCalleeProxy(buf []byte) *SACPCalleeProxy {
	return &SACPCalleeProxy{buffer: buf}
}

func (p *SACPCalleeProxy) GetOriginalUUID() string {
	return string(bytes.Trim(p.buffer[UUIDOffset:UUIDOffset+SlotSize], "\x00"))
}

func (p *SACPCalleeProxy) GetOriginalAuthority() AuthorityLevel {
	return AuthorityLevel(p.buffer[OrigAuthOffset])
}

func (p *SACPCalleeProxy) GetCurrentAuthority() AuthorityLevel {
	return AuthorityLevel(p.buffer[CurrAuthOffset])
}

func (p *SACPCalleeProxy) GetPayload() string {
	return string(bytes.Trim(p.buffer[PayloadOffset:], "\x00"))
}
