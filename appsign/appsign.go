package appsign

// AppSign 应用签名结构详情
type AppSign struct {
	Accept           string
	ContentMD5       string
	ContentType      string
	Date             string
	UserAgent        string
	Key              string
	Nonce            string
	Signature        string
	SignatureHeaders []string
	SignatureMethod  string
	Timestamp        int64
}
