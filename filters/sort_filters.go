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
		values.SortByProperty(result, fmt.Sprint(key), false)
	}

	return result
}

func sortNaturalFilter(array []any, key any) any {
	result := make([]any, len(array))
	copy(result, array)

	switch {
	case reflect.ValueOf(array).Len() == 0:
	case key != nil:
		sort.SliceStable(result, func(i, j int) bool {
			getKey := func(m any) string {
				rv := reflect.ValueOf(m)
				if rv.Kind() != reflect.Map {
					return ""
				}
				ev := rv.MapIndex(reflect.ValueOf(key))
				if ev.IsValid() && ev.CanInterface() {
					if s, ok := ev.Interface().(string); ok {
						return strings.ToLower(s)
					}
				}
				return ""
			}
			ki, kj := getKey(result[i]), getKey(result[j])
			// Empty key (nil or missing) goes last.
			if ki == "" && kj == "" {
				return false
			}
			if ki == "" {
				return false
			}
			if kj == "" {
				return true
			}
			return ki < kj
		})
	default:
		// Find the first non-nil element to determine the element type.
		firstNonNil := -1
		for i, v := range result {
			if v != nil {
				firstNonNil = i
				break
			}
		}
		if firstNonNil == -1 {
			// All nils — nothing to sort.
			break
		}
		if reflect.TypeOf(result[firstNonNil]).Kind() == reflect.String {
			sort.SliceStable(result, func(i, j int) bool {
				a, b := result[i], result[j]
				if a == nil && b == nil {
					return false
				}
				if a == nil {
					return false // nil goes last
				}
				if b == nil {
					return true
				}
				return strings.ToUpper(a.(string)) < strings.ToUpper(b.(string))
			})
		}
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
