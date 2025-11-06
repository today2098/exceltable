package exceltable

import (
	"path/filepath"
	"testing"
)

func BenchmarkWriteBlank(b *testing.B) {
	for b.Loop() {
		f, err := NewFile()
		if err != nil {
			b.Error(err)
		}

		path := filepath.Join(b.TempDir(), "tmp.xlsx")
		if err := f.SaveAs(path); err != nil {
			b.Error(err)
		}
	}
}
