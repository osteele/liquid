// Package evaluator is an interim internal package that forwards to package values.
package evaluator

import (
	"reflect"
	"time"

	"github.com/osteele/liquid/values"
)

// Convert should be replaced by values.Convert.
func Convert(value any, typ reflect.Type) (any, error) {
	return values.Convert(value, typ)
}

// MustConvertItem should be replaced by values.Convert.
func MustConvertItem(item any, array any) any {
	return values.MustConvertItem(item, array)
}

// Sort should be replaced by values.
func Sort(data []any) {
	values.Sort(data)
}

// SortByProperty should be replaced by values.SortByProperty
func SortByProperty(data []any, key string, nilFirst bool) {
	values.SortByProperty(data, key, nilFirst)
}

// ParseDate should be replaced by values.SortByProperty
func ParseDate(s string) (time.Time, error) {
	return values.ParseDate(s)
}
