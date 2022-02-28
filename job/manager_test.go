package job

import (
	"testing"
	"time"

	"golang.org/x/exp/errors/fmt"
)

func TestJob(t *testing.T) {
	go func() {
		for {
			t.Logf("data: %v\r\n", Manager.State())
			time.Sleep(1 * time.Second)
		}
	}()
	Manager.Add(NewJob("1", 2*time.Second, func() {
		fmt.Println("Hello world", time.Now().Format("15:04:05"))
		// time.Sleep(2 * time.Second)
	}))

	time.Sleep(50 * time.Second)
	Manager.Del("1")
	time.Sleep(2 * time.Second)
}
