# 第三方库和工具

## 第三方库

- **contextx** go-zero 框架自带 context 拓展库：
  - 项目地址：https://github.com/zeromicro/go-zero/tree/master/core/contextx
- **mapping** go-zero 框架自带文本序列化和反序列化库：
  - 文档地址：https://github.com/zeromicro/zero-doc/blob/main/go-zero.dev/cn/mapping.md
  - 使用例子：https://github.com/zeromicro/go-zero/blob/master/core/mapping/unmarshaler_test.go
- **threading** go-zero 框架自带多线程库：
  - 在一次调用中另起 goroutine 时，需要增加 recover 机制，而不应该只是 go func()，threading 包里面已经追加了 recover 的逻辑，可以这样使用 `threading.GoSafe(func() { _ = os.RemoveAll(dstDir) })`
  - 项目地址：https://github.com/zeromicro/go-zero/tree/master/core/threading
  - 案例分享：[goroutine 和 panic 不得不说的故事](https://blog.csdn.net/RA681t58CJxsgCkJ31/article/details/83005923)
- **gorm** 通用数据库相关操作库：
  - v2 文档地址：https://gorm.io/zh_CN/docs/index.html
  - v1 文档地址：https://v1.gorm.io/zh_CN/docs/index.html
- **kv** 通用 redis 键值相关操作库：
  - 项目地址：https://github.com/zeromicro/go-zero/tree/master/core/stores/kv
  - 封装：[micro-pkg/xkv](../xkv)
- **cache** 通用缓存相关操作库：
  - 项目地址：https://github.com/zeromicro/go-zero/tree/master/core/stores/cache
- **agollo** 阿波罗配置中心 go 客户端：
  - 项目地址：https://github.com/philchia/agollo
  - 封装：[micro-pkg/apollo](../apollo)
- **validator** 结构体字段参数校验库：
  - 项目地址：https://github.com/go-playground/validator
  - v10 文档地址：https://pkg.go.dev/github.com/go-playground/validator/v10
  - 封装：[go-tool/validator](https://github.com/sliveryou/go-tool#validator)
- **casbin** 支持多种访问控制模型的访问控制框架：
  - 项目地址：https://github.com/casbin/casbin
  - 文档地址：https://casbin.org/zh/docs/overview
  - model 语法：https://casbin.org/zh/docs/syntax-for-models
- **go-tool** 常用工具函数集合：
  - 项目地址：https://github.com/sliveryou/go-tool

## 第三方工具

- **grpcui** grpc 服务端可视化测试工具：
  - 项目地址：https://github.com/fullstorydev/grpcui
  - 启动示例：`grpcui -plaintext localhost:12345`
- **grpc_health_probe** 通用 grpc 健康检查工具
  - 项目地址：https://github.com/grpc-ecosystem/grpc-health-probe
  - 使用示例：`grpc_health_probe -addr=localost:12345 -connect-timeout 250ms -rpc-timeout 100ms`
- **gofumpt** 加强版 gofmt：
  - 项目地址：https://github.com/mvdan/gofumpt
- **goimports-reviser** 加强版 goimports：
  - 项目地址：https://github.com/incu6us/goimports-reviser
- **golangci-lint** 静态代码检测工具：
  - 项目地址：https://github.com/golangci/golangci-lint
  - 文档地址：https://golangci-lint.run
- **swag** swagger 文档快速生成工具：
  - 项目地址：https://github.com/swaggo/swag
  - 文档地址：https://github.com/swaggo/swag/blob/master/README_zh-CN.md
- **goctl** 定制化 goctl：
  - 项目地址：https://github.com/sliveryou/goctl
- **swag2md** 基于 swagger 文档快速生成 markdown 文档工具：
  - 项目地址：https://github.com/sliveryou/swag2md

客户端可视化工具：

1. [ETCD Manager](https://github.com/gtamas/etcdmanager/releases)
2. [Another Redis Desktop Manager](https://github.com/qishibo/AnotherRedisDesktopManager/releases)
