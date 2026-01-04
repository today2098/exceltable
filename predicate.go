package exceltable

import (
	"maps"
	"reflect"
	"sync"
)

var defaultPredicates = newPredicateMap()

type predicate struct {
	val      reflect.Value
	typ      reflect.Type
	isMethod bool
}

func newPredicate(val reflect.Value, isMethod bool) (*predicate, error) {
	typ := val.Type()
	if typ.Kind() != reflect.Func {
		return nil, ErrNotFuncType
	}

	if isMethod {
		n, m := typ.NumIn(), typ.NumOut()
		in, out := make([]reflect.Type, n), make([]reflect.Type, m)
		for i := range n {
			in[i] = typ.In(i)
		}
		for i := range m {
			out[i] = typ.Out(i)
		}
		typ = reflect.FuncOf(in[1:], out, false)
	}

	if !validatePredicateType(typ) {
		return nil, ErrInvalidPredicate
	}

	return &predicate{
		val:      val,
		typ:      typ,
		isMethod: isMethod,
	}, nil
}

func (p *predicate) isAcceptableType(argType reflect.Type) bool {
	if p.typ.NumIn() == 0 {
		return true
	}
	return argType.AssignableTo(p.typ.In(0))
}

func (p *predicate) bind(ptrV reflect.Value) reflect.Value {
	if !p.isMethod {
		return p.val
	}

	return reflect.MakeFunc(p.typ, func(in []reflect.Value) []reflect.Value {
		return p.val.Call(append([]reflect.Value{ptrV}, in...))
	})
}

func (p *predicate) callWithReceiver(rx, arg reflect.Value) bool {
	// NOTE: pred must be a nullary or unary function taking arg's type and returning a bool.
	pred := p.bind(rx)

	if p.typ.NumIn() == 0 {
		return pred.Call([]reflect.Value{})[0].Bool() // nullary predicate
	}

	return pred.Call([]reflect.Value{arg})[0].Bool() // unary predicate
}

func (p *predicate) copy(pred *predicate) {
	p.val = pred.val
	p.typ = pred.typ
	p.isMethod = pred.isMethod
}

func (p *predicate) clone() *predicate {
	return &predicate{
		val:      p.val,
		typ:      p.typ,
		isMethod: p.isMethod,
	}
}

func validatePredicateType(typ reflect.Type) bool {
	if typ.NumIn() >= 2 {
		return false
	}

	if typ.NumOut() != 1 || typ.Out(0).Kind() != reflect.Bool {
		return false
	}

	return true
}

type predicateList []*predicate

func (pl *predicateList) append(pred *predicate) {
	*pl = append(*pl, pred)
}

func (pl *predicateList) callWithReceiver(rx, arg reflect.Value) bool {
	for _, pred := range *pl {
		if pred.callWithReceiver(rx, arg) {
			return true
		}
	}
	return false
}

type predicateMap struct {
	sync.Mutex
	v map[string]*predicate
}

func newPredicateMap() *predicateMap {
	return &predicateMap{
		v: make(map[string]*predicate),
	}
}

func (pm *predicateMap) store(key string, pred *predicate) {
	pm.Lock()
	defer pm.Unlock()

	if oldPred, ok := pm.v[key]; ok {
		// NOTE: Reuse the existing pointer to maintain consistency with other references.
		oldPred.copy(pred)
		return
	}

	pm.v[key] = pred
}

func (pm *predicateMap) load(key string) (pred *predicate, ok bool) {
	pm.Lock()
	defer pm.Unlock()

	pred, ok = pm.v[key]
	return
}

func (pm *predicateMap) clone() *predicateMap {
	pm.Lock()
	defer pm.Unlock()

	mp := maps.Clone(pm.v)
	for _, pred := range mp {
		pred = pred.clone()
	}

	return &predicateMap{
		v: mp,
	}
}

// RegisterDefaultPredicate registers a new default predicate function with key:
//
//	exceltable.RegisterDefaultPredicate("isAlice", func(name string) bool {
//		return name == "Alice"
//	})
func RegisterDefaultPredicate(key string, predFunc any) error {
	pred, err := newPredicate(reflect.ValueOf(predFunc), false)
	if err != nil {
		return err
	}

	defaultPredicates.store(key, pred)
	return nil
}
