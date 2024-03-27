package captcha

import (
	"time"

	"dario.cat/mergo"
	"github.com/mojocn/base64Captcha"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/conf"

	"github.com/sliveryou/micro-pkg/xkv"
)

const (
	// DefaultKeyPrefix 验证码答案缓存 key 前缀
	DefaultKeyPrefix = "micro.pkg:captcha:answer:"
)

// Config 验证码配置
type Config struct {
	KeyPrefix      string        `json:",optional"`    // 验证码答案缓存 key 前缀，为空则使用 DefaultKeyPrefix
	ImageWidth     int           `json:",default=240"` // 验证码图片宽度
	ImageHeight    int           `json:",default=80"`  // 验证码图片高度
	CodeLength     int           `json:",default=6"`   // 验证码编码长度
	CodeExpiration time.Duration `json:",default=5m"`  // 验证码编码过期时间
}

// Captcha 验证码校验器
type Captcha struct {
	c       Config
	captcha *base64Captcha.Captcha
}

// NewCaptcha 新建验证码校验器
func NewCaptcha(c Config, kvStore *xkv.Store) (*Captcha, error) {
	if kvStore == nil {
		return nil, errors.New("captcha: illegal captcha config")
	}
	if err := c.fillDefault(); err != nil {
		return nil, errors.WithMessage(err, "captcha: fill default config err")
	}

	driver := base64Captcha.NewDriverDigit(
		c.ImageHeight, c.ImageWidth, c.CodeLength, 0.7, 80,
	)
	store := NewStore(
		kvStore, c.KeyPrefix, int(c.CodeExpiration.Seconds()),
	)
	captcha := base64Captcha.NewCaptcha(driver, store)

	return &Captcha{c: c, captcha: captcha}, nil
}

// MustNewCaptcha 新建验证码校验器
func MustNewCaptcha(c Config, kvStore *xkv.Store) *Captcha {
	captcha, err := NewCaptcha(c, kvStore)
	if err != nil {
		panic(err)
	}

	return captcha
}

// Generate 生成验证码信息
func (c *Captcha) Generate() (id, b64s, answer string, err error) {
	return c.captcha.Generate()
}

// Verify 校验验证码信息
func (c *Captcha) Verify(id, answer string, clear bool) (ok bool) {
	return c.captcha.Verify(id, answer, clear)
}

// fillDefault 填充默认值
func (c *Config) fillDefault() error {
	fill := &Config{}
	if err := conf.FillDefault(fill); err != nil {
		return err
	}

	if c.KeyPrefix == "" {
		c.KeyPrefix = DefaultKeyPrefix
	}

	return mergo.Merge(c, fill)
}
