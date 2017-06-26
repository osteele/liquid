package expressions

import (
	"fmt"
	"reflect"
)

func GenericCompare(a, b reflect.Value) int {
	if a.Interface() == b.Interface() {
		return 0
	}
	ak, bk := a.Kind(), b.Kind()
	// _ = ak.Convert
	switch a.Kind() {
	case reflect.Bool:
		if b.Kind() == reflect.Bool {
			switch {
			case a.Bool() && b.Bool():
				return 0
			case a.Bool():
				return 1
			case b.Bool():
				return -1
			default:
				return 0
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if hasIntKind(b) {
			if a.Int() < b.Int() {
				return -1
			}
			if a.Int() > b.Int() {
				return 1
			}
			return 0
		}
		if hasFloatKind(b) {
			return GenericCompare(reflect.ValueOf(float64(a.Int())), b)
		}
	case reflect.Float32, reflect.Float64:
		if hasIntKind(b) {
			b = reflect.ValueOf(float64(b.Int()))
		}
		if hasFloatKind(b) {
			if a.Float() < b.Float() {
				return -1
			}
			if a.Float() > b.Float() {
				return 1
			}
			return 0
		}
	}
	panic(fmt.Errorf("unimplemented: comparison of %v<%s> with %v<%s>", a, ak, b, bk))
}

func hasIntKind(n reflect.Value) bool {
	switch n.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func hasFloatKind(n reflect.Value) bool {
	switch n.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
