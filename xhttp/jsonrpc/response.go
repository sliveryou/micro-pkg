package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

var rpcRespKeySet = map[string]struct{}{"jsonrpc": {}, "id": {}, "result": {}, "error": {}}

// RPCError 通用 JSON-RPC 错误
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Error 实现 Error 方法
func (e *RPCError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// RPCResponses 通用 JSON-RPC 响应体列表
type RPCResponses []*RPCResponse

// AsMap 将响应体列表作为 map 返回，其中 key 为响应对应的 id
func (resps RPCResponses) AsMap() map[int]*RPCResponse {
	respMap := make(map[int]*RPCResponse)
	for _, resp := range resps {
		respMap[resp.ID] = resp
	}

	return respMap
}

// GetByID 返回给定 id 的响应体，如果响应体不存在，则返回 nil
func (resps RPCResponses) GetByID(id int) *RPCResponse {
	for _, resp := range resps {
		if resp.ID == id {
			return resp
		}
	}

	return nil
}

// HasError 判断响应体列表中是否有响应体发生了错误
func (resps RPCResponses) HasError() bool {
	for _, resp := range resps {
		if resp.Error != nil {
			return true
		}
	}

	return false
}

// RPCResponse 通用 JSON-RPC 响应体
type RPCResponse struct {
	JSONRPC    string         `json:"jsonrpc"`
	ID         int            `json:"id"`
	Result     any            `json:"result,omitempty"`
	Error      any            `json:"error,omitempty"`
	Extensions map[string]any `json:"-"`
}

// GetRPCError 获取通用 JSON-RPC 错误
func (resp *RPCResponse) GetRPCError() *RPCError {
	if resp.Error == nil {
		return nil
	}

	switch e := resp.Error.(type) {
	case string:
		return &RPCError{Message: e}
	case map[string]any:
		to := &RPCError{}
		if err := readTo(e, to); err == nil {
			return to
		}
	}

	return &RPCError{Message: fmt.Sprintf("%v", resp.Error)}
}

// GetInt64 获取响应结果的 int64 类型值
func (resp *RPCResponse) GetInt64() (int64, error) {
	if resp.Error != nil {
		return 0, resp.GetRPCError()
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
		return 0, resp.GetRPCError()
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
		return false, resp.GetRPCError()
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
		return "", resp.GetRPCError()
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
		return resp.GetRPCError()
	}

	return readTo(resp.Result, to)
}

// MarshalJSON 实现 json.Marshaler 接口的 MarshalJSON 方法
func (resp RPCResponse) MarshalJSON() ([]byte, error) {
	// 创建新类型防止递归调用
	type _RPCResponse RPCResponse
	base, err := json.Marshal(_RPCResponse(resp))
	if err != nil {
		return nil, err
	}

	if len(resp.Extensions) > 0 {
		clone := make(map[string]any)
		for k, v := range resp.Extensions {
			// 忽略重复的 key
			if _, ok := rpcRespKeySet[k]; !ok {
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
func (resp *RPCResponse) UnmarshalJSON(data []byte) error {
	// 创建新类型防止递归调用
	type _RPCResponse RPCResponse
	_resp := _RPCResponse(*resp)
	if err := unmarshal(data, &_resp); err != nil {
		return err
	}
	*resp = RPCResponse(_resp)

	if err := unmarshal(data, &resp.Extensions); err != nil {
		return err
	}
	for k := range rpcRespKeySet {
		delete(resp.Extensions, k)
	}

	return nil
}
