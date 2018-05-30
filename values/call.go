package values

import (
	"fmt"
	"reflect"
)

// Call applies a function to arguments, converting them as necessary.
//
// The conversion follows Liquid (Ruby?) semantics, which are more aggressive than
// Go conversion.
//
// The function should return one or two values; the second value,
// if present, should be an error.
func Call(fn reflect.Value, args []interface{}) (interface{}, error) {
	in, err := convertCallArguments(fn, args)
	if err != nil {
		return nil, err
	}
	results := fn.Call(in)
	return convertCallResults(results)
}

// A CallParityError is a mismatch between the argument and parameter counts.
type CallParityError struct{ NumArgs, NumParams int }

func (e *CallParityError) Error() string {
	return fmt.Sprintf("wrong number of arguments (given %d, expected %d)", e.NumArgs, e.NumParams)
}

func convertCallResults(results []reflect.Value) (interface{}, error) {
	if len(results) > 1 && results[1].Interface() != nil {
		switch e := results[1].Interface().(type) {
		case error:
			return nil, e
		default:
			panic(e)
		}
	}
	return results[0].Interface(), nil
}

// Convert args to match the input types of function fn.
func convertCallArguments(fn reflect.Value, args []interface{}) (results []reflect.Value, err error) {
	rt := fn.Type()
	if len(args) > rt.NumIn() && !rt.IsVariadic() {
		return nil, &CallParityError{NumArgs: len(args), NumParams: rt.NumIn()}
	}
	if rt.IsVariadic() {
		numArgs, minArgs := len(args), rt.NumIn()-1
		if numArgs < minArgs {
			numArgs = minArgs
		}
		results = make([]reflect.Value, numArgs)
	} else {
		results = make([]reflect.Value, rt.NumIn())
	}
	for i, arg := range args {
		var typ reflect.Type
		if rt.IsVariadic() && i >= rt.NumIn()-1 {
			typ = rt.In(rt.NumIn() - 1).Elem()
		} else {
			typ = rt.In(i)
		}
		switch {
		case isDefaultFunctionType(typ):
			results[i] = makeConstantFunction(typ, arg)
		case arg == nil:
			results[i] = reflect.Zero(typ)
		default:
			results[i] = reflect.ValueOf(MustConvert(arg, typ))
		}
	}

	// create zeros and default functions for parameters without arguments
	for i := len(args); i < len(results); i++ {
		typ := rt.In(i)
		switch {
		case isDefaultFunctionType(typ):
			results[i] = makeIdentityFunction(typ)
		default:
			results[i] = reflect.Zero(typ)
		}
	}
	return
}

func isDefaultFunctionType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Func && typ.NumIn() == 1 && typ.NumOut() == 1
}

func makeConstantFunction(typ reflect.Type, arg interface{}) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(MustConvert(arg, typ.Out(0)))}
	})
}

func makeIdentityFunction(typ reflect.Type) reflect.Value {
	return reflect.MakeFunc(typ, func(args []reflect.Value) []reflect.Value {
		return args
	})
}
