package filters

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/osteele/liquid/values"
)

func sortFilter(array []any, key any) []any {
	result := make([]any, len(array))
	copy(result, array)
	if key == nil {
		values.Sort(result)
	} else {
		values.SortByProperty(result, fmt.Sprint(key), true)
	}
	return result
}

func sortNaturalFilter(array []any, key any) any {
	result := make([]any, len(array))
	copy(result, array)
	switch {
	case reflect.ValueOf(array).Len() == 0:
	case key != nil:
		sort.Sort(keySortable{result, func(m any) string {
			rv := reflect.ValueOf(m)
			if rv.Kind() != reflect.Map {
				return ""
			}
			ev := rv.MapIndex(reflect.ValueOf(key))
			if ev.CanInterface() {
				if s, ok := ev.Interface().(string); ok {
					return strings.ToLower(s)
				}
			}
			return ""
		}})
	case reflect.TypeOf(array[0]).Kind() == reflect.String:
		sort.Sort(keySortable{result, func(s any) string {
			return strings.ToUpper(s.(string))
		}})
	}
	return result
}

type keySortable struct {
	slice []any
	keyFn func(any) string
}

// Len is part of sort.Interface.
func (s keySortable) Len() int {
	return len(s.slice)
}

// Swap is part of sort.Interface.
func (s keySortable) Swap(i, j int) {
	a := s.slice
	a[i], a[j] = a[j], a[i]
}

// Less is part of sort.Interface.
func (s keySortable) Less(i, j int) bool {
	k, sl := s.keyFn, s.slice
	a, b := k(sl[i]), k(sl[j])
	return a < b
}
