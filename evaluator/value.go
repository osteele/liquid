package evaluator

import (
	"fmt"
	"reflect"
	"strings"
)

// A Value is a Liquid runtime value.
type Value interface {
	// Value retrieval
	Interface() interface{}
	Int() int

	// Comparison
	Equal(Value) bool
	Less(Value) bool

	Contains(Value) bool
	IndexValue(Value) Value
	PropertyValue(Value) Value

	// Predicate
	Test() bool
}

// ValueOf returns a Value that wraps its argument.
// If the argument is already a Value, it returns this.
func ValueOf(value interface{}) Value { // nolint: gocyclo
	// interned values
	switch value {
	case nil:
		return nilValue
	case true:
		return trueValue
	case false:
		return falseValue
	}
	switch v := value.(type) {
	case drop:
		return dropWrapper{d: v}
	case Value:
		return v
	}
	rk := reflect.TypeOf(value).Kind()
	if rk <= reflect.Float64 {
		return wrapperValue{value}
	}
	switch rk {
	case reflect.Ptr:
		rv := reflect.ValueOf(value)
		if rv.Type().Elem().Kind() == reflect.Struct {
			return structValue{wrapperValue{value}}
		}
		return ValueOf(rv.Elem().Interface())
	case reflect.String:
		return stringValue{wrapperValue{value}}
	case reflect.Array, reflect.Slice:
		return arrayValue{wrapperValue{value}}
	case reflect.Map:
		return mapValue{wrapperValue{value}}
	case reflect.Struct:
		return structValue{wrapperValue{value}}
	default:
		return wrapperValue{value}
	}
}

const (
	firstKey = "first"
	lastKey  = "last"
	sizeKey  = "size"
)

// embed this in a struct to give it default implementations of the Value interface
type valueEmbed struct{}

func (v valueEmbed) Equal(Value) bool          { return false }
func (v valueEmbed) Less(Value) bool           { return false }
func (v valueEmbed) IndexValue(Value) Value    { return nilValue }
func (v valueEmbed) Contains(Value) bool       { return false }
func (v valueEmbed) Int() int                  { panic(conversionError("", v, reflect.TypeOf(1))) }
func (v valueEmbed) Interface() interface{}    { return nil }
func (v valueEmbed) PropertyValue(Value) Value { return nilValue }
func (v valueEmbed) Test() bool                { return true }

type wrapperValue struct{ basis interface{} }

func (v wrapperValue) Equal(other Value) bool    { return Equal(v.basis, other.Interface()) }
func (v wrapperValue) Less(other Value) bool     { return Less(v.basis, other.Interface()) }
func (v wrapperValue) IndexValue(Value) Value    { return nilValue }
func (v wrapperValue) Contains(Value) bool       { return false }
func (v wrapperValue) Interface() interface{}    { return v.basis }
func (v wrapperValue) PropertyValue(Value) Value { return nilValue }
func (v wrapperValue) Test() bool                { return v.basis != nil && v.basis != false }

func (v wrapperValue) Int() int {
	if n, ok := v.basis.(int); ok {
		return n
	}
	panic(conversionError("", v.basis, reflect.TypeOf(1)))
}

// interned values
var nilValue = wrapperValue{nil}
var falseValue = wrapperValue{false}
var trueValue = wrapperValue{true}

type arrayValue struct{ wrapperValue }
type mapValue struct{ wrapperValue }
type stringValue struct{ wrapperValue }

func (v arrayValue) Contains(elem Value) bool {
	rv := reflect.ValueOf(v.basis)
	e := elem.Interface()
	for i, len := 0, rv.Len(); i < len; i++ {
		if Equal(rv.Index(i).Interface(), e) {
			return true
		}
	}
	return false
}

func (v mapValue) Contains(index Value) bool {
	rv := reflect.ValueOf(v.basis)
	iv := reflect.ValueOf(index.Interface())
	if rv.Type().Key() == iv.Type() {
		return rv.MapIndex(iv).IsValid()
	}
	return false
}

func (v stringValue) Contains(substr Value) bool {
	s, ok := substr.Interface().(string)
	if !ok {
		s = fmt.Sprint(substr.Interface())
	}
	return strings.Contains(v.basis.(string), s)
}

func (v arrayValue) IndexValue(index Value) Value {
	rv := reflect.ValueOf(v.basis)
	var n int
	switch ix := index.Interface().(type) {
	case int:
		n = ix
	case float32:
		// this how Ruby arrays, and therefore Liquid arrays, work
		n = int(ix)
	case float64:
		n = int(ix)
	default:
		return nilValue
	}
	if n < 0 {
		n += rv.Len()
	}
	if 0 <= n && n < rv.Len() {
		return ValueOf(rv.Index(n).Interface())
	}
	return nilValue
}

func (v mapValue) IndexValue(index Value) Value {
	rv := reflect.ValueOf(v.basis)
	iv := reflect.ValueOf(index.Interface())
	if rv.Type().Key() == iv.Type() {
		ev := rv.MapIndex(iv)
		if ev.IsValid() {
			return ValueOf(ev.Interface())
		}
	}
	return nilValue
}

func (v arrayValue) PropertyValue(index Value) Value {
	rv := reflect.ValueOf(v.basis)
	switch index.Interface() {
	case firstKey:
		if rv.Len() > 0 {
			return ValueOf(rv.Index(0).Interface())
		}
	case lastKey:
		if rv.Len() > 0 {
			return ValueOf(rv.Index(rv.Len() - 1).Interface())
		}
	case sizeKey:
		return ValueOf(rv.Len())
	}
	return nilValue
}

func (v mapValue) PropertyValue(index Value) Value {
	rv := reflect.ValueOf(v.Interface())
	iv := reflect.ValueOf(index.Interface())
	ev := rv.MapIndex(iv)
	switch {
	case ev.IsValid():
		return ValueOf(ev.Interface())
	case index.Interface() == sizeKey:
		return ValueOf(rv.Len())
	default:
		return nilValue
	}
}

func (v stringValue) PropertyValue(index Value) Value {
	if index.Interface() == sizeKey {
		return ValueOf(len(v.basis.(string)))
	}
	return nilValue
}
