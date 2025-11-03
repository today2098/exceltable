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
	// rule tag name.
	warnTag  ruleTagType = "warn"
	errorTag ruleTagType = "error"

	// rule predicate key.
	always        predKeyType = "always"
	zero          predKeyType = "zero"
	notZero       predKeyType = "notZero"
	nilPredKey    predKeyType = "nil"
	notNilPredKey predKeyType = "notNil"
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
	predicates sync.Map // pair of (rule name, function).
)

func init() {
	RegisterRule(errorTag, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffaaaa"},
		},
	}, -10001)
	RegisterRule(warnTag, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffffaa"},
		},
	}, -10002)

	RegisterPredicate(always, func() bool { return true })
	RegisterPredicate(zero, func(arg any) bool {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		}
		return v.IsZero()
	})
	RegisterPredicate(notZero, func(arg any) bool {
		v := reflect.ValueOf(arg)
		if v.Kind() == reflect.Pointer && !v.IsNil() {
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
		return v.Kind() == reflect.Pointer && !v.IsNil()
	})
}

func RegisterRule(tag ruleTagType, style *excelize.Style, priority int) {
	rules.Lock()
	defer rules.Unlock()

	rules.v = append(rules.v, &rule{priority, tag, style})
	sort.SliceStable(rules.v, func(i, j int) bool {
		return rules.v[i].priority < rules.v[j].priority // NOTE: Priority is in ascending order.
	})
}

func RegisterPredicate(key predKeyType, pred any) {
	predicates.Store(key, pred)
}

func CountByRule[M any](obj *M, tag string) (int, error) {
	t := reflect.TypeFor[M]()
	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()
	if v.Kind() != reflect.Struct {
		return 0, ErrNotStructType
	}

	res := 0
	for i := range t.NumField() {
		keys := strings.Split(t.Field(i).Tag.Get(tag), ",")
		b, err := verifyByPreds(ptrV, v, v.Field(i), keys)
		if err != nil {
			return 0, err
		}
		if b {
			res++
		}
	}

	return res, nil
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
		pred := ptrV.MethodByName(key)
		if !pred.IsValid() {
			pred = v.MethodByName(key)
			if !pred.IsValid() {
				tmp, ok := predicates.Load(key)
				if !ok {
					return false, ErrUnknownMethod
				}
				pred = reflect.ValueOf(tmp)
			}
		}

		b, err := callPredicate(pred, field)
		if err != nil {
			return false, err
		}

		return b, nil
	}
}

func callPredicate(pred, filed reflect.Value) (bool, error) {
	if !(pred.Type().NumOut() == 1 && pred.Type().Out(0).Kind() == reflect.Bool) {
		return false, ErrInvalidMethod
	}
	if pred.Type().NumIn() == 0 {
		return pred.Call([]reflect.Value{})[0].Bool(), nil
	}
	if pred.Type().NumIn() == 1 && filed.Type().AssignableTo(pred.Type().In(0)) {
		return pred.Call([]reflect.Value{filed})[0].Bool(), nil
	}
	return false, ErrInvalidMethod
}
