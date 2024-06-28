package xwebsocket

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	assert.NotNil(t, m)
	assert.Equal(t, 2048, m.Upgrader.ReadBufferSize)
	assert.Equal(t, 2048, m.Upgrader.WriteBufferSize)
}

func TestManager_UpgradeClient(t *testing.T) {
	m := NewManager()
	c, _ := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))
	assert.NotNil(t, c)
	assert.NotNil(t, c.Conn())
	assert.NotEmpty(t, c.ConnID())
	assert.Equal(t, "1", c.UserID())
	assert.Equal(t, "web", c.AppID())
	assert.True(t, c.IsAlive())
}

func TestManager_GetClient(t *testing.T) {
	m := NewManager()
	c, _ := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))

	get, ok := m.GetClient(c.ConnID())
	assert.True(t, ok)
	assert.NotNil(t, get)
	assert.Equal(t, c.ConnID(), get.ConnID())
	assert.Equal(t, c.UserID(), get.UserID())
	assert.Equal(t, c.AppID(), get.AppID())
}

func TestManager_GetUserClient(t *testing.T) {
	m := NewManager()
	c, _ := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))

	get, ok := m.GetUserClient(c.UserID(), c.AppID())
	assert.True(t, ok)
	assert.NotNil(t, get)
	assert.Equal(t, c.ConnID(), get.ConnID())
	assert.Equal(t, c.UserID(), get.UserID())
	assert.Equal(t, c.AppID(), get.AppID())
}

func TestManager_GetUserClients(t *testing.T) {
	m := NewManager()
	c, _ := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))

	gets, ok := m.GetUserClients(c.UserID())
	assert.True(t, ok)
	assert.Len(t, gets, 1)

	get := gets[0]
	assert.Equal(t, c.ConnID(), get.ConnID())
	assert.Equal(t, c.UserID(), get.UserID())
	assert.Equal(t, c.AppID(), get.AppID())
}

func TestManager_GetClients(t *testing.T) {
	m := NewManager()
	_, _ = getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))
	_, _ = getClientAndChan(t, m, WithUserID("2"), WithAppID("web"))

	gets := m.GetClients()
	assert.Len(t, gets, 2)
}

func TestManager_SendMsgToClient(t *testing.T) {
	m := NewManager()
	c, cc := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))

	m.SendMsgToClient(c.ConnID(), []byte("Hello, SliverYou!"))

	bs, ok := getBytes(cc)
	assert.True(t, ok)
	assert.Equal(t, "Hello, SliverYou!", string(bs))
}

func TestManager_SendMsgToUser(t *testing.T) {
	m := NewManager()
	_, cc1 := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))
	_, cc2 := getClientAndChan(t, m, WithUserID("1"), WithAppID("ios"))

	m.SendMsgToUser("1", []byte("Hello, SliverYou!"))

	bs1, ok1 := getBytes(cc1)
	assert.True(t, ok1)
	assert.Equal(t, "Hello, SliverYou!", string(bs1))

	bs2, ok2 := getBytes(cc2)
	assert.True(t, ok2)
	assert.Equal(t, "Hello, SliverYou!", string(bs2))
}

func TestManager_BroadcastMsg(t *testing.T) {
	m := NewManager()
	_, cc1 := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))
	_, cc2 := getClientAndChan(t, m, WithUserID("2"), WithAppID("web"))

	m.BroadcastMsg([]byte("Hello, SliverYou!"))

	bs1, ok1 := getBytes(cc1)
	assert.True(t, ok1)
	assert.Equal(t, "Hello, SliverYou!", string(bs1))

	bs2, ok2 := getBytes(cc2)
	assert.True(t, ok2)
	assert.Equal(t, "Hello, SliverYou!", string(bs2))
}

func TestManager_DelClient(t *testing.T) {
	m := NewManager()
	_, cc1 := getClientAndChan(t, m, WithUserID("1"), WithAppID("web"))
	_, cc2 := getClientAndChan(t, m, WithUserID("2"), WithAppID("web"))
	c3, cc3 := getClientAndChan(t, m, WithUserID("3"), WithAppID("web"))

	m.DelClient(c3.ConnID())
	assert.False(t, c3.IsAlive())
	m.BroadcastMsg([]byte("Hello, SliverYou!"))

	bs1, ok1 := getBytes(cc1)
	assert.True(t, ok1)
	assert.Equal(t, "Hello, SliverYou!", string(bs1))

	bs2, ok2 := getBytes(cc2)
	assert.True(t, ok2)
	assert.Equal(t, "Hello, SliverYou!", string(bs2))

	_, ok3 := getBytes(cc3)
	assert.False(t, ok3)
}

func getClientAndChan(t *testing.T, m *Manager, opts ...Option) (*Client, chan []byte) {
	t.Helper()

	var client *Client
	var wg sync.WaitGroup
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := m.UpgradeClient(w, r, opts...)
		require.NoError(t, err)
		client = c
		wg.Done()
	}))

	var dialer websocket.Dialer
	conn, _, err := dialer.Dial(makeWsProto(server.URL), nil)
	require.NoError(t, err)

	wg.Wait()
	assert.NotNil(t, client)

	// 模拟前端接收到的消息
	clientChan := make(chan []byte, 1)
	go func() {
		defer close(clientChan)

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
			clientChan <- message
		}
	}()

	client.AddHandlers(func(client *Client, msg *Msg) {
		fmt.Printf("received message: %s\n", msg.Data)
	})

	return client, clientChan
}

func makeWsProto(s string) string {
	return "ws" + strings.TrimPrefix(s, "http")
}

func getBytes(ch <-chan []byte) ([]byte, bool) {
	select {
	case bs, ok := <-ch:
		return bs, ok
	case <-time.After(1 * time.Second):
		return nil, false
	}
}
