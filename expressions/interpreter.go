package expressions

import (
	"reflect"
)

func makeObjectPropertyEvaluator(obj func(Context) interface{}, attr string) func(Context) interface{} {
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(obj(ctx))
		switch ref.Kind() {
		case reflect.Map:
			value := ref.MapIndex(reflect.ValueOf(attr))
			if value.Kind() != reflect.Invalid {
				return value.Interface()
			}
		}
		return nil
	}
}

func makeIndexEvaluator(obj, indexfn func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		ref := reflect.ValueOf(obj(ctx))
		index := reflect.ValueOf(indexfn(ctx))
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			switch index.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				n := int(index.Int())
				if n < 0 {
					n = ref.Len() + n
				}
				if 0 <= n && n < ref.Len() {
					return ref.Index(n).Interface()
				}
			}
		}
		return nil
	}
}
