package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sliveryou/micro-pkg/xhttp"
)

func TestRPCOption(t *testing.T) {
	endpoint := "https://www.test.com/jsonrpc"
	hc := xhttp.NewHTTPClient()
	hds := map[string]string{
		xhttp.HeaderUserAgent: "Test",
	}
	c := NewRPCClient(endpoint,
		WithHTTPClient(hc),
		WithCustomHeaders(hds),
		WithDefaultRequestID(1),
	)
	assert.NotNil(t, c)

	dc, ok := c.(*rpcClient)
	assert.True(t, ok)
	assert.Equal(t, endpoint, dc.endpoint)
	assert.Equal(t, hc, dc.client.GetHTTPClient())
	assert.Equal(t, 1, dc.defaultRequestID)
}
