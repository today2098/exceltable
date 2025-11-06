package exceltable

import (
	"os"
	"slices"
	"testing"

	"github.com/xuri/excelize/v2"
)

var specialID = []string{
	"",
	"",
	"abcdef",
}

var persons = []*person{
	{
		ID:            "123456",
		Name:          "Alice",
		Year:          17,
		Address:       "",
		AccountNumber: "0000-0000-0000-0000",
		SpecialID:     &specialID[0],
	},
	{
		ID:            "112358",
		Name:          "Bob",
		Year:          32,
		Address:       "Boston",
		AccountNumber: "1111-1111-1111-1111",
		SpecialID:     nil,
	},
	{
		ID:            "",
		Name:          "Carol",
		Year:          100,
		Address:       "Kyoto",
		AccountNumber: "",
		SpecialID:     &specialID[2],
	},
}

type person struct {
	ID            string  `error:"zero"`
	Name          string  `csv:"name" excel:"氏名" newface:"isNewFace" error:"zero"`
	Year          int     `csv:"year" excel:"年齢" warn:"IsChild,IsOld"`
	Address       string  `csv:"address" excel:"住所" warn:"-"`
	AccountNumber string  `csv:"account_number" excel:"-"`
	SpecialID     *string `warn:"notZero" error:"nil"`
}

func (p *person) IsChild() bool {
	return !p.IsAdult()
}

func (p *person) IsAdult() bool {
	return p.Year >= 18
}

func (p *person) IsOld() bool {
	return p.Year >= 75
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	RegisterRule(0, "newface", &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#aaffaa"},
		},
	})
	RegisterPredicate("isNewFace", func(name string) bool {
		var newFaces = []string{"Alice"}
		return slices.Contains(newFaces, name)
	})
}
