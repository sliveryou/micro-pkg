# oss

通用对象存储服务（Object Storage Service，OSS）客户端。
   - authored by sliveryou
  
## 背景

对象存储非常适合存储大容量非结构化的数据，例如图片、视频、日志文件、备份数据和容器/虚拟机镜像等，  
而一个对象文件可以是任意大小，一般从几 KB 到 5T 不等。

详细概述可参考如下地址：

- [阿里云 OSS 产品概述](https://help.aliyun.com/zh/oss/product-overview/what-is-oss)
- [华为云 OBS 产品概述](https://support.huaweicloud.com/productdesc-obs/zh-cn_topic_0045829060.html)
- [腾讯云 COS 产品概述](https://cloud.tencent.com/document/product/436/6222)

## 支持服务

- 付费服务：
  - [阿里云 OSS](https://www.aliyun.com/product/oss)
  - [华为云 OBS](https://support.huaweicloud.com/obs/index.html)
  - [腾讯云 COS](https://cloud.tencent.com/product/cos)

- 免费服务：
  - [开源对象存储 MinIO](https://minio.org.cn/docs/minio/container/operations/installation.html) 一套开源的对象存储服务，支持单机部署和分布式部署，纯内网环境的微服务架构中可搭建 MinIO 集群提供对象存储服务
  - 本地存储 Local 推荐只在纯内网环境的单体服务架构中使用，本地存储实际是将文件对象存储在指定本地文件目录下，可以使用 nginx 或 gin 框架的 Static 路由对该文件目录进行服务
  - 模拟存储 Mock 不做任何文件对象的增删改查操作，一般在测试或不需要对象存储功能时使用

## 设计思路

```go
// OSS 客户端接口
type OSS interface {
	// Cloud 获取云服务商名称
	Cloud() string
	// GetURL 获取对象在 OSS 上的完整访问 URL
	GetURL(key string) string
	// GetObject 获取对象在 OSS 的存储数据
	GetObject(key string) (io.ReadCloser, error)
	// PutObject 上传对象至 OSS
	PutObject(key string, reader io.Reader) (string, error)
	// DeleteObjects 批量删除 OSS 上的对象
	DeleteObjects(keys ...string) error
	// UploadFile 上传文件至 OSS，filePath：文件路径，partSize：分块大小（字节），routines：并发数
	UploadFile(key, filePath string, partSize int64, routines int) (string, error)
	// AuthorizedUpload 授权上传至 OSS，expires：过期时间（秒）
	AuthorizedUpload(key string, expires int) (string, error)
	// GetThumbnailSuffix 获取缩略图后缀，如果只传一个值则进行等比缩放，两个值都传时会强制缩放，可能会导致图片变形
	GetThumbnailSuffix(width, height int, size int64) string
}
```

关于文件的需求一般是：  

1. 前端上传文件 
2. 后端接收并存储 
3. 后端返回前端可访问文件的地址或提供文件下载接口

文件托管可以是一块很大的课题，获取文件服务一般不应该由业务服务器提供，主要文件属于一种静态的数据，  
由服务器来处理不免显得有些浪费资源，并且，将文件托管于对象存储服务上，在容灾、分担业务服务器压力等方面具有不少的好处。

所以，OSS 客户端的接口设计主要从对文件对象的增删改查入手：

- 增、改：PutObject，UploadFile，AuthorizedUpload
  - PutObject 上传较小的 io.Reader 对象
  - UploadFile 上传较大的文件对象，会并发的对文件进行分块和断点续传
  - AuthorizedUpload 授权给客户端上传，不经由服务端上传，减少传输文件的 IO 并分摊服务端压力
  - 当 key 相同时相当于执行覆盖更新操作
- 删：DeleteObjects
  - DeleteObjects 批量根据 key 进行对象删除
- 查：GetURL，GetObject
  - GetURL 根据 key 获取对象在 OSS 上的完整访问 URL
  - GetObject 根据 key 获取对象在 OSS 的存储数据
