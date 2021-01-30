package query

import (
	"testing"
)

func checkPanic(t *testing.T, msg string, f func()) {
	t.Helper()
	defer func() {
		if e := recover(); e == nil {
			t.Error("no panic")
		} else if e.(string) != msg {
			t.Errorf("got panic %q, want %q", e, msg)
		}
	}()
	f()
}
