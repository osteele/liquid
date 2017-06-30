package generics

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var timeType = reflect.TypeOf(time.Now())

func conversionError(modifier string, value interface{}, typ reflect.Type) error {
	if modifier != "" {
		modifier += " "
	}
	switch ref := value.(type) {
	case reflect.Value:
		value = ref.Interface()
	}
	return genericErrorf("can't convert %s%T(%v) to type %s", modifier, value, value, typ)
}

// Convert value to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func Convert(value interface{}, target reflect.Type) (interface{}, error) {
	r := reflect.ValueOf(value)
	if r.Type().ConvertibleTo(target) {
		return r.Convert(target).Interface(), nil
	}
	if reflect.PtrTo(r.Type()) == target {
		return &value, nil
	}
	if r.Kind() == reflect.String && target == timeType {
		return ParseTime(value.(string))
	}
	switch target.Kind() {
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
		case int:
			return float64(value), nil
		case string:
			return strconv.ParseFloat(value, 64)
		}
	case reflect.Map:
		out := reflect.MakeMap(target)
		for _, key := range r.MapKeys() {
			if target.Key().Kind() == reflect.String {
				key = reflect.ValueOf(fmt.Sprint(key))
			}
			if !key.Type().ConvertibleTo(target.Key()) {
				return nil, conversionError("map key", key, target.Key())
			}
			key = key.Convert(target.Key())
			value := r.MapIndex(key)
			if target.Elem().Kind() == reflect.String {
				value = reflect.ValueOf(fmt.Sprint(value))
			}
			if !value.Type().ConvertibleTo(target.Elem()) {
				return nil, conversionError("map value", value, target.Elem())
			}
			out.SetMapIndex(key, value.Convert(target.Elem()))
		}
		return out.Interface(), nil
	case reflect.Slice:
		if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
			break
		}
		out := reflect.MakeSlice(target, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			item, err := Convert(r.Index(i).Interface(), target.Elem())
			if err != nil {
				return nil, err
			}
			out = reflect.Append(out, reflect.ValueOf(item))
		}
		return out.Interface(), nil
	}
	return nil, conversionError("", value, target)
}

// MustConvert wraps Convert, but panics on error.
func MustConvert(value interface{}, t reflect.Type) interface{} {
	out, err := Convert(value, t)
	if err != nil {
		panic(err)
	}
	return out
}

// MustConvertItem converts item to conform to array, else panics.
func MustConvertItem(item interface{}, array []interface{}) interface{} {
	item, err := Convert(item, reflect.TypeOf(array).Elem())
	if err != nil {
		panic(fmt.Errorf("can't convert %#v to %s: %s", item, reflect.TypeOf(array).Elem(), err))
	}
	return item
}
