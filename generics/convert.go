package generics

import (
	"reflect"
	"strconv"
	"time"
)

var timeType = reflect.TypeOf(time.Now())

// Convert value to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
//
// TODO It's weird that this takes an interface{} but returns a Value
func Convert(value interface{}, t reflect.Type) reflect.Value {
	r := reflect.ValueOf(value)
	if r.Type().ConvertibleTo(t) {
		return r.Convert(t)
	}
	if reflect.PtrTo(r.Type()) == t {
		return reflect.ValueOf(&value)
	}
	if r.Kind() == reflect.String && t == timeType {
		v, err := ParseTime(value.(string))
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(v)
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.Atoi(value.(string))
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(n)
	case reflect.Slice:
		if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
			break
		}
		x := reflect.MakeSlice(t, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			c := Convert(r.Index(i).Interface(), t.Elem())
			x = reflect.Append(x, c)
		}
		return x
	}
	panic(genericErrorf("generic.Convert can't convert %#v<%s> to %v", value, r.Type(), t))
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
