package exceltable

import (
	"path/filepath"
	"testing"
)

func BenchmarkWriteWithStreamWriter(b *testing.B) {
	for b.Loop() {
		f, err := NewFile()
		if err != nil {
			b.Error(err)
		}

		ssw, err := NewSheetWithStreamWriter[person](f, "test", "A1", true)
		if err != nil {
			b.Error(err)
		}

		if err := ssw.SetHeader(); err != nil {
			b.Error(err)
		}

		for range 10000 {
			if err := ssw.SetRow(persons[0]); err != nil {
				b.Error(err)
			}
			if err := ssw.SetRow(persons[1]); err != nil {
				b.Error(err)
			}
			if err := ssw.SetRow(persons[2]); err != nil {
				b.Error(err)
			}
		}

		if err := ssw.AddDefaultTable(); err != nil {
			b.Error(err)
		}

		if err := ssw.Flush(); err != nil {
			b.Error(err)
		}

		path := filepath.Join(b.TempDir(), "test.xlsx")
		if err := f.SaveAs(path); err != nil {
			b.Error(err)
		}
	}
}
