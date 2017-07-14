package evaluator

import (
	"reflect"
	"strings"
)

// Contains returns a boolean indicating whether array is a sequence that contains item.
func Contains(array interface{}, item interface{}) bool {
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

// ContainsString returns a bool indicating whether a string or array contains an object.
func ContainsString(container interface{}, item string) bool {
	switch container := container.(type) {
	case string:
		return strings.Contains(container, item)
	case []string:
		for _, s := range container {
			if s == item {
				return true
			}
		}
	case []interface{}:
		for _, k := range container {
			if s, ok := k.(string); ok && s == item {
				return true
			}
		}
	default:
		return false
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
