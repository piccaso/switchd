package tests

import (
	"testing"

	"github.com/piccaso/switchd/lib"
)

func TestSyncString(t *testing.T) {
	s := lib.NewSyncString("lala")
	if "lala" != s.Get() {
		t.Fail()
	}

	if !s.Equals("lala") {
		t.Fail()
	}
}
