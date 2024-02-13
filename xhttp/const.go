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
	// HeaderContentType 请求头：Content-Type
	HeaderContentType = "Content-Type"
	// HeaderDate 请求头：Date
	HeaderDate = "Date"
	// HeaderUserAgent 请求头：User-Agent
	HeaderUserAgent = "User-Agent"
	// HeaderAuthorization 请求头：Authorization
	HeaderAuthorization = "Authorization"
	// HeaderLocation 请求头：Location
	HeaderLocation = "Location"
	// HeaderContentDisposition 请求头：Content-Disposition
	HeaderContentDisposition = "Content-Disposition"

	// HeaderGWErrorCode 自定义网关请求头：X-GW-Error-Code
	HeaderGWErrorCode = "X-GW-Error-Code"
	// HeaderGWErrorMessage 自定义网关请求头：X-GW-Error-Message
	HeaderGWErrorMessage = "X-GW-Error-Message"

	// ApplicationForm 应用类型：x-www-form-urlencoded
	ApplicationForm = "application/x-www-form-urlencoded"
	// ApplicationStream 应用类型：octet-stream
	ApplicationStream = "application/octet-stream"
	// ApplicationJSON 应用类型：json
	ApplicationJSON = "application/json"
	// ApplicationXML 应用类型：xml
	ApplicationXML = "application/xml"
	// ApplicationText 应用类型：text
	ApplicationText = "application/text"
	// ApplicationZip 应用类型：zip
	ApplicationZip = "application/zip"
)
