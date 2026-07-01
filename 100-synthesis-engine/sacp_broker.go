package synthesis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// SACPBroker coordinates SACP raw QUIC/UDP connections and Monotonic Authority validation.
type SACPBroker struct {
	mu               sync.RWMutex
	udpAddr          string
	wsAddr           string
	activeTokens     map[string]AuthorityLevel
	activeSession    *QAPCHeader
	listener         *net.UDPConn
	shutdownCtx      context.Context
	shutdownFn       context.CancelFunc
	idleTimeout      time.Duration
	watchdogTimer    *time.Timer
	watchdogMu       sync.Mutex
	activeStreams    int32
	shutdownCallback func()
}

func NewSACPBroker(udpAddr string, wsAddr string) *SACPBroker {
	ctx, cancel := context.WithCancel(context.Background())
	return &SACPBroker{
		udpAddr:      udpAddr,
		wsAddr:       wsAddr,
		activeTokens: make(map[string]AuthorityLevel),
		shutdownCtx:  ctx,
		shutdownFn:   cancel,
		idleTimeout:  30 * time.Minute, // 30 minutes default
	}
}

// RegisterTransitionToken registers a single-use metabolic hot-swap token on the broker.
func (b *SACPBroker) RegisterTransitionToken(token string, auth AuthorityLevel) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activeTokens[token] = auth
	slog.Info(fmt.Sprintf("[SACPBroker] Registered Metabolic transition token %q with level %s", token, auth))
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
	slog.Info(fmt.Sprintf("[SACPBroker] raw QUIC/UDP Broker listening on %s", b.udpAddr))

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
				slog.Info(fmt.Sprintf("[SACPBroker] Read error: %v", err))
				continue
			}

			// Parse incoming frame using the Symmetric Callee Proxy
			callee := NewSACPCalleeProxy(buf[:n])
			if callee.GetOriginalUUID() == "" {
				slog.Info(fmt.Sprintf("[SACPBroker] Ignoring malformed frame of size %d", n))
				continue
			}

			header := &QAPCHeader{
				OriginalUUID:      callee.GetOriginalUUID(),
				OriginalAuthority: callee.GetOriginalAuthority(),
				CurrentAuthority:  callee.GetCurrentAuthority(),
			}

			// Validate Monotonic Authority Invariant
			err = b.ValidateMonotonicInvariant(header)
			if err != nil {
				slog.Info(fmt.Sprintf("[SACPBroker] [BLOCK] Packet rejected: %v", err))
				continue
			}

			b.mu.Lock()
			b.activeSession = header
			b.mu.Unlock()

			b.ResetIdleWatchdog() // Reset on SACP frame activity

			slog.Info(fmt.Sprintf("[SACPBroker] [PASS] Verified SACP Frame for %s with authority %s", header.OriginalUUID, header.CurrentAuthority))
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

	b.ResetIdleWatchdog() // Reset on metabolic hot-swap task goal arrival

	slog.Info(fmt.Sprintf("[SACPBroker] Metabolic Hot-Swap success! Session %q elevated to QUIC SACP with level %s", clientUUID, auth))
	return nil
}

func (b *SACPBroker) SetIdleTimeout(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.idleTimeout = d
}

func (b *SACPBroker) StartIdleWatchdog(onShutdown func()) {
	b.watchdogMu.Lock()
	defer b.watchdogMu.Unlock()

	b.mu.Lock()
	b.shutdownCallback = onShutdown
	timeout := b.idleTimeout
	b.mu.Unlock()

	if b.watchdogTimer != nil {
		b.watchdogTimer.Stop()
	}

	b.watchdogTimer = time.AfterFunc(timeout, func() {
		b.mu.Lock()
		cb := b.shutdownCallback
		b.mu.Unlock()
		if cb != nil {
			cb()
		}
	})
}

func (b *SACPBroker) ResetIdleWatchdog() {
	b.watchdogMu.Lock()
	defer b.watchdogMu.Unlock()

	b.mu.Lock()
	timeout := b.idleTimeout
	b.mu.Unlock()

	if b.watchdogTimer != nil {
		b.watchdogTimer.Stop()
		b.watchdogTimer.Reset(timeout)
	}
}

func (b *SACPBroker) OpenInvocationStream() int32 {
	b.ResetIdleWatchdog()
	return atomic.AddInt32(&b.activeStreams, 1)
}

func (b *SACPBroker) CloseInvocationStream() int32 {
	b.ResetIdleWatchdog()
	return atomic.AddInt32(&b.activeStreams, -1)
}

func (b *SACPBroker) GetActiveStreamsCount() int32 {
	return atomic.LoadInt32(&b.activeStreams)
}

func (b *SACPBroker) Close() {
	b.shutdownFn()
	b.watchdogMu.Lock()
	if b.watchdogTimer != nil {
		b.watchdogTimer.Stop()
	}
	b.watchdogMu.Unlock()

	if b.listener != nil {
		b.listener.Close()
	}
}
