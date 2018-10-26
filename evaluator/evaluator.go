// Package evaluator is an interim internal package that forwards to package values.
package evaluator

import (
	"reflect"
	"time"

	"github.com/urbn8/liquid/values"
)

// Convert should be replaced by values.Convert.
func Convert(value interface{}, typ reflect.Type) (interface{}, error) {
	return values.Convert(value, typ)
}

// MustConvertItem should be replaced by values.Convert.
func MustConvertItem(item interface{}, array interface{}) interface{} {
	return values.MustConvertItem(item, array)
}

// Sort should be replaced by values.
func Sort(data []interface{}) {
	values.Sort(data)
}

// SortByProperty should be replaced by values.SortByProperty
func SortByProperty(data []interface{}, key string, nilFirst bool) {
	values.SortByProperty(data, key, nilFirst)
}

// ParseDate should be replaced by values.SortByProperty
func ParseDate(s string) (time.Time, error) {
	return values.ParseDate(s)
}
