package evaluator

import (
	"math"
	"reflect"
)

// Index returns sequence[ix] according to Liquid semantics.
func Index(sequence, ix interface{}) interface{} { // nolint: gocyclo
	ref := reflect.ValueOf(ToLiquid(sequence))
	ixRef := reflect.ValueOf(ix)
	if !ref.IsValid() || !ixRef.IsValid() {
		return nil
	}
	switch ref.Kind() {
	case reflect.Array, reflect.Slice:
		switch ixRef.Kind() {
		case reflect.Float32, reflect.Float64:
			if n, frac := math.Modf(ixRef.Float()); frac == 0 {
				ix = int(n)
				ixRef = reflect.ValueOf(ix)
			}
		}
		switch ixRef.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := int(ixRef.Int())
			if n < 0 {
				n = ref.Len() + n
			}
			if 0 <= n && n < ref.Len() {
				return ToLiquid(ref.Index(n).Interface())
			}
		}
	case reflect.Map:
		if ixRef.Type().ConvertibleTo(ref.Type().Key()) {
			item := ref.MapIndex(ixRef.Convert(ref.Type().Key()))
			if item.IsValid() {
				return item.Interface()
			}
		}
	}
	return nil
}

const (
	sizeProperty  = "size"
	firstProperty = "first"
	lastProperty  = "last"
)

// ObjectProperty object.name according to Liquid semantics.
func ObjectProperty(object interface{}, name string) interface{} { // nolint: gocyclo
	ref := reflect.ValueOf(ToLiquid(object))
	switch ref.Kind() {
	case reflect.Array, reflect.Slice:
		if ref.Len() == 0 {
			return nil
		}
		switch name {
		case firstProperty:
			return ToLiquid(ref.Index(0).Interface())
		case lastProperty:
			return ToLiquid(ref.Index(ref.Len() - 1).Interface())
		case sizeProperty:
			return ref.Len()
		}
	case reflect.String:
		if name == sizeProperty {
			return ref.Len()
		}
	case reflect.Map:
		value := ref.MapIndex(reflect.ValueOf(name))
		if value.Kind() != reflect.Invalid {
			return ToLiquid(value.Interface())
		}
		if name == sizeProperty {
			return reflect.ValueOf(name).Len()
		}
	}
	return nil
}
