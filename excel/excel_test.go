package excel

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	excelize "github.com/xuri/excelize/v2"
)

func TestGetRows(t *testing.T) {
	f, err := os.Open("../testdata/test.xlsx")
	require.NoError(t, err)
	defer f.Close()

	// 现在解压后大小最大为 20MB
	rows, err := GetRows(f, DefaultSheet, excelize.Options{UnzipSizeLimit: 20 << 20, UnzipXMLSizeLimit: 20 << 20})
	require.NoError(t, err)

	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
	assert.Len(t, rows, 4)
}

func TestGetFilteredRows(t *testing.T) {
	f, err := os.Open("../testdata/test.xlsx")
	require.NoError(t, err)
	defer f.Close()

	rows, err := GetFilteredRows(f, DefaultSheet, 7, 1)
	require.NoError(t, err)

	for _, row := range rows {
		for _, colCell := range row {
			fmt.Print(colCell, "\t")
		}
		fmt.Println()
	}
	assert.Len(t, rows, 3)
}

func TestReadRows(t *testing.T) {
	f, err := os.Open("../testdata/test.xlsx")
	require.NoError(t, err)
	defer f.Close()

	var rows [][]string
	err = ReadRows(f, DefaultSheet, func(rowNum int, columns []string) bool {
		if rowNum == 4 {
			return false
		}
		if rowNum > 1 && len(columns) >= 7 {
			rows = append(rows, columns)
			for _, column := range columns {
				fmt.Print(column, "\t")
			}
			fmt.Println()
		}
		return true
	})
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestWriteRows(t *testing.T) {
	f, err := os.Open("../testdata/test.xlsx")
	require.NoError(t, err)
	defer f.Close()

	rows := [][]any{
		{"赵六", "Java开发", "13400000000", "2021-12-06", 1, 1, "d"},
		{"陈七", "产品", "13500000000", "2021-12-06", 1, 1, "e"},
		{"杨八", "财务", "13600000000", "2021-12-06", 1, 1, "f"},
	}

	handler := func(rowNum int) (columns []any, needWrite, isContinue bool) {
		if rowNum > 1 && rowNum < 5 {
			return rows[rowNum-2], true, true
		} else if rowNum >= 5 {
			return nil, false, false
		}
		return nil, false, true
	}

	err = WriteRows(f, "Sheet1", "../testdata/test-write.xlsx", handler)
	require.NoError(t, err)
}
