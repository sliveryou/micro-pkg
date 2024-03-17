package xhttp

import "net/http"

// 协议前缀
const (
	SchemeHTTPPrefix  = "http://"
	SchemeHTTPSPrefix = "https://"
)

// HTTP 方法
const (
	MethodGet     = http.MethodGet
	MethodHead    = http.MethodHead
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodConnect = http.MethodConnect
	MethodOptions = http.MethodOptions
	MethodTrace   = http.MethodTrace
)

// HTTP 头部
const (
	HeaderAccept               = "Accept"
	HeaderAcceptEncoding       = "Accept-Encoding"
	HeaderAcceptLanguage       = "Accept-Language"
	HeaderAccessToken          = "Access-Token"
	HeaderAllow                = "Allow"
	HeaderAuthorization        = "Authorization"
	HeaderCacheControl         = "Cache-Control"
	HeaderContentDisposition   = "Content-Disposition"
	HeaderContentEncoding      = "Content-Encoding"
	HeaderContentLength        = "Content-Length"
	HeaderContentType          = "Content-Type"
	HeaderDate                 = "Date"
	HeaderHost                 = "Host"
	HeaderLocation             = "Location"
	HeaderOrigin               = "Origin"
	HeaderRange                = "Range"
	HeaderReferer              = "Referer"
	HeaderToken                = "Token"
	HeaderUserAgent            = "User-Agent"
	HeaderVary                 = "Vary"
	HeaderXAppEngineRemoteAddr = "X-Appengine-Remote-Addr"
	HeaderXCSRFToken           = "X-Csrf-Token"
	HeaderXForwardedFor        = "X-Forwarded-For"
	HeaderXHealthSecret        = "X-Health-Secret"
	HeaderXRealIP              = "X-Real-Ip"
	HeaderXRequestedWith       = "X-Requested-With"

	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"

	HeaderCaErrorCode    = "X-Ca-Error-Code"
	HeaderCaErrorMessage = "X-Ca-Error-Message"
)

// MIME 类型
const (
	MIMEForm                      = "application/x-www-form-urlencoded"
	MIMEMultipartForm             = "multipart/form-data"
	MIMEMultipartFormWithBoundary = MIMEMultipartForm + "; boundary="
	MIMETextPlain                 = "text/plain"
	MIMEApplicationJSON           = "application/json"
	MIMEApplicationXML            = "application/xml"
	MIMEApplicationPDF            = "application/pdf"
	MIMEApplicationZip            = "application/zip"
	MIMEOctetStream               = "application/octet-stream"
)
