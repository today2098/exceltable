package exceltable

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	csvTag   string = "csv"
	excelTag string = "excel"

	defaultTableStyle = "TableStyleMedium6"
)

type sheetRule struct {
	predKey predKeyType
	styleID int
}

type Sheet[M any] struct {
	File            *File
	name            string
	x, y            int
	row, tableWidth int
	numField        int
	skip            []bool
	header          []any
	rulesList       [][]*sheetRule
}

func NewSheet[M any](f *File, name, cell string, active bool) (*Sheet[M], error) {
	idx, err := f.File.NewSheet(name)
	if err != nil {
		return nil, err
	}
	if active {
		f.File.SetActiveSheet(idx)
	}

	x, y, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return nil, err
	}

	tableWidth, numFiled, skip, header, rulesList, err := getFieldValue(reflect.TypeFor[M](), f.rules)
	if err != nil {
		return nil, err
	}

	return &Sheet[M]{
		File:       f,
		name:       name,
		x:          x,
		y:          y,
		row:        0,
		tableWidth: tableWidth,
		numField:   numFiled,
		skip:       skip,
		header:     header,
		rulesList:  rulesList,
	}, nil
}

func (s *Sheet[M]) SetHeader() error {
	for col := range s.tableWidth {
		if err := s.setCellValue(col, 0, s.header[col]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sheet[M]) SetRow(obj *M) error {
	s.row++

	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	col := 0
	for i := range s.numField {
		if s.skip[i] {
			continue
		}

		field := v.Field(i)
		for field.Kind() == reflect.Pointer && !field.IsNil() {
			field = field.Elem()
		}
		if err := s.setCellValue(col, s.row, field.Interface()); err != nil {
			return err
		}

		for _, rule := range s.rulesList[col] {
			b, err := verifyByPred(ptrV, v, field, rule.predKey)
			if err != nil {
				return err
			}
			if b {
				if err := s.setCellStyle(col, s.row, rule.styleID); err != nil {
					return err
				}
				break // NOTE: Break to prevent overwriting.
			}
		}

		col++
	}

	return nil
}

func (s *Sheet[M]) setCellValue(col, row int, val any) error {
	return s.File.File.SetCellValue(s.name, s.coordinatesToCellName(col, row), val)
}

func (s *Sheet[M]) setCellStyle(col, row, styleId int) error {
	cell := s.coordinatesToCellName(col, row)
	return s.File.File.SetCellStyle(s.name, cell, cell, styleId)
}

func (s *Sheet[M]) AddDefaultTable() error {
	return s.AddTable(defaultTableStyle)
}

func (s *Sheet[M]) AddTable(styleName string) error {
	return s.File.File.AddTable(s.name, s.newTable(styleName))
}

func (s *Sheet[M]) newTable(styleName string) *excelize.Table {
	topLeftCell := s.coordinatesToCellName(0, 0)
	bottomRightCell := s.coordinatesToCellName(max(s.tableWidth-1, 1), max(s.row, 1))
	return &excelize.Table{
		Range:     fmt.Sprintf("%s:%s", topLeftCell, bottomRightCell),
		Name:      fmt.Sprintf("%sTable", s.name),
		StyleName: styleName,
	}
}

func (s *Sheet[M]) coordinatesToCellName(col, row int) string {
	cell, err := excelize.CoordinatesToCellName(s.x+col, s.y+row)
	if err != nil {
		panic(err)
	}
	return cell
}

func getFieldValue(t reflect.Type, fileRules []*fileRule) (tableWidth, numFiled int, skip []bool, header []any, rulesList [][]*sheetRule, err error) {
	if t.Kind() != reflect.Struct {
		return 0, 0, nil, nil, nil, ErrNotStructType
	}

	tableWidth, numFiled = 0, t.NumField()
	skip = make([]bool, numFiled)
	header = make([]any, 0, numFiled)
	rulesList = make([][]*sheetRule, 0, numFiled)
	for i := range numFiled {
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
		for _, fileRule := range fileRules {
			for key := range strings.SplitSeq(field.Tag.Get(fileRule.tag), ",") {
				rules = append(rules, &sheetRule{key, fileRule.styleID})
			}
		}
		rulesList = append(rulesList, rules)
	}

	return
}
