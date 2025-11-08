package exceltable

import (
	"os"
	"slices"
	"testing"

	"github.com/xuri/excelize/v2"
)

var specialIDs = []string{
	"",
	"",
	"SID-999999",
}

var persons = []*person{
	{
		ID:            "ID-123456",           //
		Name:          "Alice",               // newface
		Age:           17,                    // warn
		Address:       "",                    //
		AccountNumber: "0000-0000-0000-0000", // omitted
		SpecialID:     &specialIDs[0],        // warn
	},
	{
		ID:            "ID-112358",           //
		Name:          "Bob",                 //
		Age:           32,                    //
		Address:       "Boston",              //
		AccountNumber: "1111-1111-1111-1111", // omitted
		SpecialID:     nil,                   // error
	},
	{
		ID:            "",             // error
		Name:          "Carol",        //
		Age:           100,            // warn
		Address:       "京都",           //
		AccountNumber: "",             // omitted
		SpecialID:     &specialIDs[2], // warn
	},
}

type person struct {
	ID              string  `error:"zero"`
	unexportedValue int     //lint:ignore U1000 Ignore unused for test
	Name            string  `csv:"name" excel:"氏名" newface:"isNewFace" error:"zero"`
	Age             int     `csv:"age" excel:"年齢" warn:"IsChild,IsOld"`
	Address         string  `csv:"address" excel:"住所" warn:"-"`
	AccountNumber   string  `csv:"account_number" excel:"-"`
	SpecialID       *string `warn:"notZero" error:"nil"`
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
	RegisterRule(0, "newface", &excelize.Style{ // custom style rule.
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
