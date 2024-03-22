package appsign

import "github.com/sliveryou/micro-pkg/internal/bizerr"

var (
	// ErrInvalidSignParams 签名参数错误
	ErrInvalidSignParams = bizerr.ErrInvalidSignParams
	// ErrInvalidContentMD5 Content-MD5 错误
	ErrInvalidContentMD5 = bizerr.ErrInvalidContentMD5
	// ErrBodyTooLarge 请求体过大错误
	ErrBodyTooLarge = bizerr.ErrBodyTooLarge
)
