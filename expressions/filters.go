package expressions

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type valueFn func(Context) interface{}

func joinFilter(in []interface{}) interface{} {
	a := make([]string, len(in))
	for i, x := range in {
		a[i] = fmt.Sprint(x)
	}
	return strings.Join(a, ", ")
}

func sortFilter(in []interface{}) []interface{} {
	a := make([]interface{}, len(in))
	for i, x := range in {
		a[i] = x
	}
	sort.Sort(sortable(a))
	return a
}

var filters = map[string]interface{}{
	"join": joinFilter,
	"sort": sortFilter,
}

func makeFilter(f valueFn, name string) valueFn {
	fn, ok := filters[name]
	if !ok {
		panic(fmt.Errorf("unknown filter: %s", name))
	}
	fr := reflect.ValueOf(fn)
	return func(ctx Context) interface{} {
		args := []interface{}{f(ctx)}
		in := convertArguments(fr, args)
		r := fr.Call(in)[0]
		return r.Interface()
	}
}
