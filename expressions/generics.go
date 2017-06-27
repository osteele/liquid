package expressions

import (
	"fmt"
	"reflect"
	"sort"
)

type genericError string

func (e genericError) Error() string { return string(e) }

func genericErrorf(format string, a ...interface{}) error {
	return genericError(fmt.Sprintf(format, a...))
}

type genericSortable []interface{}

func genericSort(data []interface{}) {
	sort.Sort(genericSortable(data))
}

// Len is part of sort.Interface.
func (s genericSortable) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s genericSortable) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s genericSortable) Less(i, j int) bool {
	return genericSameTypeCompare(s[i], s[j]) < 0
}

func sortByProperty(data []interface{}, key string) {
	sort.Sort(sortableByProperty{data, key})
}

type sortableByProperty struct {
	data []interface{}
	key  string
}

// Len is part of sort.Interface.
func (s sortableByProperty) Len() int {
	return len(s.data)
}

// Swap is part of sort.Interface.
func (s sortableByProperty) Swap(i, j int) {
	data := s.data
	data[i], data[j] = data[j], data[i]
}

// Less is part of sort.Interface.
func (s sortableByProperty) Less(i, j int) bool {
	// index returns the value at the s.key, if in is a map that contains this key
	index := func(in interface{}) interface{} {
		rt := reflect.ValueOf(in)
		if rt.Kind() == reflect.Map && rt.Type().Key().Kind() == reflect.String {
			return rt.MapIndex(reflect.ValueOf(s.key)).Interface()
		}
		return nil
	}
	a, b := index(s.data[i]), index(s.data[j])
	// TODO implement nil-first vs. nil last
	switch {
	case a == nil:
		return true
	case b == nil:
		return false
	default:
		// TODO relax same type requirement
		return genericSameTypeCompare(a, b) < 0
	}
}

// genericApply applies a function to arguments, converting them as necessary.
// The conversion follows Liquid semantics, which are more aggressive than
// Go conversion. The function should return one or two values; the second value,
// if present, should be an error.
func genericApply(fn reflect.Value, args []interface{}) (interface{}, error) {
	in := convertArguments(fn, args)
	outs := fn.Call(in)
	if len(outs) > 1 && outs[1].Interface() != nil {
		switch e := outs[1].Interface().(type) {
		case error:
			return nil, e
		default:
			panic(e)
		}
	}
	return outs[0].Interface(), nil
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

// Convert args to match the input types of function fn.
func convertArguments(fn reflect.Value, in []interface{}) []reflect.Value {
	rt := fn.Type()
	out := make([]reflect.Value, rt.NumIn())
	for i, arg := range in {
		if i < rt.NumIn() {
			out[i] = convertType(arg, rt.In(i))
		}
	}
	for i := len(in); i < rt.NumIn(); i++ {
		out[i] = reflect.Zero(rt.In(i))
	}
	return out
}

func genericSameTypeCompare(av, bv interface{}) int {
	a, b := reflect.ValueOf(av), reflect.ValueOf(bv)
	if a.Kind() != b.Kind() {
		panic(fmt.Errorf("genericSameTypeCompare called on different types: %v and %v", a, b))
	}
	if a == b {
		return 0
	}
	switch a.Kind() {
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
		panic(genericErrorf("unimplemented generic same-type comparison for %v and %v", a, b))
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
