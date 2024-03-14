package jsonrpc

import (
	"context"

	"github.com/sliveryou/micro-pkg/xhttp"
)

const (
	// RPCVersion 默认 JSON-RPC 版本
	RPCVersion = "2.0"
)

// RPCClient 通用 JSON-RPC 客户端接口
//
//	规范：https://wiki.geekdream.com/Specification/json-rpc_2.0.html
type RPCClient interface {
	// Call 进行 JSON-RPC 调用
	Call(ctx context.Context, method string, params ...any) (*RPCResponse, error)
	// CallRaw 基于所给请求体进行 JSON-RPC 调用
	CallRaw(ctx context.Context, request *RPCRequest) (*RPCResponse, error)
	// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallFor(ctx context.Context, out any, method string, params ...any) error
	// CallRawFor 基于所给请求体进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallRawFor(ctx context.Context, out any, request *RPCRequest) error
	// CallBatch 进行 JSON-RPC 批量调用（会自动设置 JSONRPC 与 ID 字段，ID 将从 1 开始递增）
	CallBatch(ctx context.Context, reqs RPCRequests) (RPCResponses, error)
	// CallBatchRaw 基于所给请求体进行 JSON-RPC 批量调用
	CallBatchRaw(ctx context.Context, reqs RPCRequests) (RPCResponses, error)
}

// NewRPCClient 新建通用 JSON-RPC 客户端
func NewRPCClient(endpoint string, opts ...RPCOption) RPCClient {
	c := &rpcClient{endpoint: endpoint}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = xhttp.NewHTTPClient()
	}

	return c
}
