package exceltable

import (
	"fmt"
	"reflect"

	"github.com/xuri/excelize/v2"
)

// Default table style name.
const DefaultTableStyle = "TableStyleMedium6"

type cellValue struct {
	styleID int
	value   any
}

type sheetBase[M any] struct {
	file       *File
	name       string // sheet name
	x, y       int    // starting cell coordinates
	tableWidth int    // table width (number of columns)
	row        int    // current number of rows
	field      *field // field of type M
}

func newSheetBase[M any](f *File, name, cell string, active bool) (*sheetBase[M], error) {
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

	field, err := f.cache.cache(reflect.TypeFor[M]())
	if err != nil {
		return nil, err
	}

	return &sheetBase[M]{
		file:       f,
		name:       name,
		x:          x,
		y:          y,
		tableWidth: field.r,
		row:        1,
		field:      field,
	}, nil
}

func (sb *sheetBase[M]) getHeader() []any {
	header := make([]any, 0, sb.tableWidth)
	var dfs func(field *field)
	dfs = func(field *field) {
		for _, child := range field.children {
			if child.tag.inline {
				dfs(child)
				continue
			}
			header = append(header, child.tag.columnName)
		}
	}

	dfs(sb.field)
	return header
}

func (sb *sheetBase[M]) parseToCellValueList(obj *M) ([]*cellValue, error) {
	return sb.parseToCellValueListInternal(reflect.ValueOf(obj).Elem(), sb.field, nil)
}

func (sb *sheetBase[M]) parseToCellValueListInternal(v reflect.Value, field *field, parentRule *rule) ([]*cellValue, error) {
	ptrV := v.Addr()

	cellValues := make([]*cellValue, 0, sb.tableWidth)
	for _, child := range field.children {
		fieldV := v.Field(child.fieldIndex)
		rule := parentRule

		for r := range sb.file.rules.iter() {
			if rule != nil && rule.name == r.name {
				break
			}

			if child.rules[r.name].callWithReceiver(ptrV, fieldV) {
				rule = r
				break
			}
		}

		baseFieldV := walkValue(fieldV)
		if child.tag.inline && baseFieldV.Type().Kind() != reflect.Pointer {
			childCellValues, err := sb.parseToCellValueListInternal(baseFieldV, child, rule)
			if err != nil {
				return nil, err
			}
			cellValues = append(cellValues, childCellValues...)
			continue
		}

		value := fmt.Sprint(baseFieldV.Interface())
		if assignableToStringer(child.typ) {
			value = fmt.Sprint(fieldV.Interface())
		}
		if (child.tag.omitEmpty && baseFieldV.Type() == child.typ && baseFieldV.IsZero()) ||
			(child.tag.omitZero && baseFieldV.IsZero()) ||
			(!child.tag.specifyNil && baseFieldV.Type().Kind() == reflect.Pointer) {
			value = ""
		}

		// NOTE: Invalid style ID is greater than or equal to 0.
		// If styleID is -1, no style is applied.
		styleID := -1
		if rule != nil {
			styleID = rule.styleID
		}

		for i := child.l; i < child.r; i++ {
			cellValues = append(cellValues, &cellValue{
				styleID: styleID,
				value:   value,
			})
		}
	}

	return cellValues, nil
}

func (sb *sheetBase[M]) newTable(styleName string) *excelize.Table {
	topLeftCell := sb.coordinatesToCellName(0, 0)
	bottomRightCell := sb.coordinatesToCellName(max(sb.tableWidth-1, 1), max(sb.row-1, 1))
	return &excelize.Table{
		Range:     fmt.Sprintf("%s:%s", topLeftCell, bottomRightCell),
		Name:      fmt.Sprintf("%sTable", sb.name),
		StyleName: styleName,
	}
}

func (sb *sheetBase[M]) coordinatesToCellName(col, row int, abs ...bool) string {
	cell, err := excelize.CoordinatesToCellName(sb.x+col, sb.y+row, abs...)
	if err != nil {
		panic(err) // This should never happen when col and row are non-negative.
	}
	return cell
}
