package values

import (
	"reflect"
)

// TODO Length is now only used by the "size" filter.
// Maybe it should go somewhere else.

// Length returns the length of a string or array. In keeping with Liquid semantics,
// and contra Go, it does not return the size of a map.
func Length(value interface{}) int {
	value = ToLiquid(value)
	ref := reflect.ValueOf(value)
	switch ref.Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return ref.Len()
	default:
		return 0
	}
}
