package appsign

import (
	"bytes"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/convert"
	"github.com/sliveryou/go-tool/v2/sliceg"
)

// AppSign 应用签名
//
// 签名规则参考：阿里云使用摘要签名认证方式调用 API
// https://help.aliyun.com/zh/api-gateway/user-guide/use-digest-authentication-to-call-an-api
// 的规则并实现
//
// 下列请求头部不参与请求头部签名计算：
//
//	"Accept"、"Content-MD5"、"Content-Type"、"Date"、"X-Ca-Signature" 和 "X-Ca-Signature-Headers"
//
// 下列请求头部是必填项：
//
//	"X-Ca-Key"、"X-Ca-Nonce"、"X-Ca-Signature" 和 "X-Ca-Timestamp"
type AppSign struct {
	Method             string            // 请求方法，必填
	Accept             string            // "Accept" 请求头部的值，非必填
	ContentMD5         string            // "Content-MD5" 请求头部的值， 非必填
	ContentType        string            // "Content-Type" 请求头部的值，非必填
	Date               string            // "Data" 请求头部的值，非必填
	Key                string            // 应用 AppKey，必填
	Nonce              string            // 随机数，必填
	Signature          string            // 签名，必填
	SignatureHeaders   []string          // 所有签名请求头部的 slice，非必填
	SignatureHeaderMap map[string]string // 所有签名请求头部的 map，非必填
	SignatureMethod    string            // 签名算法，非必填，支持 HmacSHA256 和 HmacSHA1，默认为 HmacSHA256
	Timestamp          int64             // 毫秒级时间戳，必填
	Params             string            // 参数，计算得到
	StringToSign       string            // 待签名字符串，计算得到
}

// FromRequest 从请求中解析应用签名
func FromRequest(r *http.Request) (*AppSign, error) {
	h := r.Header

	// 初始化应用签名
	as := &AppSign{
		Method:          strings.ToUpper(r.Method),
		Accept:          h.Get(HeaderAccept),
		ContentMD5:      h.Get(HeaderContentMD5),
		ContentType:     h.Get(HeaderContentType),
		Date:            h.Get(HeaderDate),
		Key:             h.Get(HeaderCAKey),
		Nonce:           h.Get(HeaderCANonce),
		Signature:       h.Get(HeaderCASignature),
		SignatureMethod: h.Get(HeaderCASignatureMethod),
		Timestamp:       convert.ToInt64(h.Get(HeaderCATimestamp)),
	}

	// 必填项校验
	if as.Method == "" || as.Key == "" || as.Nonce == "" ||
		as.Signature == "" || as.Timestamp < 1 ||
		!sliceg.Contain(signMethods, as.SignatureMethod) {
		return nil, ErrInvalidSignParams
	}

	// Content-MD5 值校验
	if err := as.checkContentMD5(r); err != nil {
		return nil, errors.WithMessage(err, "check content md5 err")
	}

	// 获取参数
	params, err := getParams(r)
	if err != nil {
		return nil, errors.WithMessage(err, "get params err")
	}
	as.Params = params

	// 获取所有签名请求头信息
	var shs []string
	shm := make(map[string]string)
	splits := strings.Split(h.Get(HeaderCASignatureHeaders), defaultSep)
	for i := range splits {
		splits[i] = strings.TrimSpace(splits[i])
	}
	for _, sh := range splits {
		if _, ok := notSignHeaders[http.CanonicalHeaderKey(sh)]; !ok {
			shs = append(shs, sh)
			shm[sh] = h.Get(sh)
		}
	}
	sort.Strings(shs)
	as.SignatureHeaders, as.SignatureHeaderMap = shs, shm

	// 计算待签名字符串
	as.StringToSign = as.CalcStringToSign()

	return as, nil
}

// CalcStringToSign 计算待签名字符串
func (as *AppSign) CalcStringToSign() string {
	var s strings.Builder
	s.WriteString(as.Method + defaultLF)
	s.WriteString(as.Accept + defaultLF)
	s.WriteString(as.ContentMD5 + defaultLF)
	s.WriteString(as.ContentType + defaultLF)
	s.WriteString(as.Date + defaultLF)
	for _, sh := range as.SignatureHeaders {
		s.WriteString(sh + ":" + as.SignatureHeaderMap[sh] + defaultLF)
	}
	s.WriteString(as.Params)

	return s.String()
}

// CheckSign 校验签名，并返回正确签名
//
// 建议：即使验签成功，也要根据 Timestamp 和 Nonce 字段进行再次校验，避免重放攻击
func (as *AppSign) CheckSign(secret string) (sign string, ok bool) {
	if as.SignatureMethod == SignatureMethodHmacSHA1 {
		sign = hmacSHA1([]byte(as.StringToSign), []byte(secret))
	} else {
		sign = hmacSHA256([]byte(as.StringToSign), []byte(secret))
	}
	if as.Signature == sign {
		ok = true
	}

	return
}

// checkContentMD5 校验 Content-MD5 值
func (as *AppSign) checkContentMD5(r *http.Request) error {
	if r.Body != nil && as.ContentMD5 != "" && as.ContentType != MIMEForm &&
		!strings.HasPrefix(as.ContentType, MIMEMultipartFormWithBoundary) {
		reader := &io.LimitedReader{R: r.Body, N: maxBodyLen}
		b, err := io.ReadAll(reader)
		if err != nil {
			return errors.WithMessage(err, "read all request body err")
		}
		if reader.N <= 0 {
			return ErrBodyTooLarge
		}
		if err := r.Body.Close(); err != nil {
			return errors.WithMessage(err, "request body close err")
		}
		if md5(b) != as.ContentMD5 {
			return ErrInvalidContentMD5
		}

		r.Body = io.NopCloser(bytes.NewBuffer(b))
	}

	return nil
}
