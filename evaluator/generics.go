// Package generics defines methods such as sorting, comparison, and type conversion, that apply to interface types.
//
// It is similar to, and makes heavy use of, the reflect package.
//
// Since the intent is to provide runtime services for the Liquid expression interpreter,
// this package does not implement "generic" generics.
// It attempts to implement Liquid semantics (which are largely Ruby semantics).
package evaluator

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

// Contains returns a boolean indicating whether array is a sequence that contains item.
func Contains(array interface{}, item interface{}) bool {
	item = ToLiquid(item)
	ref := reflect.ValueOf(array)
	switch ref.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < ref.Len(); i++ {
			if ref.Index(i).Interface() == item {
				return true
			}
		}
	}
	return false
}

// IsEmpty returns a bool indicating whether the value is empty according to Liquid semantics.
func IsEmpty(value interface{}) bool {
	value = ToLiquid(value)
	if value == nil {
		return false
	}
	r := reflect.ValueOf(value)
	switch r.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return r.Len() == 0
	case reflect.Bool:
		return !r.Bool()
	default:
		return false
	}
}

// IsTrue returns a bool indicating whether the value is true according to Liquid semantics.
func IsTrue(value interface{}) bool {
	value = ToLiquid(value)
	return value != nil && value != false
}

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
