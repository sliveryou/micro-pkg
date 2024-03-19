package xreq

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	stdurl "net/url"
	stdpath "path"
	"reflect"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/convert"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// -------------------- 可选参数接口 -------------------- //

type (
	// Marshaler 序列化函数
	Marshaler = func(v any) ([]byte, error)
	// Unmarshaler 反序列化函数
	Unmarshaler = func(data []byte, v any) error
)

// Option http 请求可选参数应用器接口
type Option interface {
	// Apply 将可选参数应用于 http 请求中
	Apply(request *http.Request) (*http.Request, error)
}

// OptionCollection 可选参数集合
type OptionCollection []Option

// Apply 将可选参数集合应用于 http 请求中
func (oc OptionCollection) Apply(request *http.Request) (*http.Request, error) {
	for _, option := range oc {
		var err error
		request, err = option.Apply(request)
		if err != nil {
			message := "option collection apply err"
			if optionName := getOptionName(option); optionName != "" {
				message = fmt.Sprintf("option collection apply %s err", optionName)
			}

			return nil, errors.WithMessage(err, message)
		}
	}

	return request, nil
}

// With 添加可选参数列表
func (oc OptionCollection) With(withOptions ...Option) OptionCollection {
	return copyAndAppend(oc, withOptions...)
}

// OptionFunc 可选参数生成函数
type OptionFunc func(request *http.Request) (*http.Request, error)

// Apply 将可选参数生成函数应用于 http 请求中
func (f OptionFunc) Apply(request *http.Request) (*http.Request, error) {
	return f(request)
}

// Apply 将可选参数列表应用于 http 请求中
func Apply(request *http.Request, options ...Option) (*http.Request, error) {
	return OptionCollection(options).Apply(request)
}

// -------------------- 预定义可选参数 -------------------- //

// Method 将 method 应用于 http 请求中
func Method(method string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.Method = method

		return request, nil
	})
}

// RawURL 将 *url.URL 应用于 http 请求中
func RawURL(url *stdurl.URL) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.URL = url

		return request, nil
	})
}

// URL 将 url 应用于 http 请求中
func URL(url string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		u, err := stdurl.Parse(url)
		if err != nil {
			return nil, errors.WithMessagef(err, "parse url: %s err", url)
		}

		return RawURL(u).Apply(request)
	})
}

// Scheme 将 scheme 应用于 http 请求的 url 中
func Scheme(scheme string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.URL.Scheme = scheme

		return request, nil
	})
}

// User 将 *url.Userinfo 应用于 http 请求的 url 中
func User(user *stdurl.Userinfo) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.URL.User = user

		return request, nil
	})
}

// Username 将 username 应用于 http 请求的 url 中
func Username(username string) Option {
	return User(stdurl.User(username))
}

// UserPassword 将 username 和 password 应用于 http 请求的 url 中
func UserPassword(username, password string) Option {
	return User(stdurl.UserPassword(username, password))
}

// Host 将 host 应用于 http 请求中
func Host(host string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.Host = host
		request.URL.Host = host

		return request, nil
	})
}

// Path 使用 path.Join 将路径分段连接并应用于 http 请求的 url path 中
func Path(segments ...string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		path := stdpath.Join(segments...)
		if !stdpath.IsAbs(path) {
			path = "/" + path
		}

		request.URL.Path = path

		return request, nil
	})
}

// AddPath 使用 path.Join 将路径分段连接并添加到 http 请求的 url path 中
func AddPath(segments ...string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		elems := append([]string{request.URL.Path}, segments...)

		return Apply(request, Path(elems...))
	})
}

// Query 将 key 和 value 应用于 http 请求的 url query 中
func Query(key string, value any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		q.Set(key, convert.ToString(value))

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// AddQuery 将 key 和 value 添加到 http 请求的 url query 中
func AddQuery(key string, value any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		q.Add(key, convert.ToString(value))

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// Queries 将 url.Values 应用于 http 请求的 url query 中
func Queries(queries stdurl.Values) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		for key, values := range queries {
			q[key] = values
		}

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// AddQueries 将 url.Values 添加到 http 请求的 url query 中
func AddQueries(queries stdurl.Values) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		for key, values := range queries {
			q[key] = append(q[key], values...)
		}

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// QueryMap 将 map[string]any 应用于 http 请求的 url query 中
func QueryMap(queryMap map[string]any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		for key, value := range queryMap {
			q.Set(key, convert.ToString(value))
		}

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// AddQueryMap 将 map[string]any 添加到 http 请求的 url query 中
func AddQueryMap(queryMap map[string]any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		q := request.URL.Query()
		for key, value := range queryMap {
			q.Add(key, convert.ToString(value))
		}

		request.URL.RawQuery = q.Encode()

		return request, nil
	})
}

// Header 将 key 和 values 应用于 http 请求的 header 中
func Header(key string, values ...string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.Header.Set(key, strings.Join(values, ", "))

		return request, nil
	})
}

// Headers 将 http.Header 应用于 http 请求的 header 中
func Headers(headers http.Header) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		for key, values := range headers {
			request.Header.Set(key, strings.Join(values, ", "))
		}

		return request, nil
	})
}

// HeaderMap 将 map[string]string 应用于 http 请求的 header 中
func HeaderMap(headerMap map[string]string) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		for key, value := range headerMap {
			request.Header.Set(key, value)
		}

		return request, nil
	})
}

// Accept 将 "Accept" 头部应用于 http 请求的 header 中
func Accept(accept string) Option {
	return Header(xhttp.HeaderAccept, accept)
}

// Authorization 将 "Authorization" 头部应用于 http 请求的 header 中
func Authorization(authorization string) Option {
	return Header(xhttp.HeaderAuthorization, authorization)
}

// CacheControl 将 "Cache-Control" 头部应用于 http 请求的 header 中
func CacheControl(cacheControl string) Option {
	return Header(xhttp.HeaderCacheControl, cacheControl)
}

// ContentLength 将 contentLength 应用于 http 请求中
func ContentLength(contentLength int) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.ContentLength = int64(contentLength)

		return request, nil
	})
}

// ContentType 将 "Content-Type" 头部应用于 http 请求的 header 中
func ContentType(contentType string) Option {
	return Header(xhttp.HeaderContentType, contentType)
}

// Referer 将 "Referer" 头部应用于 http 请求的 header 中
func Referer(referer string) Option {
	return Header(xhttp.HeaderReferer, referer)
}

// UserAgent 将 "User-Agent" 头部应用于 http 请求的 header 中
func UserAgent(userAgent string) Option {
	return Header(xhttp.HeaderUserAgent, userAgent)
}

// AddCookies 将 *http.Cookie 列表添加到 http 请求中
func AddCookies(cookies ...*http.Cookie) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}

		return request, nil
	})
}

// BasicAuth 构建 "Authorization: Basic <base64Encode(username:password)>" 头部并应用于 http 请求的 header 中
func BasicAuth(username, password string) Option {
	auth := username + ":" + password
	token := base64.StdEncoding.EncodeToString([]byte(auth))

	return Authorization("Basic " + token)
}

// BearerAuth 构建 "Authorization: Bearer <token>" 头部并应用于 http 请求的 header 中
func BearerAuth(token string) Option {
	return Authorization("Bearer " + token)
}

// TokenAuth 构建 "Authorization: Token <token>" 头部并应用于 http 请求的 header 中
func TokenAuth(token string) Option {
	return Authorization("Token " + token)
}

// Context 将 context.Context 应用于 http 请求中
func Context(ctx context.Context) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		return request.WithContext(ctx), nil
	})
}

// ContextValue 将 key 和 value 应用于 http 请求的 context 中
func ContextValue(key, value any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		return request.WithContext(context.WithValue(request.Context(), key, value)), nil
	})
}

// Body 将 io.ReadCloser 应用于 http 请求的 body 中
func Body(body io.ReadCloser) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		request.Body = body
		if nc, ok := body.(nopCloser); ok && request.ContentLength == 0 {
			request.ContentLength = nc.Size()
		}

		return request, nil
	})
}

// BodyReader 将 io.Reader 应用于 http 请求的 body 中
func BodyReader(body io.Reader) Option {
	readCloser, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		readCloser = rc(body)
	}

	return Body(readCloser)
}

// BodyBytes 将 []byte 应用于 http 请求的 body 中
func BodyBytes(body []byte) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		return Apply(request,
			BodyReader(bytes.NewBuffer(body)),
			// 设置 ContentLength 和 GetBody
			OptionFunc(func(request *http.Request) (*http.Request, error) {
				request.ContentLength = int64(len(body))
				request.GetBody = func() (io.ReadCloser, error) {
					return rc(bytes.NewReader(body)), nil
				}

				return request, nil
			}),
		)
	})
}

// BodyString 将 string 应用于 http 请求的 body 中
func BodyString(body string) Option {
	return BodyBytes([]byte(body))
}

// BodyForm 将 url.Values 应用于 http 请求的 body 中，并设置 "Content-Type" 头部为 "application/x-www-form-urlencoded"
func BodyForm(body stdurl.Values) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		return Apply(request,
			ContentType(xhttp.MIMEForm),
			BodyString(body.Encode()),
		)
	})
}

// BodyFormMap 将 map[string]any 应用于 http 请求的 body 中，并设置 "Content-Type" 头部为 "application/x-www-form-urlencoded"
func BodyFormMap(formMap map[string]any) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		body := make(stdurl.Values)
		for key, value := range formMap {
			body.Set(key, convert.ToString(value))
		}

		return Apply(request, BodyForm(body))
	})
}

// BodyJSON 使用 Marshaler 将 obj 序列化为 json 格式应用于 http 请求的 body 中，并设置 "Content-Type" 头部为 "application/json"
func BodyJSON(obj any, marshaler ...Marshaler) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		jsonMarshal := json.Marshal
		if len(marshaler) > 0 && marshaler[0] != nil {
			jsonMarshal = marshaler[0]
		}

		body, err := jsonMarshal(obj)
		if err != nil {
			return nil, errors.WithMessagef(err, "json marshal obj: %v err", obj)
		}

		return Apply(request,
			ContentType(xhttp.MIMEApplicationJSON),
			BodyBytes(body),
		)
	})
}

// BodyXML 使用 Marshaler 将 obj 序列化为 xml 格式应用于 http 请求的 body 中，并设置 "Content-Type" 头部为 "application/xml"
func BodyXML(obj any, marshaler ...Marshaler) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		xmlMarshal := xml.Marshal
		if len(marshaler) > 0 && marshaler[0] != nil {
			xmlMarshal = marshaler[0]
		}

		body, err := xmlMarshal(obj)
		if err != nil {
			return nil, errors.WithMessagef(err, "xml marshal obj: %v err", obj)
		}

		return Apply(request,
			ContentType(xhttp.MIMEApplicationXML),
			BodyBytes(body),
		)
	})
}

// Dump 转储 http 请求信息并将其写入 io.Writer 中
func Dump(w io.Writer) Option {
	return OptionFunc(func(request *http.Request) (*http.Request, error) {
		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			return nil, errors.WithMessage(err, "dump request err")
		}

		if _, err := w.Write(dump); err != nil {
			return nil, errors.WithMessage(err, "write dump err")
		}

		return request, nil
	})
}

var _ io.ReadCloser = nopCloser{}

type nopCloser struct {
	io.Reader
}

// rc 与 io.NopCloser() 函数类似，实现该方法是为了能便于拿到底层的 io.Reader
func rc(r io.Reader) nopCloser {
	return nopCloser{r}
}

func (nopCloser) Close() error { return nil }

func (nc nopCloser) Size() int64 {
	l, _ := xhttp.GetReaderLen(nc.Reader)

	return l
}

func getOptionName(o Option) string {
	t := reflect.TypeOf(o)
	selfPkgName := "github.com/sliveryou/micro-pkg/xhttp/xreq"

	if t.Kind() == reflect.Func {
		n := runtime.FuncForPC(reflect.ValueOf(o).Pointer()).Name()
		// 去除自身包前缀
		n = strings.TrimPrefix(n, selfPkgName+".")
		// 去除匿名函数后缀
		n = strings.TrimSuffix(n, ".func1")

		return n
	} else if pkgName := t.PkgPath(); pkgName != "" && pkgName != selfPkgName {
		return pkgName + "." + t.Name()
	}

	return t.Name()
}

func copyAndAppend[T any](olds []T, withs ...T) []T {
	news := make([]T, len(olds), len(olds)+len(withs))
	copy(news, olds)

	return append(news, withs...)
}
