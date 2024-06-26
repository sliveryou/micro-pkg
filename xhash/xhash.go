package xhash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/sliveryou/micro-pkg/xhash/sm3"
)

// 摘要名称
const (
	MD5    = "md5"
	SM3    = "sm3"
	SHA1   = "sha1"
	SHA224 = "sha224"
	SHA256 = "sha256"
	SHA384 = "sha384"
	SHA512 = "sha512"
)

// New 根据提供的摘要名称新建 hash.Hash 对象
func New(digest string) hash.Hash {
	switch strings.ToLower(digest) {
	case MD5:
		return md5.New()
	case SM3:
		return sm3.New()
	case SHA1:
		return sha1.New()
	case SHA224:
		return sha256.New224()
	case SHA256:
		return sha256.New()
	case SHA384:
		return sha512.New384()
	case SHA512:
		return sha512.New()
	default:
		return md5.New()
	}
}

// HashReader 根据提供的 hash.Hash 对象计算 io.Reader 具体内容的 hash
func HashReader(h hash.Hash, r io.Reader) (string, error) {
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// Base64HashReader 根据提供的 hash.Hash 对象计算 io.Reader 具体内容的 base64 hash
func Base64HashReader(h hash.Hash, r io.Reader) (string, error) {
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// HashString 根据提供的 hash.Hash 对象计算所给 s 字符串具体内容的 hash
func HashString(h hash.Hash, s string) (string, error) {
	return HashReader(h, strings.NewReader(s))
}

// Base64HashString 根据提供的 hash.Hash 对象计算所给 s 字符串具体内容的 base64 hash
func Base64HashString(h hash.Hash, s string) (string, error) {
	return Base64HashReader(h, strings.NewReader(s))
}

// HashFile 根据提供的 hash.Hash 对象计算文件路径指向的文件具体内容的 hash，
// 若传递不为空的文件名称参数，则会在文件具体内容后追加文件名称然后计算 hash
func HashFile(h hash.Hash, filePath string, fileName ...string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", errors.WithMessagef(err, "open %s err", filePath)
	}
	defer f.Close()

	_, err = io.Copy(h, f)
	if err != nil {
		return "", errors.WithMessage(err, "calculate hash err")
	}

	if len(fileName) > 0 && fileName[0] != "" {
		h.Write([]byte(fileName[0]))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// GenFromPwd 生成密码的 bcrypt hash
func GenFromPwd(pwd string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPwd), nil
}

// CmpHashAndPwd 将密码的 bcrypt hash 与其可能的等效明文密码进行比较
func CmpHashAndPwd(hashedPwd, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}
