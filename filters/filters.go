// Package filters defines the standard Liquid filters.
package filters

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/leekchan/timeutil"
	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/generics"
)

// DefineStandardFilters defines the standard Liquid filters.
func DefineStandardFilters() {
	// values
	expressions.DefineFilter("default", func(in, defaultValue interface{}) interface{} {
		if in == nil || in == false || generics.IsEmpty(in) {
			in = defaultValue
		}
		return in
	})

	// dates
	expressions.DefineFilter("date", func(in, iformat interface{}) interface{} {
		format, ok := iformat.(string)
		if !ok {
			format = "%a, %b %d, %y"
		}
		switch date := in.(type) {
		case string:
			d, err := generics.ParseTime(date)
			if err != nil {
				panic(err)
			}
			return timeutil.Strftime(&d, format)
		case time.Time:
			return timeutil.Strftime(&date, format)
		default:
			panic(expressions.UnimplementedError(fmt.Sprintf("date conversion from %v", date)))
		}
	})

	// lists
	expressions.DefineFilter("join", joinFilter)
	expressions.DefineFilter("sort", sortFilter)

	// strings
	expressions.DefineFilter("split", splitFilter)

	// Jekyll
	expressions.DefineFilter("inspect", json.Marshal)
}

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

func sortFilter(in []interface{}, key interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	if key == nil {
		generics.Sort(out)
	} else {
		generics.SortByProperty(out, key.(string))
	}
	return out
}

func splitFilter(in, sep string) interface{} {
	return strings.Split(in, sep)
}
