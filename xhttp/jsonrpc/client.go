package jsonrpc

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

// rpcClient 默认 JSON-RPC 客户端
type rpcClient struct {
	endpoint         string
	client           *xreq.Client
	options          xreq.OptionCollection
	defaultRequestID int
}

// Call 进行 JSON-RPC 调用
func (c *rpcClient) Call(ctx context.Context, method string, params ...any) (*RPCResponse, error) {
	return c.doCall(ctx, NewRPCRequestWithID(c.defaultRequestID, method, params...))
}

// CallRaw 基于所给请求体进行 JSON-RPC 调用
func (c *rpcClient) CallRaw(ctx context.Context, req *RPCRequest) (*RPCResponse, error) {
	return c.doCall(ctx, req)
}

// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
func (c *rpcClient) CallFor(ctx context.Context, out any, method string, params ...any) error {
	rpcResp, err := c.Call(ctx, method, params...)
	if err != nil {
		return err
	}

	return rpcResp.ReadToObject(out)
}

// CallRawFor 基于所给请求体进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
func (c *rpcClient) CallRawFor(ctx context.Context, out any, req *RPCRequest) error {
	rpcResp, err := c.CallRaw(ctx, req)
	if err != nil {
		return err
	}

	return rpcResp.ReadToObject(out)
}

// CallBatch 进行 JSON-RPC 批量调用（内部会自动设置 JSONRPC 与 ID 字段，ID 将从 1 开始递增）
func (c *rpcClient) CallBatch(ctx context.Context, reqs RPCRequests) (RPCResponses, error) {
	if len(reqs) == 0 {
		return nil, errors.New("empty request list")
	}

	for i, req := range reqs {
		req.JSONRPC = RPCVersion
		req.ID = i + 1
	}

	return c.doBatchCall(ctx, reqs)
}

// CallBatchRaw 基于所给请求体列表进行 JSON-RPC 批量调用
func (c *rpcClient) CallBatchRaw(ctx context.Context, reqs RPCRequests) (RPCResponses, error) {
	if len(reqs) == 0 {
		return nil, errors.New("empty request list")
	}

	return c.doBatchCall(ctx, reqs)
}

// NewHTTPRequest 新建 HTTP 请求体（req 可以为 *RPCRequest 或 []*RPCRequest）
func (c *rpcClient) NewHTTPRequest(ctx context.Context, req any) (*http.Request, error) {
	httpReq, err := xreq.NewPost(c.endpoint,
		xreq.Context(ctx),
		xreq.BodyJSON(req),
		c.options,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	return httpReq, nil
}

// CallWithHTTPRequest 使用 http.Request 进行 JSON-RPC 调用
func (c *rpcClient) CallWithHTTPRequest(httpReq *http.Request) (*http.Response, *RPCResponse, error) {
	resp, err := c.client.DoWithRequest(httpReq)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "call on %s err", httpReq.URL.String())
	}

	var rpcResp *RPCResponse
	if err := resp.JSONUnmarshal(&rpcResp); err != nil {
		return resp.RawResponse, nil, errors.WithMessagef(err, "call on %s status code: %d, decode body err",
			httpReq.URL.String(), resp.StatusCode())
	}
	if rpcResp == nil {
		return resp.RawResponse, nil, errors.WithMessagef(err, "call on %s status code: %d, rpc response missing err",
			httpReq.URL.String(), resp.StatusCode())
	}

	return resp.RawResponse, rpcResp, nil
}

// CallBatchWithHTTPRequest 使用 http.Request 进行 JSON-RPC 批量调用
func (c *rpcClient) CallBatchWithHTTPRequest(httpReq *http.Request) (*http.Response, []*RPCResponse, error) {
	resp, err := c.client.DoWithRequest(httpReq)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "batch call on %s err", httpReq.URL.String())
	}

	var rpcResponses RPCResponses
	if err := resp.JSONUnmarshal(&rpcResponses); err != nil {
		return resp.RawResponse, nil, errors.WithMessagef(err, "batch call on %s status code: %d, decode body err",
			httpReq.URL.String(), resp.StatusCode())
	}
	if len(rpcResponses) == 0 {
		return resp.RawResponse, nil, errors.WithMessagef(err, "batch call on %s status code: %d, rpc response missing err",
			httpReq.URL.String(), resp.StatusCode())
	}

	return resp.RawResponse, rpcResponses, nil
}

// doCall 执行 JSON-RPC 调用
func (c *rpcClient) doCall(ctx context.Context, req *RPCRequest) (*RPCResponse, error) {
	httpReq, err := c.NewHTTPRequest(ctx, req)
	if err != nil {
		return nil, errors.WithMessagef(err, "call %s method on %s err, new request err", req.Method, c.endpoint)
	}

	_, rpcResp, err := c.CallWithHTTPRequest(httpReq)

	return rpcResp, err
}

// doBatchCall 执行 JSON-RPC 批量调用
func (c *rpcClient) doBatchCall(ctx context.Context, reqs []*RPCRequest) ([]*RPCResponse, error) {
	httpReq, err := c.NewHTTPRequest(ctx, reqs)
	if err != nil {
		return nil, errors.WithMessagef(err, "batch call on %s err, new request err", c.endpoint)
	}

	_, rpcResp, err := c.CallBatchWithHTTPRequest(httpReq)

	return rpcResp, err
}
