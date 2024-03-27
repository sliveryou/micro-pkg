package shorturl

import "github.com/sliveryou/micro-pkg/shorturl/mur3shorter"

var _ Shorter = (*mur3shorter.Shorter)(nil)

// Shorter 短地址映射器接口
type Shorter interface {
	// Mapping 长地址映射成短地址标识
	Mapping(longURL string) (shortURL string)
}

// Config 短地址配置
type Config struct {
	Length   int64  // 生成短地址标识长度，不建议超过 12
	Alphabet string `json:",optional"` // 对应进制数字符表，为空则使用默认，不为空长度须为 62
}

// NewShorter 新建短地址映射器
func NewShorter(c Config) (Shorter, error) {
	return mur3shorter.NewShorter(c.Length, c.Alphabet)
}

// MustNewShorter 新建短地址映射器
func MustNewShorter(c Config) Shorter {
	s, err := mur3shorter.NewShorter(c.Length, c.Alphabet)
	if err != nil {
		panic(err)
	}

	return s
}
