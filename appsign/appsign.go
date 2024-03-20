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

// AppSign 应用签名结构详情
//
// 签名规则参考：阿里云使用摘要签名认证方式调用 API
// https://help.aliyun.com/zh/api-gateway/user-guide/use-digest-authentication-to-call-an-api
// 的规则并实现
//
// 以下 Header 不参与 Header 签名计算：
//
//	"Accept"
//	"Content-MD5"
//	"Content-Type"
//	"Date"
//	"X-Ca-Signature"
//	"X-Ca-Signature-Headers"
type AppSign struct {
	Method             string            // HTTP 方法，必填
	Accept             string            // "Accept" 请求头的值，非必填
	ContentMD5         string            // "Content-MD5" 请求头的值， 非必填
	ContentType        string            // "Content-Type" 请求头的值，非必填
	Date               string            // "Data" 请求头的值，非必填
	Key                string            // 应用 AppKey，必填
	Nonce              string            // 随机数，必填
	Signature          string            // 签名，必填
	SignatureHeaders   []string          // 所有签名请求头的 slice，非必填
	SignatureHeaderMap map[string]string // 所有签名请求头的 map，非必填
	SignatureMethod    string            // 签名算法，非必填，支持 HmacSHA256 和 HmacSHA1，默认为 HmacSHA256
	Timestamp          int64             // 时间戳，必填
	Params             string            // 参数，必填
}

// FromRequest 从请求中解析应用签名
func FromRequest(r *http.Request) (*AppSign, error) {
	h := r.Header

	params, err := getParams(r)
	if err != nil {
		return nil, errors.WithMessage(err, "get params err")
	}

	splits := strings.Split(h.Get(HeaderCASignatureHeaders), defaultSep)
	for i := range splits {
		splits[i] = strings.TrimSpace(splits[i])
	}

	var shs []string
	shMap := make(map[string]string)
	for _, sh := range splits {
		if _, ok := notSignHeaders[http.CanonicalHeaderKey(sh)]; !ok {
			shs = append(shs, sh)
			shMap[sh] = h.Get(sh)
		}
	}
	sort.Strings(shs)

	as := &AppSign{
		Method:             strings.ToUpper(r.Method),
		Accept:             h.Get(HeaderAccept),
		ContentMD5:         h.Get(HeaderContentMD5),
		ContentType:        h.Get(HeaderContentType),
		Date:               h.Get(HeaderDate),
		Key:                h.Get(HeaderCAKey),
		Nonce:              h.Get(HeaderCANonce),
		Signature:          h.Get(HeaderCASignature),
		SignatureHeaders:   shs,
		SignatureHeaderMap: shMap,
		SignatureMethod:    h.Get(HeaderCASignatureMethod),
		Timestamp:          convert.ToInt64(h.Get(HeaderCATimestamp)),
		Params:             params,
	}

	if err := as.checkContentMD5(r); err != nil {
		return nil, errors.WithMessage(err, "check content md5 err")
	}

	if as.Method == "" || as.Key == "" || as.Nonce == "" ||
		as.Signature == "" || as.Timestamp < 1 ||
		!sliceg.Contain(signMethods, as.SignatureMethod) {
		return nil, ErrInvalidSignParams
	}

	return as, nil
}

// CalcSignString 计算待签名字符串
func (as *AppSign) CalcSignString() string {
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
func (as *AppSign) CheckSign(secret string) (string, error) {
	var sign string
	signStr := as.CalcSignString()

	if as.SignatureMethod == SignatureMethodHmacSHA1 {
		sign = hmacSHA1([]byte(signStr), []byte(secret))
	} else {
		sign = hmacSHA256([]byte(signStr), []byte(secret))
	}
	if as.Signature != sign {
		return sign, ErrSignVerify
	}

	return sign, nil
}

// checkContentMD5 校验 Content-MD5 值
func (as *AppSign) checkContentMD5(r *http.Request) error {
	if as.ContentMD5 != "" && r.Body != nil && as.ContentType != MIMEForm &&
		!strings.HasPrefix(as.ContentType, MIMEMultipartFormWithBoundary) {
		reader := io.LimitReader(r.Body, maxBodyLen)
		b, err := io.ReadAll(reader)
		if err != nil {
			return errors.WithMessage(err, "read all request body err")
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
