package exceltable

import "reflect"

type Stringer interface {
	String() string
}

func assignableToStringer(typ reflect.Type) bool {
	return typ.AssignableTo(reflect.TypeFor[Stringer]())
}

func walkType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func walkValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Pointer && !val.IsNil() {
		val = val.Elem()
	}
	return val
}
