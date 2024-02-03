# 开发规范

三个原则：

- Clarity（清晰）

`Hal Abelson and Gerald Sussman` 说过：

> Programs must be written for people to read, and only incidentally for machines to execute.

程序是什么，程序必须是为了开发人员阅读而编写的，只是偶尔给机器去执行，99% 的时间程序代码面向的是开发人员，
而只有 1% 的时间可能是机器在执行，这里比例不是重点，从中我们可以看出，清晰的代码是多么的重要，
因为所有程序，不仅是 Go 语言，都是由开发人员编写，供其他人阅读和维护。

- Simplicity（简单）

可靠的前提条件就是简单，我们在实际开发中都遇到过，这段代码在写什么，想要完成什么事情，往往因为：

1. 没有注释，代码密密麻麻，根本不知道从何入手
2. 调用名称千奇百怪，不能见名知义
3. 内部逻辑极其复杂等

令开发人员不理解这段代码，因此也不知道如何去维护，这就带来了复杂性，
程序越是复杂就越难维护，越难维护就会是程序变得越来越复杂，
因此，遇到程序变复杂时首先应该想到的是——重构，重构会重新设计程序，让程序变得简单。

- Productivity（生产力）

开发中往往有许多不必要的重复劳动和简单但麻烦的操作步骤，需要利用各种有效的途径来利用有限的时间完成开发效率最大化，
`go-zero` 有个准则是 `工具大于约定和文档`，我个人是非常认同的，尽量使用工具来补足重复劳动和简单但麻烦的步骤，提升开发效率。

## 命名规范

好的命名规范主要可以对应到上面的 `清晰` 和 `简单` 原则，因为它可以：

- 降低代码阅读成本
- 降低维护难度
- 降低代码复杂度

### 命名准则

- 当变量名称在定义和最后依次使用之间的距离很短时，简短的名称看起来会更好
- 变量命名应尽量描述其内容，而不是类型
- 常量命名应尽量描述其值，而不是如何使用这个值
- 在遇到 for，if 等循环或分支时，推荐简短字母命名来标识参数和返回值
- package 名称也是命名的一部分，请尽量将其利用起来
- 使用一致的命名风格

### 文件命名规范

- 全部小写
- 除单元测试 `*_test.go` 外尽量避免文件中出现下划线 `_`
- 文件名称不宜过长，一些单词可以选择缩写

### 变量命名规范

- 首字母小写
- 驼峰命名
- 见名知义，避免拼音替代英文
- 不建议包含下划线 `_`
- 不建议包含数字

**适用范围**
- 局部变量
- 函数出参、入参

### 函数、常量命名规范

- 驼峰式命名
- 可导出的必须首字母大写
- 不可导出的必须首字母小写
- 避免全部大写与下划线 `_` 组合

## 路由规范

### 路由命名规范

- 简单增删改查接口尽量组合 HTTP Method 的语意，如 GET（获取），POST（新建），PUT（更新），DELETE（删除）
- 带上版本号
- 脊柱式命名
- 小写单数单词、横杠 `-` 组合
- 见名知义
- 复杂接口统一为 POST 请求

```
GET /v1/personal-auth/1 获取用户 id 为 1 的个人认证信息
GET /v1/order?page=1&page_size=10&name=VIP1 分页获取订单信息列表
PUT /v1/user/1/password 更新用户 id 为 1 的用户的密码
GET /v1/proof/1 获取存证 id 为 1 的存证信息
POST /v1/proof 添加存证
DELETE /v1/proof/1 删除存证
```

### 路由注意事项

- 请求体注意标对正确的参数来源，如 `form, path 和 json`
- 请求体显示设置必填参数和可选参数，必填参数附上 `validate` 和 `label` 标签，可选参数在其数据来源后加上 `,optional`
- 响应体统一以 `json` 形式返回
- 不推荐没有响应，如果不知道返回什么就返回该资源的 `id`
- 一般整数类型只选择 `int32` 或 `int64`，偏向枚举含义的整数类型选择 `int32`，其他选择 `int64` 
- 请求体和响应体各个参数均须附上注释，枚举类型须附上各个数值对应的含义

例子：

```go
// Department 部门
type Department struct {
	Id          int64  `json:"id"`          // 部门id
	Name        string `json:"name"`        // 部门名称
	Description string `json:"description"` // 部门描述
}

// DepartmentPage 部门分页
type DepartmentPage struct {
	Department
	JobCount       int64 `json:"job_count"`       // 岗位数量
	PersonnelCount int64 `json:"personnel_count"` // 人员数量
	UnableDelete   bool  `json:"unable_delete"`   // 不能删除
}

// GetDepartmentReq 查询部门请求
type GetDepartmentReq struct {
	Id int64 `path:"id" validate:"required" label:"部门id"` // 部门id
}

// GetDepartmentResp 查询部门响应
type GetDepartmentResp struct {
	Department
}

// GetDepartmentPagesReq 查询部门分页请求
type GetDepartmentPagesReq struct {
	Type     *int32 `form:"type,optional" validate:"omitempty,oneof=0 1 2" label:"部门类型"` // 部门类型（0-内部 1-外部 2-特殊）
	Name     string `form:"name,optional"`                               // 部门名称
	Search   string `form:"search,optional"`                             // 搜索 
	Page     int64  `form:"page" validate:"required" label:"页数"`        // 页数
	PageSize int64  `form:"page_size" validate:"required" label:"每条页数"` // 每条页数
}

// GetDepartmentPagesResp 查询部门分页响应
type GetDepartmentPagesResp struct {
	Count     int64             `json:"count"`      // 总数
	PageCount int64             `json:"page_count"` // 页数
	Results   []*DepartmentPage `json:"results"`    // 结果
}

// CreateDepartmentReq 创建部门请求
type CreateDepartmentReq struct {
	Name        string `json:"name" validate:"required" label:"部门名称"` // 部门名称
	Description string `json:"description,optional"`                  // 部门描述
}

// CreateDepartmentResp 创建部门响应
type CreateDepartmentResp struct {
	Department
}

// UpdateDepartmentReq 更新部门请求
type UpdateDepartmentReq struct {
	Id          int64   `path:"id" validate:"required" label:"部门id" swaggerignore:"true"` // 部门id
	Name        string  `json:"name" validate:"required" label:"部门名称"`                    // 部门名称
	Description *string `json:"description,optional"`                                     // 部门描述
}

// UpdateDepartmentResp 更新部门响应
type UpdateDepartmentResp struct {
	Department
}
```

PS:

- **validator** 结构体字段参数校验库：
  - 项目地址：https://github.com/go-playground/validator
  - v10 文档地址：https://pkg.go.dev/github.com/go-playground/validator/v10
  - 封装：[go-tool/validator](https://github.com/sliveryou/go-tool#validator)
- **swag** swagger 文档快速生成工具：
  - 项目地址：https://github.com/swaggo/swag
  - 文档地址：https://github.com/swaggo/swag/blob/master/README_zh-CN.md
- **swag2md** 基于 swagger 文档快速生成 markdown 文档工具：
  - 项目地址：https://github.com/sliveryou/swag2md

## 编码规范

### import

- 单个 `import` 不建议用圆括号包裹
- 按照 `标准库`、`第三方包`、`其他项目包` 和 `项目包` 顺序分组引入（可以使用 goimports-reviser 工具）：

```go
import (
    "context"
    
    "github.com/zero-micro/go-zero/core/logx"
    "gorm.io/gorm"
	
    "gitlab.xxx.cn/go-tool/timex"
    
    "gitlab.xxx.cn/my-project/model"
    "gitlab.xxx.cn/my-project/pkg/db"
)
```

### 函数开发

- 一般在 pkg 下的包，最好能业务无关化，达到复制粘贴到别的项目下也能开箱即用
- 不建议包调用设计成单例模式，最好设计成依赖注入的模式，减少包与包之间调用的耦合性
- 必须格式化（可以使用 gofumpt 工具）
- 必须有注释
- 必须有单元测试文件，复杂的包要附上 README.md
- 保持不信任原则，要有参数合法性检查或者默认参数机制

可以参考 [go-tool](https://github.com/sliveryou/go-tool) 的包代码

### 函数调用

- 另起协程要有 recover 机制，不能直接 `go func`，建议使用 [github.com/zeromicro/go-zero/core/threading](https://github.com/zeromicro/go-zero/tree/master/core/threading) 起协程

### 函数返回

- 对象避免非指针返回
- 遵循有正常值返回则一定无 error，有 error 则一定无正常值返回的原则

### 错误处理

- 有 error 必须处理，如果不能处理就必须抛出
- 避免下划线 `_` 接收 error

### 函数体编码

- 建议一个代码块结束空一行，如 if、for 等

```go
func demo () {
    if xxx {
        // do something
    }
    
    for _, x := range xxx {
        // do something
    }
    
    fmt.Println("xxx")
}
```
  
- return 前空一行

```go
func demo(id string) (string, error) {
    ......

    return "xxx", nil
}
```

- return 结构体时，直接声明并赋值，然后返回

```go
// 推荐
func demo() (*user.GetUserInfoResp, error) {
    ......
    
    return &user.GetUserInfoResp{
        Id:   in.Id,
        Name: in.Name,
    }, nil
}
```

PS:

- **gofumpt** 加强版 gofmt：
  - 项目地址：https://github.com/mvdan/gofumpt
- **goimports-reviser** 加强版 goimports：
  - 项目地址：https://github.com/incu6us/goimports-reviser
- **goctl** 定制化 goctl：
  - 项目地址：https://github.com/sliveryou/goctl
