package expressions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/osteele/liquid/errors"
)

type InterpreterError string

func (e InterpreterError) Error() string { return string(e) }

type valueFn func(Context) interface{}

func joinFilter(in []interface{}, sep interface{}) interface{} {
	a := make([]string, len(in))
	s := ", "
	if sep != nil {
		s = fmt.Sprint(sep)
	}
	for i, x := range in {
		a[i] = fmt.Sprint(x)
	}
	return strings.Join(a, s)
}

// func sortFilter(in []interface{}, key interface{}) []interface{} {
// 	fmt.Println("sort", in, key)
func sortFilter(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	sort.Sort(genericSortable(out))
	return out
}

func splitFilter(in, sep string) interface{} {
	return strings.Split(in, sep)
}

var filters = map[string]interface{}{}

func init() {
	DefineStandardFilters()
}

func DefineStandardFilters() {
	// lists
	DefineFilter("join", joinFilter)
	DefineFilter("sort", sortFilter)

	// strings
	DefineFilter("split", splitFilter)

	// Jekyll
	DefineFilter("inspect", json.Marshal)
}

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
				case genericError:
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
		in := convertArguments(fr, args)
		outs := fr.Call(in)
		if len(outs) > 1 && outs[1].Interface() != nil {
			err := outs[1].Interface()
			panic(err)
		}
		switch out := outs[0].Interface().(type) {
		case []byte:
			return string(out)
		default:
			return out
		}
	}
}
