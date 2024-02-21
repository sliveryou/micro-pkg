package xfield

import (
	"reflect"
	"unsafe"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

// 实现相关接口
var (
	_ field.Expr    = Raw{}
	_ gen.Condition = RawCondition{}
)

// Raw 原始字段
type Raw struct {
	Column string
	field.Expr
}

// NewRaw 新建原始字段
func NewRaw(column string) Raw {
	r := Raw{Column: column}
	r.replace()

	return r
}

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
	// 修改自定义 expr 类型的 e 字段为 Raw
	eElem.Set(reflect.ValueOf(*r))

	// 获取 Raw 的 field.Expr 字段
	rawExpr := reflect.ValueOf(r).Elem().FieldByName("Expr")
	rawExprElem := reflect.NewAt(rawExpr.Type(), unsafe.Pointer(rawExpr.UnsafeAddr())).Elem()
	// 修改 Raw 的 field.Expr 字段为自定义 expr
	rawExprElem.Set(expr)
}

// Build 实现 Build 方法
func (r Raw) Build(builder clause.Builder) {
	_, _ = builder.WriteString(r.Column)
}

// Tabler 表接口
type Tabler interface {
	Alias() string
	TableName() string
}

// Field 字段结构
type Field struct {
	Expr  field.Expr
	Table Tabler
}

// RawCondition 原始条件
type RawCondition struct {
	field.Field
	sql  string
	vars []any
}

// NewRawCondition 新建原始条件
func NewRawCondition(sql string, vars ...any) RawCondition {
	return RawCondition{
		sql:  sql,
		vars: vars,
	}
}

// BeCond 实现 BeCond 方法
func (m RawCondition) BeCond() any {
	var vars []any

	for _, v := range m.vars {
		switch vt := v.(type) {
		case Field:
			column := clause.Column{
				Name: vt.Expr.ColumnName().String(),
				Raw:  false,
			}
			if vt.Table != nil {
				column.Table = vt.Table.TableName()
				if vt.Table.Alias() != "" {
					column.Table = vt.Table.Alias()
				}
			}
			vars = append(vars, column)
		case field.Expr:
			column := clause.Column{
				Name: vt.ColumnName().String(),
				Raw:  false,
			}
			vars = append(vars, column)
		default:
			vars = append(vars, v)
		}
	}

	return clause.NamedExpr{SQL: m.sql, Vars: vars}
}

// CondError 实现 CondError 方法
func (RawCondition) CondError() error { return nil }
