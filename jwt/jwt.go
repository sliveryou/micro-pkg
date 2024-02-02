package jwt

import (
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

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
)

var (
	standardClaims = []string{
		jwtAudience, jwtExpire, jwtID, jwtIssueAt,
		jwtIssuer, jwtNotBefore, jwtSubject,
	}

	errInvalidToken = errors.New("invalid jwt token")
	errNoClaims     = errors.New("no token claims")
)

// Config JWT 相关配置
type Config struct {
	Issuer         string        // 签发者
	SecretKey      string        // 密钥
	ExpirationTime time.Duration // 过期时间
}

// JWT 结构详情
type JWT struct {
	c Config
}

// NewJWT 新建 JWT
func NewJWT(c Config) (*JWT, error) {
	if c.Issuer == "" || c.SecretKey == "" || c.ExpirationTime.Seconds() <= 0 {
		return nil, errors.New("jwt: illegal jwt configure")
	}

	return &JWT{c: c}, nil
}

// MustNewJWT 新建 JWT
func MustNewJWT(c Config) *JWT {
	j, err := NewJWT(c)
	if err != nil {
		panic(err)
	}

	return j
}

// GenToken 根据 payloads 生成 JWT token
func (j *JWT) GenToken(payloads map[string]any, expirationTime ...time.Duration) (string, error) {
	et := j.c.ExpirationTime
	if len(expirationTime) > 0 && expirationTime[0].Seconds() > 0 {
		et = expirationTime[0]
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
		if !slices.Contains(standardClaims, k) {
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
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func (j *JWT) ParseToken(tokenString string, token any) error {
	payloads, err := j.ParseTokenPayloads(tokenString)
	if err != nil {
		return err
	}

	return errors.WithMessage(decodePayloads(payloads, token), "decode payloads err")
}

// ParseTokenFromRequest 从请求头解析 JWT token，并将其反序列化至指定 token 结构体中
// 注意：token 必须为结构体指针，名称以 json tag 对应的名称与 payloads 进行映射
func (j *JWT) ParseTokenFromRequest(r *http.Request, token any) error {
	payloads, err := j.ParseTokenPayloadsFromRequest(r)
	if err != nil {
		return err
	}

	return errors.WithMessage(decodePayloads(payloads, token), "decode payloads err")
}

// ParseTokenPayloads 解析 JWT token，返回 payloads
func (j *JWT) ParseTokenPayloads(tokenString string) (map[string]any, error) {
	token, err := j.newParser().Parse(trimBearerPrefix(tokenString), j.keyFunc())
	if err != nil {
		return nil, errors.WithMessage(err, "jwt.Parse err")
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
		return nil, errors.WithMessage(err, "request.ParseFromRequest err")
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
		if !slices.Contains(standardClaims, k) {
			payloads[k] = v
		}
	}

	return payloads, nil
}

// decodePayloads 反序列化 payloads 至 dst 结构体中
func decodePayloads(payloads map[string]any, dst any) error {
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

	return d.Decode(payloads)
}

// trimBearerPrefix 去除 token 的 'Bearer ' 前缀
func trimBearerPrefix(tok string) string {
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:]
	}

	return tok
}
