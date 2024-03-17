package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRPCClient(t *testing.T) {
	endpoint := "https://www.test.com/jsonrpc"
	c := NewRPCClient(endpoint,
		WithDefaultRequestID(100),
	)
	assert.NotNil(t, c)

	dc, ok := c.(*rpcClient)
	assert.True(t, ok)
	assert.Equal(t, endpoint, dc.endpoint)
	assert.Equal(t, 100, dc.defaultRequestID)
}

type person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
}

type object struct {
	Name       string `json:"name"`
	Age        int    `json:"age"`
	Country    string `json:"country"`
	UUID       string `json:"uuid"`
	RawRequest string `json:"raw_request"`
}
