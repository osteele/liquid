package expressions

import (
	"fmt"
	"reflect"
)

type genericError struct{ message string }

func (e *genericError) Error() string { return e.message }
func genericErrorf(format string, a ...interface{}) error {
	return &genericError{message: fmt.Sprintf(format, a...)}
}

type sortable []interface{}

// Len is part of sort.Interface.
func (s sortable) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s sortable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s sortable) Less(i, j int) bool {
	return genericSameTypeCompare(s[i], s[j]) < 0
}

// Convert val to the type. This is a more aggressive conversion, that will
// recursively create new map and slice values as necessary. It doesn't
// handle circular references.
func convertType(val interface{}, t reflect.Type) reflect.Value {
	r := reflect.ValueOf(val)
	if r.Type().ConvertibleTo(t) {
		return r.Convert(t)
	}
	if reflect.PtrTo(r.Type()) == t {
		return reflect.ValueOf(&val)
	}
	switch t.Kind() {
	case reflect.Slice:
		if r.Kind() != reflect.Array && r.Kind() != reflect.Slice {
			break
		}
		x := reflect.MakeSlice(t, 0, r.Len())
		for i := 0; i < r.Len(); i++ {
			c := convertType(r.Index(i).Interface(), t.Elem())
			x = reflect.Append(x, c)
		}
		return x
	}
	panic(genericErrorf("convertType: can't convert %#v<%s> to %v", val, r.Type(), t))
}

// Convert args to match the input types of fr, which should be a function reflection.
func convertArguments(fv reflect.Value, args []interface{}) []reflect.Value {
	rt := fv.Type()
	rs := make([]reflect.Value, rt.NumIn())
	for i, arg := range args {
		if i < rt.NumIn() {
			rs[i] = convertType(arg, rt.In(i))
		}
	}
	return rs
}

func genericSameTypeCompare(av, bv interface{}) int {
	a, b := reflect.ValueOf(av), reflect.ValueOf(bv)
	if a.Kind() != b.Kind() {
		panic(genericErrorf("different types: %v and %v", a, b))
	}
	if a == b {
		return 0
	}
	switch a.Kind() {
	case reflect.String:
		if a.String() < b.String() {
			return -1
		}
	default:
		panic(genericErrorf("unimplemented generic comparison for %s", a.Kind()))
	}
	return 1
}

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
