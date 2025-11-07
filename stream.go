package exceltable

import (
	"reflect"

	"github.com/xuri/excelize/v2"
)

type SheetWithStreamWriter[M any] struct {
	sheetBase[M]
	StreamWriter *excelize.StreamWriter
}

func NewSheetWithStreamWriter[M any](f *File, name, cell string, active bool) (*SheetWithStreamWriter[M], error) {
	s := &SheetWithStreamWriter[M]{}
	err := s.construct(f, name, cell, active)
	if err != nil {
		return nil, err
	}

	s.StreamWriter, err = s.File.File.NewStreamWriter(s.name)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (ssw *SheetWithStreamWriter[M]) SetHeader() error {
	cell := ssw.coordinatesToCellName(0, 0)
	return ssw.StreamWriter.SetRow(cell, ssw.header)
}

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
			b, err := verifyByPred(ptrV, field, rule.predKey)
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

	cell := ssw.coordinatesToCellName(0, ssw.row)
	if err := ssw.StreamWriter.SetRow(cell, values); err != nil {
		return err
	}

	ssw.row++
	return nil
}

func (ssw *SheetWithStreamWriter[M]) AddDefaultTable() error {
	return ssw.AddTable(defaultTableStyle)
}

func (ssw *SheetWithStreamWriter[M]) AddTable(styleName string) error {
	return ssw.StreamWriter.AddTable(ssw.newTable(styleName))
}

func (ssw *SheetWithStreamWriter[M]) Flush() error {
	return ssw.StreamWriter.Flush()
}
