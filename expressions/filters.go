package expressions

import (
	"fmt"
	"reflect"

	"github.com/osteele/liquid/errors"
	"github.com/osteele/liquid/generics"
)

type InterpreterError string

func (e InterpreterError) Error() string { return string(e) }

type valueFn func(Context) interface{}

var filters = map[string]interface{}{}

func DefineFilter(name string, fn interface{}) {
	rf := reflect.ValueOf(fn)
	switch {
	case rf.Kind() != reflect.Func:
		panic(fmt.Errorf("a filter must be a function"))
	case rf.Type().NumIn() < 1:
		panic(fmt.Errorf("a filter function must have at least one input"))
	case rf.Type().NumOut() > 2:
		panic(fmt.Errorf("a filter must be have one or two outputs"))
		// case rf.Type().Out(1).Implements(…):
		// 	panic(fmt.Errorf("a filter's second output must be type error"))
	}
	filters[name] = fn
}

func makeFilter(f valueFn, name string, param valueFn) valueFn {
	fn, ok := filters[name]
	if !ok {
		panic(errors.UndefinedFilter(name))
	}
	fr := reflect.ValueOf(fn)
	return func(ctx Context) interface{} {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case generics.GenericError:
					panic(InterpreterError(e.Error()))
				default:
					// fmt.Println(string(debug.Stack()))
					panic(e)
				}
			}
		}()
		args := []interface{}{f(ctx)}
		if param != nil {
			args = append(args, param(ctx))
		}
		out, err := generics.Apply(fr, args)
		if err != nil {
			panic(err)
		}
		switch out := out.(type) {
		case []byte:
			return string(out)
		default:
			return out
		}
	}
}
