package generics

import (
	"fmt"
	"reflect"
)

// GenericError is an error regarding generic conversion.
type GenericError string

func (e GenericError) Error() string { return string(e) }

func genericErrorf(format string, a ...interface{}) error {
	return GenericError(fmt.Sprintf(format, a...))
}

// IsEmpty returns a bool indicating whether the value is empty according to Liquid semantics.
func IsEmpty(value interface{}) bool {
	if value == nil {
		return false
	}
	r := reflect.ValueOf(value)
	switch r.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return r.Len() == 0
	case reflect.Bool:
		return r.Bool() == false
	default:
		return false
	}
}

// IsTrue returns a bool indicating whether the value is true according to Liquid semantics.
func IsTrue(value interface{}) bool {
	return value != nil && value != false
}

// Length returns the length of a string or array. In keeping with Liquid semantics,
// and contra Go, it does not return the size of a map.
func Length(value interface{}) int {
	ref := reflect.ValueOf(value)
	switch ref.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
		return ref.Len();
		default:
		return 0
	}
}
