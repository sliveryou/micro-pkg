package openapi

import (
	"context"
	"path"
	"strings"
)

// Format 命名空间格式
type Format string

const (
	// FormatProperties 命名空间格式：properties
	FormatProperties Format = "properties"
	// FormatXML 命名空间格式：xml
	FormatXML Format = "xml"
	// FormatYML 命名空间格式：yml
	FormatYML Format = "yml"
	// FormatYAML 命名空间格式：yaml
	FormatYAML Format = "yaml"
	// FormatJSON 命名空间格式：json
	FormatJSON Format = "json"
)

// apiOption 接口可选参数配置
type apiOption struct {
	env       string
	appID     string
	cluster   string
	namespace string
}

// APIOption 接口可选参数配置
type APIOption func(o *apiOption)

// WithEnv 设置环境
func WithEnv(env string) APIOption {
	return func(o *apiOption) {
		o.env = env
	}
}

// WithAppID 设置应用ID
func WithAppID(appID string) APIOption {
	return func(o *apiOption) {
		o.appID = appID
	}
}

// WithCluster 设置集群
func WithCluster(cluster string) APIOption {
	return func(o *apiOption) {
		o.cluster = cluster
	}
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) APIOption {
	return func(o *apiOption) {
		o.namespace = namespace
	}
}

// OpenAPI 阿波罗配置中心开放平台接口
type OpenAPI interface {
	// GetEnvClusters 获取对应应用环境下的所有集群信息（appId 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_321-%e8%8e%b7%e5%8f%96app%e7%9a%84%e7%8e%af%e5%a2%83%ef%bc%8c%e9%9b%86%e7%be%a4%e4%bf%a1%e6%81%af
	GetEnvClusters(ctx context.Context, opts ...APIOption) ([]*EnvClusters, error)
	// GetNamespaces 获取对应集群下的所有命名空间信息（env、appId 和 cluster 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_325-%e8%8e%b7%e5%8f%96%e9%9b%86%e7%be%a4%e4%b8%8b%e6%89%80%e6%9c%89namespace%e4%bf%a1%e6%81%af%e6%8e%a5%e5%8f%a3
	GetNamespaces(ctx context.Context, opts ...APIOption) ([]*Namespace, error)
	// GetNamespace 获取指定命名空间信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_326-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e4%bf%a1%e6%81%af%e6%8e%a5%e5%8f%a3
	GetNamespace(ctx context.Context, opts ...APIOption) (*Namespace, error)
	// CreateNamespace 创建命名空间信息（appId 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_327-%e5%88%9b%e5%bb%banamespace
	CreateNamespace(ctx context.Context, r CreateNamespaceReq, opts ...APIOption) (*CreateNamespaceResp, error)
	// GetNamespaceLock 获取指定命名空间锁定信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_328-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e5%bd%93%e5%89%8d%e7%bc%96%e8%be%91%e4%ba%ba%e6%8e%a5%e5%8f%a3
	GetNamespaceLock(ctx context.Context, opts ...APIOption) (*NamespaceLock, error)
	// AddItem 添加配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3210-%e6%96%b0%e5%a2%9e%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	AddItem(ctx context.Context, r AddItemReq, opts ...APIOption) (*Item, error)
	// UpdateItem 更新配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3211-%e4%bf%ae%e6%94%b9%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	UpdateItem(ctx context.Context, r UpdateItemReq, opts ...APIOption) error
	// CreateOrUpdateItem 创建或者更新配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3211-%e4%bf%ae%e6%94%b9%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	CreateOrUpdateItem(ctx context.Context, r UpdateItemReq, opts ...APIOption) error
	// DeleteItem 删除配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3212-%e5%88%a0%e9%99%a4%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	DeleteItem(ctx context.Context, r DeleteItemReq, opts ...APIOption) error
	// PublishRelease 发布版本配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3213-%e5%8f%91%e5%b8%83%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	PublishRelease(ctx context.Context, r PublishReleaseReq, opts ...APIOption) (*Release, error)
	// GetRelease 获取版本配置信息（env、appId、cluster 和 namespace 必填）
	// https://www.apolloconfig.com/#/zh/usage/apollo-open-api-platform?id=_3214-%e8%8e%b7%e5%8f%96%e6%9f%90%e4%b8%aanamespace%e5%bd%93%e5%89%8d%e7%94%9f%e6%95%88%e7%9a%84%e5%b7%b2%e5%8f%91%e5%b8%83%e9%85%8d%e7%bd%ae%e6%8e%a5%e5%8f%a3
	GetRelease(ctx context.Context, opts ...APIOption) (*Release, error)
}

// EnvClusters 环境及其下集群信息
type EnvClusters struct {
	Env      string   `json:"env"`
	Clusters []string `json:"clusters"`
}

// Item 配置信息
type Item struct {
	Key                        string `json:"key"`
	Value                      string `json:"value"`
	Comment                    string `json:"comment"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

// Namespace 命名空间信息
type Namespace struct {
	AppID                      string `json:"appId"`
	ClusterName                string `json:"clusterName"`
	NamespaceName              string `json:"namespaceName"`
	Comment                    string `json:"comment"`
	Format                     string `json:"format"`
	IsPublic                   bool   `json:"isPublic"`
	Items                      []Item `json:"items"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

// CreateNamespaceReq 创建命名空间请求
type CreateNamespaceReq struct {
	Name                string `json:"name"`                // namespace 的名字
	AppID               string `json:"appId"`               // namespace 所属的 AppId
	Format              Format `json:"format"`              // namespace 的格式，只能是以下类型：properties、xml、json、yml 和 yaml
	IsPublic            bool   `json:"isPublic"`            // 是否是公共文件
	Comment             string `json:"comment"`             // namespace 说明（非必填）
	DataChangeCreatedBy string `json:"dataChangeCreatedBy"` // namespace 的创建人，格式为域账号，也就是 sso 系统的 User Id
}

// CreateNamespaceResp 创建命名空间响应
type CreateNamespaceResp struct {
	Name                       string `json:"name"`
	AppID                      string `json:"appId"`
	Format                     string `json:"format"`
	IsPublic                   bool   `json:"isPublic"`
	Comment                    string `json:"comment"`
	DataChangeCreatedBy        string `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string `json:"dataChangeLastModifiedTime"`
}

// NamespaceLock 命名空间锁定信息
type NamespaceLock struct {
	NamespaceName string `json:"namespaceName"`
	IsLocked      bool   `json:"isLocked"`
	LockedBy      string `json:"lockedBy"`
}

// AddItemReq 添加配置信息请求
type AddItemReq struct {
	Key                 string `json:"key"`                 // 配置的 key，长度不能超过 128 个字符。非 properties 格式，key 固定为 content
	Value               string `json:"value"`               // 配置的 value，长度不能超过 20000 个字符，非 properties 格式，value 为文件全部内容
	Comment             string `json:"comment"`             // 配置的备注，长度不能超过 1024 个字符（非必填）
	DataChangeCreatedBy string `json:"dataChangeCreatedBy"` // item 的创建人，格式为域账号，也就是 sso 系统的 User Id
}

// UpdateItemReq 更新配置信息请求
type UpdateItemReq struct {
	Key                      string `json:"key"`                      // 配置的 key，需和 url 中的 key 值一致。非 properties 格式，key 固定为content
	Value                    string `json:"value"`                    // 配置的 value，长度不能超过 20000 个字符，非 properties 格式，value 为文件全部内容
	Comment                  string `json:"comment"`                  // 配置的备注，长度不能超过 256 个字符（非必填）
	DataChangeLastModifiedBy string `json:"dataChangeLastModifiedBy"` // item 的修改人，格式为域账号，也就是 sso 系统的 User Id
	DataChangeCreatedBy      string `json:"dataChangeCreatedBy"`      // 当 createIfNotExists 为 true 时必选。item 的创建人，格式为域账号，也就是 sso 系统的 User ID
}

// DeleteItemReq 删除配置信息请求
type DeleteItemReq struct {
	Key      string `json:"key"`      // 配置的 key。非 properties 格式，key 固定为 content
	Operator string `json:"operator"` // 删除配置的操作者，域账号
}

// PublishReleaseReq 发布版本配置信息请求
type PublishReleaseReq struct {
	ReleaseTitle   string `json:"releaseTitle"`   // 此次发布的标题，长度不能超过 64 个字符
	ReleaseComment string `json:"releaseComment"` // 发布的备注，长度不能超过 256 个字符（非必填）
	ReleasedBy     string `json:"releasedBy"`     // 发布人，域账号，注意：如果 ApolloConfigDB.ServerConfig 中的 namespace.lock.switch 设置为 true 的话（默认是 false），那么该环境不允许发布人和编辑人为同一人。所以如果编辑人是 zhanglea，发布人就不能再是 zhanglea。
}

// Release 版本配置信息
type Release struct {
	AppID                      string            `json:"appId"`
	ClusterName                string            `json:"clusterName"`
	NamespaceName              string            `json:"namespaceName"`
	Name                       string            `json:"name"`
	Configurations             map[string]string `json:"configurations"`
	Comment                    string            `json:"comment"`
	DataChangeCreatedBy        string            `json:"dataChangeCreatedBy"`
	DataChangeLastModifiedBy   string            `json:"dataChangeLastModifiedBy"`
	DataChangeCreatedTime      string            `json:"dataChangeCreatedTime"`
	DataChangeLastModifiedTime string            `json:"dataChangeLastModifiedTime"`
}

// ParseFormat 解析命名空间格式
func ParseFormat(namespace string) (Format, bool) {
	if namespace == "" {
		return "", false
	}

	suffix := path.Ext(namespace)
	switch suffix {
	case "", ".properties":
		return FormatProperties, true
	case ".xml":
		return FormatXML, true
	case ".yml":
		return FormatYML, true
	case ".yaml":
		return FormatYAML, true
	case ".json":
		return FormatJSON, true
	default:
		return "", false
	}
}

// TrimFormat 去除命名空间格式
func TrimFormat(namespace string) string {
	if namespace == "" {
		return ""
	}

	suffix := path.Ext(namespace)

	return strings.TrimSuffix(namespace, suffix)
}
