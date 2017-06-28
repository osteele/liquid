package generics

import (
	"fmt"
	"reflect"
)

// Equal returns a bool indicating whether a == b after conversion.
func Equal(a, b interface{}) bool {
	return genericCompare(reflect.ValueOf(a), reflect.ValueOf(b)) == 0
}

// Less returns a bool indicating whether a < b.
func Less(a, b interface{}) bool {
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return true
	case b == nil:
		return false
	}
	c := genericCompare(reflect.ValueOf(a), reflect.ValueOf(b)) < 0
	return c
}

func genericSameTypeCompare(av, bv interface{}) int {
	switch {
	case av == nil && bv == nil:
		return 0
	case av == nil:
		return -1
	case bv == nil:
		return 1
	}
	a, b := reflect.ValueOf(av), reflect.ValueOf(bv)
	if a.Kind() != b.Kind() {
		panic(fmt.Errorf("genericSameTypeCompare called on different types: %v and %v", a, b))
	}
	if a == b {
		return 0
	}
	switch a.Kind() {
	case reflect.Bool:
		if !a.Bool() && b.Bool() {
			return -1
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() < b.Int() {
			return -1
		}
	case reflect.Float32, reflect.Float64:
		if a.Float() < b.Float() {
			return -1
		}
	case reflect.String:
		if a.String() < b.String() {
			return -1
		}
	default:
		panic(genericErrorf("unimplemented generic same-type comparison for %v<%s> and %v<%s>", a, a.Type(), b, b.Type()))
	}
	return 1
}

func genericCompare(a, b reflect.Value) int {
	if a.Interface() == b.Interface() {
		return 0
	}
	if a.Type() == b.Type() {
		return genericSameTypeCompare(a.Interface(), b.Interface())
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
			return genericCompare(reflect.ValueOf(float64(a.Int())), b)
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
	panic(genericErrorf("unimplemented: comparison of %v<%s> with %v<%s>", a, ak, b, bk))
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
