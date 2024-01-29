# micro-pkg

[![Github License](https://img.shields.io/github/license/sliveryou/go-pkg.svg?style=flat)](https://github.com/sliveryou/go-pkg/blob/master/LICENSE)
[![Go Doc](https://godoc.org/github.com/sliveryou/go-pkg?status.svg)](https://pkg.go.dev/github.com/sliveryou/go-pkg)
[![Go Report](https://goreportcard.com/badge/github.com/sliveryou/go-pkg)](https://goreportcard.com/report/github.com/sliveryou/go-pkg)
[![Github Latest Release](https://img.shields.io/github/release/sliveryou/go-pkg.svg?style=flat)](https://github.com/sliveryou/go-pkg/releases/latest)
[![Github Latest Tag](https://img.shields.io/github/tag/sliveryou/go-pkg.svg?style=flat)](https://github.com/sliveryou/go-pkg/tags)
[![Github Stars](https://img.shields.io/github/stars/sliveryou/go-pkg.svg?style=flat)](https://github.com/sliveryou/go-pkg/stargazers)

go 微服务常用公共包

## 简介

- `apollo` 阿波罗配置中心 go 客户端
- `errcode` 通用业务错误码包，记录了业务状态码、业务消息和 HTTP 状态码，并实现了 GRPCStatus 接口，可在 grpc 调用中流转
- `gstream` grpc 流式消息内容读写器，利用反射动态创建消息对象，流式读写消息内容
- `limit` 基于 redis lua 脚本编写的时间段限流器和令牌桶限流器
- `retry` 通用操作重试包，对操作进行失败重试，可以组合不同的策略
- `xhash` 通用 hash 校验和计算包，常用 hash 计算，基于 bcrypt hash 的密码生成与校验等
- `xhttp` http 相关操作库，请求参数反序列化和响应参数序列化，http 通用客户端和 ip 获取等
- `xkv` 通用 redis 集群键值相关操作库
