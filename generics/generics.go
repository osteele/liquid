package generics

import (
	"fmt"
	"reflect"
	"time"
)

// GenericError is an error regarding generic conversion.
type GenericError string

func (e GenericError) Error() string { return string(e) }

func genericErrorf(format string, a ...interface{}) error {
	return GenericError(fmt.Sprintf(format, a...))
}

// Convert val to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func convertType(val interface{}, t reflect.Type) reflect.Value {
	r := reflect.ValueOf(val)
	if r.Type().ConvertibleTo(t) {
		return r.Convert(t)
	}
	if reflect.PtrTo(r.Type()) == t {
		return reflect.ValueOf(&val)
	}
	// if r.Kind() == reflect.String && t.Name() == "time.Time" {
	// 	fmt.Println("ok")
	// }
	switch t.Kind() {
	case reflect.Slice:
		if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
			break
		}
		x := reflect.MakeSlice(t, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			c := convertType(r.Index(i).Interface(), t.Elem())
			x = reflect.Append(x, c)
		}
		return x
	}
	panic(genericErrorf("convertType: can't convert %#v<%s> to %v", val, r.Type(), t))
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

var dateLayouts = []string{
	"2006-01-02 15:04:05 -07:00",
	"January 2, 2006",
	"2006-01-02",
}

// ParseTime tries a few heuristics to parse a date from a string
func ParseTime(value string) (time.Time, error) {
	if value == "now" {
		return time.Now(), nil
	}
	for _, layout := range dateLayouts {
		// fmt.Println(layout, time.Now().Format(layout), value)
		time, err := time.Parse(layout, value)
		if err == nil {
			return time, nil
		}
	}
	return time.Now(), genericErrorf("can't convert %s to a time", value)
}
