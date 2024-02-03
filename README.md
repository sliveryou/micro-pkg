# micro-pkg

[![Github License](https://img.shields.io/github/license/sliveryou/micro-pkg.svg?style=flat)](https://github.com/sliveryou/micro-pkg/blob/master/LICENSE)
[![Go Doc](https://godoc.org/github.com/sliveryou/micro-pkg?status.svg)](https://pkg.go.dev/github.com/sliveryou/micro-pkg)
[![Go Report](https://goreportcard.com/badge/github.com/sliveryou/micro-pkg)](https://goreportcard.com/report/github.com/sliveryou/micro-pkg)
[![Github Latest Release](https://img.shields.io/github/release/sliveryou/micro-pkg.svg?style=flat)](https://github.com/sliveryou/micro-pkg/releases/latest)
[![Github Latest Tag](https://img.shields.io/github/tag/sliveryou/micro-pkg.svg?style=flat)](https://github.com/sliveryou/micro-pkg/tags)
[![Github Stars](https://img.shields.io/github/stars/sliveryou/micro-pkg.svg?style=flat)](https://github.com/sliveryou/micro-pkg/stargazers)

go 微服务常用公共包

## 简介

- **apollo** 阿波罗配置中心 go 客户端
- **errcode** 通用业务错误码包，记录了业务状态码、业务消息和 HTTP 状态码，并实现了 `GRPCStatus() *status.Status` 接口，可在 grpc 调用中流转
- **gstream** grpc 流式消息内容读写器，利用反射动态创建消息对象，流式读写消息内容
- **jwt** jwt token 生成和解析包，支持返回 `map[string]any` 类型的 payloads 或反序列化至指定 token 结构体，另外包含 grpc 拦截器，可以自动在 metadata 中传递和解析 token 信息
- **limit** 基于 redis lua 脚本编写的时间段限流器和令牌桶限流器
- **retry** 通用操作重试包，对操作进行失败重试，可以组合不同的策略
- **xgrpc** 包含常用 grpc 拦截器
- **xhash** 通用 hash 校验和计算包，常用 hash 计算，基于 bcrypt hash 的密码生成与校验等
- **xhttp** http 相关操作库，包含请求参数反序列化和响应参数序列化、http 通用客户端、http 通用中间件 和 ip 获取等
- **xkv** 通用 redis 集群键值相关操作库
- **xonce** 操作执行器，只执行一次成功操作，失败可以再次执行

## 文档

- [开发规范](docs/dev-specification.md)
- [Go 安全指南](docs/security-guide.md)
- [Uber Go 语言编码规范](https://github.com/xxjwxc/uber_go_guide_cn)
- [第三方库和工具](docs/third-party-libraries.md)
- [接口错误码](docs/errcode.md)
- [通用 grpc 流式消息内容读写器](gstream/README.md)
- [参考文献](docs/references.md)
