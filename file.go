package exceltable

import (
	"slices"

	"github.com/xuri/excelize/v2"
)

type fileRule struct {
	tag     ruleTagType
	styleID int
}

type File struct {
	File  *excelize.File
	rules []*fileRule
}

func NewFile() (f *File, err error) {
	f = &File{
		File:  excelize.NewFile(),
		rules: make([]*fileRule, 0, len(rules.v)),
	}

	if err := f.registerRuleTags(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *File) SaveAs(name string) error {
	return f.File.SaveAs(name)
}

func (f *File) Close() error {
	return f.File.Close()
}

func (f *File) registerRuleTags() error {
	rules.Lock()
	defer rules.Unlock()

	for _, r := range slices.Backward(rules.v) { // NOTE: Priority is in ascending order.
		styleID, err := f.File.NewStyle(r.style)
		if err != nil {
			return err
		}
		f.rules = append(f.rules, &fileRule{r.tag, styleID})
	}
	return nil
}
