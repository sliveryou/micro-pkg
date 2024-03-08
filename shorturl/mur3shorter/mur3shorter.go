package mur3shorter

import (
	"errors"

	"github.com/spaolacci/murmur3"
)

const (
	// Index 十进制为 61，用于截取索引数
	Index = 0b111101
	// Shift 右移位数，为 Index 所占位数
	Shift = 6
	// DefaultAlphabet 默认对应进制数字符表，长度须为 62
	DefaultAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// Shorter 基于 murmur3 的短地址映射器结构详情
type Shorter struct {
	Length   int    // 生成短地址标识长度
	Alphabet []byte // 对应进制数字符表
}

// NewShorter 新建 murmur3 短地址映射器
func NewShorter(length int64, alphabet ...string) (*Shorter, error) {
	a := DefaultAlphabet
	if length < 1 {
		return nil, errors.New("mur3shorter: illegal shorter config")
	}
	if len(alphabet) > 0 && len(alphabet[0]) == 62 {
		a = alphabet[0]
	}

	return &Shorter{
		Length:   int(length),
		Alphabet: []byte(a),
	}, nil
}

// MustNewShorter 新建 murmur3 短地址映射器
func MustNewShorter(length int64, alphabet ...string) *Shorter {
	s, err := NewShorter(length, alphabet...)
	if err != nil {
		panic(err)
	}

	return s
}

// Mapping 长地址映射成短地址标识
func (s *Shorter) Mapping(longURL string) string {
	sum := murmur3.Sum64([]byte(longURL))

	var tempIndex uint64
	shortURL := make([]byte, s.Length)
	for i := range shortURL {
		tempIndex = Index & sum
		shortURL[i] = s.Alphabet[tempIndex]
		sum >>= Shift // 右移位操作
	}

	return string(shortURL)
}
