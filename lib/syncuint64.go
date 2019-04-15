package lib

import "sync"

// SyncUint64 ...
type SyncUint64 struct {
	lock  sync.RWMutex
	value uint64
}

// NewSyncUint64 ...
func NewSyncUint64(value uint64) *SyncUint64 {
	var s SyncUint64
	s.Set(value)
	return &s
}

// Get value
func (t *SyncUint64) Get() uint64 {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.value
}

// Set value
func (t *SyncUint64) Set(value uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.value = value
}

// Add value
func (t *SyncUint64) Add(value uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.value += value
}

// Substract value
func (t *SyncUint64) Substract(value uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.value -= value
}
