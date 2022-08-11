package scheduler

import (
	"fmt"
	"sync"
	"time"
)

type ScheduledExecutor struct {
	delay  time.Duration
	ticker time.Ticker
	quit   chan struct{}
	once   sync.Once
}

func NewTimedExecutor(initialDelay time.Duration, delay time.Duration) *ScheduledExecutor {
	return &ScheduledExecutor{
		delay:  delay,
		ticker: *time.NewTicker(initialDelay),
		quit:   make(chan struct{}),
	}
}

func (se *ScheduledExecutor) Start(task func(), runAsync bool) {
	go se.once.Do(func() {
		defer func() {
			fmt.Println("Scheduler stopped!!")
		}()
		firstExec := true
		for {
			select {
			case <-se.ticker.C:
				if firstExec {
					se.ticker.Stop()
					se.ticker = *time.NewTicker(se.delay)
					firstExec = false
				}
				if runAsync {
					go task()
				} else {
					task()
				}
				break
			case <-se.quit:
				close(se.quit)
				se.ticker.Stop()
				return
			}
		}
	})
}

func (se *ScheduledExecutor) Stop() {
	go func() {
		se.quit <- struct{}{}
	}()
}
