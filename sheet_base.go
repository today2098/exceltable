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
	pred     reflect.Value
	funcT    reflect.Type
	isMethod bool
	styleID  int
}

func newSheetRule(pred reflect.Value, isMethod bool, styleID int) *sheetRule {
	if !isMethod {
		return &sheetRule{
			pred:     pred,
			isMethod: false,
			styleID:  styleID,
		}
	}

	n, m := pred.Type().NumIn(), pred.Type().NumOut()
	in, out := make([]reflect.Type, n), make([]reflect.Type, m)
	for i := range n {
		in[i] = pred.Type().In(i)
	}
	for i := range m {
		out[i] = pred.Type().Out(i)
	}

	return &sheetRule{
		pred:     pred,
		funcT:    reflect.FuncOf(in[1:], out, false),
		isMethod: true,
		styleID:  styleID,
	}
}

func (sr *sheetRule) bind(ptrV reflect.Value) reflect.Value {
	if !sr.isMethod {
		return sr.pred
	}

	return reflect.MakeFunc(sr.funcT, func(in []reflect.Value) []reflect.Value {
		return sr.pred.Call(append([]reflect.Value{ptrV}, in...))
	})
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
	ptrT := reflect.PointerTo(t)

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
				switch key {
				case "", "-":
					// ignore
				default:
					if method, ok := ptrT.MethodByName(key); ok {
						rules = append(rules, newSheetRule(method.Func, true, rule.styleID))
						break
					}

					if function, ok := predicates.Load(key); ok {
						rules = append(rules, newSheetRule(reflect.ValueOf(function), false, rule.styleID))
						break
					}

					return nil, ErrUnknownPredicate
				}
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

func getUnderlyingValue(field reflect.Value) any {
	for field.Kind() == reflect.Pointer && !field.IsNil() {
		field = field.Elem()
	}
	return field.Interface()
}
