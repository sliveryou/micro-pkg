package apollo

import (
	"fmt"
	"path"

	"github.com/mitchellh/mapstructure"
	agollo "github.com/philchia/agollo/v4"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Config 阿波罗配置中心相关配置
type Config struct {
	IsEnable        bool     // 是否启用
	AppID           string   // 应用ID
	Cluster         string   // 集群
	NameSpaceNames  []string // 命名空间
	CacheDir        string   // 配置缓存目录
	MetaAddr        string   // 服务地址
	AccessKeySecret string   // 访问鉴权密钥
}

// Apollo 阿波罗配置中心客户端
type Apollo struct {
	c             Config // 相关配置
	agollo.Client        // 客户端
}

// NewApollo 新建阿波罗配置中心客户端
func NewApollo(c Config) (*Apollo, error) {
	a := &Apollo{c: c}

	if !c.IsEnable {
		a.Client = &MockClient{}
	} else {
		if c.AppID == "" || c.Cluster == "" ||
			len(c.NameSpaceNames) == 0 || c.MetaAddr == "" {
			return nil, errors.New("apollo: illegal apollo configure")
		}

		client := agollo.NewClient(&agollo.Conf{
			AppID:           c.AppID,
			Cluster:         c.Cluster,
			NameSpaceNames:  c.NameSpaceNames,
			CacheDir:        c.CacheDir,
			MetaAddr:        c.MetaAddr,
			AccesskeySecret: c.AccessKeySecret,
		})

		err := client.Start()
		if err != nil {
			return nil, errors.WithMessage(err, "apollo: client start err")
		}

		a.Client = client
	}

	return a, nil
}

// MustNewApollo 新建阿波罗配置中心客户端
func MustNewApollo(c Config) *Apollo {
	a, err := NewApollo(c)
	if err != nil {
		panic(err)
	}

	return a
}

// GetNamespaceValue 获取给定 namespace 的 key 所对应的 value
func (a *Apollo) GetNamespaceValue(namespace, key string) string {
	return a.GetString(key, agollo.WithNamespace(namespace))
}

// GetNamespaceContent 获取给定 namespace 的具体 content
func (a *Apollo) GetNamespaceContent(namespace string) string {
	return a.GetContent(agollo.WithNamespace(namespace))
}

// GetDumpFileName 获取备份文件名称
func (a *Apollo) GetDumpFileName() string {
	fileName := fmt.Sprintf(".%s_%s", a.c.AppID, a.c.Cluster)
	return path.Join(a.c.CacheDir, fileName)
}

// UnmarshalYaml 将 content 字符串 yaml 反序列化到 value 中
func UnmarshalYaml(content string, value any, zeroFields ...bool) error {
	needZeroFields := false
	if len(zeroFields) > 0 {
		needZeroFields = zeroFields[0]
	}

	m := make(map[string]any)
	err := yaml.Unmarshal([]byte(content), m)
	if err != nil {
		return errors.WithMessage(err, "yaml unmarshal content err")
	}

	dc := &mapstructure.DecoderConfig{
		Result:           value,
		Squash:           true,
		WeaklyTypedInput: true,
		ZeroFields:       needZeroFields,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	d, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return errors.WithMessage(err, "new map structure decoder err")
	}

	return d.Decode(m)
}

// MustUnmarshalYaml 将 content 字符串 yaml 反序列化到 value 中
func MustUnmarshalYaml(content string, value any, zeroFields ...bool) {
	err := UnmarshalYaml(content, value, zeroFields...)
	if err != nil {
		panic(err)
	}
}
