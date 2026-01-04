package exceltable

// Sheet provides methods to write data of type M into a spreadsheet table.
type Sheet[M any] struct {
	*sheetBase[M]
}

// NewSheet creates a new *exceltable.Sheet with the given sheet name and starting cell.
//
//	s, _ := exceltable.NewSheet[YourStruct](f, "NewSheet", "A1", true)
func NewSheet[M any](f *File, name, cell string, active bool) (*Sheet[M], error) {
	sb, err := newSheetBase[M](f, name, cell, active)
	if err != nil {
		return nil, err
	}
	return &Sheet[M]{
		sheetBase: sb,
	}, nil
}

// SetHeader writes the header row to the table.
func (s *Sheet[M]) SetHeader() error {
	for col, columnName := range s.getHeader() {
		if err := s.setCellValue(col, 0, columnName); err != nil {
			return err
		}
	}
	return nil
}

// SetRow writes a row of data to the table.
func (s *Sheet[M]) SetRow(obj *M) error {
	cellValues, err := s.parseToCellValueList(obj)
	if err != nil {
		return err
	}

	for col, v := range cellValues {
		if err := s.setCellValue(col, s.row, v.value); err != nil {
			return err
		}

		if v.styleID >= 0 {
			if err := s.setCellStyle(col, s.row, v.styleID); err != nil {
				return err
			}
		}
	}

	s.row++
	return nil
}

func (s *Sheet[M]) setCellValue(col, row int, val any) error {
	return s.file.File.SetCellValue(s.name, s.coordinatesToCellName(col, row), val)
}

func (s *Sheet[M]) setCellStyle(col, row, styleID int) error {
	cell := s.coordinatesToCellName(col, row)
	return s.file.File.SetCellStyle(s.name, cell, cell, styleID)
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
	return s.file.File.AddTable(s.name, s.newTable(styleName))
}
