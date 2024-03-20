package appsign

import (
	"crypto/hmac"
	stdmd5 "crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"

	"github.com/sliveryou/micro-pkg/xhttp"
)

var notSignHeaders = map[string]struct{}{
	http.CanonicalHeaderKey(HeaderAccept):             {},
	http.CanonicalHeaderKey(HeaderContentMD5):         {},
	http.CanonicalHeaderKey(HeaderContentType):        {},
	http.CanonicalHeaderKey(HeaderDate):               {},
	http.CanonicalHeaderKey(HeaderCASignature):        {},
	http.CanonicalHeaderKey(HeaderCASignatureHeaders): {},
}

var signMethods = []string{
	"", SignatureMethodHmacSHA256, SignatureMethodHmacSHA1,
}

// getParams 获取请求参数
func getParams(r *http.Request) (string, error) {
	clone, err := xhttp.CopyRequest(r, maxBodyLen)
	if err != nil {
		return "", errors.WithMessage(err, "copy request err")
	}

	if err := clone.ParseForm(); err != nil {
		return "", errors.WithMessage(err, "parse form err")
	}

	paramKeys := make([]string, 0, len(clone.Form))
	for key := range clone.Form {
		paramKeys = append(paramKeys, key)
	}
	sort.Strings(paramKeys)

	paramList := make([]string, 0)
	for _, key := range paramKeys {
		value := clone.Form.Get(key)
		if value == "" {
			paramList = append(paramList, key)
		} else {
			paramList = append(paramList, key+"="+value)
		}
	}

	params := strings.Join(paramList, "&")
	if params != "" {
		params = "?" + params
	}

	return clone.URL.Path + params, nil
}

func md5(b []byte) string {
	m := stdmd5.New()
	m.Write(b)

	return base64.StdEncoding.EncodeToString(m.Sum(nil))
}

func hmacSHA1(b, key []byte) string {
	h := hmac.New(sha1.New, key)
	h.Write(b)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func hmacSHA256(b, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(b)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
