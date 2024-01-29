package gstream

import (
	"io"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// StreamWriter grpc 流式消息内容写入器
type StreamWriter struct {
	stream    grpc.ClientStream
	t         reflect.Type
	field     string
	err       error
	buf       []byte
	bufNum    int
	chunkSize int
}

// NewStreamWriter 新建 grpc 流式消息内容写入器
func NewStreamWriter(stream grpc.ClientStream, target interface{}, field string, chunkSize int) (io.WriteCloser, error) {
	v, err := checkAndGetTargetValue(target, field)
	if err != nil {
		return nil, err
	}

	return &StreamWriter{
		stream:    stream,
		t:         v.Type(),
		field:     field,
		buf:       make([]byte, chunkSize),
		bufNum:    0,
		chunkSize: chunkSize,
	}, nil
}

// MustNewStreamWriter 新建 grpc 流式消息内容写入器
func MustNewStreamWriter(stream grpc.ClientStream, target interface{}, field string, chunkSize int) io.WriteCloser {
	w, err := NewStreamWriter(stream, target, field, chunkSize)
	if err != nil {
		panic(err)
	}

	return w
}

// Write 写入 grpc 流式消息内容
func (w *StreamWriter) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, w.err
	}

	if w.bufNum > 0 {
		i := copy(w.buf[w.bufNum:], p)
		w.bufNum += i
		n += i
		p = p[i:]
		if w.bufNum < w.chunkSize {
			return n, err
		}

		target := reflect.New(w.t).Elem()
		target.FieldByName(w.field).SetBytes(w.buf)

		if w.err = w.stream.SendMsg(target.Addr().Interface()); w.err != nil {
			return n, w.err
		}
		w.bufNum = 0
	}

	for len(p) >= w.chunkSize {
		target := reflect.New(w.t).Elem()
		target.FieldByName(w.field).SetBytes(p[:w.chunkSize])

		if w.err = w.stream.SendMsg(target.Addr().Interface()); w.err != nil {
			return n, w.err
		}
		n += w.chunkSize
		p = p[w.chunkSize:]
	}

	w.bufNum = copy(w.buf, p)
	n += w.bufNum
	return n, err
}

// Close 关闭 grpc 流式消息内容写入器，并刷新写入器缓存内容
func (w *StreamWriter) Close() error {
	if w.err == nil && w.bufNum > 0 {
		req := reflect.New(w.t).Elem()
		req.FieldByName(w.field).SetBytes(w.buf[:w.bufNum])

		w.err = w.stream.SendMsg(req.Addr().Interface())
		w.bufNum = 0
	}
	return w.err
}

// StreamReader grpc 流式消息内容读取器
type StreamReader struct {
	stream grpc.ServerStream
	t      reflect.Type
	field  string
	index  int64
	buf    []byte
}

// NewStreamReader 新建 grpc 流式消息内容读取器
func NewStreamReader(stream grpc.ServerStream, target interface{}, field string, size int64) (io.Reader, error) {
	v, err := checkAndGetTargetValue(target, field)
	if err != nil {
		return nil, err
	}

	return io.LimitReader(&StreamReader{
		stream: stream,
		t:      v.Type(),
		field:  field,
	}, size), nil
}

// MustNewStreamReader 新建 grpc 流式消息内容读取器
func MustNewStreamReader(stream grpc.ServerStream, target interface{}, field string, size int64) io.Reader {
	r, err := NewStreamReader(stream, target, field, size)
	if err != nil {
		panic(err)
	}

	return r
}

// Read 读取 grpc 流式消息内容
func (r *StreamReader) Read(p []byte) (n int, err error) {
	if r.index >= int64(len(r.buf)) {
		target := reflect.New(r.t).Elem()

		err := r.stream.RecvMsg(target.Addr().Interface())
		if err != nil {
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}
			return 0, err
		}
		r.index = 0
		r.buf = target.FieldByName(r.field).Bytes()
	}

	n = copy(p, r.buf[r.index:])
	r.index += int64(n)
	return
}

// checkAndGetTargetValue 检查并获取目标对象的反射值对象
func checkAndGetTargetValue(target interface{}, field string) (reflect.Value, error) {
	// 目标对象不能为 nil
	if target == nil {
		return reflect.Value{}, errors.New("gstream: target cannot be nil")
	}
	// 所给字段名称不能为空
	if field == "" {
		return reflect.Value{}, errors.New("gstream: field cannot be empty")
	}
	// 所给字段须为可导出字段（首字母大写）
	if r, _ := utf8.DecodeRuneInString(field); !unicode.IsUpper(r) {
		return reflect.Value{}, errors.New("gstream: field should be exported")
	}

	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 目标对象值须不为 nil
	if !v.IsValid() {
		return reflect.Value{}, errors.New("gstream: target value is invalid")
	}
	// 目标对象须为结构体种类
	if v.Kind() != reflect.Struct {
		return reflect.Value{}, errors.New("gstream: target type should be struct")
	}
	f := v.FieldByName(field)
	// 所给字段须为目标结构体字段成员
	if !f.IsValid() {
		return reflect.Value{}, errors.New("gstream: field should be target struct field")
	}
	// 所给字段类型须为 []byte 类型
	_, ok := f.Interface().([]byte)
	if !ok {
		return reflect.Value{}, errors.New("gstream: field should be []byte type")
	}

	return v, nil
}
