package xonce

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOnceSuccess_Success(t *testing.T) {
	os := new(OnceSuccess)
	err := os.Do(func() error {
		return errors.New("something err")
	})
	require.Error(t, err)
	assert.False(t, os.Success())

	err = os.Do(func() error {
		return nil
	})
	require.NoError(t, err)
	assert.True(t, os.Success())

	err = os.Do(func() error {
		return errors.New("something err")
	})
	require.NoError(t, err)
	assert.True(t, os.Success())
}

type one int

func (o *one) Increment() {
	*o++
}

func TestOnceSuccess_Do(t *testing.T) {
	o := new(one)
	os := new(OnceSuccess)
	wg := new(sync.WaitGroup)
	n := 10

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(o *one, os *OnceSuccess, wg *sync.WaitGroup) {
			fmt.Println("run goroutine")
			_ = os.Do(func() error {
				fmt.Println("run once success")
				o.Increment()
				return nil
			})
			wg.Done()
		}(o, os, wg)
	}
	wg.Wait()

	assert.True(t, os.Success())
	assert.Equal(t, 1, int(*o))
}
