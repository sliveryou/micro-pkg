package bizerr

import (
	"net/http"

	"github.com/sliveryou/micro-pkg/errcode"
)

// errcode 包预定义错误
var (
	// OK 成功
	OK = errcode.OK
	// ErrCommon 通用错误
	ErrCommon = errcode.ErrCommon
	// ErrRecordNotFound 记录不存在错误
	ErrRecordNotFound = errcode.ErrRecordNotFound
	// ErrUnexpected 意外错误
	ErrUnexpected = errcode.ErrUnexpected
	// ErrInvalidParams 请求参数错误
	ErrInvalidParams = errcode.ErrInvalidParams
)

// appsign 包预定义错误
var (
	// ErrInvalidSignParams 签名参数错误
	ErrInvalidSignParams = errcode.New(110, "签名参数错误")
	// ErrInvalidContentMD5 Content-MD5 错误
	ErrInvalidContentMD5 = errcode.New(111, "Content-MD5 错误")
	// ErrBodyTooLarge 请求体过大错误
	ErrBodyTooLarge = errcode.New(112, "请求体过大")
)

// auth 包预定义错误
const (
	// CodeBankCard4CAuth 银行卡四要素认证失败错误业务状态码
	CodeBankCard4CAuth = 115

	// CodeCorpAccountAuth 企业银行卡账户认证失败错误业务状态码
	CodeCorpAccountAuth = 116

	// CodeFaceVideoAuth 视频活体检测失败错误业务状态码
	CodeFaceVideoAuth = 117
	// CodeFacePersonAuth 人脸实名认证失败错误业务状态码
	CodeFacePersonAuth = 118
)

var (
	// ErrBankCard4CAuth 银行卡四要素认证失败错误
	ErrBankCard4CAuth = errcode.New(CodeBankCard4CAuth, "银行卡四要素认证失败")

	// ErrCorpAccountAuth 企业银行卡账户认证失败错误
	ErrCorpAccountAuth = errcode.New(CodeCorpAccountAuth, "企业银行卡账户认证失败")

	// ErrFaceVideoAuth 视频活体检测失败错误
	ErrFaceVideoAuth = errcode.New(CodeFaceVideoAuth, "视频活体检测失败")
	// ErrFacePersonAuth 人脸实名认证失败错误
	ErrFacePersonAuth = errcode.New(CodeFacePersonAuth, "人脸与公民身份证小图相似度匹配过低")
)

// express 包预定义错误

// CodeGetExpressFailed 查询快递失败错误业务状态码
const CodeGetExpressFailed = 120

// ErrGetExpressFailed 查询快递失败错误
var ErrGetExpressFailed = errcode.New(CodeGetExpressFailed, "查询快递失败")

// notify 包预定义错误
var (
	// ErrProviderOverQuota 提供方请求超过配额错误
	ErrProviderOverQuota = errcode.New(130, "该提供方今日请求超过配额，请明日再试")
	// ErrIPSourceOverQuota IP 地址请求超过配额错误
	ErrIPSourceOverQuota = errcode.New(131, "该 IP 地址今日请求超过配额，请明日再试")
	// ErrReceiverOverQuota 接收方请求超过配额错误
	ErrReceiverOverQuota = errcode.New(132, "该接收方今日请求超过配额，请明日再试")

	// ErrSendTooFrequently 发送过于频繁错误
	ErrSendTooFrequently = errcode.New(133, "发送过于频繁，请稍后再试")
	// ErrVerifyTooFrequently 验证过于频繁错误
	ErrVerifyTooFrequently = errcode.New(134, "验证过于频繁，请稍后再试")

	// ErrEmailUnsupported 暂不支持邮件通知服务错误
	ErrEmailUnsupported = errcode.New(135, "暂不支持邮件通知服务")
	// ErrSmsUnsupported 暂不支持短信通知服务错误
	ErrSmsUnsupported = errcode.New(136, "暂不支持短信通知服务")

	// ErrEmailTmplNotFound 邮件模板不存在错误
	ErrEmailTmplNotFound = errcode.New(137, "邮件模板信息不存在")
	// ErrSmsTmplNotFound 短信模板不存在错误
	ErrSmsTmplNotFound = errcode.New(138, "短信模板信息不存在")

	// ErrInvalidCaptcha 验证码错误
	ErrInvalidCaptcha = errcode.New(139, "验证码错误")
	// ErrCaptchaNotFound 验证码不存在或已过期错误
	ErrCaptchaNotFound = errcode.New(140, "验证码不存在或已过期")
)

// xhttp/xmiddleware 和 xgrpc/xinterceptor 包预定义错误

// CodeInvalidSign 签名错误业务状态码
const CodeInvalidSign = 150

var (
	// ErrInvalidSign 签名错误
	ErrInvalidSign = errcode.New(CodeInvalidSign, "签名错误", http.StatusUnauthorized)
	// ErrSignExpired 签名已过期错误
	ErrSignExpired = errcode.New(151, "签名已过期", http.StatusUnauthorized)
	// ErrNonceExpired 随机数已过期错误
	ErrNonceExpired = errcode.New(152, "随机数已过期", http.StatusUnauthorized)

	// ErrInvalidToken Token 错误
	ErrInvalidToken = errcode.New(153, "Token 错误", http.StatusUnauthorized)

	// ErrAPINotAllowed 暂不支持该 API 错误
	ErrAPINotAllowed = errcode.New(154, "暂不支持该 API")
	// ErrRPCNotAllowed 暂不支持该 RPC 错误
	ErrRPCNotAllowed = errcode.New(155, "暂不支持该 RPC")
)
