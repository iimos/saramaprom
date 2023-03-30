package saramaprom

import (
	"testing"
	"time"
)

func TestStartScheduler(t *testing.T) {
	t.Run("Execute job", func(t *testing.T) {
		interval := time.Millisecond * 5
		executed := make(chan struct{})
		s := StartScheduler(interval, func() { executed <- struct{}{} })

		select {
		case <-executed:
		case <-time.After(interval * 2):
			t.Error("Expected job to be executed but was not.")
		}
		s.Stop()
	})
}
