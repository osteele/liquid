package values

import (
	"reflect"
	"strings"
)

// IsEmpty returns true if value is an empty string, empty array, or empty map
// according to Liquid semantics. nil and false are NOT empty.
func IsEmpty(value any) bool {
	value = ToLiquid(value)
	if value == nil {
		return false
	}

	r := reflect.ValueOf(value)
	switch r.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return r.Len() == 0
	default:
		return false
	}
}

// IsBlank returns true if value is nil, false, an empty or whitespace-only
// string, an empty array, or an empty map — matching Liquid blank? semantics.
func IsBlank(value any) bool {
	value = ToLiquid(value)
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case bool:
		return !v
	case string:
		return strings.TrimSpace(v) == ""
	}

	r := reflect.ValueOf(value)
	switch r.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return r.Len() == 0
	default:
		return false
	}
}

