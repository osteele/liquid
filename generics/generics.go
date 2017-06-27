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

// Apply applies a function to arguments, converting them as necessary.
// The conversion follows Liquid semantics, which are more aggressive than
// Go conversion. The function should return one or two values; the second value,
// if present, should be an error.
func Apply(fn reflect.Value, args []interface{}) (interface{}, error) {
	in := convertArguments(fn, args)
	outs := fn.Call(in)
	if len(outs) > 1 && outs[1].Interface() != nil {
		switch e := outs[1].Interface().(type) {
		case error:
			return nil, e
		default:
			panic(e)
		}
	}
	return outs[0].Interface(), nil
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

// Convert args to match the input types of function fn.
func convertArguments(fn reflect.Value, in []interface{}) []reflect.Value {
	rt := fn.Type()
	out := make([]reflect.Value, rt.NumIn())
	for i, arg := range in {
		if i < rt.NumIn() {
			out[i] = convertType(arg, rt.In(i))
		}
	}
	for i := len(in); i < rt.NumIn(); i++ {
		out[i] = reflect.Zero(rt.In(i))
	}
	return out
}
