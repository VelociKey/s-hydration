package synthesis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// SACPBroker coordinates SACP raw QUIC/UDP connections and Monotonic Authority validation.
type SACPBroker struct {
	mu            sync.RWMutex
	udpAddr       string
	wsAddr        string
	activeTokens  map[string]AuthorityLevel
	activeSession *QAPCHeader
	listener      *net.UDPConn
	shutdownCtx   context.Context
	shutdownFn    context.CancelFunc
}

func NewSACPBroker(udpAddr string, wsAddr string) *SACPBroker {
	ctx, cancel := context.WithCancel(context.Background())
	return &SACPBroker{
		udpAddr:      udpAddr,
		wsAddr:       wsAddr,
		activeTokens: make(map[string]AuthorityLevel),
		shutdownCtx:  ctx,
		shutdownFn:   cancel,
	}
}

// RegisterTransitionToken registers a single-use metabolic hot-swap token on the broker.
func (b *SACPBroker) RegisterTransitionToken(token string, auth AuthorityLevel) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activeTokens[token] = auth
	log.Printf("[SACPBroker] Registered Metabolic transition token %q with level %s", token, auth)
}

// ValidateMonotonicInvariant asserts that delegates do not escalate authority.
func (b *SACPBroker) ValidateMonotonicInvariant(header *QAPCHeader) error {
	if header.CurrentAuthority > header.OriginalAuthority {
		return fmt.Errorf("MONOTONIC VIOLATION: CurrentAuthority %s exceeds OriginalAuthority %s", header.CurrentAuthority, header.OriginalAuthority)
	}
	return nil
}

// StartUDPServer launches the raw UDP QUIC receiver listener.
func (b *SACPBroker) StartUDPServer() error {
	addr, err := net.ResolveUDPAddr("udp", b.udpAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	b.listener = conn
	log.Printf("[SACPBroker] raw QUIC/UDP Broker listening on %s", b.udpAddr)

	go b.listenLoop()
	return nil
}

func (b *SACPBroker) listenLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-b.shutdownCtx.Done():
			return
		default:
			n, _, err := b.listener.ReadFromUDP(buf)
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Printf("[SACPBroker] Read error: %v", err)
				continue
			}

			if n < PayloadOffset {
				log.Printf("[SACPBroker] Ignoring malformed frame of size %d", n)
				continue
			}

			// Parse incoming frame using the Symmetric Callee Proxy
			callee := NewSACPCalleeProxy(buf[:n])
			header := &QAPCHeader{
				OriginalUUID:      callee.GetOriginalUUID(),
				OriginalAuthority: callee.GetOriginalAuthority(),
				CurrentAuthority:  callee.GetCurrentAuthority(),
			}

			// Validate Monotonic Authority Invariant
			err = b.ValidateMonotonicInvariant(header)
			if err != nil {
				log.Printf("[SACPBroker] [BLOCK] Packet rejected: %v", err)
				continue
			}

			b.mu.Lock()
			b.activeSession = header
			b.mu.Unlock()

			log.Printf("[SACPBroker] [PASS] Verified SACP Frame for %s with authority %s", header.OriginalUUID, header.CurrentAuthority)
		}
	}
}

// ProcessMetabolicSwap simulates the live session hot-swap transition.
func (b *SACPBroker) ProcessMetabolicSwap(token string, clientUUID string) error {
	b.mu.Lock()
	auth, ok := b.activeTokens[token]
	if !ok {
		b.mu.Unlock()
		return fmt.Errorf("swap failed: invalid transition token %q", token)
	}
	delete(b.activeTokens, token) // single-use token consumption
	b.mu.Unlock()

	// Establish new active qAPC session header
	header := NewQAPCHeader(clientUUID, auth)
	
	b.mu.Lock()
	b.activeSession = header
	b.mu.Unlock()

	log.Printf("[SACPBroker] Metabolic Hot-Swap success! Session %q elevated to QUIC SACP with level %s", clientUUID, auth)
	return nil
}

func (b *SACPBroker) Close() {
	b.shutdownFn()
	if b.listener != nil {
		b.listener.Close()
	}
}
