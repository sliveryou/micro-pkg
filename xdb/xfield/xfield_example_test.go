package xfield_test

import (
	"gorm.io/gen/field"

	"github.com/sliveryou/micro-pkg/xdb/xfield"
)

//nolint:testableexamples
func ExampleNewRaw() {
	name1 := field.NewString("my_table", "name")
	name2 := field.NewString("", "name")

	// 使用方法1
	xfield.NewRaw("GROUP_CONCAT(DISTINCT `name` ORDER BY `name` ASC SEPARATOR ',') AS `names`")

	// 使用方法2
	xfield.NewRaw("GROUP_CONCAT(DISTINCT ? ORDER BY ? ASC SEPARATOR ',') AS ?",
		name1, name1, "names")

	// 使用方法3
	xfield.NewRaw("GROUP_CONCAT(DISTINCT ? ORDER BY ? ASC SEPARATOR ',') AS ?",
		name2, name2, "names")
}

//nolint:testableexamples
func ExampleNewRawCondition() {
	version := field.NewString("my_table", "version")

	// 使用方法1
	xfield.NewRawCondition("UPPER(version) = 'SOME_VERSION'")

	// 使用方法2
	xfield.NewRawCondition("UPPER(?) = 'SOME_VERSION'", version)

	// 使用方法3
	xfield.NewRawCondition("UPPER(?) = ?", version, "SOME_VERSION")
}
