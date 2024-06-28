package sysctl

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestFileLock_Lock(t *testing.T) {
	var (
		l    = 10
		wg   sync.WaitGroup
		path = "flock.lock"
	)
	wg.Add(l)

	for i := 0; i < l; i++ {
		go func(i int) {
			fl := NewFileLock(path)
			defer func() {
				wg.Done()
				fl.Unlock()
			}()

			err := fl.Lock()
			if err == nil {
				t.Log(i)
				fl.WriteString(strconv.Itoa(i) + "\n")
				time.Sleep(100 * time.Millisecond)
			} else {
				t.Log(i, err)
			}
		}(i)
	}

	wg.Wait()
	os.Remove(path)
}
