package job

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Job struct {
	interval time.Duration
	ticker   *time.Ticker
	id       string
	do       func()
	count    int64
	sig      chan struct{}
	nextexec time.Time
	lastCost time.Duration
	mut      sync.RWMutex
}

func NewJob(id string, interval time.Duration, todo func()) *Job {
	var j = &Job{
		interval: interval,
		nextexec: time.Now(),
		id:       id,
		do:       todo,
		sig:      make(chan struct{}, 1),
	}
	return j
}

func (j *Job) startup() {
	j.do()
	j.ticker = time.NewTicker(j.interval)
loop:
	for {
		select {
		case <-j.ticker.C:
			var begin = time.Now()
			j.do()
			var end = time.Now()
			j.mut.Lock()
			j.nextexec = end
			if j.count >= math.MaxInt64-1 {
				j.count = 0
			}
			j.lastCost = end.Sub(begin)
			j.count++
			j.nextexec = j.nextexec.Add(j.interval)
			j.mut.Unlock()
		case <-j.sig:
			j.ticker.Stop()
			break loop
		}
	}
}

func (j *Job) stop() {
	j.sig <- struct{}{}
}

func (j *Job) state() string {
	j.mut.RLock()
	defer j.mut.RUnlock()
	return fmt.Sprintf("id: %s, total-count: %d, next-predict: %s, last-cost: %v", j.id, j.count, j.nextexec.Format("2006-01-02 15:04:05"), j.lastCost)
}
