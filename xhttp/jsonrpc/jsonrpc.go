package jsonrpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

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
	CallRaw(ctx context.Context, req *RPCRequest) (*RPCResponse, error)
	// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallFor(ctx context.Context, out any, method string, params ...any) error
	// CallRawFor 基于所给请求体进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallRawFor(ctx context.Context, out any, req *RPCRequest) error

	// CallBatch 进行 JSON-RPC 批量调用（内部会自动设置 JSONRPC 与 ID 字段，ID 将从 1 开始递增）
	CallBatch(ctx context.Context, reqs RPCRequests) (RPCResponses, error)
	// CallBatchRaw 基于所给请求体列表进行 JSON-RPC 批量调用
	CallBatchRaw(ctx context.Context, reqs RPCRequests) (RPCResponses, error)

	// NewHTTPRequest 新建 HTTP 请求体（req 可以为 *RPCRequest 或 RPCRequests）
	NewHTTPRequest(ctx context.Context, req any) (*http.Request, error)
	// CallWithHTTPRequest 使用 http.Request 进行 JSON-RPC 调用
	CallWithHTTPRequest(httpReq *http.Request) (*http.Response, *RPCResponse, error)
	// CallBatchWithHTTPRequest 使用 http.Request 进行 JSON-RPC 批量调用
	CallBatchWithHTTPRequest(httpReq *http.Request) (*http.Response, []*RPCResponse, error)
}

// NewRPCClient 新建通用 JSON-RPC 客户端
func NewRPCClient(endpoint string, opts ...RPCOption) RPCClient {
	if endpoint == "" {
		panic(errors.New("jsonrpc: empty endpoint is invalid"))
	}

	c := &rpcClient{endpoint: addHTTPPrefix(endpoint)}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = xhttp.NewHTTPClient()
	}

	return c
}
