package xhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/pkg/errors"
)

const (
	// jsonrpcVersion 默认 JSON-RPC 版本
	jsonrpcVersion = "2.0"
)

// RPCClient 通用 JSON-RPC 客户端接口
type RPCClient interface {
	// Call 进行 JSON-RPC 调用
	Call(ctx context.Context, method string, params ...any) (*RPCResponse, error)
	// CallRaw 基于所给请求体进行 JSON-RPC 调用
	CallRaw(ctx context.Context, request *RPCRequest) (*RPCResponse, error)
	// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
	CallFor(ctx context.Context, out any, method string, params ...any) error
}

// RPCOption JSON-RPC 客户端可选配置
type RPCOption func(server *rpcClient)

// WithHTTPClient 使用配置的 HTTP 客户端
func WithHTTPClient(hc *http.Client) RPCOption {
	return func(c *rpcClient) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithCustomHeaders 使用配置的 HTTP 请求头
func WithCustomHeaders(m map[string]string) RPCOption {
	return func(c *rpcClient) {
		c.customHeaders = make(map[string]string)
		for k, v := range m {
			c.customHeaders[k] = v
		}
	}
}

// NewRPCClient 新建通用 JSON-RPC 客户端
func NewRPCClient(endpoint string, opts ...RPCOption) RPCClient {
	c := &rpcClient{endpoint: endpoint}

	for _, opt := range opts {
		opt(c)
	}

	if c.httpClient == nil {
		c.httpClient = NewHTTPClient()
	}

	return c
}

// rpcClient 默认 JSON-RPC 客户端
type rpcClient struct {
	endpoint      string
	httpClient    *http.Client
	customHeaders map[string]string
}

// newRequest 新建 HTTP 请求体
func (c *rpcClient) newRequest(ctx context.Context, req any) (*http.Request, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.WithMessagef(err, "json marshal %v err", req)
	}
	// fmt.Println(string(body))

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.WithMessage(err, "new http request err")
	}

	request.Header.Set(HeaderAccept, ApplicationJSON)
	request.Header.Set(HeaderContentType, ApplicationJSON)

	for k, v := range c.customHeaders {
		request.Header.Set(k, v)
	}

	return request, nil
}

// doCall 执行 JSON-RPC 调用
func (c *rpcClient) doCall(ctx context.Context, req *RPCRequest) (*RPCResponse, error) {
	httpReq, err := c.newRequest(ctx, req)
	if err != nil {
		return nil, errors.WithMessagef(err, "call %s method on %s err",
			req.Method, c.endpoint)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.WithMessagef(err, "call %s method on %s err",
			req.Method, httpReq.URL.String())
	}
	defer httpResp.Body.Close()

	d := json.NewDecoder(httpResp.Body)
	d.DisallowUnknownFields()
	d.UseNumber()

	var rpcResp *RPCResponse
	err = d.Decode(&rpcResp)
	if err != nil {
		return nil, errors.WithMessagef(err, "call %s method on %s status code: %d, decode body err",
			req.Method, httpReq.URL.String(), httpResp.StatusCode)
	}
	if rpcResp == nil {
		return nil, errors.WithMessagef(err, "call %s method on %s status code: %d, rpc response missing err",
			req.Method, httpReq.URL.String(), httpResp.StatusCode)
	}

	return rpcResp, nil
}

// Call 进行 JSON-RPC 调用
func (c *rpcClient) Call(ctx context.Context, method string, params ...any) (*RPCResponse, error) {
	req := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		JSONRPC: jsonrpcVersion,
	}

	return c.doCall(ctx, req)
}

// CallRaw 基于所给请求体进行 JSON-RPC 调用
func (c *rpcClient) CallRaw(ctx context.Context, request *RPCRequest) (*RPCResponse, error) {
	return c.doCall(ctx, request)
}

// CallFor 进行 JSON-RPC 调用并将响应结果反序列化到所给类型对象中
func (c *rpcClient) CallFor(ctx context.Context, out any, method string, params ...any) error {
	rpcResp, err := c.Call(ctx, method, params...)
	if err != nil {
		return err
	}

	return rpcResp.ReadToObject(out)
}

// RPCRequest 通用 JSON-RPC 请求体
type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// NewRPCRequest 新建通用 JSON-RPC 请求体
func NewRPCRequest(method string, params ...any) *RPCRequest {
	req := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		JSONRPC: jsonrpcVersion,
	}

	return req
}

// Params 构建请求参数
func Params(params ...any) any {
	var ps any

	if params != nil {
		switch len(params) {
		case 0:
		case 1:
			if param := params[0]; param != nil {
				typeOf := reflect.TypeOf(param)
				for typeOf != nil && typeOf.Kind() == reflect.Ptr {
					typeOf = typeOf.Elem()
				}

				// array、slice、interface 和 map 不改变其参数方式，其余类型都包装在数组中
				if typeOf != nil {
					switch typeOf.Kind() {
					case reflect.Array:
						ps = param
					case reflect.Slice:
						ps = param
					case reflect.Interface:
						ps = param
					case reflect.Map:
						ps = param
					default:
						ps = params
					}
				}
			} else {
				ps = params
			}
		default:
			ps = params
		}
	}

	return ps
}

// RPCResponse 通用 JSON-RPC 响应体
type RPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// GetInt64 获取响应结果的 int64 类型值
func (resp *RPCResponse) GetInt64() (int64, error) {
	if resp.Error != nil {
		return 0, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(json.Number)
	if !ok {
		return 0, errors.Errorf("parse number from %v err", resp.Result)
	}

	i, err := val.Int64()
	if err != nil {
		return 0, errors.Errorf("parse int64 from %v err", resp.Result)
	}

	return i, nil
}

// GetFloat64 获取响应结果的 float64 类型值
func (resp *RPCResponse) GetFloat64() (float64, error) {
	if resp.Error != nil {
		return 0, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(json.Number)
	if !ok {
		return 0, errors.Errorf("parse number from %v err", resp.Result)
	}

	f, err := val.Float64()
	if err != nil {
		return 0, errors.Errorf("parse float64 from %v err", resp.Result)
	}

	return f, nil
}

// GetBool 获取响应结果的 bool 类型值
func (resp *RPCResponse) GetBool() (bool, error) {
	if resp.Error != nil {
		return false, errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(bool)
	if !ok {
		return false, errors.Errorf("parse bool from %v err", resp.Result)
	}

	return val, nil
}

// GetString 获取响应结果的 string 类型值
func (resp *RPCResponse) GetString() (string, error) {
	if resp.Error != nil {
		return "", errors.Errorf("%v", resp.Error)
	}

	val, ok := resp.Result.(string)
	if !ok {
		return "", errors.Errorf("parse string from %v err", resp.Result)
	}

	return val, nil
}

// ReadToObject 将响应结果反序列化到所给类型对象中
func (resp *RPCResponse) ReadToObject(to any) error {
	if resp.Error != nil {
		return errors.Errorf("%v", resp.Error)
	}

	from, err := json.Marshal(resp.Result)
	if err != nil {
		return errors.WithMessagef(err, "json marshal %v err", resp.Result)
	}

	err = json.Unmarshal(from, to)
	if err != nil {
		return errors.WithMessagef(err, "json unmarshal %s err", from)
	}

	return nil
}
