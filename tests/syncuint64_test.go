package tests

import (
	"testing"

	"github.com/piccaso/switchd/lib"
)

func TestSyncuint64(t *testing.T) {
	i := lib.NewSyncUint64(1)
	if i.Get() != uint64(1) {
		t.Fail()
	}
}
