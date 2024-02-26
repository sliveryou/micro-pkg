package watcher

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/client/v3/mock/mockserver"
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
	time.Sleep(5 * time.Millisecond)
}

func teardown() {
	ms.Stop()
}

func getWatcher() *Watcher {
	c := Config{
		Key:       "/watcher/test-key",
		Endpoints: endpoints,
	}

	return MustNewWatcher(c)
}

func TestNewWatcher(t *testing.T) {
	c := Config{
		Key:       "/watcher/test-key",
		Endpoints: endpoints,
	}

	w, err := NewWatcher(c)
	require.NoError(t, err)
	t.Log(w)
}

func TestMustNewWatcher(t *testing.T) {
	assert.NotPanics(t, func() {
		c := Config{
			Key:       "/watcher/test-key",
			Endpoints: endpoints,
		}

		w := MustNewWatcher(c)
		t.Log(w)
	})
}

func TestWatcher_SetUpdateCallback(t *testing.T) {
	w := getWatcher()
	assert.NotNil(t, w)

	err := w.SetUpdateCallback(func(rev string) {
		logx.Info("get new rev: ", rev)
	})
	require.NoError(t, err)
}

func TestWatcher_Update(t *testing.T) {
	w := getWatcher()
	assert.NotNil(t, w)

	err := w.SetUpdateCallback(func(rev string) {
		logx.Info("get new rev: ", rev)
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = w.Update()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
}

func TestWatcher_Close(t *testing.T) {
	w := getWatcher()
	assert.NotNil(t, w)

	err := w.SetUpdateCallback(func(rev string) {
		logx.Info("get new rev:", rev)
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = w.Update()
	require.NoError(t, err)

	w.Close()

	time.Sleep(1 * time.Second)
}

func TestWatcher(t *testing.T) {
	// c := Config{
	// 	Key:       "/watcher/test-key",
	// 	Endpoints: []string{"127.0.0.1:2379"},
	// 	Username:  "root",
	// 	Password:  "ABC123456",
	// }
	c := Config{
		Key:       "/watcher/test-key",
		Endpoints: endpoints,
	}

	w1, err := NewWatcher(c)
	require.NoError(t, err)

	err = w1.SetUpdateCallback(func(rev string) {
		logx.Info("w1 get new rev:", rev)
	})
	require.NoError(t, err)

	w2, err := NewWatcher(c)
	require.NoError(t, err)

	err = w2.SetUpdateCallback(func(rev string) {
		logx.Info("w2 get new rev:", rev)
	})
	require.NoError(t, err)

	time.Sleep(1 * time.Second)

	err = w1.Update()
	require.NoError(t, err)

	time.Sleep(1 * time.Second)
}
