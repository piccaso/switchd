package tests

import (
	"testing"

	"github.com/piccaso/switchd/lib"
)

func TestSyncTimestamp(t *testing.T) {
	ts := lib.NewSyncTimestamp()
	ts.Substract(200) // set 200 seconds in the past
	els := ts.ElapsedSeconds()
	t.Logf("elapsed: %v", els)
	if els > 201 || els < 199 {
		t.Fail()
	}
}
