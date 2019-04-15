package lib

import "sync"

// NewSyncString ...
func NewSyncString(value string) *SyncString {
	var p SyncString
	p.Set(value)
	return &p
}

// SyncString ...
type SyncString struct {
	lock  sync.RWMutex
	value string
}

// Get ...
func (t *SyncString) Get() string {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.value
}

// Set ...
func (t *SyncString) Set(value string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.value = value
}

// Equals ...
func (t *SyncString) Equals(value string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.value == value
}
