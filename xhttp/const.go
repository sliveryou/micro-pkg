package xhttp

import "net/http"

const (
	// MethodGet 请求方法：GET
	MethodGet = http.MethodGet
	// MethodHead 请求方法：HEAD
	MethodHead = http.MethodHead
	// MethodPost 请求方法：POST
	MethodPost = http.MethodPost
	// MethodPut 请求方法：PUT
	MethodPut = http.MethodPut
	// MethodPatch 请求方法：PATCH
	MethodPatch = http.MethodPatch
	// MethodDelete 请求方法：DELETE
	MethodDelete = http.MethodDelete
	// MethodConnect 请求方法：CONNECT
	MethodConnect = http.MethodConnect
	// MethodOptions 请求方法：OPTIONS
	MethodOptions = http.MethodOptions
	// MethodTrace 请求方法：TRACE
	MethodTrace = http.MethodTrace

	// HeaderAccept 请求头：Accept
	HeaderAccept = "Accept"
	// HeaderAcceptLanguage 请求头：Accept-Language
	HeaderAcceptLanguage = "Accept-Language"
	// HeaderContentType 请求头：Content-Type
	HeaderContentType = "Content-Type"
	// HeaderContentDisposition 请求头：Content-Disposition
	HeaderContentDisposition = "Content-Disposition"
	// HeaderDate 请求头：Date
	HeaderDate = "Date"
	// HeaderLocation 请求头：Location
	HeaderLocation = "Location"
	// HeaderUserAgent 请求头：User-Agent
	HeaderUserAgent = "User-Agent"
	// HeaderAuthorization 请求头：Authorization
	HeaderAuthorization = "Authorization"

	// HeaderCaErrorCode 自定义网关请求头：X-Ca-Error-Code
	HeaderCaErrorCode = "X-Ca-Error-Code"
	// HeaderCaErrorMessage 自定义网关请求头：X-Ca-Error-Message
	HeaderCaErrorMessage = "X-Ca-Error-Message"

	// ContentTypeForm 内容类型：x-www-form-urlencoded
	ContentTypeForm = "application/x-www-form-urlencoded"
	// ContentTypeMultipartForm 内容类型：multipart/form-data
	ContentTypeMultipartForm = "multipart/form-data"
	// ContentTypeMultipartFormWithBoundary 内容类型：multipart/form-data; boundary=
	ContentTypeMultipartFormWithBoundary = "multipart/form-data; boundary="
	// ContentTypeText 内容类型：text
	ContentTypeText = "text/plain"
	// ContentTypeJSON 内容类型：json
	ContentTypeJSON = "application/json"
	// ContentTypeXML 内容类型：xml
	ContentTypeXML = "application/xml"
	// ContentTypePDF 内容类型：pdf
	ContentTypePDF = "application/pdf"
	// ContentTypeZip 内容类型：zip
	ContentTypeZip = "application/zip"
	// ContentTypeStream 内容类型：octet-stream
	ContentTypeStream = "application/octet-stream"
)
