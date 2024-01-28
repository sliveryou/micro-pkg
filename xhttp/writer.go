package xhttp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/h2non/filetype"
)

// LoggedResponseWriter 日志记录响应写入器
type LoggedResponseWriter struct {
	W    http.ResponseWriter
	R    *http.Request
	Code int
}

// NewLoggedResponseWriter 新建日志记录响应写入器
func NewLoggedResponseWriter(w http.ResponseWriter, r *http.Request) *LoggedResponseWriter {
	return &LoggedResponseWriter{
		W:    w,
		R:    r,
		Code: http.StatusOK,
	}
}

// Flush 实现Flush方法
func (w *LoggedResponseWriter) Flush() {
	if flusher, ok := w.W.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Header 实现Header方法
func (w *LoggedResponseWriter) Header() http.Header {
	return w.W.Header()
}

// Hijack 实现Hijack方法
func (w *LoggedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacked, ok := w.W.(http.Hijacker); ok {
		return hijacked.Hijack()
	}

	return nil, nil, errors.New("server doesn't support hijacking")
}

// Write 实现 Write 方法
func (w *LoggedResponseWriter) Write(bs []byte) (int, error) {
	return w.W.Write(bs)
}

// WriteHeader 实现 WriteHeader 方法
func (w *LoggedResponseWriter) WriteHeader(code int) {
	w.W.WriteHeader(code)
	w.Code = code
}

// DetailLoggedResponseWriter 详细日志记录响应写入器
type DetailLoggedResponseWriter struct {
	Writer *LoggedResponseWriter
	Buf    *bytes.Buffer
}

// NewDetailLoggedResponseWriter 新建详细日志记录响应写入器
func NewDetailLoggedResponseWriter(w http.ResponseWriter, r *http.Request) *DetailLoggedResponseWriter {
	return &DetailLoggedResponseWriter{
		Writer: NewLoggedResponseWriter(w, r),
		Buf:    &bytes.Buffer{},
	}
}

// Flush 实现 Flush 方法
func (w *DetailLoggedResponseWriter) Flush() {
	w.Writer.Flush()
}

// Header 实现 Header 方法
func (w *DetailLoggedResponseWriter) Header() http.Header {
	return w.Writer.Header()
}

// Hijack 实现 Hijack 方法
func (w *DetailLoggedResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.Writer.Hijack()
}

// Write 实现 Write 方法
func (w *DetailLoggedResponseWriter) Write(bs []byte) (int, error) {
	w.Buf.Write(bs)
	return w.Writer.Write(bs)
}

// WriteHeader 实现 WriteHeader 方法
func (w *DetailLoggedResponseWriter) WriteHeader(code int) {
	w.Writer.WriteHeader(code)
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// A MultipartWriter generates multipart messages.
type MultipartWriter struct {
	w *multipart.Writer
}

// NewMultipartWriter returns a new multipart Writer with a random boundary,
// writing to w.
func NewMultipartWriter(w io.Writer) *MultipartWriter {
	return &MultipartWriter{
		w: multipart.NewWriter(w),
	}
}

// Boundary returns the Writer's boundary.
func (w *MultipartWriter) Boundary() string {
	return w.w.Boundary()
}

// SetBoundary overrides the Writer's default randomly-generated
// boundary separator with an explicit value.
//
// SetBoundary must be called before any parts are created, may only
// contain certain ASCII characters, and must be non-empty and
// at most 70 bytes long.
func (w *MultipartWriter) SetBoundary(boundary string) error {
	return w.w.SetBoundary(boundary)
}

// FormDataContentType returns the Content-Type for an HTTP
// multipart/form-data with this Writer's Boundary.
func (w *MultipartWriter) FormDataContentType() string {
	return w.w.FormDataContentType()
}

// CreatePart creates a new multipart section with the provided
// header. The body of the part should be written to the returned
// Writer. After calling CreatePart, any previous part may no longer
// be written to.
func (w *MultipartWriter) CreatePart(header textproto.MIMEHeader) (io.Writer, error) {
	return w.w.CreatePart(header)
}

// CreateFormFile is a convenience wrapper around CreatePart. It creates
// a new form-data header with the provided field name and file name.
func (w *MultipartWriter) CreateFormFile(fieldName, fileName string, contentType ...string) (io.Writer, error) {
	ct := ApplicationStream
	if len(contentType) > 0 {
		ct = contentType[0]
	}

	h := make(textproto.MIMEHeader)
	h.Set(HeaderContentDisposition,
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes(fieldName), escapeQuotes(fileName)))
	h.Set(HeaderContentType, ct)

	return w.CreatePart(h)
}

// WriteFile calls CreateFormFile and then writes the given file reader.
// It also returns the file content type.
func (w *MultipartWriter) WriteFile(fieldName, fileName string, fileReader io.Reader, contentType ...string) (string, error) {
	var head []byte
	var ct string
	if len(contentType) > 0 {
		ct = contentType[0]
	} else {
		// https://github.com/h2non/filetype#file-header
		head = make([]byte, 261)
		n, err := fileReader.Read(head)
		if err != nil {
			return "", err
		}

		kind, err := filetype.Match(head)
		if err != nil {
			return "", err
		}

		if kind != filetype.Unknown {
			ct = kind.MIME.Value
		} else {
			ct = TypeByExtension(fileName)
		}

		head = head[:n]
	}

	p, err := w.CreateFormFile(fieldName, fileName, ct)
	if err != nil {
		return "", err
	}

	if len(head) > 0 {
		_, err = p.Write(head)
		if err != nil {
			return "", err
		}
	}

	_, err = io.Copy(p, fileReader)
	if err != nil {
		return "", err
	}

	return ct, nil
}

// CreateFormField calls CreatePart with a header using the
// given field name.
func (w *MultipartWriter) CreateFormField(fieldName string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set(HeaderContentDisposition,
		fmt.Sprintf(`form-data; name="%s"`, escapeQuotes(fieldName)))

	return w.CreatePart(h)
}

// WriteField calls CreateFormField and then writes the given value.
func (w *MultipartWriter) WriteField(fieldName, value string) error {
	p, err := w.CreateFormField(fieldName)
	if err != nil {
		return err
	}

	_, err = p.Write([]byte(value))
	return err
}

// Close finishes the multipart message and writes the trailing
// boundary end line to the output.
func (w *MultipartWriter) Close() error {
	return w.w.Close()
}
