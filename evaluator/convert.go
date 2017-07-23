package evaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// A TypeError is an error during type conversion.
type TypeError string

func (e TypeError) Error() string { return string(e) }

func typeErrorf(format string, a ...interface{}) TypeError {
	return TypeError(fmt.Sprintf(format, a...))
}

var timeType = reflect.TypeOf(time.Now())

func conversionError(modifier string, value interface{}, typ reflect.Type) error {
	if modifier != "" {
		modifier += " "
	}
	switch ref := value.(type) {
	case reflect.Value:
		value = ref.Interface()
	}
	return typeErrorf("can't convert %s%T(%v) to type %s", modifier, value, value, typ)
}

// Convert value to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func Convert(value interface{}, typ reflect.Type) (interface{}, error) { // nolint: gocyclo
	value = ToLiquid(value)
	r := reflect.ValueOf(value)
	// int.Convert(string) returns "\x01" not "1", so guard against that in the following test
	if typ.Kind() != reflect.String && value != nil && r.Type().ConvertibleTo(typ) {
		return r.Convert(typ).Interface(), nil
	}
	if typ == timeType && r.Kind() == reflect.String {
		return ParseDate(value.(string))
	}
	// currently unused:
	// case reflect.PtrTo(r.Type()) == typ:
	// 	return &value, nil
	// }
	switch typ.Kind() {
	case reflect.Bool:
		return !(value == nil || value == false), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch value := value.(type) {
		case bool:
			if value {
				return 1, nil
			}
			return 0, nil
		case string:
			return strconv.Atoi(value)
		}
	case reflect.Float32, reflect.Float64:
		switch value := value.(type) {
		// case int is handled by r.Convert(type) above
		case string:
			return strconv.ParseFloat(value, 64)
		}
	case reflect.Map:
		out := reflect.MakeMap(typ)
		for _, key := range r.MapKeys() {
			if typ.Key().Kind() == reflect.String {
				key = reflect.ValueOf(fmt.Sprint(key))
			}
			if !key.Type().ConvertibleTo(typ.Key()) {
				return nil, conversionError("map key", key, typ.Key())
			}
			key = key.Convert(typ.Key())
			value := r.MapIndex(key)
			if typ.Elem().Kind() == reflect.String {
				value = reflect.ValueOf(fmt.Sprint(value))
			}
			if !value.Type().ConvertibleTo(typ.Elem()) {
				return nil, conversionError("map value", value, typ.Elem())
			}
			out.SetMapIndex(key, value.Convert(typ.Elem()))
		}
		return out.Interface(), nil
	case reflect.Slice:
		switch r.Kind() {
		case reflect.Array, reflect.Slice:
			out := reflect.MakeSlice(typ, 0, r.Len())
			for i := 0; i < r.Len(); i++ {
				item, err := Convert(r.Index(i).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				out = reflect.Append(out, reflect.ValueOf(item))
			}
			return out.Interface(), nil
		case reflect.Map:
			out := reflect.MakeSlice(typ, 0, r.Len())
			for _, key := range r.MapKeys() {
				item, err := Convert(r.MapIndex(key).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				out = reflect.Append(out, reflect.ValueOf(item))
			}
			return out.Interface(), nil
		}
	case reflect.String:
		switch value := value.(type) {
		case []byte:
			return string(value), nil
		case fmt.Stringer:
			return value.String(), nil
		default:
			return fmt.Sprint(value), nil
		}
	}
	return nil, conversionError("", value, typ)
}

// MustConvert is like Convert, but panics if conversion fails.
func MustConvert(value interface{}, t reflect.Type) interface{} {
	out, err := Convert(value, t)
	if err != nil {
		panic(err)
	}
	return out
}

// MustConvertItem converts item to conform to the type array's element, else panics.
// Unlike MustConvert, the second argument is a value not a type.
func MustConvertItem(item interface{}, array interface{}) interface{} {
	item, err := Convert(item, reflect.TypeOf(array).Elem())
	if err != nil {
		panic(typeErrorf("can't convert %#v to %s: %s", item, reflect.TypeOf(array).Elem(), err))
	}
	return item
}
