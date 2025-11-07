// Package exceltable provides helper functions for writing Go structs to Excel tables.
//
// It wraps [excelize] to offer a lightweight interface for table generation,
// including header customization via struct tags and predicate-based conditional cell styling.
//
// [excelize]: https://github.com/qax-os/excelize
package exceltable

import (
	"slices"

	"github.com/xuri/excelize/v2"
)

type fileRule struct {
	tag     ruleTagType
	styleID int
}

// File wraps excelize.File and saves pairs of rule tag and style ID.
type File struct {
	File  *excelize.File
	rules []*fileRule
}

// NewFile creates a new exceltable.File:
//
//	f, _ := exceltable.NewFile()
//
// It is equivalent to:
//
//	f, _ := exceltable.Wrap(excelize.NewFile())
func NewFile(opts ...excelize.Options) (*File, error) {
	return Wrap(excelize.NewFile(opts...))
}

// Wrap wraps an existing excelize.File into exceltable.File:
//
//	file, _ := excelize.OpenFile("Book1.xlsx")
//	f, _ := exceltable.Wrap(file)
func Wrap(file *excelize.File) (*File, error) {
	f := &File{
		File:  file,
		rules: make([]*fileRule, 0, len(rules.v)),
	}

	if err := f.registeRuleTags(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *File) registeRuleTags() error {
	rules.Lock()
	defer rules.Unlock()

	for _, r := range slices.Backward(rules.v) { // NOTE: Rules are sorted in ascending order of priority.
		styleID, err := f.File.NewStyle(r.style)
		if err != nil {
			return err
		}
		f.rules = append(f.rules, &fileRule{r.tag, styleID})
	}

	return nil
}

// SaveAs saves contents to the Excel file specified by name.
// It is equivalent to excelize.File.SaveAs:
//
//	err := f.SaveAs("Book1.xlsx")
func (f *File) SaveAs(name string, opts ...excelize.Options) error {
	return f.File.SaveAs(name, opts...)
}
