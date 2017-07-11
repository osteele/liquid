package expression

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

func makeContainsExpr(e1, e2 func(Context) interface{}) func(Context) interface{} { // nolint: gocyclo
	return func(ctx Context) interface{} {
		search, ok := e2((ctx)).(string)
		if !ok {
			return false
		}
		switch container := e1((ctx)).(type) {
		case string:
			return strings.Contains(container, search)
		case []string:
			for _, s := range container {
				if s == search {
					return true
				}
			}
		case []interface{}:
			for _, k := range container {
				if s, ok := k.(string); ok && s == search {
					return true
				}
			}
		default:
			return false
		}
		return false
	}
}

func makeFilter(fn valueFn, name string, args []valueFn) valueFn {
	return func(ctx Context) interface{} {
		return ctx.ApplyFilter(name, fn, args)
	}
}

func makeIndexExpr(objFn, indexFn func(Context) interface{}) func(Context) interface{} { // nolint: gocyclo
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(objFn(ctx))
		ix := indexFn(ctx)
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
			fmt.Println("map")
			if ixRef.Type().ConvertibleTo(ref.Type().Key()) {
				item := ref.MapIndex(ixRef.Convert(ref.Type().Key()))
				if item.IsValid() {
					return ToLiquid(item.Interface())
				}
			}
		}
		return nil
	}
}

func makeObjectPropertyExpr(objFn func(Context) interface{}, attr string) func(Context) interface{} { // nolint: gocyclo
	const sizeString = "size"
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(objFn(ctx))
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			if ref.Len() == 0 {
				return nil
			}
			switch attr {
			case "first":
				return ToLiquid(ref.Index(0).Interface())
			case "last":
				return ToLiquid(ref.Index(ref.Len() - 1).Interface())
			case sizeString:
				return ToLiquid(ref.Len())
			}
		case reflect.String:
			if attr == sizeString {
				return ToLiquid(ref.Len())
			}
		case reflect.Map:
			value := ref.MapIndex(reflect.ValueOf(attr))
			if value.Kind() != reflect.Invalid {
				return ToLiquid(value.Interface())
			}
			if attr == sizeString {
				return reflect.ValueOf(attr).Len()
			}
		}
		return nil
	}
}
