package main

import (
	"slices"

	"github.com/today2098/exceltable"
	"github.com/xuri/excelize/v2"
)

type Person struct {
	ID            string  `error:"zero"`
	Name          string  `csv:"name" excel:"氏名" newface:"isNewFace" error:"zero"`
	Age           int     `csv:"age" excel:"年齢" warn:"IsChild,IsOld"`
	Address       string  `csv:"address" excel:"住所"`
	AccountNumber string  `csv:"account_number" excel:"-"`
	SpecialID     *string `warn:"notZero" error:"nil"`
}

func (p *Person) IsChild() bool { // pointer receiver.
	return p.Age < 18
}

func (p Person) IsOld() bool { // value receiver.
	return 75 <= p.Age
}

func init() {
	exceltable.RegisterRule(0, "newface", &excelize.Style{ // custom style rule.
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#aaffaa"},
		},
	})

	exceltable.RegisterPredicate("isNewFace", func(name string) bool { // predicate function.
		newFaces := []string{"Alice"}
		return slices.Contains(newFaces, name)
	})
}

func main() {
	aliceSpecialID := ""
	alice := &Person{
		ID:            "ID-123456",
		Name:          "Alice",
		Age:           17,
		Address:       "",
		AccountNumber: "0000-0000-0000-0000",
		SpecialID:     &aliceSpecialID,
	}

	bob := &Person{
		ID:            "ID-112358",
		Name:          "Bob",
		Age:           32,
		Address:       "Boston",
		AccountNumber: "1111-1111-1111-1111",
		SpecialID:     nil,
	}

	carolSpecialID := "SID-999999"
	carol := &Person{
		ID:            "",
		Name:          "Carol",
		Age:           100,
		Address:       "京都",
		AccountNumber: "",
		SpecialID:     &carolSpecialID,
	}

	f, _ := exceltable.NewFile()
	s, _ := exceltable.NewSheetWithStreamWriter[Person](f, "NewSheet", "A1", true)

	s.SetHeader()

	s.SetRow(alice)
	s.SetRow(bob)
	s.SetRow(carol)

	s.AddDefaultTable()
	s.Flush()

	f.SaveAs("NewBook.xlsx")
}
