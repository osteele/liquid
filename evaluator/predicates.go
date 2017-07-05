package evaluator

import "reflect"

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
