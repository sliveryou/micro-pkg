package excel

import (
	"io"

	"github.com/pkg/errors"
	excelize "github.com/xuri/excelize/v2"
)

const (
	// DefaultSheet 默认表
	DefaultSheet = "Sheet1"
)

// GetRows 获取 excel 表上的所有行数据
func GetRows(r io.Reader, sheet string, opt ...excelize.Options) ([][]string, error) {
	f, err := excelize.OpenReader(r, opt...)
	if err != nil {
		return nil, errors.WithMessage(err, "excelize.OpenReader err")
	}
	defer f.Close()

	rows, err := f.GetRows(sheet, opt...)
	if err != nil {
		return nil, errors.WithMessage(err, "f.GetRows err")
	}

	return rows, nil
}

// GetFilteredRows 获取 excel 表上的所有大于等于指定行长度的行数据（忽略前 skipRow 行）
func GetFilteredRows(r io.Reader, sheet string, rowLength, skipRow int, opt ...excelize.Options) ([][]string, error) {
	var filteredRows [][]string
	if err := ReadRows(r, sheet, func(rowNum int, columns []string) bool {
		if rowNum > skipRow && len(columns) >= rowLength {
			filteredRows = append(filteredRows, columns)
		}
		return true
	}); err != nil {
		return nil, errors.WithMessage(err, "ReadRows err")
	}

	return filteredRows, nil
}

// ReadHandler 流式读取处理器
type ReadHandler func(rowNum int, columns []string) (isContinue bool)

// ReadRows 流式读取处理 excel 表上的行数据
func ReadRows(r io.Reader, sheet string, handler ReadHandler, opt ...excelize.Options) error {
	f, err := excelize.OpenReader(r, opt...)
	if err != nil {
		return errors.WithMessage(err, "excelize.OpenReader err")
	}
	defer f.Close()

	rows, err := f.Rows(sheet)
	if err != nil {
		return errors.WithMessage(err, "f.Rows err")
	}
	defer rows.Close()

	rowNum := 0
	for rows.Next() {
		rowNum++
		columns, err := rows.Columns(opt...)
		if err != nil {
			return errors.WithMessagef(err, "rowNum: %d, rows.Columns err", rowNum)
		}
		if isContinue := handler(rowNum, columns); !isContinue {
			return nil
		}
	}

	return nil
}

// WriteHandler 流式写入处理器
type WriteHandler func(rowNum int) (columns []any, needWrite, isContinue bool)

// WriteRows 流式写入行数据至指定 excel 表中
func WriteRows(r io.Reader, sheet, saveAs string, handler WriteHandler, opt ...excelize.Options) error {
	f, err := excelize.OpenReader(r, opt...)
	if err != nil {
		return errors.WithMessage(err, "excelize.OpenReader err")
	}
	defer f.Close()

	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		return errors.WithMessage(err, "f.NewStreamWriter err")
	}

	rowNum := 0
	for {
		rowNum++
		columns, needWrite, isContinue := handler(rowNum)
		if !isContinue {
			break
		}
		if needWrite {
			cellName, err := excelize.CoordinatesToCellName(1, rowNum)
			if err != nil {
				return errors.WithMessagef(err, "rowNum: %d, excelize.CoordinatesToCellName err", rowNum)
			}
			if err := sw.SetRow(cellName, columns); err != nil {
				return errors.WithMessagef(err, "rowNum: %d, cellName: %s, sw.SetRow err", rowNum, cellName)
			}
		}
	}

	if err := sw.Flush(); err != nil {
		return errors.WithMessage(err, "sw.Flush err")
	}

	if err := f.SaveAs(saveAs); err != nil {
		return errors.WithMessage(err, "f.SaveAs err")
	}

	return nil
}
