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

func NewFile(opts ...excelize.Options) (*File, error) {
	return Wrap(excelize.NewFile(opts...))
}

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

func (f *File) SaveAs(name string, opts ...excelize.Options) error {
	return f.File.SaveAs(name, opts...)
}
