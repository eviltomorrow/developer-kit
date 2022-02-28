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
	go prepared.(*Job).startup()
}

func (m *manager) Del(id string) {
	exsited, ok := Manager.jobs.LoadAndDelete(id)
	if ok {
		exsited.(*Job).stop()
	}
}

func (m *manager) State() []string {
	var data = make([]string, 0, 64)
	m.jobs.Range(func(key, value interface{}) bool {
		data = append(data, value.(*Job).state())
		return true
	})
	return data
}
