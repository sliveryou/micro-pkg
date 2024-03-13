package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCResponse_GetInt64(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":19,"id":4}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 4, resp.ID)

	result, err := resp.GetInt64()
	require.NoError(t, err)
	assert.Equal(t, int64(19), result)
}

func TestRPCResponse_GetFloat64(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":16.66,"id":2}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 2, resp.ID)

	result, err := resp.GetFloat64()
	require.NoError(t, err)
	assert.InDelta(t, 16.66, result, 0.001)
}

func TestRPCResponse_GetBool(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":true,"id":2}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 2, resp.ID)

	result, err := resp.GetBool()
	require.NoError(t, err)
	assert.True(t, result)
}

func TestRPCResponse_GetString(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":"sliveryou","id":666}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 666, resp.ID)

	result, err := resp.GetString()
	require.NoError(t, err)
	assert.Equal(t, "sliveryou", result)
}

func TestRPCResponse_ReadToObject(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":{"name":"sliveryou","age":18,"country":"China"},"id":2,"auth":"auth","ext":"ext"}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 2, resp.ID)
	assert.Len(t, resp.Extensions, 2)
	assert.Equal(t, "auth", resp.Extensions["auth"])
	assert.Equal(t, "ext", resp.Extensions["ext"])

	var result _Person
	err = resp.ReadToObject(&result)
	require.NoError(t, err)
	assert.Equal(t, _Person{Name: "sliveryou", Age: 18, Country: "China"}, result)
	assert.Len(t, resp.Extensions, 2)
}

func TestRPCResponse_MarshalJSON(t *testing.T) {
	resp := &RPCResponse{
		JSONRPC:    RPCVersion,
		ID:         2,
		Result:     _Person{Name: "sliveryou", Age: 18, Country: "China"},
		Extensions: map[string]any{"auth": "auth", "user": "sliveryou"},
	}
	body, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":2,"result":{"name":"sliveryou","age":18,"country":"China"},"auth":"auth","user":"sliveryou"}`, string(body))

	resp = &RPCResponse{
		JSONRPC: RPCVersion,
		ID:      2,
		Error: RPCError{
			Code:    99,
			Message: "rpc error",
			Data:    "rpc error",
		},
		Extensions: map[string]any{"id": 3},
	}
	body, err = json.Marshal(resp)
	require.NoError(t, err)
	assert.Equal(t, `{"jsonrpc":"2.0","id":2,"error":{"code":99,"message":"rpc error","data":"rpc error"}}`, string(body))
}

func TestRPCResponse_UnmarshalJSON(t *testing.T) {
	body := `{"jsonrpc":"2.0","result":19,"id":4}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Nil(t, resp.Error)
	assert.Empty(t, resp.Extensions)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 4, resp.ID)

	number, ok := resp.Result.(json.Number)
	assert.True(t, ok)
	result, err := number.Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(19), result)
}

func TestRPCResponse_UnmarshalJSON2(t *testing.T) {
	body := `{"jsonrpc":"2.0","error":{"code":-32600,"message":"Invalid Request","data":"Invalid Request"},"id":3,"uuid":"ee82ba1dde7148e29b755fc03d1638ba","token":"token"}`
	var resp *RPCResponse
	err := json.Unmarshal([]byte(body), &resp)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Error)
	assert.Len(t, resp.Extensions, 2)
	assert.Equal(t, RPCVersion, resp.JSONRPC)
	assert.Equal(t, 3, resp.ID)

	rpcErr := resp.GetRPCError()
	assert.NotNil(t, rpcErr)
	assert.Equal(t, -32600, rpcErr.Code)
	assert.Equal(t, "Invalid Request", rpcErr.Message)
	assert.Equal(t, "Invalid Request", rpcErr.Data)
}
