package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sliveryou/micro-pkg/xhttp/jsonrpc/internal/mockserver"
)

var (
	addr           = ":18090"
	singleEndpoint = fmt.Sprintf("http://localhost%s/jsonrpc", addr)
	batchEndpoint  = fmt.Sprintf("http://localhost%s/jsonrpc/batch", addr)
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	go mockserver.Run(addr)
	// wait for mock server to run
	time.Sleep(5 * time.Millisecond)
}

func teardown() {
	mockserver.Close()
}

func TestRpcClient_Call(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	resp, err := client.Call(context.Background(), mockserver.MethodGetInt64, 1, 2, 3)
	require.NoError(t, err)
	assert.Nil(t, resp.Error)
	result, err := resp.GetInt64()
	require.NoError(t, err)
	t.Log(result)
	rawRequest, ok := resp.Extensions[mockserver.ExtensionRawRequest].(string)
	assert.True(t, ok)
	assert.Equal(t, `{"jsonrpc":"2.0","id":0,"method":"mock.GetInt64","params":[1,2,3]}`, rawRequest)
}

func TestRpcClient_Call2(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	resp, err := client.Call(context.Background(), "unknown method", 1, 2, 3)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
	rpcErr := resp.GetRPCError()
	assert.NotNil(t, rpcErr)
	assert.Equal(t, mockserver.DefaultErrCode, rpcErr.Code)
	assert.Equal(t, "method not found", rpcErr.Message)
	assert.Equal(t, "method not found", rpcErr.Data)
}

func TestRpcClient_Call3(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	resp, err := client.Call(context.Background(), mockserver.MethodGetStdErr, 1, 2, 3)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
	rpcErr := resp.GetRPCError()
	assert.NotNil(t, rpcErr)
	assert.Equal(t, mockserver.DefaultErrCode, rpcErr.Code)
	assert.Equal(t, "method get std err", rpcErr.Message)
	assert.Equal(t, "method get std err", rpcErr.Data)
}

func TestRpcClient_Call4(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	resp, err := client.Call(context.Background(), mockserver.MethodGetStringErr, 1, 2, 3)
	require.NoError(t, err)
	assert.NotNil(t, resp.Error)
	rpcErrString, ok := resp.Error.(string)
	assert.True(t, ok)
	assert.Equal(t, "method get string err", rpcErrString)

	rpcErr := resp.GetRPCError()
	assert.NotNil(t, rpcErr)
	assert.Equal(t, 0, rpcErr.Code)
	assert.Equal(t, "method get string err", rpcErr.Message)
	assert.Nil(t, rpcErr.Data)
}

func TestRpcClient_CallRaw(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	req := NewRPCRequestWithID(123, mockserver.MethodGetFloat64, map[string]any{"1.1": 1.1}).
		WithExtensions(map[string]any{"auth": "auth"})

	resp, err := client.CallRaw(context.Background(), req)
	require.NoError(t, err)
	assert.Nil(t, resp.Error)
	result, err := resp.GetFloat64()
	require.NoError(t, err)
	t.Log(result)
	rawRequest, ok := resp.Extensions[mockserver.ExtensionRawRequest].(string)
	assert.True(t, ok)
	assert.Equal(t, `{"jsonrpc":"2.0","id":123,"method":"mock.GetFloat64","params":{"1.1":1.1},"auth":"auth"}`, rawRequest)
}

func TestRpcClient_CallFor(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	var m map[string]any
	err := client.CallFor(context.Background(), &m, mockserver.MethodReadToObject, 1, 2, 3)
	require.NoError(t, err)
	uuid, ok := m["uuid"].(string)
	assert.True(t, ok)
	t.Log(uuid)
	rawRequest, ok := m[mockserver.ExtensionRawRequest].(string)
	assert.True(t, ok)
	assert.Equal(t, `{"jsonrpc":"2.0","id":0,"method":"mock.ReadToObject","params":[1,2,3]}`, rawRequest)
}

func TestRpcClient_CallRawFor(t *testing.T) {
	client := NewRPCClient(singleEndpoint)
	assert.NotNil(t, client)

	req := NewRPCRequestWithID(123, mockserver.MethodReadToObject, map[string]any{"1.1": 1.1}).
		WithExtensions(map[string]any{"auth": "auth"})

	var m map[string]any
	err := client.CallRawFor(context.Background(), &m, req)
	require.NoError(t, err)
	uuid, ok := m["uuid"].(string)
	assert.True(t, ok)
	t.Log(uuid)
	rawRequest, ok := m[mockserver.ExtensionRawRequest].(string)
	assert.True(t, ok)
	assert.Equal(t, `{"jsonrpc":"2.0","id":123,"method":"mock.ReadToObject","params":{"1.1":1.1},"auth":"auth"}`, rawRequest)
}

func TestRpcClient_CallBatch(t *testing.T) {
	client := NewRPCClient(batchEndpoint)
	assert.NotNil(t, client)

	req1 := NewRPCRequest(mockserver.MethodReadToObject, map[string]any{"action": "read_to_object"}).
		WithExtensions(map[string]any{"auth": "auth"})
	req2 := NewRPCRequest(mockserver.MethodGetBool, map[string]any{"action": "get_bool"}).
		WithExtensions(map[string]any{"auth": "auth"})
	req3 := NewRPCRequest(mockserver.MethodGetString, map[string]any{"action": "get_string"}).
		WithExtensions(map[string]any{"auth": "auth"})
	reqs := []*RPCRequest{req1, req2, req3}

	resps, err := client.CallBatch(context.Background(), reqs)
	require.NoError(t, err)
	assert.Len(t, resps, 3)
	assert.False(t, resps.HasError())

	b1, err := json.Marshal(reqs)
	require.NoError(t, err)

	resp1 := resps.GetByID(1)
	assert.NotNil(t, resp1)
	var result1 _Object
	err = resp1.ReadToObject(&result1)
	require.NoError(t, err)
	assert.Equal(t, string(b1), result1.RawRequest)
	fmt.Println(result1)

	b2, err := json.MarshalIndent(resps, "", "    ")
	require.NoError(t, err)
	fmt.Println(string(b2))
}

func TestRpcClient_CallBatchRaw(t *testing.T) {
	client := NewRPCClient(batchEndpoint)
	assert.NotNil(t, client)

	req1 := NewRPCRequestWithID(100, mockserver.MethodReadToObject, map[string]any{"action": "read_to_object"}).
		WithExtensions(map[string]any{"auth": "auth"})
	req2 := NewRPCRequestWithID(101, mockserver.MethodGetBool, map[string]any{"action": "get_bool"}).
		WithExtensions(map[string]any{"auth": "auth"})
	req3 := NewRPCRequestWithID(102, mockserver.MethodGetString, map[string]any{"action": "get_string"}).
		WithExtensions(map[string]any{"auth": "auth"})
	reqs := []*RPCRequest{req1, req2, req3}

	resps, err := client.CallBatchRaw(context.Background(), reqs)
	require.NoError(t, err)
	assert.Len(t, resps, 3)
	assert.False(t, resps.HasError())

	b1, err := json.Marshal(reqs)
	require.NoError(t, err)

	respMap := resps.AsMap()
	assert.NotNil(t, respMap)
	resp1 := respMap[100]
	assert.NotNil(t, resp1)
	var result1 _Object
	err = resp1.ReadToObject(&result1)
	require.NoError(t, err)
	assert.Equal(t, string(b1), result1.RawRequest)
	fmt.Println(result1)

	b2, err := json.MarshalIndent(resps, "", "    ")
	require.NoError(t, err)
	fmt.Println(string(b2))
}
