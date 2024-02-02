package xonce

import (
	"sync"
	"sync/atomic"
)

// OnceSuccess 操作执行器（只执行一次成功操作）
type OnceSuccess struct {
	done uint32
	m    sync.Mutex
}

// Success 判断操作执行是否成功
func (o *OnceSuccess) Success() bool {
	return atomic.LoadUint32(&o.done) == 1
}

// Do 操作执行
func (o *OnceSuccess) Do(f func() error) error {
	if o.Success() {
		return nil
	}

	o.m.Lock()
	defer o.m.Unlock()

	if o.done == 0 {
		if err := f(); err != nil {
			return err
		}

		atomic.StoreUint32(&o.done, 1)
	}

	return nil
}
