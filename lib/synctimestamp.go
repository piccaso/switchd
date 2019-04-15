package lib

import "time"

func timeNow() uint64 {
	return uint64(time.Now().Unix())
}

// NewSyncTimestamp form 'now'
func NewSyncTimestamp() *SyncTimestamp {
	var t SyncTimestamp
	t.Reset()
	return &t
}

// SyncTimestamp Synchronized Unix Timestamp
type SyncTimestamp struct {
	SyncUint64
}

// Reset to 'now'
func (t *SyncTimestamp) Reset() {
	t.Set(timeNow())
}

// GetTime (precision = seconds)
func (t *SyncTimestamp) GetTime() time.Time {
	return time.Unix(int64(t.Get()), 0)
}

// ElapsedSeconds ...
func (t *SyncTimestamp) ElapsedSeconds() uint64 {
	var e = time.Since(t.GetTime())
	return uint64(e.Seconds())
}
