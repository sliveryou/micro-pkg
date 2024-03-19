package xreq

import (
	"encoding/json"
	"encoding/xml"
)

type (
	// Marshaler 序列化函数
	Marshaler = func(v any) ([]byte, error)
	// Unmarshaler 反序列化函数
	Unmarshaler = func(data []byte, v any) error
)

var (
	// JSONMarshaler json 序列化函数
	JSONMarshaler = json.Marshal
	// XMLMarshaler xml 序列化函数
	XMLMarshaler = xml.Marshal

	// JSONUnmarshaler json 反序列化函数
	JSONUnmarshaler = json.Unmarshal
	// XMLUnmarshaler xml 反序列化函数
	XMLUnmarshaler = xml.Unmarshal
)
