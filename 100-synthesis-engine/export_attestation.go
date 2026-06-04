//go:build !production

package synthesis

func (b *SACPBroker) GetActiveSession() *QAPCHeader {
	return b.activeSession
}

func (b *SACPBroker) LockMu() {
	b.mu.Lock()
}

func (b *SACPBroker) UnlockMu() {
	b.mu.Unlock()
}

func (b *SACPBroker) RLockMu() {
	b.mu.RLock()
}

func (b *SACPBroker) RUnlockMu() {
	b.mu.RUnlock()
}
