package timewheel

import (
	"log"
	"testing"
	"time"
)

func TestTimewheel(t *testing.T) {
	var tw = NewTimeWheel(1*time.Second, 5, nil)
	tw.Start()
	defer tw.Stop()

	tw.AddTask(1*time.Second, 1, time.Now(), 10, func(key interface{}) {
		log.Printf("key: %v, Hello world", key)
	})
	time.Sleep(15 * time.Second)
}
