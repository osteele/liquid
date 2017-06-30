package expressions

import (
	"reflect"
	"strings"
)

func makeContainsExpr(e1, e2 func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		a, aok := e1(ctx).(string)
		b, bok := e2(ctx).(string)
		return aok && bok && strings.Contains(a, b)
	}
}

func makeFilter(fn valueFn, name string, args []valueFn) valueFn {
	return func(ctx Context) interface{} {
		return ctx.Filters().runFilter(ctx, fn, name, args)
	}
}

func makeIndexExpr(obj, index func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(obj(ctx))
		i := reflect.ValueOf(index(ctx))
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			switch i.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n := int(i.Int())
				if n < 0 {
					n = ref.Len() + n
				}
				if 0 <= n && n < ref.Len() {
					return ref.Index(n).Interface()
				}
			}
		case reflect.Map:
			if i.Type().ConvertibleTo(ref.Type().Key()) {
				item := ref.MapIndex(i.Convert(ref.Type().Key()))
				if item.IsValid() {
					return item.Interface()
				}
			}
		}
		return nil
	}
}

func makeObjectPropertyExpr(obj func(Context) interface{}, attr string) func(Context) interface{} {
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(obj(ctx))
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			if ref.Len() == 0 {
				return nil
			}
			switch attr {
			case "first":
				return ref.Index(0).Interface()
			case "last":
				return ref.Index(ref.Len() - 1).Interface()
			case "size":
				return ref.Len()
			}
		case reflect.String:
			if attr == "size" {
				return ref.Len()
			}
		case reflect.Map:
			value := ref.MapIndex(reflect.ValueOf(attr))
			if value.Kind() != reflect.Invalid {
				return value.Interface()
			}
		}
		return nil
	}
}
