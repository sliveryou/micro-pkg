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
  - gen 文档地址：https://gorm.io/zh_CN/gen/index.html
  - 封装：[micro-pkg/xdb](../xdb)，[micro-pkg/xdb/xfield](../xdb/xfield)
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
  - 封装：[micro-pkg/enforcer](../enforcer)
- **excelize** 操作 office excel 文档基础库：
  - 项目地址：https://github.com/qax-os/excelize
  - 文档地址：https://xuri.me/excelize/zh-hans
  - 封装：[micro-pkg/excel](../excel)
- **base64Captcha** base64 编码图形验证码库：
  - 项目地址：https://github.com/mojocn/base64Captcha
  - 文档地址：https://zh.mojotv.cn/go/refactor-base64-captcha
  - base64 图片转换：https://tool.chinaz.com/tools/imgtobase
  - 封装：[micro-pkg/captcha](../captcha)
- **go-tool** 常用工具函数集合：
  - 项目地址：https://github.com/sliveryou/go-tool
  - 文档地址：https://pkg.go.dev/github.com/sliveryou/go-tool/v2

## 第三方工具

- **gvm** go 版本管理工具：
  - 项目地址：https://github.com/moovweb/gvm
  - 使用示例：
    - `gvm listall` 
    - `gvm install go1.21.7` 
    - `gvm use go1.21.7 --default`
- **grpcui** grpc 服务端可视化测试工具：
  - 项目地址：https://github.com/fullstorydev/grpcui
  - 启动示例：`grpcui -plaintext localhost:12345`
- **grpc_health_probe** 通用 grpc 健康检查工具
  - 项目地址：https://github.com/grpc-ecosystem/grpc-health-probe
  - 使用示例：`grpc_health_probe -addr=localost:12345 -connect-timeout 250ms -rpc-timeout 100ms`
- **scc** 代码行数统计工具：
  - 项目地址：https://github.com/boyter/scc
  - 使用示例：`scc -i go`
- **shfmt** shell 脚本格式化工具：
  - 项目地址：https://mvdan.cc/sh/v3/cmd/shfmt
  - 使用示例：`shfmt -w -s -i 2 -ci -bn -sr dep.sh`
- **gofumpt** 加强版 gofmt：
  - 项目地址：https://github.com/mvdan/gofumpt
  - 使用示例：`gofumpt -w -extra main.go`
- **goimports-reviser** 加强版 goimports：
  - 项目地址：https://github.com/incu6us/goimports-reviser
  - 使用示例：`goimports-reviser -rm-unused -set-alias -company-prefixes "github.com/sliveryou" -project-name "github.com/sliveryou/micro-pkg" main.go`
- **golangci-lint** 静态代码检测工具：
  - 项目地址：https://github.com/golangci/golangci-lint
  - 文档地址：https://golangci-lint.run
  - 使用示例：`golangci-lint run ./...`
- **swag** swagger 文档快速生成工具：
  - 项目地址：https://github.com/swaggo/swag
  - 文档地址：https://github.com/swaggo/swag/blob/master/README_zh-CN.md
  - 使用示例：`swag init -d . -g main.go -p snakecase --ot go,json,yaml -o docs`
- **enumer** go 枚举方法生成工具
  - 项目地址：https://github.com/alvaroloes/enumer
  - 使用示例：`//go:generate enumer -type Status -json -linecomment -output health_string.go`
  - [使用教程](enumer.md)
- **grom** 基于 mysql 数据表生成 go-zero 相关项目文件工具：
  - 项目地址：https://github.com/sliveryou/grom/tree/feat-go-zero
  - 文档地址：https://github.com/sliveryou/grom/blob/feat-go-zero/README_zh-CN.md
  - 使用示例：
    - `grom api config -n config.yaml`
    - `grom api generate -n config.yaml`
- **goctl** 定制化 goctl：
  - 项目地址：https://github.com/sliveryou/goctl
  - 使用示例：`goctl api proto --api base.api --dir .`
- **swag2md** 基于 swagger 文档快速生成 markdown 文档工具：
  - 项目地址：https://github.com/sliveryou/swag2md
  - 使用示例：
    - `swag2md -t "接口文档" -s swagger.json -o api.md`
    - `swag2md casbin -s swagger.json -o policy.csv --sub ADMIN --deny`

## 客户端可视化工具

1. [ETCD Manager](https://github.com/gtamas/etcdmanager/releases)
2. [Another Redis Desktop Manager](https://github.com/qishibo/AnotherRedisDesktopManager/releases)
