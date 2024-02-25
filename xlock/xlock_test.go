package xlock

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/xlock/internal/mockserver"
)

var (
	ms        *mockserver.MockServers
	endpoints []string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	servers, err := mockserver.StartMockServers(2)
	if err != nil {
		log.Fatalf("mockserver.StartMockServers err: %v", err)
	}
	ms = servers

	endpoints = make([]string, 0, len(ms.Servers))
	for _, s := range ms.Servers {
		endpoints = append(endpoints, s.Address)
	}

	// wait for mock server to run
	time.Sleep(time.Millisecond * 3)
}

func teardown() {
	ms.Stop()
}

func getLocker() *Locker {
	c := Config{
		Prefix:    "/xlock/",
		Endpoints: endpoints,
	}

	return MustNewLocker(c)
}

func TestNewLocker(t *testing.T) {
	c := Config{
		Prefix:    "/xlock/",
		Endpoints: endpoints,
	}

	l, err := NewLocker(c)
	require.NoError(t, err)
	t.Log(l)
}

func TestMustNewLocker(t *testing.T) {
	c := Config{
		Prefix:    "/xlock/",
		Endpoints: endpoints,
	}

	assert.NotPanics(t, func() {
		MustNewLocker(c)
	})
}

func TestLocker_NewLock(t *testing.T) {
	l := getLocker()
	key := "/test-key/"

	lock, err := l.NewLock(key)
	require.NoError(t, err)
	assert.NotNil(t, lock)
	assert.Equal(t, "/xlock/test-key", lock.lockKey)
}

func TestLock_Lock(t *testing.T) {
	l := getLocker()
	key := "test-key"

	lock, err := l.NewLock(key)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	err = lock.Lock(ctx)
	require.NoError(t, err)
	t.Log("lock successfully")

	err = lock.Unlock(ctx)
	require.NoError(t, err)
	t.Log("unlock successfully")

	cancel()
}

func TestLock_TryLock(t *testing.T) {
	l := getLocker()
	key := "test-key"

	lock, err := l.NewLock(key)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

	err = lock.TryLock(ctx)
	require.NoError(t, err)
	t.Log("lock successfully")

	err = lock.Unlock(ctx)
	require.NoError(t, err)
	t.Log("unlock successfully")

	cancel()
}
