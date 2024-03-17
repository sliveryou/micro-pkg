package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCRequest_WithExtensions(t *testing.T) {
	req := &RPCRequest{}
	b, err := json.Marshal(req.WithExtensions(map[string]any{"auth": "auth"}))
	require.NoError(t, err)
	assert.Len(t, req.Extensions, 1)
	assert.Equal(t, `{"jsonrpc":"","id":0,"method":"","auth":"auth"}`, string(b))
}

func TestNewRPCRequest(t *testing.T) {
	method := "test.get"
	req := NewRPCRequest(method, 1, 2, 3)
	assert.NotNil(t, req)
	body, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":0,"method":"test.get","params":[1,2,3]}`, string(body))
}

func TestNewRPCRequestWithID(t *testing.T) {
	method := "test.get"
	req := NewRPCRequestWithID(100, method, map[string]any{"a": "a", "1": 1, "ok": true})
	assert.NotNil(t, req)
	body, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":100,"method":"test.get","params":{"1":1,"a":"a","ok":true}}`, string(body))

	req = NewRPCRequestWithID(666, method, []person{{Name: "sliveryou", Age: 18, Country: "China"}}).
		WithExtensions(map[string]any{"ext": "ext", "auth": "auth"})
	assert.NotNil(t, req)
	body, err = json.Marshal(req)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":666,"method":"test.get","params":[{"name":"sliveryou","age":18,"country":"China"}],"auth":"auth","ext":"ext"}`, string(body))
}

func TestRPCRequest_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(RPCRequest{})
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"","id":0,"method":""}`, string(b))

	b, err = json.Marshal(&RPCRequest{})
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"","id":0,"method":""}`, string(b))

	b, err = json.Marshal(&RPCRequest{Extensions: map[string]any{"id": 1}})
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"","id":0,"method":""}`, string(b))

	b, err = json.Marshal(&RPCRequest{Extensions: map[string]any{"auth": "auth"}})
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"","id":0,"method":"","auth":"auth"}`, string(b))
}

func TestRPCRequest_UnmarshalJSON(t *testing.T) {
	method := "test.get"
	req := NewRPCRequestWithID(666, method, []person{{Name: "sliveryou", Age: 18, Country: "China"}}).
		WithExtensions(map[string]any{"ext": "ext", "auth": "auth"})
	assert.NotNil(t, req)
	body, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":666,"method":"test.get","params":[{"name":"sliveryou","age":18,"country":"China"}],"auth":"auth","ext":"ext"}`, string(body))

	reqNew := &RPCRequest{}
	err = json.Unmarshal(body, reqNew)
	require.NoError(t, err)
	assert.Len(t, req.Extensions, 2)
	assert.Equal(t, req.ID, reqNew.ID)
	assert.Equal(t, req.Method, reqNew.Method)
	assert.NotEmpty(t, reqNew.Params)
	assert.NotEmpty(t, reqNew.Extensions)
	assert.Equal(t, "auth", reqNew.Extensions["auth"])
	assert.Equal(t, "ext", reqNew.Extensions["ext"])
}
