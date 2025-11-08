package exceltable

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sheetRule_bind(t *testing.T) {
	ptrT := reflect.TypeFor[*person]()
	ptrV := reflect.ValueOf(persons[0]) // alice

	{
		method, ok := ptrT.MethodByName("IsChild")
		require.True(t, ok)

		sr := newSheetRule(method.Func, true, 0)
		b, err := callPredicate(sr.bind(ptrV), reflect.Value{})
		require.NoError(t, err)
		assert.True(t, b)
	}

	{
		function, ok := predicates.Load("isNewFace")
		require.True(t, ok)

		sr := newSheetRule(reflect.ValueOf(function), false, 0)
		b, err := callPredicate(sr.bind(ptrV), ptrV.Elem().FieldByName("Name"))
		require.NoError(t, err)
		assert.True(t, b)
	}
}

func Test_newSheetBase(t *testing.T) {
	f, err := NewFile()
	require.NoError(t, err)

	{
		_, err := newSheetBase[int](f, "test", "A1", true)
		require.Error(t, err)
		assert.Equal(t, ErrNotStructType, err)
	}

	{
		_, err := newSheetBase[person](f, "test", "", true)
		assert.Error(t, err)
	}

	{
		sb, err := newSheetBase[person](f, "test", "B2", true)
		require.NoError(t, err)
		assert.Equal(t, sb.name, "test")
		assert.Equal(t, sb.x, 2)
		assert.Equal(t, sb.y, 2)
		assert.Equal(t, sb.row, 1)
		assert.Equal(t, sb.tableWidth, 5)
		assert.Equal(t, sb.numField, 7)

		wantSkip := []bool{false, true, false, false, false, true, false}
		if diff := cmp.Diff(wantSkip, sb.skip); diff != "" {
			t.Errorf("newSheetBase[person](...) mismatch (-want +got):\n%s", diff)
		}

		wantHeader := []any{"ID", "氏名", "年齢", "住所", "SpecialID"}
		if diff := cmp.Diff(wantHeader, sb.header); diff != "" {
			t.Errorf("newSheetBase[person](...) mismatch (-want +got):\n%s", diff)
		}

		// TODO: verify rulesList.
	}
}

func Test_sheetBase_newTable(t *testing.T) {
	f, err := NewFile()
	require.NoError(t, err)
	sb, err := newSheetBase[person](f, "test", "C3", true)
	require.NoError(t, err)

	table := sb.newTable(DefaultTableStyle)
	assert.Equal(t, table.Range, "C3:G4")
	assert.Equal(t, table.Name, "testTable")
	assert.Equal(t, table.StyleName, DefaultTableStyle)
}

func Test_sheetBase_coordinatesToCellName(t *testing.T) {
	f, err := NewFile()
	require.NoError(t, err)
	sb, err := newSheetBase[person](f, "test", "D4", true)
	require.NoError(t, err)

	cell := sb.coordinatesToCellName(0, 0)
	assert.Equal(t, cell, "D4")

	cell = sb.coordinatesToCellName(2, 2)
	assert.Equal(t, cell, "F6")
}
