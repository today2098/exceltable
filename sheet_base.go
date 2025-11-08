package exceltable

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Tags indicating header values.
const (
	csvTag   string = "csv"
	excelTag string = "excel"
)

// Default table style name.
const DefaultTableStyle = "TableStyleMedium6"

// sheetRule represents relation between predicate key and style ID.
type sheetRule struct {
	predKey predKeyType
	styleID int
}

type sheetBase[M any] struct {
	File       *File
	name       string         // sheet name
	x, y       int            // starting cell coordinates
	row        int            // current number of rows
	tableWidth int            // table width (number of columns)
	numField   int            // number of struct fields
	skip       []bool         // whether to skip each struct field
	header     []any          // header values
	rulesList  [][]*sheetRule // rules for each column
}

func newSheetBase[M any](f *File, name, cell string, active bool) (*sheetBase[M], error) {
	t := reflect.TypeFor[M]()
	if t.Kind() != reflect.Struct {
		return nil, ErrNotStructType
	}

	idx, err := f.NewSheet(name)
	if err != nil {
		return nil, err
	}
	if active {
		f.SetActiveSheet(idx)
	}

	x, y, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return nil, err
	}

	tableWidth, numField := 0, t.NumField()
	skip := make([]bool, numField)
	header := make([]any, 0, numField)
	rulesList := make([][]*sheetRule, 0, numField)
	for i := range numField {
		field := t.Field(i)
		if field.PkgPath != "" { // field is unexported.
			skip[i] = true
			continue
		}

		h := field.Tag.Get(excelTag)
		if h == "" {
			h = field.Tag.Get(csvTag)
		}

		switch h {
		case "":
			h = field.Name
		case "-":
			skip[i] = true
			continue
		}

		header = append(header, h)
		tableWidth++

		rules := make([]*sheetRule, 0)
		for _, rule := range f.rules {
			for key := range strings.SplitSeq(field.Tag.Get(rule.tag), ",") {
				rules = append(rules, &sheetRule{key, rule.styleID})
			}
		}
		rulesList = append(rulesList, rules)
	}

	return &sheetBase[M]{
		File:       f,
		name:       name,
		x:          x,
		y:          y,
		row:        1,
		tableWidth: tableWidth,
		numField:   numField,
		skip:       skip,
		header:     header,
		rulesList:  rulesList,
	}, nil
}

func (s *sheetBase[M]) newTable(styleName string) *excelize.Table {
	topLeftCell := s.coordinatesToCellName(0, 0)
	bottomRightCell := s.coordinatesToCellName(max(s.tableWidth-1, 1), max(s.row-1, 1))
	return &excelize.Table{
		Range:     fmt.Sprintf("%s:%s", topLeftCell, bottomRightCell),
		Name:      fmt.Sprintf("%sTable", s.name),
		StyleName: styleName,
	}
}

func (s *sheetBase[M]) coordinatesToCellName(col, row int, abs ...bool) string {
	cell, err := excelize.CoordinatesToCellName(s.x+col, s.y+row, abs...)
	if err != nil {
		panic(err) // This should never happen when col and row are non-negative.
	}
	return cell
}
