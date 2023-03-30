package saramaprom

import (
	"time"
)

// Scheduler is used to run jobs in specific intervals.
type Scheduler struct {
	stop chan<- struct{}
}

// StartScheduler starts goroutine that will run given job in given intervals until Stop() is called.
func StartScheduler(interval time.Duration, job func()) Scheduler {
	ticker := time.NewTicker(interval)
	stop := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-stop:
				ticker.Stop()
				return
			case <-ticker.C:
				job()
			}
		}
	}()

	return Scheduler{stop}
}

// Stop stops the goroutine running scheduled job.
func (s Scheduler) Stop() {
	select {
	case s.stop <- struct{}{}:
	default:
	}
}
