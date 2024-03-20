package jsonrpc

import (
	"net/http"

	"github.com/sliveryou/micro-pkg/xhttp"
	"github.com/sliveryou/micro-pkg/xhttp/xreq"
)

// RPCOption JSON-RPC 客户端可选配置
type RPCOption func(server *rpcClient)

// WithHTTPClient 使用配置的 HTTP 客户端
func WithHTTPClient(hc *http.Client) RPCOption {
	return func(c *rpcClient) {
		c.client.SetHTTPClient(hc)
	}
}

// WithCustomHeaders 使用配置的 HTTP 请求头
func WithCustomHeaders(hds map[string]string) RPCOption {
	return func(c *rpcClient) {
		customHeaders := make(map[string]string)
		for k, v := range hds {
			if k == xhttp.HeaderHost && v != "" {
				c.options = append(c.options, xreq.Host(v))
			} else {
				customHeaders[k] = v
			}
		}
		if len(customHeaders) > 0 {
			c.options = append(c.options, xreq.HeaderMap(customHeaders))
		}
	}
}

// WithDefaultRequestID 使用配置的默认请求ID
func WithDefaultRequestID(id int) RPCOption {
	return func(c *rpcClient) {
		c.defaultRequestID = id
	}
}
