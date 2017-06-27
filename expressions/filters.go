package expressions

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
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
	DefineFilter("join", joinFilter)
	DefineFilter("sort", sortFilter)
	DefineFilter("split", splitFilter)
}

func DefineFilter(name string, fn interface{}) {
	rf := reflect.ValueOf(fn)
	if rf.Kind() != reflect.Func || rf.Type().NumIn() < 0 || rf.Type().NumOut() != 1 {
		panic(fmt.Errorf("a filter must be a function with at least one input and exactly one output"))
	}
	filters[name] = fn
}

func makeFilter(f valueFn, name string, param valueFn) valueFn {
	fn, ok := filters[name]
	if !ok {
		panic(fmt.Errorf("unknown filter: %s", name))
	}
	fr := reflect.ValueOf(fn)
	return func(ctx Context) interface{} {
		args := []interface{}{f(ctx)}
		if param != nil {
			args = append(args, param(ctx))
		}
		in := convertArguments(fr, args)
		r := fr.Call(in)[0]
		return r.Interface()
	}
}
