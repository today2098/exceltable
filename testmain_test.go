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
	"SID-999999",
}

var persons = []*person{
	{
		ID:            "ID-123456",
		Name:          "Alice",
		Age:           17,
		Address:       "",
		AccountNumber: "0000-0000-0000-0000",
		SpecialID:     &specialID[0],
	},
	{
		ID:            "ID-112358",
		Name:          "Bob",
		Age:           32,
		Address:       "Boston",
		AccountNumber: "1111-1111-1111-1111",
		SpecialID:     nil,
	},
	{
		ID:            "",
		Name:          "Carol",
		Age:           100,
		Address:       "京都",
		AccountNumber: "",
		SpecialID:     &specialID[2],
	},
}

type person struct {
	ID            string  `error:"zero"`
	Name          string  `csv:"name" excel:"氏名" newface:"isNewFace" error:"zero"`
	Age           int     `csv:"age" excel:"年齢" warn:"IsChild,IsOld"`
	Address       string  `csv:"address" excel:"住所" warn:"-"`
	AccountNumber string  `csv:"account_number" excel:"-"`
	SpecialID     *string `warn:"notZero" error:"nil"`
}

func (p *person) IsChild() bool { // pointer receiver.
	return p.Age < 18
}

func (p person) IsOld() bool { // value receiver.
	return 75 <= p.Age
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	RegisterRule(0, "newface", &excelize.Style{ // custom style tag.
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#aaffaa"},
		},
	})

	RegisterPredicate("isNewFace", func(name string) bool { // predicate function.
		newFaces := []string{"Alice"}
		return slices.Contains(newFaces, name)
	})
}
