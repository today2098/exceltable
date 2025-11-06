package exceltable

import (
	"path/filepath"
	"testing"
)

func BenchmarkWrite(b *testing.B) {
	for b.Loop() {
		f, err := NewFile()
		if err != nil {
			b.Error(err)
		}

		s, err := NewSheet[person](f, "test", "A1", true)
		if err != nil {
			b.Error(err)
		}

		if err := s.SetHeader(); err != nil {
			b.Error(err)
		}

		for range 10000 {
			if err := s.SetRow(persons[0]); err != nil {
				b.Error(err)
			}
			if err := s.SetRow(persons[1]); err != nil {
				b.Error(err)
			}
			if err := s.SetRow(persons[2]); err != nil {
				b.Error(err)
			}
		}

		if err := s.AddDefaultTable(); err != nil {
			b.Error(err)
		}

		path := filepath.Join(b.TempDir(), "test.xlsx")
		if err := f.SaveAs(path); err != nil {
			b.Error(err)
		}
	}
}
