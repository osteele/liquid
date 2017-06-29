package generics

import (
	"fmt"
	"reflect"
)

// Call applies a function to arguments, converting them as necessary.
// The conversion follows Liquid semantics, which are more aggressive than
// Go conversion. The function should return one or two values; the second value,
// if present, should be an error.
func Call(fn reflect.Value, args []interface{}) (interface{}, error) {
	in := convertArguments(fn, args)
	outs := fn.Call(in)
	if len(outs) > 1 && outs[1].Interface() != nil {
		switch e := outs[1].Interface().(type) {
		case error:
			fmt.Println("error")
			return nil, e
		default:
			panic(e)
		}
	}
	return outs[0].Interface(), nil
}

// Convert args to match the input types of function fn.
func convertArguments(fn reflect.Value, in []interface{}) []reflect.Value {
	rt := fn.Type()
	out := make([]reflect.Value, rt.NumIn())
	for i, arg := range in {
		if i < rt.NumIn() {
			if arg == nil {
				out[i] = reflect.Zero(rt.In(i))
			} else {
				out[i] = reflect.ValueOf(MustConvert(arg, rt.In(i)))
			}
		}
	}
	for i := len(in); i < rt.NumIn(); i++ {
		out[i] = reflect.Zero(rt.In(i))
	}
	return out
}
