package job

import (
	"fmt"
	"math"
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
			j.do()
			if j.count >= math.MaxInt64-1 {
				j.count = 0
			}
			j.count++
			j.nextexec = j.nextexec.Add(j.interval)
		case <-j.sig:
			j.ticker.Stop()
			break loop
		}
	}
}

func (j *Job) stop() {
	j.sig <- struct{}{}
}

func (j *Job) metrics() string {
	return fmt.Sprintf("id: %s, count: %d, next: %s", j.id, j.count, j.nextexec.Format("2006-01-02 15:04:05"))
}
