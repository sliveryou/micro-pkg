package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// rpcClient 默认 JSON-RPC 客户端
type rpcClient struct {
	endpoint         string
	httpClient       *http.Client
	customHeaders    map[string]string
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
	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.WithMessagef(err, "json marshal %v err", req)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	httpReq.Header.Set(xhttp.HeaderAccept, xhttp.ContentTypeJSON)
	httpReq.Header.Set(xhttp.HeaderContentType, xhttp.ContentTypeJSON)

	for k, v := range c.customHeaders {
		if k == xhttp.HeaderHost && v != "" {
			httpReq.Host = v
		} else {
			httpReq.Header.Set(k, v)
		}
	}

	return httpReq, nil
}

// CallWithHTTPRequest 使用 http.Request 进行 JSON-RPC 调用
func (c *rpcClient) CallWithHTTPRequest(httpReq *http.Request) (*http.Response, *RPCResponse, error) {
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "call on %s err", httpReq.URL.String())
	}
	defer httpResp.Body.Close()

	d := json.NewDecoder(httpResp.Body)
	d.UseNumber()

	var rpcResp *RPCResponse
	if err := d.Decode(&rpcResp); err != nil {
		return httpResp, nil, errors.WithMessagef(err, "call on %s status code: %d, decode body err",
			httpReq.URL.String(), httpResp.StatusCode)
	}
	if rpcResp == nil {
		return httpResp, nil, errors.WithMessagef(err, "call on %s status code: %d, rpc response missing err",
			httpReq.URL.String(), httpResp.StatusCode)
	}

	return httpResp, rpcResp, nil
}

// CallBatchWithHTTPRequest 使用 http.Request 进行 JSON-RPC 批量调用
func (c *rpcClient) CallBatchWithHTTPRequest(httpReq *http.Request) (*http.Response, []*RPCResponse, error) {
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, errors.WithMessagef(err, "batch call on %s err", httpReq.URL.String())
	}
	defer httpResp.Body.Close()

	d := json.NewDecoder(httpResp.Body)
	d.UseNumber()

	var rpcResponses RPCResponses
	if err := d.Decode(&rpcResponses); err != nil {
		return httpResp, nil, errors.WithMessagef(err, "batch call on %s status code: %d, decode body err",
			httpReq.URL.String(), httpResp.StatusCode)
	}
	if len(rpcResponses) == 0 {
		return httpResp, nil, errors.WithMessagef(err, "batch call on %s status code: %d, rpc response missing err",
			httpReq.URL.String(), httpResp.StatusCode)
	}

	return httpResp, rpcResponses, nil
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
