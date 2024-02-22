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
	field.Expr
	SQL  string
	Vars []any
}

// NewRaw 新建原始字段
func NewRaw(sql string, vars ...any) Raw {
	r := Raw{
		SQL:  sql,
		Vars: vars,
	}
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
	// 修改自定义 expr 类型的 e 字段为 clause.NamedExpr
	eElem.Set(reflect.ValueOf(clause.NamedExpr{SQL: r.SQL, Vars: convertVars(r.Vars)}))

	// 获取 Raw 的 field.Expr 字段
	rawExpr := reflect.ValueOf(r).Elem().FieldByName("Expr")
	rawExprElem := reflect.NewAt(rawExpr.Type(), unsafe.Pointer(rawExpr.UnsafeAddr())).Elem()
	// 修改 Raw 的 field.Expr 字段为自定义 expr
	rawExprElem.Set(expr)
}

// RawCondition 原始条件
type RawCondition struct {
	field.Field
	SQL  string
	Vars []any
}

// NewRawCondition 新建原始条件
func NewRawCondition(sql string, vars ...any) RawCondition {
	return RawCondition{
		SQL:  sql,
		Vars: vars,
	}
}

// BeCond 实现 BeCond 方法
func (rc RawCondition) BeCond() any {
	return clause.NamedExpr{SQL: rc.SQL, Vars: convertVars(rc.Vars)}
}

// CondError 实现 CondError 方法
func (RawCondition) CondError() error { return nil }

// convertVars 转换 vars 列表
func convertVars(vars []any) []any {
	newVars := make([]any, 0, len(vars))

	for _, v := range vars {
		switch vt := v.(type) {
		case field.Expr:
			column := clause.Column{
				Name: vt.ColumnName().String(),
				Raw:  false,
			}
			if c := getColumn(vt); c != nil {
				column = *c
			}
			newVars = append(newVars, column)
		default:
			newVars = append(newVars, v)
		}
	}

	return newVars
}

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
