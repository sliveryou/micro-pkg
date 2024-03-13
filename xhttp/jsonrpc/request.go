package jsonrpc

import "encoding/json"

var rpcReqKeySet = map[string]struct{}{"jsonrpc": {}, "id": {}, "method": {}, "params": {}}

// RPCRequests 通用 JSON-RPC 请求体列表
type RPCRequests []*RPCRequest

// RPCRequest 通用 JSON-RPC 请求体
type RPCRequest struct {
	JSONRPC    string         `json:"jsonrpc"`
	ID         int            `json:"id"`
	Method     string         `json:"method"`
	Params     any            `json:"params,omitempty"`
	Extensions map[string]any `json:"-"`
}

// NewRPCRequest 新建通用 JSON-RPC 请求体
func NewRPCRequest(method string, params ...any) *RPCRequest {
	return NewRPCRequestWithID(0, method, params...)
}

// NewRPCRequestWithID 使用指定 id 新建通用 JSON-RPC 请求体
func NewRPCRequestWithID(id int, method string, params ...any) *RPCRequest {
	req := &RPCRequest{
		JSONRPC: RPCVersion,
		ID:      id,
		Method:  method,
		Params:  Params(params...),
	}

	return req
}

// WithExtensions 使用拓展字段
func (req *RPCRequest) WithExtensions(exts map[string]any) *RPCRequest {
	req.Extensions = exts

	return req
}

// MarshalJSON 实现 json.Marshaler 接口的 MarshalJSON 方法
func (req RPCRequest) MarshalJSON() ([]byte, error) {
	// 创建新类型防止递归调用
	type _RPCRequest RPCRequest
	base, err := json.Marshal(_RPCRequest(req))
	if err != nil {
		return nil, err
	}

	if len(req.Extensions) > 0 {
		clone := make(map[string]any)
		for k, v := range req.Extensions {
			// 忽略重复的 key
			if _, ok := rpcReqKeySet[k]; !ok {
				clone[k] = v
			}
		}
		if len(clone) > 0 {
			exts, err := json.Marshal(clone)
			if err != nil {
				return nil, err
			}

			exts[0] = ','
			base = append(base[:len(base)-1], exts...)
		}
	}

	return base, nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口的 UnmarshalJSON 方法
func (req *RPCRequest) UnmarshalJSON(data []byte) error {
	// 创建新类型防止递归调用
	type _RPCRequest RPCRequest
	_req := _RPCRequest(*req)
	if err := unmarshal(data, &_req); err != nil {
		return err
	}
	*req = RPCRequest(_req)

	if err := unmarshal(data, &req.Extensions); err != nil {
		return err
	}
	for k := range rpcReqKeySet {
		delete(req.Extensions, k)
	}

	return nil
}
