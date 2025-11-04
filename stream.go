package exceltable

import (
	"reflect"

	"github.com/xuri/excelize/v2"
)

type SheetWithStreamWriter[M any] struct {
	Sheet        *Sheet[M]
	StreamWriter *excelize.StreamWriter
}

func NewSheetWithStreamWriter[M any](f *File, name, cell string, active bool) (*SheetWithStreamWriter[M], error) {
	s, err := NewSheet[M](f, name, cell, active)
	if err != nil {
		return nil, err
	}

	streamWriter, err := s.File.File.NewStreamWriter(name)
	if err != nil {
		return nil, err
	}

	return &SheetWithStreamWriter[M]{
		Sheet:        s,
		StreamWriter: streamWriter,
	}, nil
}

func (ssw *SheetWithStreamWriter[M]) SetHeader() error {
	cell := ssw.Sheet.coordinatesToCellName(0, 0)
	return ssw.StreamWriter.SetRow(cell, ssw.Sheet.header)
}

func (ssw *SheetWithStreamWriter[M]) SetRow(obj *M) error {
	ssw.Sheet.row++

	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	values := make([]any, 0, ssw.Sheet.tableWidth)
	col := 0
	for i := range ssw.Sheet.numField {
		if ssw.Sheet.skip[i] {
			continue
		}

		field := v.Field(i)
		for field.Kind() == reflect.Pointer && !field.IsNil() {
			field = field.Elem()
		}

		styleID := 0
		for _, rule := range ssw.Sheet.rulesList[col] {
			b, err := verifyByPred(ptrV, v, field, rule.predKey)
			if err != nil {
				return err
			}
			if b {
				styleID = rule.styleID
				break
			}
		}

		values = append(values, &excelize.Cell{
			StyleID: styleID,
			Value:   field.Interface(),
		})
		col++
	}

	cell := ssw.Sheet.coordinatesToCellName(0, ssw.Sheet.row)
	return ssw.StreamWriter.SetRow(cell, values)
}

func (ssw *SheetWithStreamWriter[M]) AddTable() error {
	return ssw.StreamWriter.AddTable(ssw.Sheet.newTable(defaultTableStyle))
}

func (ssw *SheetWithStreamWriter[M]) Flush() error {
	return ssw.StreamWriter.Flush()
}
