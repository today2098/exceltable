// Package exceltable is a simple wrapper around [excelize],
// providing utilities for writing Go structs to spreadsheet tables.
//
// It supports mapping between structs and spreadsheet tables,
// customizable column headers via struct tags,
// and conditional cell styling based on predicate functions.
//
// [excelize]: https://github.com/qax-os/excelize
package exceltable

import (
	"io"

	"github.com/xuri/excelize/v2"
)

// File wraps *excelize.File and holds style rules.
type File struct {
	*excelize.File
	rules      *ruleList
	predicates *predicateMap
	cache      *fieldCache
}

// Wrap wraps an existing *excelize.File into exceltable.File and returns its pointer:
//
//	file, _ := excelize.OpenFile("Book1.xlsx")
//	f, _ := exceltable.Wrap(file)
func Wrap(file *excelize.File) (*File, error) {
	rules := defaultRules.clone()
	for _, r := range rules.v {
		styleID, err := file.NewStyle(r.style)
		if err != nil {
			return nil, err
		}
		r.styleID = styleID
	}

	predicates := defaultPredicates.clone()
	cache := newFieldCache(rules, predicates)

	return &File{
		File:       file,
		rules:      rules,
		predicates: predicates,
		cache:      cache,
	}, nil
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
	file, err := excelize.OpenFile(filename, opts...)
	if err != nil {
		return nil, err
	}
	return Wrap(file)
}

// OpenReader read data stream from io.Reader and returns *exceltable.File wrapping it.
func OpenReader(r io.Reader, opts ...excelize.Options) (*File, error) {
	file, err := excelize.OpenReader(r, opts...)
	if err != nil {
		return nil, err
	}
	return Wrap(file)
}

// CountByRule counts the number of fields in obj that satisfy the predicates associated with the rule tag name.
func (f *File) CountByRule(obj any, tagName string) (int, error) {
	return countByRule(obj, tagName, f.cache)
}
