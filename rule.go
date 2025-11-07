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

// Rule tags and predicate keys.
const (
	warnTag  ruleTagType = "warn"
	errorTag ruleTagType = "error"

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

// Global variables for rules and predicates.
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
			Color:   []string{"#ffffaa"}, // light yellow
		},
	})
	RegisterRule(99, errorTag, &excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#ffaaaa"}, // light red
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

// RegisterRule registers a new rule with the given priority, tag, and style.
// Rules with higher priority values are applied earlier:
//
//	exceltable.RegisterRule(0, "customTag", &excelize.Style{ ... })
func RegisterRule(priority int, tag ruleTagType, style *excelize.Style) {
	rules.Lock()
	defer rules.Unlock()

	rules.v = append(rules.v, &rule{priority, tag, style})
	sort.SliceStable(rules.v, func(i, j int) bool {
		return rules.v[i].priority < rules.v[j].priority // NOTE: Rules are sorted in ascending order of priority.
	})
}

// RegisterPredicate registers a new predicate function with the given key:
//
//	exceltable.RegisterPredicate("isAlice", func(name string) bool { return name == "Alice" })
func RegisterPredicate(key predKeyType, pred any) {
	predicates.Store(key, pred)
}

// CountByRule counts the number of fields in obj that satisfy the predicate associated with the given rule tag.
func CountByRule[M any](obj *M, tag string) (int, error) {
	t := reflect.TypeFor[M]()
	if t.Kind() != reflect.Struct {
		return 0, ErrNotStructType
	}

	ptrV := reflect.ValueOf(obj)
	v := ptrV.Elem()

	numField, cnt := t.NumField(), 0
	for i := range numField {
		field := v.Field(i)
		for key := range strings.SplitSeq(t.Field(i).Tag.Get(tag), ",") {
			b, err := verifyByPred(ptrV, field, key)
			if err != nil {
				return 0, err
			}
			if b {
				cnt++
				break
			}
		}
	}

	return cnt, nil
}

func verifyByPred(ptrV, field reflect.Value, key predKeyType) (bool, error) {
	switch key {
	case "", "-":
		return false, nil
	default:
		if pred := ptrV.MethodByName(key); pred.IsValid() {
			return callPredicate(pred, field)
		}

		if pred, ok := predicates.Load(key); ok {
			return callPredicate(reflect.ValueOf(pred), field)
		}
	}

	return false, ErrUnknownPredicate
}

func callPredicate(pred, arg reflect.Value) (bool, error) {
	if !(pred.Type().NumOut() == 1 && pred.Type().Out(0).Kind() == reflect.Bool) {
		return false, ErrInvalidPredicate
	}

	if pred.Type().NumIn() == 0 {
		return pred.Call([]reflect.Value{})[0].Bool(), nil // nulary predicate
	}

	if pred.Type().NumIn() == 1 && arg.Type().AssignableTo(pred.Type().In(0)) {
		return pred.Call([]reflect.Value{arg})[0].Bool(), nil // unary predicate
	}

	return false, ErrInvalidPredicate
}
