package jsonrpc

import "net/http"

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
func WithCustomHeaders(hds map[string]string) RPCOption {
	return func(c *rpcClient) {
		c.customHeaders = make(map[string]string)
		for k, v := range hds {
			c.customHeaders[k] = v
		}
	}
}

// WithDefaultRequestID 使用配置的默认请求ID
func WithDefaultRequestID(id int) RPCOption {
	return func(c *rpcClient) {
		c.defaultRequestID = id
	}
}
