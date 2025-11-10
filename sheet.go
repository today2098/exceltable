package exceltable

import "reflect"

// Sheet provides methods to write data of type M into spreadsheet table.
type Sheet[M any] struct {
	*sheetBase[M]
}

// NewSheet creates a new exceltable.Sheet with the given sheet name and starting cell.
//
//	s, _ := exceltable.NewSheet[YourStruct](f, "NewSheet", "A1", true)
func NewSheet[M any](f *File, name, cell string, active bool) (*Sheet[M], error) {
	sb, err := newSheetBase[M](f, name, cell, active)
	if err != nil {
		return nil, err
	}

	return &Sheet[M]{sb}, nil
}

// SetHeader writes the header row to the table.
func (s *Sheet[M]) SetHeader() error {
	for col := range s.tableWidth {
		if err := s.setCellValue(col, 0, s.header[col]); err != nil {
			return err
		}
	}
	return nil
}

// SetRow writes a row of data to the table.
func (s *Sheet[M]) SetRow(obj *M) error {
	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	col := 0
	for i := range s.numField {
		if s.skip[i] {
			continue
		}

		field := v.Field(i)
		if err := s.setCellValue(col, s.row, getUnderlyingValue(field)); err != nil {
			return err
		}

		for _, rule := range s.rulesList[col] {
			pred := rule.bind(ptrV)
			b, err := callPredicate(pred, field)
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

// AddDefaultTable creates a table with the default style to the sheet.
//
// It must be called after writing all data rows.
func (s *Sheet[M]) AddDefaultTable() error {
	return s.AddTable(DefaultTableStyle)
}

// AddTable creates a table with the specified style name to the sheet.
//
// It must be called after writing all data rows.
func (s *Sheet[M]) AddTable(styleName string) error {
	return s.File.File.AddTable(s.name, s.newTable(styleName))
}
