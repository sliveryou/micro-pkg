package jwt

import (
	stderrors "errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/sliveryou/go-tool/v2/timex"
)

const (
	jwtAudience  = "aud"
	jwtExpire    = "exp"
	jwtID        = "jti"
	jwtIssueAt   = "iat"
	jwtIssuer    = "iss"
	jwtNotBefore = "nbf"
	jwtSubject   = "sub"

	jwtHeaderAlg = "alg"
	jwtAlgHS256  = "HS256"

	defaultExpiration = 72 * time.Hour
)

var (
	standardClaimSet = map[string]struct{}{
		jwtAudience: {}, jwtExpire: {}, jwtID: {}, jwtIssueAt: {}, jwtIssuer: {}, jwtNotBefore: {}, jwtSubject: {},
	}

	errInvalidToken    = stderrors.New("invalid jwt token")
	errNoClaims        = stderrors.New("no token claims")
	errUnsupportedType = stderrors.New("unsupported token type")
	errNoTokenInCtx    = stderrors.New("no token present in context")
)

// Config JWT 配置
type Config struct {
	Issuer     string        // 签发者
	SecretKey  string        // 密钥
	Expiration time.Duration `json:",default=72h"` // 过期时间
}

// JWT 对象
type JWT struct {
	c Config
}

// NewJWT 新建 JWT 对象
func NewJWT(c Config) (*JWT, error) {
	if c.Issuer == "" || c.SecretKey == "" || c.Expiration < 0 {
		return nil, errors.New("jwt: illegal jwt config")
	}
	if c.Expiration == 0 {
		c.Expiration = defaultExpiration
	}

	return &JWT{c: c}, nil
}

// MustNewJWT 新建 JWT 对象
func MustNewJWT(c Config) *JWT {
	j, err := NewJWT(c)
	if err != nil {
		panic(err)
	}

	return j
}

// GenToken 根据给定 token 结构体生成 JWT token
//
// 注意：token 必须为结构体或结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func (j *JWT) GenToken(token any, expiration ...time.Duration) (string, error) {
	if !IsStruct(token) && !IsStructPointer(token) {
		return "", errUnsupportedType
	}

	payloads := make(map[string]any)
	if err := decode(token, &payloads); err != nil {
		return "", errors.WithMessage(err, "decode token to payloads err")
	}

	return j.GenTokenWithPayloads(payloads, expiration...)
}

// GenTokenWithPayloads 根据 payloads 生成 JWT token
func (j *JWT) GenTokenWithPayloads(payloads map[string]any, expiration ...time.Duration) (string, error) {
	et := j.c.Expiration
	if len(expiration) > 0 && expiration[0] > 0 {
		et = expiration[0]
	}

	claims := make(jwt.MapClaims)
	// https://www.iana.org/assignments/jwt/jwt.xhtml
	// 预定义载荷
	now := timex.Now()
	claims[jwtIssuer] = j.c.Issuer         // issuer，签发者
	claims[jwtIssueAt] = now.Unix()        // issued at，签发时间
	claims[jwtNotBefore] = now.Unix()      // not before，生效时间
	claims[jwtExpire] = now.Add(et).Unix() // expiration time，过期时间

	for k, v := range payloads {
		if _, ok := standardClaimSet[k]; !ok {
			claims[k] = v
		}
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ts, err := t.SignedString([]byte(j.c.SecretKey))
	if err != nil {
		return "", errors.WithMessage(err, "sign token err")
	}

	return ts, nil
}

// ParseToken 解析 JWT token，并将其反序列化至指定 token 结构体中
//
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func (j *JWT) ParseToken(tokenString string, token any) error {
	if !IsStructPointer(token) {
		return errUnsupportedType
	}

	payloads, err := j.ParseTokenPayloads(tokenString)
	if err != nil {
		return err
	}

	return errors.WithMessage(decode(payloads, token), "decode payloads to token err")
}

// ParseTokenFromRequest 从请求头解析 JWT token，并将其反序列化至指定 token 结构体中
//
// 注意：token 必须为结构体指针类型，名称以 json tag 对应的名称与 payloads 进行映射
func (j *JWT) ParseTokenFromRequest(r *http.Request, token any) error {
	if !IsStructPointer(token) {
		return errUnsupportedType
	}

	payloads, err := j.ParseTokenPayloadsFromRequest(r)
	if err != nil {
		return err
	}

	return errors.WithMessage(decode(payloads, token), "decode payloads to token err")
}

// ParseTokenPayloads 解析 JWT token，返回 payloads
func (j *JWT) ParseTokenPayloads(tokenString string) (map[string]any, error) {
	token, err := j.newParser().Parse(trimBearerPrefix(tokenString), j.keyFunc())
	if err != nil {
		return nil, errors.WithMessage(err, "parse from token string err")
	}

	payloads, err := extractPayloads(token)
	if err != nil {
		return nil, errors.WithMessage(err, "extract payloads err")
	}

	return payloads, nil
}

// ParseTokenPayloadsFromRequest 从请求头解析 JWT token，返回 payloads
func (j *JWT) ParseTokenPayloadsFromRequest(r *http.Request) (map[string]any, error) {
	token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
		j.keyFunc(), request.WithParser(j.newParser()))
	if err != nil {
		return nil, errors.WithMessage(err, "parse from request err")
	}

	payloads, err := extractPayloads(token)
	if err != nil {
		return nil, errors.WithMessage(err, "extract payloads err")
	}

	return payloads, nil
}

// newParser 新建 JWT 解析器
func (j *JWT) newParser() *jwt.Parser {
	return jwt.NewParser(
		jwt.WithIssuer(j.c.Issuer),
		jwt.WithValidMethods([]string{jwtAlgHS256}),
		jwt.WithExpirationRequired(),
		jwt.WithJSONNumber(),
	)
}

// keyFunc JWT 签名密钥函数
func (j *JWT) keyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signed method: %v", token.Header[jwtHeaderAlg])
		}
		return []byte(j.c.SecretKey), nil
	}
}

// extractPayloads 提取 token 里包含的 payloads
func extractPayloads(token *jwt.Token) (map[string]any, error) {
	if token == nil || !token.Valid {
		return nil, errInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errNoClaims
	}

	payloads := make(map[string]any)
	for k, v := range claims {
		if _, ok := standardClaimSet[k]; !ok {
			payloads[k] = v
		}
	}

	return payloads, nil
}

// decode 反序列化 src 至 dst
func decode(src, dst any) error {
	dc := &mapstructure.DecoderConfig{
		Result:           dst,
		Squash:           true,
		WeaklyTypedInput: true,
		ZeroFields:       false,
		TagName:          "json",
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	d, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return errors.WithMessage(err, "new map structure decoder err")
	}

	return d.Decode(src)
}

// trimBearerPrefix 去除 token 的 'Bearer ' 前缀
func trimBearerPrefix(tok string) string {
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:]
	}

	return tok
}

// IsStructPointer 判断是否为结构体指针
func IsStructPointer(obj any) bool {
	if obj == nil {
		return false
	}

	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return false
	}

	val = val.Elem()
	if !val.CanAddr() {
		return false
	}

	return val.Kind() == reflect.Struct
}

// IsStruct 判断是否为结构体
func IsStruct(obj any) bool {
	if obj == nil {
		return false
	}

	return reflect.ValueOf(obj).Kind() == reflect.Struct
}
