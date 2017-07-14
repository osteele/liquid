// Package evaluator defines methods such as sorting, comparison, and type conversion, that apply to interface types.
//
// It is similar to, and makes heavy use of, the reflect package.
//
// Since the intent is to provide runtime services for the Liquid expression interpreter,
// this package does not implement "generic" generics.
// It attempts to implement Liquid semantics (which are largely Ruby semantics).
package evaluator

import (
	"reflect"
)

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
