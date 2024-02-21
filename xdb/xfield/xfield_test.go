package xfield_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils/tests"

	"github.com/sliveryou/micro-pkg/xdb/xfield"
)

func TestNewRaw(t *testing.T) {
	f := xfield.NewRaw("GROUP_CONCAT(DISTINCT `name` ORDER BY `name` ASC SEPARATOR ',') AS `names`")
	assert.NotNil(t, f.Expr)
	t.Log(f)
}

func TestNewRawCondition(t *testing.T) {
	db, _ := gorm.Open(tests.DummyDialector{}, nil)
	version := field.NewString("my_table", "version")

	cases := []struct {
		sql    string
		vars   []any
		expect string
	}{
		{
			sql: "UPPER(version) = 'SOME_VERSION'", vars: nil,
			expect: "UPPER(version) = 'SOME_VERSION'",
		},
		{
			sql: "UPPER(?) = ?", vars: []any{version, "SOME_VERSION"},
			expect: "UPPER(`version`) = \"SOME_VERSION\"",
		},
		{
			sql: "UPPER(?) = ?", vars: []any{version, "SOME_VERSION"},
			expect: "UPPER(`version`) = \"SOME_VERSION\"",
		},
		{
			sql: "UPPER(?) = ?",
			vars: []any{
				xfield.Field{
					Expr:  version,
					Table: newMyTabler("", "my_table"),
				},
				"SOME_VERSION",
			},
			expect: "UPPER(`my_table`.`version`) = \"SOME_VERSION\"",
		},
		{
			sql: "UPPER(?) = ?",
			vars: []any{
				xfield.Field{
					Expr:  version,
					Table: newMyTabler("my_table_alias", "my_table"),
				},
				"SOME_VERSION",
			},
			expect: "UPPER(`my_table_alias`.`version`) = \"SOME_VERSION\"",
		},
	}

	for _, c := range cases {
		condition := xfield.NewRawCondition(c.sql, c.vars...)
		namedExpr, ok := condition.BeCond().(clause.NamedExpr)
		assert.True(t, ok)

		stmt := &gorm.Statement{DB: db, Clauses: map[string]clause.Clause{}}
		namedExpr.Build(stmt)
		assert.Equal(t, c.expect, db.Dialector.Explain(stmt.SQL.String(), stmt.Vars...))
	}
}

type myTabler struct {
	alias     string
	tableName string
}

func newMyTabler(alias, tableName string) myTabler {
	return myTabler{alias: alias, tableName: tableName}
}

func (mt myTabler) Alias() string {
	return mt.alias
}

func (mt myTabler) TableName() string {
	return mt.tableName
}
