package exceltable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
