package exceltable

import (
	"reflect"

	"github.com/xuri/excelize/v2"
)

// SheetWithStreamWriter provides methods to write data of type M into spreadsheet table using excelize.StreamWriter.
type SheetWithStreamWriter[M any] struct {
	*sheetBase[M]
	*excelize.StreamWriter
}

// NewSheetWithStreamWriter creates a new exceltable.SheetWithStreamWriter with the given sheet name and starting cell.
//
//	ssw, _ := exceltable.NewSheetWithStreamWriter[YourStruct](f, "NewSheet", "A1", true)
func NewSheetWithStreamWriter[M any](f *File, name, cell string, active bool) (*SheetWithStreamWriter[M], error) {
	sb, err := newSheetBase[M](f, name, cell, active)
	if err != nil {
		return nil, err
	}

	streamWriter, err := f.File.NewStreamWriter(name)
	if err != nil {
		return nil, err
	}

	return &SheetWithStreamWriter[M]{sb, streamWriter}, nil
}

// SetHeader writes the header row to the table.
//
// It must be called before writing any data rows.
func (ssw *SheetWithStreamWriter[M]) SetHeader() error {
	return ssw.StreamWriter.SetRow(ssw.coordinatesToCellName(0, 0), ssw.header)
}

// SetRow writes a row of data to the table.
func (ssw *SheetWithStreamWriter[M]) SetRow(obj *M) error {
	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	values := make([]any, 0, ssw.tableWidth)
	col := 0
	for i := range ssw.numField {
		if ssw.skip[i] {
			continue
		}

		field := v.Field(i)
		for field.Kind() == reflect.Pointer && !field.IsNil() {
			field = field.Elem()
		}

		styleID := 0
		for _, rule := range ssw.rulesList[col] {
			pred := rule.bind(ptrV)
			b, err := callPredicate(pred, field)
			if err != nil {
				return err
			}

			if b {
				styleID = rule.styleID
				break // NOTE: Break to prevent overwriting.
			}
		}

		values = append(values, &excelize.Cell{
			StyleID: styleID,
			Value:   field.Interface(),
		})
		col++
	}

	cell := ssw.coordinatesToCellName(0, ssw.row)
	if err := ssw.StreamWriter.SetRow(cell, values); err != nil {
		return err
	}

	ssw.row++
	return nil
}

// AddDefaultTable creates a table with the default style to the sheet.
//
// It must be called after writing all data rows.
func (ssw *SheetWithStreamWriter[M]) AddDefaultTable() error {
	return ssw.AddTable(DefaultTableStyle)
}

// AddTable creates a table with the specified style name to the sheet.
//
// It must be called after writing all data rows.
func (ssw *SheetWithStreamWriter[M]) AddTable(styleName string) error {
	return ssw.StreamWriter.AddTable(ssw.newTable(styleName))
}
