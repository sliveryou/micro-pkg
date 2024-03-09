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
func GetRows(r io.Reader, sheet string, opts ...excelize.Options) ([][]string, error) {
	if sheet == "" {
		sheet = DefaultSheet
	}

	f, err := excelize.OpenReader(r, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "open reader err")
	}
	defer f.Close()

	rows, err := f.GetRows(sheet, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "get rows err")
	}

	return rows, nil
}

// GetFilteredRows 获取 excel 表上的所有大于等于指定行长度的行数据（忽略前 skipRow 行）
func GetFilteredRows(r io.Reader, sheet string, rowLength, skipRow int, opts ...excelize.Options) ([][]string, error) {
	var filteredRows [][]string
	if err := ReadRows(r, sheet, func(rowNum int, columns []string) bool {
		if rowNum > skipRow && len(columns) >= rowLength {
			filteredRows = append(filteredRows, columns)
		}
		return true
	}, opts...); err != nil {
		return nil, errors.WithMessage(err, "read rows err")
	}

	return filteredRows, nil
}

// ReadHandler 流式读取处理器
type ReadHandler func(rowNum int, columns []string) (isContinue bool)

// ReadRows 流式读取处理 excel 表上的行数据
func ReadRows(r io.Reader, sheet string, handler ReadHandler, opts ...excelize.Options) error {
	if sheet == "" {
		sheet = DefaultSheet
	}

	f, err := excelize.OpenReader(r, opts...)
	if err != nil {
		return errors.WithMessage(err, "open reader err")
	}
	defer f.Close()

	rows, err := f.Rows(sheet)
	if err != nil {
		return errors.WithMessage(err, "get rows iterator err")
	}
	defer rows.Close()

	rowNum := 0
	for rows.Next() {
		rowNum++
		columns, err := rows.Columns(opts...)
		if err != nil {
			return errors.WithMessagef(err, "get row columns err, row num: %d", rowNum)
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
func WriteRows(r io.Reader, sheet, saveAs string, handler WriteHandler, opts ...excelize.Options) error {
	if sheet == "" {
		sheet = DefaultSheet
	}

	f, err := excelize.OpenReader(r, opts...)
	if err != nil {
		return errors.WithMessage(err, "open reader err")
	}
	defer f.Close()

	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		return errors.WithMessage(err, "new stream writer err")
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
				return errors.WithMessagef(err, "coordinates to cell name err, row num: %d", rowNum)
			}
			if err := sw.SetRow(cellName, columns); err != nil {
				return errors.WithMessagef(err, "set row err, row num: %d, cell name: %s", rowNum, cellName)
			}
		}
	}

	if err := sw.Flush(); err != nil {
		return errors.WithMessage(err, "stream writer flush err")
	}

	if err := f.SaveAs(saveAs); err != nil {
		return errors.WithMessagef(err, "sheet save as %s err", saveAs)
	}

	return nil
}
