package values

import (
	"fmt"
	"reflect"
	"strings"

	yaml "gopkg.in/yaml.v2"
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
	case 0:
		return zeroValue
	case 1:
		return oneValue
	}
	// interfaces
	switch v := value.(type) {
	case drop:
		return &dropWrapper{d: v}
	case yaml.MapSlice:
		return mapSliceValue{slice: v}
	case Value:
		return v
	}
	switch reflect.TypeOf(value).Kind() {
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

// embed this in a struct to "inherit" default implementations of the Value interface
type valueEmbed struct{}

func (v valueEmbed) Equal(Value) bool          { return false }
func (v valueEmbed) Less(Value) bool           { return false }
func (v valueEmbed) IndexValue(Value) Value    { return nilValue }
func (v valueEmbed) Contains(Value) bool       { return false }
func (v valueEmbed) Int() int                  { panic(conversionError("", v, reflect.TypeOf(1))) }
func (v valueEmbed) PropertyValue(Value) Value { return nilValue }
func (v valueEmbed) Test() bool                { return true }

// A wrapperValue wraps a Go value.
type wrapperValue struct{ value interface{} }

func (v wrapperValue) Equal(other Value) bool    { return Equal(v.value, other.Interface()) }
func (v wrapperValue) Less(other Value) bool     { return Less(v.value, other.Interface()) }
func (v wrapperValue) IndexValue(Value) Value    { return nilValue }
func (v wrapperValue) Contains(Value) bool       { return false }
func (v wrapperValue) Interface() interface{}    { return v.value }
func (v wrapperValue) PropertyValue(Value) Value { return nilValue }
func (v wrapperValue) Test() bool                { return v.value != nil && v.value != false }

func (v wrapperValue) Int() int {
	if n, ok := v.value.(int); ok {
		return n
	}
	panic(conversionError("", v.value, reflect.TypeOf(1)))
}

// interned values
var nilValue = wrapperValue{nil}
var falseValue = wrapperValue{false}
var trueValue = wrapperValue{true}
var zeroValue = wrapperValue{0}
var oneValue = wrapperValue{1}

// container values
type arrayValue struct{ wrapperValue }
type mapValue struct{ wrapperValue }
type stringValue struct{ wrapperValue }

func (av arrayValue) Contains(ev Value) bool {
	ar := reflect.ValueOf(av.value)
	e := ev.Interface()
	for i, len := 0, ar.Len(); i < len; i++ {
		if Equal(ar.Index(i).Interface(), e) {
			return true
		}
	}
	return false
}

func (av arrayValue) IndexValue(iv Value) Value {
	ar := reflect.ValueOf(av.value)
	var n int
	switch ix := iv.Interface().(type) {
	case int:
		n = ix
	case float32:
		// Ruby array indexing truncates floats
		n = int(ix)
	case float64:
		n = int(ix)
	default:
		return nilValue
	}
	if n < 0 {
		n += ar.Len()
	}
	if 0 <= n && n < ar.Len() {
		return ValueOf(ar.Index(n).Interface())
	}
	return nilValue
}

func (av arrayValue) PropertyValue(iv Value) Value {
	ar := reflect.ValueOf(av.value)
	switch iv.Interface() {
	case firstKey:
		if ar.Len() > 0 {
			return ValueOf(ar.Index(0).Interface())
		}
	case lastKey:
		if ar.Len() > 0 {
			return ValueOf(ar.Index(ar.Len() - 1).Interface())
		}
	case sizeKey:
		return ValueOf(ar.Len())
	}
	return nilValue
}

func (mv mapValue) Contains(iv Value) bool {
	mr := reflect.ValueOf(mv.value)
	ir := reflect.ValueOf(iv.Interface())
	if ir.IsValid() && mr.Type().Key() == ir.Type() {
		return mr.MapIndex(ir).IsValid()
	}
	return false
}

func (mv mapValue) IndexValue(iv Value) Value {
	mr := reflect.ValueOf(mv.value)
	ir := reflect.ValueOf(iv.Interface())
	kt := mr.Type().Key()
	if ir.IsValid() && ir.Type().ConvertibleTo(kt) && ir.Type().Comparable() {
		er := mr.MapIndex(ir.Convert(kt))
		if er.IsValid() {
			return ValueOf(er.Interface())
		}
	}
	return nilValue
}

func (mv mapValue) PropertyValue(iv Value) Value {
	mr := reflect.ValueOf(mv.Interface())
	ir := reflect.ValueOf(iv.Interface())
	if !ir.IsValid() {
		return nilValue
	}
	er := mr.MapIndex(ir)
	switch {
	case er.IsValid():
		return ValueOf(er.Interface())
	case iv.Interface() == sizeKey:
		return ValueOf(mr.Len())
	default:
		return nilValue
	}
}

func (sv stringValue) Contains(substr Value) bool {
	s, ok := substr.Interface().(string)
	if !ok {
		s = fmt.Sprint(substr.Interface())
	}
	return strings.Contains(sv.value.(string), s)
}

func (sv stringValue) PropertyValue(iv Value) Value {
	if iv.Interface() == sizeKey {
		return ValueOf(len(sv.value.(string)))
	}
	return nilValue
}
