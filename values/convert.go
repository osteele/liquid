package values

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	yaml "gopkg.in/yaml.v2"
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
	rv := reflect.ValueOf(value)
	// int.Convert(string) returns "\x01" not "1", so guard against that in the following test
	if typ.Kind() != reflect.String && value != nil && rv.Type().ConvertibleTo(typ) {
		return rv.Convert(typ).Interface(), nil
	}
	if typ == timeType && rv.Kind() == reflect.String {
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
		et := typ.Elem()
		result := reflect.MakeMap(typ)
		for _, key := range rv.MapKeys() {
			if typ.Key().Kind() == reflect.String {
				key = reflect.ValueOf(fmt.Sprint(key))
			}
			if !key.Type().ConvertibleTo(typ.Key()) {
				return nil, conversionError("map key", key, typ.Key())
			}
			key = key.Convert(typ.Key())
			ev := rv.MapIndex(key)
			if et.Kind() == reflect.String {
				ev = reflect.ValueOf(fmt.Sprint(ev))
			}
			if !ev.Type().ConvertibleTo(et) {
				return nil, conversionError("map value", ev, et)
			}
			result.SetMapIndex(key, ev.Convert(et))
		}
		return result.Interface(), nil
	case reflect.Slice:
		et := typ.Elem()
		if ms, ok := value.(yaml.MapSlice); ok {
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for _, item := range ms {
				// TODO something more nuanced
				if item.Value == nil {
					if et.Kind() >= reflect.Array {
						ev := reflect.Zero(et)
						result = reflect.Append(result, ev.Convert(et))
					}
					continue
				}
				ev := reflect.ValueOf(item.Value)
				if et.Kind() == reflect.String {
					ev = reflect.ValueOf(fmt.Sprint(ev))
				}
				if !ev.Type().ConvertibleTo(et) {
					return nil, conversionError("map value", ev, et)
				}
				result = reflect.Append(result, ev.Convert(et))
			}
			return result.Interface(), nil
		}
		switch rv.Kind() {
		case reflect.Array, reflect.Slice:
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				item, err := Convert(rv.Index(i).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				result = reflect.Append(result, reflect.ValueOf(item))
			}
			return result.Interface(), nil
		case reflect.Map:
			result := reflect.MakeSlice(typ, 0, rv.Len())
			for _, key := range rv.MapKeys() {
				item, err := Convert(rv.MapIndex(key).Interface(), typ.Elem())
				if err != nil {
					return nil, err
				}
				result = reflect.Append(result, reflect.ValueOf(item))
			}
			return result.Interface(), nil
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
