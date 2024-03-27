package xwebsocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Cancel(t *testing.T) {
	m := NewManager()
	c, _ := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))

	var wg sync.WaitGroup
	wg.Add(1)
	var cancelled atomic.Bool
	c.cancel = func() {
		cancelled.Store(true)
	}
	go func() {
		assert.False(t, cancelled.Load())
		m.DelClient(c.connID)
		time.Sleep(200 * time.Millisecond)
		assert.True(t, cancelled.Load())
		wg.Done()
	}()
	wg.Wait()

	assert.NotPanics(t, func() { m.DelClient(c.connID) })
}

func TestClient_Handlers(t *testing.T) {
	m := NewManager()

	msgChan := make(chan []byte, 1)
	handler1 := func(client *Client, msg *Msg) {
		msgChan <- msg.Data
	}
	handler2 := func(client *Client, msg *Msg) {
		fmt.Printf("received message: %s\n", msg.Data)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := m.UpgradeClient(w, r)
		require.NoError(t, err)
		assert.NotNil(t, c)
		c.AddHandlers(handler1, handler2)
		wg.Done()
	}))
	defer server.Close()

	var dialer websocket.Dialer
	conn, _, err := dialer.Dial(makeWsProto(server.URL), nil)
	require.NoError(t, err)
	wg.Wait()

	err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	require.NoError(t, err)

	bs, ok := getBytes(msgChan)
	assert.True(t, ok)
	assert.Equal(t, "hello", string(bs))
}
