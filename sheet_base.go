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

type sheetBaseInterface interface {
	AddTable(string) error
}

type sheetBase[M any] struct {
	sheetBaseInterface
	File            *File
	name            string
	x, y            int
	row, tableWidth int
	numField        int
	skip            []bool
	header          []any
	rulesList       [][]*sheetRule
}

func (s *sheetBase[M]) construct(f *File, name, cell string, active bool) error {
	s.File, s.name, s.row = f, name, 1

	idx, err := s.File.File.NewSheet(s.name)
	if err != nil {
		return err
	}
	if active {
		f.File.SetActiveSheet(idx)
	}

	if s.x, s.y, err = excelize.CellNameToCoordinates(cell); err != nil {
		return err
	}

	t := reflect.TypeFor[M]()
	if t.Kind() != reflect.Struct {
		return ErrNotStructType
	}

	s.tableWidth, s.numField = 0, t.NumField()
	s.skip = make([]bool, s.numField)
	s.header = make([]any, 0, s.numField)
	s.rulesList = make([][]*sheetRule, 0, s.numField)
	for i := range s.numField {
		field := t.Field(i)
		if field.PkgPath != "" { // field is unexported.
			s.skip[i] = true
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
			s.skip[i] = true
			continue
		}

		s.header = append(s.header, h)
		s.tableWidth++

		rules := make([]*sheetRule, 0)
		for _, fileRule := range s.File.rules {
			for key := range strings.SplitSeq(field.Tag.Get(fileRule.tag), ",") {
				rules = append(rules, &sheetRule{key, fileRule.styleID})
			}
		}
		s.rulesList = append(s.rulesList, rules)
	}

	return nil
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

func (s *sheetBase[M]) coordinatesToCellName(col, row int) string {
	cell, err := excelize.CoordinatesToCellName(s.x+col, s.y+row)
	if err != nil {
		panic(err)
	}
	return cell
}
