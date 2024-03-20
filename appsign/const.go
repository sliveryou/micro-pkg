package appsign

// 请求头
const (
	HeaderAccept      = "Accept"
	HeaderContentMD5  = "Content-Md5"
	HeaderContentType = "Content-Type"
	HeaderDate        = "Date"

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

// MIME 类型
const (
	MIMEForm                      = "application/x-www-form-urlencoded"
	MIMEMultipartForm             = "multipart/form-data"
	MIMEMultipartFormWithBoundary = MIMEMultipartForm + "; boundary="
)

// 默认值
const (
	defaultLF  = "\n"
	defaultSep = ","
	maxBodyLen = 8 << 20 // 8 MB
)
