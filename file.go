// Package exceltable is a simple wrapper around [excelize],
// providing utilities for writing Go structs to spreadsheet tables.
//
// It supports customizable column headers via struct tags
// and conditional cell styling based on predicate functions.
//
// [excelize]: https://github.com/qax-os/excelize
package exceltable

import (
	"io"
	"slices"

	"github.com/xuri/excelize/v2"
)

// fileRule represents relation between rule tag and style ID.
type fileRule struct {
	tag     ruleTagType
	styleID int
}

// File wraps excelize.File and holds style rules.
type File struct {
	*excelize.File
	rules []*fileRule // NOTE: Rules are stored in descending order of priority.
}

// NewFile creates a new exceltable.File and returns its pointer.
// It is equivalent to:
//
//	f, _ := exceltable.Wrap(excelize.NewFile())
func NewFile(opts ...excelize.Options) (*File, error) {
	return Wrap(excelize.NewFile(opts...))
}

// OpenFile opens an existing spreadsheet file and returns *exceltable.File wrapping it.
func OpenFile(filename string, opts ...excelize.Options) (*File, error) {
	f, err := excelize.OpenFile(filename, opts...)
	if err != nil {
		return nil, err
	}
	return Wrap(f)
}

// OpenReader read data stream from io.Reader and returns *exceltable.File wrapping it.
func OpenReader(r io.Reader, opts ...excelize.Options) (*File, error) {
	f, err := excelize.OpenReader(r, opts...)
	if err != nil {
		return nil, err
	}
	return Wrap(f)
}

// Wrap wraps an existing excelize.File into exceltable.File and returns its pointer:
//
//	file, _ := excelize.OpenFile("Book1.xlsx")
//	f, _ := exceltable.Wrap(file)
func Wrap(file *excelize.File) (*File, error) {
	rules, err := createFileRules(file)
	if err != nil {
		return nil, err
	}

	return &File{
		File:  file,
		rules: rules,
	}, nil
}

func createFileRules(file *excelize.File) ([]*fileRule, error) {
	rules.Lock()
	defer rules.Unlock()

	fileRules := make([]*fileRule, 0, len(rules.v))
	for _, r := range slices.Backward(rules.v) { // NOTE: Rules are sorted in ascending order of priority.
		styleID, err := file.NewStyle(r.style)
		if err != nil {
			return nil, err
		}
		fileRules = append(fileRules, &fileRule{r.tag, styleID})
	}

	return fileRules, nil
}
