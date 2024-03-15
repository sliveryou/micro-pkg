package appsign

// 请求头
const (
	HeaderAccept      = "Accept"
	HeaderContentMD5  = "Content-Md5"
	HeaderContentType = "Content-Type"
	HeaderDate        = "Date"
	HeaderUserAgent   = "User-Agent"

	HeaderCAKey              = "X-Ca-Key"
	HeaderCANonce            = "X-Ca-Nonce"
	HeaderCASignature        = "X-Ca-Signature"
	HeaderCASignatureHeaders = "X-Ca-Signature-Headers"
	HeaderCASignatureMethod  = "X-Ca-Signature-Method"
	HeaderCATimestamp        = "X-Ca-Timestamp"
)

// 签名算法
const (
	SignatureMethodHmacSHA256 = "HmacSHA256"
	SignatureMethodHmacSHA1   = "HmacSHA1"
)
