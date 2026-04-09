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

// Truthy returns true if value is truthy in Liquid semantics.
// Only nil and false (or defined bool types whose value is false) are falsy.
// This is the canonical truthiness gate for all condition evaluation.
// Uses reflect.Kind so defined types (e.g. type MyBool bool) work correctly.
func Truthy(value any) bool {
	value = ToLiquid(value)
	if value == nil {
		return false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Bool {
		return rv.Bool()
	}
	return true
}

// IsBlank returns true if value is nil, false, an empty or whitespace-only
// string, an empty array, or an empty map — matching Liquid blank? semantics.
// Uses reflect.Kind so that defined types (e.g. type MyBool bool, type MyStr string)
// are handled the same way as their underlying built-in types.
func IsBlank(value any) bool {
	value = ToLiquid(value)
	if value == nil {
		return true
	}

	r := reflect.ValueOf(value)
	switch r.Kind() {
	case reflect.Bool:
		return !r.Bool()
	case reflect.String:
		return strings.TrimSpace(r.String()) == ""
	case reflect.Array, reflect.Map, reflect.Slice:
		return r.Len() == 0
	default:
		return false
	}
}
