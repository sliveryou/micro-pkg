package jsonrpc

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

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

				// struct、array、slice、interface 和 map 不改变其参数方式，其余类型都包装在数组中
				if typeOf != nil {
					switch typeOf.Kind() {
					case reflect.Struct:
						ps = param
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

// readTo 自定义 readTo 函数
func readTo(from, to any) error {
	fromBytes, err := json.Marshal(from)
	if err != nil {
		return errors.WithMessagef(err, "json marshal %v err", from)
	}

	if err := json.Unmarshal(fromBytes, to); err != nil {
		return errors.WithMessagef(err, "json unmarshal %s err", fromBytes)
	}

	return nil
}

// unmarshal 自定义 unmarshal 函数
func unmarshal(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()

	return d.Decode(v)
}

// addHTTPPrefix 为端节点添加 http 协议前缀
func addHTTPPrefix(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}

	return "http://" + endpoint
}
