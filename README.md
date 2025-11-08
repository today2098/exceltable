# exceltable

[![Go Report Card](https://goreportcard.com/badge/github.com/today2098/exceltable)](https://goreportcard.com/report/github.com/today2098/exceltable)
[![Go Reference](https://pkg.go.dev/badge/github.com/today2098/exceltable.svg)](https://pkg.go.dev/github.com/today2098/exceltable)

[README (日本語)](docs/README.ja.md)

A simple wrapper around [excelize](https://github.com/qax-os/excelize) (`github.com/xuri/excelize/v2`), providing utilities for writing Go structs to spreadsheet tables.

## Features

- Map Go structs to spreadsheet tables
- Customize column headers and visibility via struct tags (`excel`)
- Apply conditional cell styles based on predicate functions (e.g., highlight cells with background colors)

## Example

Output Example:

![spreadsheet_example](docs/image.png)

Code Example:

```go
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
```

## Installation

```bash
go get github.com/today2098/exceltable@latest
```

## Usage

### 1. Register Style Rules

First, register a style rule by specifying the tag name, `excelize.Style`, and priority.

Rules are evaluated in order of descending priority, and once a rule returns `true`, subsequent rules are not evaluated.

```go
exceltable.RegisterRule(0, "newface", &excelize.Style{
    Fill: excelize.Fill{
        Type:    "pattern",
        Pattern: 1,
        Color:   []string{"#aaffaa"},
    },
})
```

Default rules:

|Tag|Style|Priority|
|---|---|---|
|`warn`|Yellow background (`#ffffaa`)|98|
|`error`|Red background (`#ffaaaa`)|99|

### 2. Register Style Predicates

Next, register predicates that define the conditions under which a style is applied.

A predicate can be a method on the struct or a standalone function.

Predicate functions must either take no arguments or a single argument corresponding to the field type.

```go
exceltable.RegisterPredicate("isNewFace", func(name string) bool {
    newFaces := []string{"Alice"}
    return slices.Contains(newFaces, name)
})
```

Default predicates:

|Name|Description|
|---|---|
|`always`|Always true|
|`never`|Always false|
|`zero`|Field value is zero value|
|`notZero`|Field value is non-zero value|
|`nil`|Pointer field is nil|
|`notNil`|Pointer field is not nil|

### 3. Add Struct Tags

Add tags to struct fields to specify column headers and style application conditions.

Header names are resolved in the following order: `excel` > `csv` > field name.
To hide a field, use `excel:"-"`.

Multiple predicates can be specified as a comma-separated list (OR condition).

```go
type Person struct {
    ID            string  `error:"zero"`
    Name          string  `csv:"name" excel:"氏名" newface:"isNewFace" error:"zero"`
    Age           int     `csv:"age" excel:"年齢" warn:"IsChild,IsOld"`
    Address       string  `csv:"address" excel:"住所"`
    AccountNumber string  `csv:"account_number" excel:"-"`
    SpecialID     *string `warn:"notZero" error:"nil"`
}

func (p *Person) IsChild() bool {
    return p.Age < 18
}

func (p Person) IsOld() bool {
    return 75 <= p.Age
}
```

### 4. Write to a Spreadsheet

```go
f, _ := exceltable.NewFile()
s, _ := exceltable.NewSheetWithStreamWriter[Person](f, "NewSheet", "A1", true)

s.SetHeader()

s.SetRow(alice)
s.SetRow(bob)
s.SetRow(carol)

s.AddDefaultTable()
s.Flush()

f.SaveAs("NewBook.xlsx")
```

## License

This project is licensed under the MIT License.

It depends on [excelize](https://github.com/qax-os/excelize), which is licensed under the BSD 3-Clause License.
