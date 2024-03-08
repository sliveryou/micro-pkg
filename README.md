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
- **auth** 身份认证包，包含阿里云银行卡四要素认证、阿里云企业银行卡账户认证和百度云人脸识别认证
- **balancer** grpc 平衡器，包含了一致性 hash 平衡器
- **captcha** base64 编码的图形验证码包，使用 redis 缓存验证码答案
- **enforcer** 基于 casbin 实现的接口决策规则执行器
- **errcode** 通用业务错误码包，记录了业务状态码、业务消息和 HTTP 状态码，并实现了 `GRPCStatus() *status.Status` 接口，可在 grpc 调用中流转
- **excel** 常用 excel 操作包，包含获取所有行数据、流式读取行数据和流式写入行数据等操作 
- **express** 通用快递查询客户端，支持 express100（快递100）和 expressBird（快递鸟）
- **gstream** grpc 流式消息内容读写器，利用反射动态创建消息对象，流式读写消息内容
- **health** 健康检查包，实现了 [grpc_health_v1](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) 定义的健康检查服务端和客户端，并包含了一些常用中间件的健康检查器
- **jwt** jwt token 生成和解析包，支持返回 `map[string]any` 类型的 payloads 或反序列化至指定 token 结构体，另外包含 grpc 拦截器，可以自动在 metadata 中传递和解析 token 信息
- **limit** 基于 redis lua 脚本编写的时间段限流器和令牌桶限流器
- **lock** 基于 etcd 实现的分布式锁
- **notify** 通用通知服务包，包含短信、邮件验证码发送与短信、邮件验证码校验等功能，并支持对发送间隔、验证间隔、一天内同一接收方、一天内同一 IP 和一天内总发送量的限制与监控
- **oss** 通用对象存储服务客户端，支持 aliyun、huawei、tencent、minio、local 和 mock
- **retry** 通用操作重试包，对操作进行失败重试，可以组合不同的策略
- **shorturl** 基于 murmur3 hash 的短地址标识符生成包
- **watcher** 基于 etcd 的键值更新观察器，当观察到键发生创建或更新事件时，会触发回调函数，并实现了 casbin 的 `persist.Watcher` 接口
- **xdb** 通用数据库连接包，返回 `*gorm.DB` 对象，支持 MySQL、PostgreSQL、SQLite 和 SQLServer
- **xdb/xfield** gorm gen 字段拓展包，支持构建原始 sql 字段和原始 sql 条件
- **xgrpc** 常用 grpc 拦截器，包含请求响应日志打印和恐慌捕获恢复等
- **xhash** 通用 hash 校验和计算包，包含常用 hash 计算和基于 bcrypt hash 的密码生成与校验等
- **xhttp** http 相关操作库，包含请求参数反序列化和响应参数序列化、http 通用客户端、http 通用中间件 和 ip 获取等
- **xkv** 通用 redis 集群键值相关操作库
- **xonce** 操作执行器，只执行一次成功操作，失败可以再次执行

## 文档

- [开发规范](docs/dev-specification.md)
- [Go 安全指南](docs/security-guide.md)
- [Uber Go 语言编码规范](https://github.com/xxjwxc/uber_go_guide_cn)
- [第三方库和工具](docs/third-parties.md)
- [接口错误码](docs/errcode.md)
- [通用 grpc 流式消息内容读写器](gstream/README.md)
- [通用通知服务包](notify/README.md)
- [通用对象存储服务客户端](oss/README.md)
- [短地址标识符生成](shorturl/README.md)
- [gorm/gen 字段拓展包](xdb/xfield/README.md)
- [枚举类型方法生成工具](docs/enumer.md)
- [参考文献](docs/references.md)
