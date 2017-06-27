package expressions

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/osteele/liquid/errors"
)

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

func sortFilter(in []interface{}) []interface{} {
	a := make([]interface{}, len(in))
	for i, x := range in {
		a[i] = x
	}
	sort.Sort(sortable(a))
	return a
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
	DefineFilter("inspect", func(in interface{}) string {
		b, err := json.Marshal(in)
		if err != nil {
			panic(err)
		}
		return string(b)
	})
}

func DefineFilter(name string, fn interface{}) {
	rf := reflect.ValueOf(fn)
	if rf.Kind() != reflect.Func || rf.Type().NumIn() < 0 || rf.Type().NumOut() != 1 {
		panic(fmt.Errorf("a filter must be a function with at least one input and exactly one output"))
	}
	filters[name] = fn
}

type InterpreterError string

func (e InterpreterError) Error() string { return string(e) }

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
				case *genericError:
					panic(InterpreterError(e.Error()))
				default:
					panic(e)
				}
			}
		}()
		args := []interface{}{f(ctx)}
		if param != nil {
			args = append(args, param(ctx))
		}
		in := convertArguments(fr, args)
		r := fr.Call(in)[0]
		return r.Interface()
	}
}
