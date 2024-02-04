package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestNewHealthClient(t *testing.T) {
	client := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{"foo"},
		NonBlock:  true,
	})
	assert.NotNil(t, client)
	healthClient := NewHealthClient(client)
	assert.NotNil(t, healthClient)
}
