package xfield

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils/tests"
)

func TestNewRaw(t *testing.T) {
	db, _ := gorm.Open(tests.DummyDialector{})
	name1 := field.NewString("my_table", "name")
	name2 := field.NewString("", "name")

	cases := []struct {
		sql    string
		vars   []any
		expect string
	}{
		{
			sql:    "GROUP_CONCAT(DISTINCT `name` ORDER BY `name` ASC SEPARATOR ',') AS `names`",
			vars:   nil,
			expect: "GROUP_CONCAT(DISTINCT `name` ORDER BY `name` ASC SEPARATOR ',') AS `names`",
		},
		{
			sql:    "GROUP_CONCAT(DISTINCT ? ORDER BY ? ASC SEPARATOR ',') AS ?",
			vars:   []any{name1, name1, "names"},
			expect: "GROUP_CONCAT(DISTINCT `my_table`.`name` ORDER BY `my_table`.`name` ASC SEPARATOR ',') AS \"names\"",
		},
		{
			sql:    "GROUP_CONCAT(DISTINCT ? ORDER BY ? ASC SEPARATOR ',') AS ?",
			vars:   []any{name2, name2, "names"},
			expect: "GROUP_CONCAT(DISTINCT `name` ORDER BY `name` ASC SEPARATOR ',') AS \"names\"",
		},
	}

	for _, c := range cases {
		r := NewRaw(c.sql, c.vars...)
		assert.NotNil(t, r.Expr)

		e := getField(r.Expr, "e")
		namedExpr, ok := e.(clause.NamedExpr)
		assert.True(t, ok)

		stmt := &gorm.Statement{DB: db}
		namedExpr.Build(stmt)
		assert.Equal(t, c.expect, db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...))
	}
}

func TestNewRawCondition(t *testing.T) {
	db, _ := gorm.Open(tests.DummyDialector{})
	version1 := field.NewString("my_table", "version")
	version2 := field.NewString("", "version")

	cases := []struct {
		sql    string
		vars   []any
		expect string
	}{
		{
			sql:    "UPPER(version) = 'SOME_VERSION'",
			vars:   nil,
			expect: "UPPER(version) = 'SOME_VERSION'",
		},
		{
			sql:    "UPPER(?) = 'SOME_VERSION'",
			vars:   []any{version1},
			expect: "UPPER(`my_table`.`version`) = 'SOME_VERSION'",
		},
		{
			sql:    "UPPER(?) = ?",
			vars:   []any{version1, "SOME_VERSION"},
			expect: "UPPER(`my_table`.`version`) = \"SOME_VERSION\"",
		},
		{
			sql:    "UPPER(?) = 'SOME_VERSION'",
			vars:   []any{version2},
			expect: "UPPER(`version`) = 'SOME_VERSION'",
		},
		{
			sql:    "UPPER(?) = ?",
			vars:   []any{version2, "SOME_VERSION"},
			expect: "UPPER(`version`) = \"SOME_VERSION\"",
		},
	}

	for _, c := range cases {
		condition := NewRawCondition(c.sql, c.vars...)
		namedExpr, ok := condition.BeCond().(clause.NamedExpr)
		assert.True(t, ok)

		stmt := &gorm.Statement{DB: db}
		namedExpr.Build(stmt)
		assert.Equal(t, c.expect, db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...))
	}
}

func Test_getField(t *testing.T) {
	cases := []struct {
		s         any
		field     string
		expectNil bool
	}{
		{s: nil, field: "xxx", expectNil: true},
		{s: 100, field: "xxx", expectNil: true},
		{s: new(int64), field: "xxx", expectNil: true},
		{s: field.NewInt("my_table", "my_column"), field: "xxx", expectNil: true},
		{s: field.NewInt("my_table", "my_column"), field: "expr", expectNil: false},
		{s: field.NewString("my_table", "my_column"), field: "expr", expectNil: false},
		{s: NewRaw("sss"), field: "Expr", expectNil: false},
	}

	for _, c := range cases {
		get := getField(c.s, c.field)
		if c.expectNil {
			assert.Nil(t, get)
		} else {
			assert.NotNil(t, get)
			t.Logf("%+v", get)
		}
	}
}

func Test_getColumn(t *testing.T) {
	fs := field.NewString("my_table", "my_column")
	c := getColumn(fs)
	assert.NotNil(t, c)
	assert.Equal(t, "my_table", c.Table)
	assert.Equal(t, "my_column", c.Name)
	t.Logf("%+v", c)

	db, _ := gorm.Open(tests.DummyDialector{})
	fsq, ok := fs.In([]string{}...).BeCond().(clause.Expression)
	assert.True(t, ok)
	stmt := &gorm.Statement{DB: db}
	fsq.Build(stmt)
	t.Log(db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...))
}
