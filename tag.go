package exceltable

import (
	"reflect"
	"strings"
)

const (
	csvTagName   string = "csv"
	excelTagName string = "excel"
)

type tag struct {
	columnName string
	prefix     string
	empty      bool
	ignore     bool
	omitEmpty  bool
	omitZero   bool
	specifyNil bool
	inline     bool
}

func parseTag(sf reflect.StructField) *tag {
	tag := &tag{}

	tagStr := sf.Tag.Get(excelTagName)
	if tagStr == "" {
		tagStr = sf.Tag.Get(csvTagName)
	}

	if tagStr == "" {
		tag.columnName = sf.Name
		tag.empty = true
		return tag
	}

	elems := strings.Split(tagStr, ",")
	switch elems[0] {
	case "-":
		tag.ignore = true
		return tag
	case "":
		tag.columnName = sf.Name
	default:
		tag.columnName = elems[0]
	}

	for _, opt := range elems[1:] {
		switch opt {
		case "omitempty":
			tag.omitEmpty = true
		case "omitzero":
			tag.omitZero = true
		case "specifynil":
			tag.specifyNil = true
		case "inline":
			if walkType(sf.Type).Kind() == reflect.Struct {
				tag.inline = true
				tag.prefix = elems[0]
			}
		}
	}

	return tag
}
