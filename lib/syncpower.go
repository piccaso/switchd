package lib

// NewSyncPower ...
func NewSyncPower(power string) *SyncPower {
	var p SyncPower
	p.Set(power)
	return &p
}

// SyncPower ...
type SyncPower struct {
	SyncString
}

// On ...
func (t *SyncPower) On() {
	t.Set("on")
}

// Off ...
func (t *SyncPower) Off() {
	t.Set("off")
}

// IsOn ...
func (t *SyncPower) IsOn() bool {
	return t.Get() == "on"
}

// IsOff ...
func (t *SyncPower) IsOff() bool {
	return !t.IsOn()
}
