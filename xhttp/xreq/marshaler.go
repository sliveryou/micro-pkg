package xreq

import (
	"encoding/json"
	"encoding/xml"
)

// Marshaler 序列化函数
type Marshaler func(v any) ([]byte, error)

var (
	// JSONMarshaler json 序列化函数
	JSONMarshaler Marshaler = json.Marshal
	// XMLMarshaler xml 序列化函数
	XMLMarshaler Marshaler = xml.Marshal
)
