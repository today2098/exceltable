package exceltable

import (
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

type ruleTagType = string
type predKeyType = string

const (
	// Rule tag name.
	warnTag  ruleTagType = "warn"
	errorTag ruleTagType = "error"

	// Rule predicate key name.
	alwaysPredKey  predKeyType = "always"
	neverPredKey   predKeyType = "never"
	zeroPredKey    predKeyType = "zero"
	notZeroPredKey predKeyType = "notZero"
	nilPredKey     predKeyType = "nil"
	notNilPredKey  predKeyType = "notNil"
)

type rule struct {
	priority int
	tag      ruleTagType
	style    *excelize.Style
}

var (
	rules = struct {
		sync.Mutex
		v []*rule
	}{
		v: make([]*rule, 0),
	}

	predicates sync.Map // pair of (key, function).
)

func init() {
	RegisterRule(98, warnTag, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffffaa"},
		},
	})
	RegisterRule(99, errorTag, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffaaaa"},
		},
	})

	RegisterPredicate(alwaysPredKey, func() bool { return true })
	RegisterPredicate(neverPredKey, func() bool { return false })
	RegisterPredicate(zeroPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		for v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		}
		return v.IsZero()
	})
	RegisterPredicate(notZeroPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		for v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		}
		return !v.IsZero()
	})
	RegisterPredicate(nilPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		return v.Kind() == reflect.Pointer && v.IsNil()
	})
	RegisterPredicate(notNilPredKey, func(arg any) bool {
		v := reflect.ValueOf(arg)
		return v.Kind() != reflect.Pointer || !v.IsNil()
	})
}

func RegisterRule(priority int, tag ruleTagType, style *excelize.Style) {
	rules.Lock()
	defer rules.Unlock()

	rules.v = append(rules.v, &rule{priority, tag, style})
	sort.SliceStable(rules.v, func(i, j int) bool {
		return rules.v[i].priority < rules.v[j].priority // NOTE: Rules are sorted in ascending order of priority.
	})
}

func RegisterPredicate(key predKeyType, pred any) {
	predicates.Store(key, pred)
}

func CountByRule[M any](obj *M, tag string) (int, error) {
	t := reflect.TypeFor[M]()
	if t.Kind() != reflect.Struct {
		return 0, ErrNotStructType
	}

	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	numField, cnt := t.NumField(), 0
	for i := range numField {
		keys := strings.Split(t.Field(i).Tag.Get(tag), ",")
		b, err := verifyByPreds(ptrV, v, v.Field(i), keys)
		if err != nil {
			return 0, err
		}
		if b {
			cnt++
		}
	}

	return cnt, nil
}

func verifyByPreds(ptrV, v, field reflect.Value, keys []predKeyType) (bool, error) {
	for _, key := range keys {
		b, err := verifyByPred(ptrV, v, field, key)
		if err != nil {
			return false, err
		}
		if b {
			return true, err
		}
	}

	return false, nil
}

func verifyByPred(ptrV, v, field reflect.Value, key predKeyType) (bool, error) {
	switch key {
	case "", "-":
		return false, nil
	default:
		if pred := ptrV.MethodByName(key); pred.IsValid() {
			return callPredicate(pred, field)
		}

		if pred := v.MethodByName(key); pred.IsValid() {
			return callPredicate(pred, field)
		}

		if pred, ok := predicates.Load(key); ok {
			return callPredicate(reflect.ValueOf(pred), field)
		}
	}

	return false, ErrUnknownMethod
}

func callPredicate(pred, arg reflect.Value) (bool, error) {
	if !(pred.Type().NumOut() == 1 && pred.Type().Out(0).Kind() == reflect.Bool) {
		return false, ErrInvalidMethod
	}

	if pred.Type().NumIn() == 0 {
		return pred.Call([]reflect.Value{})[0].Bool(), nil // nulary predicate
	}

	if pred.Type().NumIn() == 1 && arg.Type().AssignableTo(pred.Type().In(0)) {
		return pred.Call([]reflect.Value{arg})[0].Bool(), nil // unary predicate
	}

	return false, ErrInvalidMethod
}
