package values

import (
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	yaml "gopkg.in/yaml.v2"
)

// A Value is a Liquid runtime value.
type Value interface {
	// Value retrieval
	Interface() any
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
func ValueOf(value any) Value { //nolint: gocyclo
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
	case *emptyDropValue:
		return v
	case *blankDropValue:
		return v
	case drop:
		return &dropWrapper{d: v}
	case yaml.MapSlice:
		return mapSliceValue{slice: v}
	case Range:
		return rangeValue{wrapperValue{value}}
	case Value:
		return v
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Chan, reflect.Func, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
		// Unsupported Go kinds: not representable in Liquid. Return an
		// invalidKindValue whose every method panics with TypeError so the
		// expression evaluator can surface it as a template error.
		return invalidKindValue{reflect.TypeOf(value)}
	case reflect.Ptr:
		rv := reflect.ValueOf(value)
		if rv.IsNil() {
			return nilValue
		}

		if rv.Type().Elem().Kind() == reflect.Struct {
			return structValue{wrapperValue{value}}
		}

		return ValueOf(rv.Elem().Interface())
	case reflect.String:
		return stringValue{wrapperValue{value}}
	case reflect.Array, reflect.Slice:
		// Byte slices and byte arrays are rendered as strings, not as numeric arrays.
		rv := reflect.ValueOf(value)
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			if rv.Kind() == reflect.Slice {
				return stringValue{wrapperValue{string(rv.Bytes())}}
			}
			// fixed-size array: copy element-by-element
			b := make([]byte, rv.Len())
			for i := range rv.Len() {
				b[i] = byte(rv.Index(i).Uint())
			}
			return stringValue{wrapperValue{string(b)}}
		}
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

// invalidKindValue represents a Go value whose kind is not representable in
// Liquid templates (chan, func, complex, unsafe pointer). Every method panics
// with a descriptive TypeError so the expression evaluator's recover can
// surface it as a template render error instead of silently rendering nothing.
type invalidKindValue struct {
	goType reflect.Type
}

func (v invalidKindValue) msg() string {
	return fmt.Sprintf("unsupported type %s: chan, func, and complex values cannot be used in Liquid templates", v.goType)
}
func (v invalidKindValue) Interface() any            { panic(TypeError(v.msg())) }
func (v invalidKindValue) Int() int                  { panic(TypeError(v.msg())) }
func (v invalidKindValue) Test() bool                { panic(TypeError(v.msg())) }
func (v invalidKindValue) Equal(Value) bool          { panic(TypeError(v.msg())) }
func (v invalidKindValue) Less(Value) bool           { panic(TypeError(v.msg())) }
func (v invalidKindValue) Contains(Value) bool       { panic(TypeError(v.msg())) }
func (v invalidKindValue) IndexValue(Value) Value    { panic(TypeError(v.msg())) }
func (v invalidKindValue) PropertyValue(Value) Value { panic(TypeError(v.msg())) }

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
type wrapperValue struct{ value any }

func (v wrapperValue) Equal(other Value) bool {
	// Symmetric comparison: delegate to EmptyDrop/BlankDrop's own Equal so it
	// can apply its emptiness/blankness semantics against this value.
	switch o := other.(type) {
	case *emptyDropValue:
		return o.Equal(v)
	case *blankDropValue:
		return o.Equal(v)
	}
	return Equal(v.value, other.Interface())
}
func (v wrapperValue) Less(other Value) bool     { return Less(v.value, other.Interface()) }
func (v wrapperValue) IndexValue(Value) Value    { return nilValue }
func (v wrapperValue) Contains(Value) bool       { return false }
func (v wrapperValue) Interface() any            { return v.value }
func (v wrapperValue) PropertyValue(Value) Value { return nilValue }
func (v wrapperValue) Test() bool {
	if v.value == nil {
		return false
	}
	// Use reflect.Kind so that defined types based on bool (e.g. type MyBool bool)
	// are treated as falsy when their value is false, matching plain bool semantics.
	rv := reflect.ValueOf(v.value)
	if rv.Kind() == reflect.Bool {
		return rv.Bool()
	}
	return true
}

func (v wrapperValue) Int() int {
	if n, ok := v.value.(int); ok {
		return n
	}

	panic(conversionError("", v.value, reflect.TypeOf(1)))
}

// interned values
var (
	nilValue   = wrapperValue{nil}
	falseValue = wrapperValue{false}
	trueValue  = wrapperValue{true}
	zeroValue  = wrapperValue{0}
	oneValue   = wrapperValue{1}
)

// container values
type (
	arrayValue  struct{ wrapperValue }
	mapValue    struct{ wrapperValue }
	stringValue struct{ wrapperValue }
)

func (av arrayValue) Contains(ev Value) bool {
	ar := reflect.ValueOf(av.value)
	e := ev.Interface()

	l := ar.Len()
	for i := range l {
		if Equal(ar.Index(i).Interface(), e) {
			return true
		}
	}

	return false
}

func (av arrayValue) IndexValue(iv Value) Value {
	ar := reflect.ValueOf(av.value)

	var n int
	raw := iv.Interface()
	rv := reflect.ValueOf(raw)
	if rv.IsValid() {
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n = int(rv.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			n = int(rv.Uint()) //nolint:gosec // G115: array indexes are never near MaxUint64
		case reflect.Float32, reflect.Float64:
			// Ruby array indexing truncates floats
			n = int(rv.Float())
		default:
			return nilValue
		}
	} else {
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
		return ValueOf(utf8.RuneCountInString(sv.value.(string)))
	}

	return nilValue
}

// SafeValue is a wrapped interface{} to mark it as being safe so that auto-escape is not applied.
// It is used by the 'safe' filter which is added when (*Engine).SetAutoEscapeReplacer() is called.
type SafeValue struct {
	Value interface{}
}
