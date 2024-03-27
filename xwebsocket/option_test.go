package xwebsocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithConnTime(t *testing.T) {
	now := time.Now()
	c := &Client{}
	opt := WithConnTime(now)
	opt(c)
	assert.Equal(t, now, c.connTime)
}

func TestWithAppID(t *testing.T) {
	c := &Client{}
	opt := WithAppID("web")
	opt(c)
	assert.Equal(t, "web", c.appID)
}

func TestWithUserID(t *testing.T) {
	c := &Client{}
	opt := WithUserID("123456")
	opt(c)
	assert.Equal(t, "123456", c.userID)
}

func TestWithHandlers(t *testing.T) {
	c := &Client{}
	opt := WithHandlers(func(client *Client, msg *Msg) {})
	opt(c)
	assert.Len(t, c.handlers, 1)
}

func TestWithChanSize(t *testing.T) {
	c := &Client{}
	opt := WithChanSize(50)
	opt(c)
	assert.Equal(t, 50, cap(c.msgChan))
}
