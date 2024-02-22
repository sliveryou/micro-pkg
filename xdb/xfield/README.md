# gorm/gen 字段拓展包 xfield

## 介绍

[xfield](https://github.com/sliveryou/micro-pkg/tree/main/xdb/xfield) 设计是作为 [gorm/gen](https://github.com/go-gorm/gen) 的一个字段拓展包。

在开发过程中，一些基础 SQL 语句，通过 [gorm/gen 文档](https://gorm.io/zh_CN/gen/index.html)介绍的方式，可以较为优雅简便的实现。

但一旦遇上一些复杂字段构建或复杂查询需求，gen 往往就无能为力了，开发小伙伴只能退而求其次使用原生 [gorm](https://gorm.io/zh_CN/docs/index.html) 方法或硬编码 sql 语句。

这不优雅，也容易产生如 sql 注入等安全问题。

究其原因是 gen 并不提供构建原始 sql 串的方法，之前社区也有讨论过，他们开发者的回复是：
> "Raw SQL may lead to some unexpected SQL Injection vulnerabilities. So we are very cautious about the use of raw SQL"

我个人感觉不提供此类方法反而更加容易导致安全问题了。

## 设计思路

我们看一个实现 `field.Expr` 接口的结构

```go
// gorm.io/gen@v0.3.25/field/expr.go

// Expr a query expression about field
type Expr interface {
	// Clause Expression interface
	Build(clause.Builder)

	As(alias string) Expr
	IColumnName
	BuildColumn(*gorm.Statement, ...BuildOpt) sql
	BuildWithArgs(*gorm.Statement) (query sql, args []interface{})
	RawExpr() expression

	// col operate expression
	AddCol(col Expr) Expr
	SubCol(col Expr) Expr
	MulCol(col Expr) Expr
	DivCol(col Expr) Expr
	ConcatCol(cols ...Expr) Expr

	// implement Condition
	BeCond() interface{}
	CondError() error

	expression() clause.Expression
}

type expr struct {
	col clause.Column

	e         clause.Expression
	buildOpts []BuildOpt
}
```

由于 gen 没有暴露 `expr` 的结构，并且其中的 e（clause.Expression）字段也是不可导出的，想要替换成 raw sql，只能通过反射的方式将 expr 中的 e 替换。

所以 `xfield.NewRaw` 函数里主要也是做这件事：

```go
// replace 替换 Raw 的 field.Expr 字段
//
//	由于 gorm.io/gen 没有暴露 expr 的结构，并且其中的 e（clause.Expression）字段也是不可导出的，
//	所以只能通过反射的方式将 expr 中的 e 替换。
func (r *Raw) replace() {
	// 新建一个空 expr
	emptyExpr := reflect.ValueOf(field.EmptyExpr())

	// 通过反射创建自定义 expr 类型
	expr := reflect.New(emptyExpr.Type()).Elem()
	expr.Set(emptyExpr)
	// 获取自定义 expr 类型的 e 字段
	e := expr.FieldByName("e")
	eElem := reflect.NewAt(e.Type(), unsafe.Pointer(e.UnsafeAddr())).Elem()
	// 修改自定义 expr 类型的 e 字段为 clause.NamedExpr
	eElem.Set(reflect.ValueOf(clause.NamedExpr{SQL: r.SQL, Vars: convertVars(r.Vars)}))

	// 获取 Raw 的 field.Expr 字段
	rawExpr := reflect.ValueOf(r).Elem().FieldByName("Expr")
	rawExprElem := reflect.NewAt(rawExpr.Type(), unsafe.Pointer(rawExpr.UnsafeAddr())).Elem()
	// 修改 Raw 的 field.Expr 字段为自定义 expr
	rawExprElem.Set(expr)
}
```

这会有个弊端：假如 gen 的开发者，修改了字段名称（毕竟它是内部不可导出字段），这段代码将会 panic，所以更新 gen 版本的时候，需要通过 xfield 包的所有测试用例或依据版本变化进行相应调整（不过有一说一，已经很久没有变化了）。

`xfield.NewRawCondition` 比上面说的要好办，只要设计实现 `gen.Condition` 的接口就行：

```go
// gorm.io/gen@v0.3.25/interface.go

type (
	// Condition query condition
	// field.Expr and subquery are expect value
	Condition interface {
		BeCond() interface{}
		CondError() error
	}
)
```

唯一有个问题，就是想要获取 `field.NewString("my_table", "name")` 字段的表名，按照目前 `field.Expr` 的接口，是拿不到的，原因和上面一样：`expr.col` 字段也是不可导出的。

当然 go 也有办法访问结构体不可导出的字段，得使用 `unsafe.Pointer`，整体逻辑如下：

```go
// getField 获取结构体对应字段
func getField(s any, fieldName string) any {
	defer func() { recover() }()

	if s == nil || fieldName == "" {
		return nil
	}

	v := reflect.ValueOf(s)
	newV := reflect.New(v.Type()).Elem()
	newV.Set(v)

	if newV.Kind() == reflect.Struct {
		f := newV.FieldByName(fieldName)
		if f.IsValid() {
			f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
			return f.Interface()
		}
	}

	return nil
}

// getColumn 获取 field.Expr 包含的 clause.Column 信息
func getColumn(f any) *clause.Column {
	fe, ok := f.(field.Expr)
	if !ok {
		return nil
	}

	expr := getField(fe, "expr")
	if expr == nil {
		return nil
	}

	col := getField(expr, "col")
	if col == nil {
		return nil
	}

	column, ok := col.(clause.Column)
	if !ok {
		return nil
	}

	return &column
}
```

## 使用

可以参考 [xfield_example_test.go](xfield_example_test.go) 文件中的例子，以下是一个使用例子：

示例：

```go
f := l.svcCtx.Q.Flow
fq := f.WithContext(l.ctx).Debug().
    Select(
        f.SrcIP,
        f.DstIP,
        f.SrcIP.Count().As("count"),
        xfield.NewRaw(
            "GROUP_CONCAT(DISTINCT ? ORDER BY ? ASC SEPARATOR ',') AS ?",
            f.Pact, f.Pact, "pacts",
        ),
    ).
    Where(
        f.SrcIP.NeqCol(f.DstIP),
        xfield.NewRawCondition(
			"? BETWEEN ? AND ?",
            f.Timestamp, time.UnixMilli(in.GetStartAt()), time.UnixMilli(in.GetEndAt()),
		), 
	).
    Group(
        f.SrcIP,
        f.DstIP,
    ).
    Order(
        field.NewField("", "count").Desc(),
    )
```
