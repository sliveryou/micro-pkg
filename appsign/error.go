package appsign

import "github.com/sliveryou/micro-pkg/errcode"

var (
	// ErrInvalidSignParams 签名参数错误
	ErrInvalidSignParams = errcode.NewCommon("签名参数错误")
	// ErrInvalidContentMD5 Content-MD5 计算错误
	ErrInvalidContentMD5 = errcode.NewCommon("Content-MD5 计算错误")
	// ErrBodyTooLarge 请求体过大错误
	ErrBodyTooLarge = errcode.NewCommon("请求体过大")
)
