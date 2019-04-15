package tests

import (
	"testing"

	"github.com/piccaso/switchd/lib"
)

func TestSyncPower(t *testing.T) {
	p := lib.NewSyncPower("on")
	
	if p.IsOff() {
		t.Fail()
	}
	p.Off()
	if p.IsOn() {
		t.Fail()
	}
}
