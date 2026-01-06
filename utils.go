package exceltable

import "reflect"

type Stringer interface {
	String() string
}

func assignableToStringer(typ reflect.Type) bool {
	return typ.AssignableTo(reflect.TypeFor[Stringer]())
}

func isNilable(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}

func walkType(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	return typ
}

func walkPtr(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return val
		}

		val = val.Elem()
	}

	return val
}

func walkValue(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return val
		}

		val = val.Elem()
	}

	return val
}

func walkValueToStringer(val reflect.Value) reflect.Value {
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if assignableToStringer(val.Type()) {
			return val
		}

		if val.IsNil() {
			return val
		}

		val = val.Elem()
	}

	return val
}
