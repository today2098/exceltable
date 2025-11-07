package exceltable

import (
	"reflect"
)

type Sheet[M any] struct {
	sheetBase[M]
}

func NewSheet[M any](f *File, name, cell string, active bool) (*Sheet[M], error) {
	s := &Sheet[M]{}
	if err := s.construct(f, name, cell, active); err != nil {
		return nil, err
	}

	return s, nil
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
			b, err := verifyByPred(ptrV, field, rule.predKey)
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
	s.row++

	return nil
}

func (s *Sheet[M]) setCellValue(col, row int, val any) error {
	return s.File.File.SetCellValue(s.name, s.coordinatesToCellName(col, row), val)
}

func (s *Sheet[M]) setCellStyle(col, row, styleID int) error {
	cell := s.coordinatesToCellName(col, row)
	return s.File.File.SetCellStyle(s.name, cell, cell, styleID)
}

func (s *Sheet[M]) AddDefaultTable() error {
	return s.AddTable(defaultTableStyle)
}

func (s *Sheet[M]) AddTable(styleName string) error {
	return s.File.File.AddTable(s.name, s.newTable(styleName))
}
