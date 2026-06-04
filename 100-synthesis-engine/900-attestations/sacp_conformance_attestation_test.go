package synthesis_attestations

import (
	. "sov.fleet/s-hydration/100-synthesis-engine"
"net"
	"sync"
	"testing"
	"time"
)

func TestMonotonicAuthority(t *testing.T) {
	header := NewQAPCHeader("guest-uuid-123", AuthWorker)

	// 1. Assert initial state
	if header.OriginalAuthority != AuthWorker || header.CurrentAuthority != AuthWorker {
		t.Errorf("Expected AuthWorker, got orig: %v, curr: %v", header.OriginalAuthority, header.CurrentAuthority)
	}

	// 2. Validate safe delegation decay (Worker -> Guest)
	delegated, err := header.Delegate(AuthGuest)
	if err != nil {
		t.Fatalf("Safe delegation failed: %v", err)
	}
	if delegated.CurrentAuthority != AuthGuest {
		t.Errorf("Expected current authority to be AuthGuest, got %s", delegated.CurrentAuthority)
	}

	// 3. Assert monotonic security invariant: escalation must fail (Worker -> Manager)
	_, err = header.Delegate(AuthManager)
	if err == nil {
		t.Error("Monotonic invariant check failed: authority escalation (Worker -> Manager) was permitted!")
	}

	// 4. Assert broker-side monotonic validation check
	broker := NewSACPBroker("127.0.0.1:9099", "127.0.0.1:8080")
	defer broker.Close()

	validHeader := &QAPCHeader{
		OriginalUUID:      "guest-uuid-123",
		OriginalAuthority: AuthManager,
		CurrentAuthority:  AuthWorker,
	}
	if err := broker.ValidateMonotonicInvariant(validHeader); err != nil {
		t.Errorf("Expected valid Monotonic Header to pass, got error: %v", err)
	}

	invalidHeader := &QAPCHeader{
		OriginalUUID:      "guest-uuid-123",
		OriginalAuthority: AuthWorker,
		CurrentAuthority:  AuthManager,
	}
	if err := broker.ValidateMonotonicInvariant(invalidHeader); err == nil {
		t.Error("Expected Monotonic Validation to reject escalation, but it passed!")
	}
}

func TestMetabolicHotSwap(t *testing.T) {
	broker := NewSACPBroker("127.0.0.1:9099", "127.0.0.1:8080")
	defer broker.Close()

	token := "metabolic-swap-token-abc"
	broker.RegisterTransitionToken(token, AuthWorker)

	// 1. First swap attempt with invalid token should fail
	err := broker.ProcessMetabolicSwap("invalid-token", "guest-uuid")
	if err == nil {
		t.Error("Expected swap with invalid token to fail, but it succeeded")
	}

	// 2. Safe swap attempt with valid token should succeed
	err = broker.ProcessMetabolicSwap(token, "guest-uuid-999")
	if err != nil {
		t.Fatalf("Metabolic Hot-Swap failed: %v", err)
	}

	if broker.GetActiveSession() == nil {
		t.Fatal("Active session is nil after successful swap")
	}
	if broker.GetActiveSession().OriginalUUID != "guest-uuid-999" || broker.GetActiveSession().CurrentAuthority != AuthWorker {
		t.Errorf("Session mismatch: %+v", broker.GetActiveSession())
	}

	// 3. Single-use enforcement check: second swap attempt with the same token must fail
	err = broker.ProcessMetabolicSwap(token, "guest-uuid-999")
	if err == nil {
		t.Error("Metabolic token single-use check failed: token was successfully reused!")
	}
}

func TestZeroCopySymmetricProxy(t *testing.T) {
	// 1. Initialize Caller Proxy
	caller := NewSACPCallerProxy(256)
	caller.SetOriginalUUID("guest-uuid-007")
	caller.SetOriginalAuthority(AuthManager)
	caller.SetCurrentAuthority(AuthWorker)
	caller.SetPayload("test-conformance-payload")

	// 2. Perform zero-copy deserialization using Callee Proxy over the same raw bytes
	callee := NewSACPCalleeProxy(caller.Buffer())

	if callee.GetOriginalUUID() != "guest-uuid-007" {
		t.Errorf("UUID mismatch: expected 'guest-uuid-007', got %q", callee.GetOriginalUUID())
	}
	if callee.GetOriginalAuthority() != AuthManager {
		t.Errorf("OrigAuth mismatch: expected AuthManager, got %s", callee.GetOriginalAuthority())
	}
	if callee.GetCurrentAuthority() != AuthWorker {
		t.Errorf("CurrAuth mismatch: expected AuthWorker, got %s", callee.GetCurrentAuthority())
	}
	if callee.GetPayload() != "test-conformance-payload" {
		t.Errorf("Payload mismatch: expected 'test-conformance-payload', got %q", callee.GetPayload())
	}
}

func TestBrokerUDPSessionTraffic(t *testing.T) {
	broker := NewSACPBroker("127.0.0.1:9099", "127.0.0.1:8080")
	err := broker.StartUDPServer()
	if err != nil {
		t.Fatalf("Failed to start UDP broker: %v", err)
	}
	defer broker.Close()

	// 1. Build and serialize SACP frame
	caller := NewSACPCallerProxy(256)
	caller.SetOriginalUUID("udp-guest-009")
	caller.SetOriginalAuthority(AuthSovereign)
	caller.SetCurrentAuthority(AuthManager)
	caller.SetPayload("live-udp-traffic-test")

	// 2. Dispatch SACP frame over local UDP loopback
	conn, err := net.Dial("udp", "127.0.0.1:9099")
	if err != nil {
		t.Fatalf("Failed to dial UDP server: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(caller.Buffer())
	if err != nil {
		t.Fatalf("Failed to write to UDP socket: %v", err)
	}

	// 3. Allow brief window for loopback processing
	time.Sleep(50 * time.Millisecond)

	broker.RLockMu()
	session := broker.GetActiveSession()
	broker.RUnlockMu()

	if session == nil {
		t.Fatal("UDP Broker failed to process incoming UDP frame: activeSession is nil")
	}

	if session.OriginalUUID != "udp-guest-009" || session.CurrentAuthority != AuthManager {
		t.Errorf("UDP Session mismatch: %+v", session)
	}
}

func TestIdleShutdownWatchdog(t *testing.T) {
	broker := NewSACPBroker("127.0.0.1:9099", "127.0.0.1:8080")
	defer broker.Close()

	// 1. Configure the test idle watchdog timeout parameter to 100 milliseconds
	broker.SetIdleTimeout(100 * time.Millisecond)

	shutdownTriggered := false
	var triggerMu sync.Mutex

	broker.StartIdleWatchdog(func() {
		triggerMu.Lock()
		shutdownTriggered = true
		triggerMu.Unlock()
	})

	// 2. Wait 50ms (less than timeout) - watchdog should not fire yet
	time.Sleep(50 * time.Millisecond)
	triggerMu.Lock()
	if shutdownTriggered {
		triggerMu.Unlock()
		t.Fatal("Watchdog fired prematurely before timeout duration elapsed!")
	}
	triggerMu.Unlock()

	// 3. Reset watchdog manually to simulate task arrival activity
	broker.ResetIdleWatchdog()

	// 4. Wait another 70ms - total elapsed is 120ms, but since reset occurred at 50ms, it should still not fire
	time.Sleep(70 * time.Millisecond)
	triggerMu.Lock()
	if shutdownTriggered {
		triggerMu.Unlock()
		t.Fatal("Watchdog fired prematurely despite active ResetIdleWatchdog call!")
	}
	triggerMu.Unlock()

	// 5. Wait 120ms (more than remaining watchdog reset duration) - it must fire now
	time.Sleep(120 * time.Millisecond)
	triggerMu.Lock()
	if !shutdownTriggered {
		triggerMu.Unlock()
		t.Fatal("Watchdog failed to fire after idle timeout period elapsed!")
	}
	triggerMu.Unlock()
}

func TestStreamSwapLifecycle(t *testing.T) {
	broker := NewSACPBroker("127.0.0.1:9099", "127.0.0.1:8080")
	defer broker.Close()

	// 1. Initial stream count should be 0
	if broker.GetActiveStreamsCount() != 0 {
		t.Errorf("Expected 0 active streams, got %d", broker.GetActiveStreamsCount())
	}

	// 2. Open Stream (Task N starts)
	broker.OpenInvocationStream()
	if broker.GetActiveStreamsCount() != 1 {
		t.Errorf("Expected 1 active stream, got %d", broker.GetActiveStreamsCount())
	}

	// 3. Open Second Stream (Parallel Task N+1 starts)
	broker.OpenInvocationStream()
	if broker.GetActiveStreamsCount() != 2 {
		t.Errorf("Expected 2 active streams, got %d", broker.GetActiveStreamsCount())
	}

	// 4. Close First Stream (Task N finishes - Stream closed cleanly via FIN/EOF)
	broker.CloseInvocationStream()
	if broker.GetActiveStreamsCount() != 1 {
		t.Errorf("Expected 1 active stream left, got %d", broker.GetActiveStreamsCount())
	}

	// 5. Close Second Stream (Task N+1 finishes)
	broker.CloseInvocationStream()
	if broker.GetActiveStreamsCount() != 0 {
		t.Errorf("Expected 0 active streams left, got %d", broker.GetActiveStreamsCount())
	}
}

