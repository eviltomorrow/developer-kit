package job

import (
	"testing"
	"time"

	"golang.org/x/exp/errors/fmt"
)

func TestJob(t *testing.T) {
	go func() {
		for metrics := range Manager.Metrics() {
			t.Logf("%s", metrics)
		}
		time.Sleep(5 * time.Second)
	}()
	Manager.Add(NewJob("1", 2*time.Second, func() {
		fmt.Println("Hello world", time.Now().Format("15:04:05"))
	}))

	select {}
}
