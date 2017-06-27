package generics

import (
	"fmt"
	"reflect"
)

// GenericError is an error regarding generic conversion.
type GenericError string

func (e GenericError) Error() string { return string(e) }

func genericErrorf(format string, a ...interface{}) error {
	return GenericError(fmt.Sprintf(format, a...))
}

// Convert val to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func convertType(val interface{}, t reflect.Type) reflect.Value {
	r := reflect.ValueOf(val)
	if r.Type().ConvertibleTo(t) {
		return r.Convert(t)
	}
	if reflect.PtrTo(r.Type()) == t {
		return reflect.ValueOf(&val)
	}
	switch t.Kind() {
	case reflect.Slice:
		if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
			break
		}
		x := reflect.MakeSlice(t, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			c := convertType(r.Index(i).Interface(), t.Elem())
			x = reflect.Append(x, c)
		}
		return x
	}
	panic(genericErrorf("convertType: can't convert %#v<%s> to %v", val, r.Type(), t))
}

// IsEmpty returns a bool indicating whether the value is empty according to Liquid semantics.
func IsEmpty(in interface{}) bool {
	if in == nil {
		return false
	}
	r := reflect.ValueOf(in)
	switch r.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return r.Len() == 0
	case reflect.Bool:
		return r.Bool() == false
	default:
		return false
	}
}
