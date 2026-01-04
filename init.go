package exceltable

import (
	"reflect"

	"github.com/xuri/excelize/v2"
)

const (
	warnTagName  string = "warn"
	errorTagName string = "error"
)

const (
	truePredKey    string = "true"
	falsePredKey   string = "false"
	zeroPredKey    string = "zero"
	notZeroPredKey string = "notZero"
	nilPredKey     string = "nil"
	notNilPredKey  string = "notNil"
)

func init() {
	RegisterDefaultRule(warnTagName, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffffaa"}, // light yellow
		},
	}, 98)
	RegisterDefaultRule(errorTagName, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffaaaa"}, // light red
		},
	}, 99)

	RegisterDefaultPredicate(truePredKey, func() bool { return true })
	RegisterDefaultPredicate(falsePredKey, func() bool { return false })
	RegisterDefaultPredicate(zeroPredKey, func(arg any) bool {
		return walkValue(reflect.ValueOf(arg)).IsZero()
	})
	RegisterDefaultPredicate(notZeroPredKey, func(arg any) bool {
		return !walkValue(reflect.ValueOf(arg)).IsZero()
	})
	RegisterDefaultPredicate(nilPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		return v.Kind() == reflect.Pointer && v.IsNil()
	})
	RegisterDefaultPredicate(notNilPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		return v.Kind() != reflect.Pointer || !v.IsNil()
	})
}
