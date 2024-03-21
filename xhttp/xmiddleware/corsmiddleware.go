package xmiddleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/sliveryou/micro-pkg/xhttp"
)

// -------------------- CORSMiddleware -------------------- //

const (
	wildcard  = "*"
	sep       = ", "
	valueTrue = "true"
	valueZero = "0"
)

// CORSConfig 跨域请求处理配置
type CORSConfig struct {
	// 跨域请求包含预检请求（preflight request）和实际请求（actual request）
	// 当客户端发送的请求不为简单请求（simple request）时，浏览器会向服务端先发送一个预检请求，以获知服务端是否允许该实际请求，
	// 预检请求的使用，可以避免跨域请求对服务端的数据产生未预期的影响
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS

	// AllowCredentials 是否允许携带认证信息，如 cookies 或 tls certificates
	//
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials bool

	// AllowHeaders 允许头部列表，用于校验预检请求中 Access-Control-Request-Headers 头部记录的头部信息
	//
	// 为空时，将使用 []string{xhttp.HeaderAuthorization, xhttp.HeaderContentType, xhttp.HeaderXRequestedWith} 列表信息
	//
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowHeaders []string

	// AllowMethods 允许方法列表，用于校验预检请求中 Access-Control-Request-Method 头部记录的方法信息
	//
	// 为空时，将使用 []string{xhttp.MethodGet, xhttp.MethodPost, xhttp.MethodHead} 列表信息；包含通配符 "*" 时，表示允许任意方法
	//
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowMethods []string

	// AllowOrigins 允许来源列表，用于校验跨域请求中 Origin 头部记录的来源信息
	//
	// 为空或者在列表中包含通配符 "*" 时，表示允许任意来源
	//
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AllowOrigins []string

	// ExposeHeaders 允许暴露头部列表，用于给客户端获取访问跨域响应默认头部定义之外的头部信息
	//
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	ExposeHeaders []string

	// MaxAge 缓存的最大时间（秒），用于浏览器缓存预检请求的返回结果
	// 详情：https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge time.Duration

	// Debug 调试标记，启用时将会打印错误日志
	Debug bool

	// UnsafeWildcardOriginWithAllowCredentials UNSAFE/INSECURE: 允许通配符 "*" 任意来源与 AllowCredentials 一起使用
	//
	// 正常情况下，当服务端设置响应头部 Access-Control-Allow-Origin 为 "*"，Access-Control-Allow-Credentials 为 "true"
	// 时，浏览器执行请求将会失败，以规避跨域请求对服务端的数据产生不良影响，而开启 UnsafeWildcardOriginWithAllowCredentials
	// 后，将会根据请求的中的 Origin 头部值设置响应的 Access-Control-Allow-Origin 头部值，而不是设置成通配符 "*"
	// 以让浏览器能继续发送客户端请求
	//
	// 注意：这是非常不安全的，也引起了 https://portswigger.net/research/exploiting-cors-misconfigurations-for-bitcoins-and-bounties
	// 文章中提到的重大安全问题，所以不建议开启
	UnsafeWildcardOriginWithAllowCredentials bool

	// AllowOriginFunc 自定义来源校验函数，它将 origin 作为参数，如果允许该 origin 则返回 true，否则返回 false
	//
	// 当设置了自定义来源校验函数时，字段 AllowOrigins 将会失效
	AllowOriginFunc func(origin string) bool
}

// DefaultCORSConfig 获取默认跨域请求处理配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowCredentials: false,
		AllowHeaders:     []string{xhttp.HeaderAccessToken, xhttp.HeaderAuthorization, xhttp.HeaderContentType, xhttp.HeaderRange, xhttp.HeaderToken, xhttp.HeaderXCSRFToken, xhttp.HeaderXHealthSecret, xhttp.HeaderXRequestedWith},
		AllowMethods:     []string{xhttp.MethodGet, xhttp.MethodPost, xhttp.MethodPut, xhttp.MethodPatch, xhttp.MethodDelete, xhttp.MethodHead},
		AllowOrigins:     []string{wildcard},
		ExposeHeaders:    []string{xhttp.HeaderContentDisposition, xhttp.HeaderContentEncoding, xhttp.HeaderCaErrorCode, xhttp.HeaderCaErrorMessage},
		MaxAge:           12 * time.Hour,
	}
}

// AllowAllCORSConfig 获取允许全部跨域请求处理配置
func AllowAllCORSConfig() CORSConfig {
	c := DefaultCORSConfig()
	c.AllowHeaders = []string{wildcard}
	c.AllowMethods = []string{wildcard}

	return c
}

// UnsafeAllowAllCORSConfig 获取不安全的允许全部跨域请求处理配置
func UnsafeAllowAllCORSConfig() CORSConfig {
	c := AllowAllCORSConfig()
	c.AllowCredentials = true
	c.UnsafeWildcardOriginWithAllowCredentials = true

	return c
}

// CORSMiddleware 跨域请求处理中间件
type CORSMiddleware struct {
	allowCredentials    bool
	allowHeaders        []string
	allowHeadersAll     bool
	allowMethods        []string
	allowMethodsAll     bool
	allowOrigins        []string
	allowOriginsAll     bool
	allowOriginFunc     func(origin string) bool
	allowOriginPatterns []string
	exposeHeaders       []string
	maxAge              string
	debug               bool

	unsafeWildcardOriginWithAllowCredentials bool
}

// AllowAllCORSMiddleware 允许全部跨域请求处理中间件
func AllowAllCORSMiddleware() *CORSMiddleware {
	return NewCORSMiddleware(AllowAllCORSConfig())
}

// UnsafeAllowAllCORSMiddleware 不安全的允许全部跨域请求处理中间件
func UnsafeAllowAllCORSMiddleware() *CORSMiddleware {
	return NewCORSMiddleware(UnsafeAllowAllCORSConfig())
}

// NewCORSMiddleware 新建跨域请求处理中间件（不传递配置时，将使用默认配置 DefaultCORSConfig）
// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS
func NewCORSMiddleware(config ...CORSConfig) *CORSMiddleware {
	c := DefaultCORSConfig()
	if len(config) > 0 {
		c = config[0]
	}

	m := &CORSMiddleware{
		allowCredentials:                         c.AllowCredentials,
		debug:                                    c.Debug,
		unsafeWildcardOriginWithAllowCredentials: c.UnsafeWildcardOriginWithAllowCredentials,
	}

	if len(c.AllowHeaders) == 0 {
		m.allowHeaders = []string{xhttp.HeaderAuthorization, xhttp.HeaderContentType, xhttp.HeaderXRequestedWith}
	} else {
		for _, h := range c.AllowHeaders {
			if h == wildcard {
				m.allowHeadersAll = true
				break
			}
		}
		if !m.allowHeadersAll {
			m.allowHeaders = convertStrings(c.AllowHeaders, http.CanonicalHeaderKey)
		}
	}

	if len(c.AllowMethods) == 0 {
		m.allowMethods = []string{xhttp.MethodGet, xhttp.MethodPost, xhttp.MethodHead}
	} else {
		for _, am := range c.AllowMethods {
			if am == wildcard {
				m.allowMethodsAll = true
				break
			}
		}
		if !m.allowMethodsAll {
			m.allowMethods = convertStrings(c.AllowMethods, strings.ToUpper)
		}
	}

	if c.AllowOriginFunc != nil {
		m.allowOriginFunc = c.AllowOriginFunc
	} else if len(c.AllowOrigins) == 0 {
		m.allowOriginsAll = true
	} else {
		for _, origin := range c.AllowOrigins {
			if origin == wildcard {
				m.allowOriginsAll = true
				m.allowOriginPatterns = nil
				break
			}

			pattern := regexp.QuoteMeta(origin)
			pattern = strings.ReplaceAll(pattern, "\\*", ".*")
			pattern = strings.ReplaceAll(pattern, "\\?", ".")

			m.allowOrigins = append(m.allowOrigins, origin)
			m.allowOriginPatterns = append(m.allowOriginPatterns, "^"+pattern+"$")
		}
	}

	if len(c.ExposeHeaders) > 0 {
		m.exposeHeaders = []string{strings.Join(convertStrings(c.ExposeHeaders, http.CanonicalHeaderKey), sep)}
	}

	if c.MaxAge > 0 {
		m.maxAge = strconv.Itoa(int(c.MaxAge.Seconds()))
	} else if c.MaxAge < 0 {
		m.maxAge = valueZero
	}

	return m
}

// Handle 跨域请求处理
func (m *CORSMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 跨域预检请求
		if r.Method == xhttp.MethodOptions && r.Header.Get(xhttp.HeaderAccessControlRequestMethod) != "" {
			m.handlePreflight(w, r)
			w.WriteHeader(http.StatusNoContent)
		} else { // 跨域实际请求
			m.handleActual(w, r)
			next(w, r)
		}
	}
}

// Handler 跨域请求处理器
func (m *CORSMiddleware) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 跨域预检请求
		if r.Method == xhttp.MethodOptions && r.Header.Get(xhttp.HeaderAccessControlRequestMethod) != "" {
			m.handlePreflight(w, r)
			w.WriteHeader(http.StatusNoContent)
		} else { // 跨域实际请求
			m.handleActual(w, r)
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

// handlePreflight 处理预检请求 https://developer.mozilla.org/zh-CN/docs/Glossary/Preflight_request
func (m CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request) {
	if err := m.dealWithPreflight(w, r); err != nil && m.debug {
		logx.WithContext(r.Context()).Errorf("cors middleware handle preflight request err: %v", err)
	}
}

// handleActual 处理实际请求 https://developer.mozilla.org/zh-CN/docs/Web/HTTP/CORS
func (m *CORSMiddleware) handleActual(w http.ResponseWriter, r *http.Request) {
	if err := m.dealWithActual(w, r); err != nil && m.debug {
		logx.WithContext(r.Context()).Errorf("cors middleware handle actual request err: %v", err)
	}
}

func (m CORSMiddleware) dealWithPreflight(w http.ResponseWriter, r *http.Request) error {
	if r.Method != xhttp.MethodOptions {
		return errors.Errorf("method %s not equaled with OPTIONS", r.Method)
	}

	headers := w.Header()
	headers.Add(xhttp.HeaderVary, xhttp.HeaderOrigin)

	origin := r.Header.Get(xhttp.HeaderOrigin)
	if origin == "" {
		return errors.New("empty origin not allowed")
	}
	// 判断请求来源是否允许
	if isAllowed := m.isOriginAllowed(origin); !isAllowed {
		return errors.Errorf("origin %s not allowed", origin)
	}

	reqMethod := strings.ToUpper(r.Header.Get(xhttp.HeaderAccessControlRequestMethod))
	// 判断请求方法是否允许
	if !m.isMethodAllowed(reqMethod) {
		return errors.Errorf("method %s not allowed", r.Method)
	}

	reqHeadersRaw := r.Header[xhttp.HeaderAccessControlRequestHeaders]
	reqHeaders := splitHeaderValues(reqHeadersRaw)
	// 判断请求头列表是否允许
	if !m.areHeadersAllowed(reqHeaders) {
		return errors.Errorf("headers %v not allowed", reqHeaders)
	}

	headers.Add(xhttp.HeaderVary, xhttp.HeaderAccessControlRequestHeaders)
	headers.Add(xhttp.HeaderVary, xhttp.HeaderAccessControlRequestMethod)

	if m.allowCredentials {
		headers.Set(xhttp.HeaderAccessControlAllowCredentials, valueTrue)
	}

	if len(reqHeaders) > 0 {
		headers.Set(xhttp.HeaderAccessControlAllowHeaders, strings.Join(reqHeaders, sep))
	}

	headers.Set(xhttp.HeaderAccessControlAllowMethods, reqMethod)

	if m.allowOriginsAll && !m.unsafeWildcardOriginWithAllowCredentials {
		headers.Set(xhttp.HeaderAccessControlAllowOrigin, wildcard)
	} else {
		headers.Set(xhttp.HeaderAccessControlAllowOrigin, origin)
	}

	if m.maxAge != "" {
		headers.Set(xhttp.HeaderAccessControlMaxAge, m.maxAge)
	}

	return nil
}

func (m *CORSMiddleware) dealWithActual(w http.ResponseWriter, r *http.Request) error {
	headers := w.Header()
	headers.Add(xhttp.HeaderVary, xhttp.HeaderOrigin)

	origin := r.Header.Get(xhttp.HeaderOrigin)
	if origin == "" {
		return errors.New("empty origin not allowed")
	}
	// 判断请求来源是否允许
	if isAllowed := m.isOriginAllowed(origin); !isAllowed {
		return errors.Errorf("origin %s not allowed", origin)
	}

	// 判断请求方法是否允许
	if !m.isMethodAllowed(r.Method) {
		return errors.Errorf("method %s not allowed", r.Method)
	}

	if m.allowCredentials {
		headers.Set(xhttp.HeaderAccessControlAllowCredentials, valueTrue)
	}

	if m.allowOriginsAll && !m.unsafeWildcardOriginWithAllowCredentials {
		headers.Set(xhttp.HeaderAccessControlAllowOrigin, wildcard)
	} else {
		headers.Set(xhttp.HeaderAccessControlAllowOrigin, origin)
	}

	if len(m.exposeHeaders) > 0 {
		headers[xhttp.HeaderAccessControlExposeHeaders] = m.exposeHeaders
	}

	return nil
}

// isOriginAllowed 判断请求来源是否允许
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	if m.allowOriginFunc != nil {
		return m.allowOriginFunc(origin)
	}

	if m.allowOriginsAll {
		return true
	}

	for _, o := range m.allowOrigins {
		if o == origin {
			return true
		}
	}

	for _, re := range m.allowOriginPatterns {
		if match, _ := regexp.MatchString(re, origin); match {
			return true
		}
	}

	return false
}

// isMethodAllowed 判断请求方法是否允许
func (m *CORSMiddleware) isMethodAllowed(method string) bool {
	if m.allowMethodsAll {
		return true
	}
	if len(m.allowMethods) == 0 {
		return false
	}

	// 允许预检请求
	if method == xhttp.MethodOptions {
		return true
	}

	for _, allowMethod := range m.allowMethods {
		if method == allowMethod {
			return true
		}
	}

	return false
}

// areHeadersAllowed 判断请求头列表是否允许
func (m *CORSMiddleware) areHeadersAllowed(headers []string) bool {
	if m.allowHeadersAll || len(headers) == 0 {
		return true
	}

	for _, header := range headers {
		found := false
		for _, h := range m.allowHeaders {
			if h == header {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func splitHeaderValues(values []string) []string {
	var out []string
	for _, v := range values {
		needsSplit := strings.IndexByte(v, ',') != -1
		if needsSplit {
			split := strings.Split(v, ",")
			for _, s := range split {
				out = append(out, http.CanonicalHeaderKey(strings.TrimSpace(s)))
			}
		} else {
			out = append(out, http.CanonicalHeaderKey(v))
		}
	}

	return out
}

type converter func(string) string

func convertStrings(s []string, c converter) []string {
	out := make([]string, 0, len(s))
	for _, v := range s {
		out = append(out, c(v))
	}

	return out
}
