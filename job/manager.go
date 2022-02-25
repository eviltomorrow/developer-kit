package job

import (
	"sync"
)

type manager struct {
	jobs sync.Map
}

var (
	Manager = &manager{}
)

func (m *manager) Add(job *Job) {
	prepared, ok := Manager.jobs.LoadOrStore(job.id, job)
	if ok {
		return
	}
	prepared.(*Job).startup()
}

func (m *manager) Del(id string) {
	exsited, ok := Manager.jobs.LoadAndDelete(id)
	if ok {
		exsited.(*Job).stop()
	}
}

func (m *manager) Metrics() chan string {
	var metrics = make(chan string, 16)

	m.jobs.Range(func(key, value interface{}) bool {
		metrics <- value.(*Job).metrics()
		return false
	})
	close(metrics)
	return metrics
}
