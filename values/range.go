package values

import (
	"fmt"
	"reflect"
)

// rangeValue wraps a Range and implements the Value interface with proper
// Contains semantics: membership test by integer value, not field name.
type rangeValue struct{ wrapperValue }

func (rv rangeValue) Contains(elem Value) bool {
	r := rv.value.(Range)
	// Convert the element to int using the same widening rules used elsewhere.
	raw := elem.Interface()
	if raw == nil {
		return false
	}
	v := reflect.ValueOf(raw)
	switch v.Kind() { //nolint:exhaustive
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return r.containsInt(int(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n := v.Uint()
		if n > uint64(^uint(0)>>1) { //nolint:gosec
			return false // overflows int
		}
		return r.containsInt(int(n)) //nolint:gosec
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		n := int(f)
		if float64(n) != f {
			return false // non-integer float
		}
		return r.containsInt(n)
	}
	return false
}

func (rv rangeValue) Equal(other Value) bool {
	switch o := other.(type) {
	case *emptyDropValue:
		return o.Equal(rv)
	case *blankDropValue:
		return o.Equal(rv)
	}
	if or, ok := other.Interface().(Range); ok {
		r := rv.value.(Range)
		return r == or
	}
	return false
}

func (rv rangeValue) Less(other Value) bool {
	return false // ranges have no natural ordering
}

// A Range is the range of integers from b to e inclusive.
type Range struct {
	b, e int
}

// NewRange returns a new Range
func NewRange(b, e int) Range {
	return Range{b, e}
}

// String renders a Range as "start..end", matching Ruby Liquid output.
func (r Range) String() string {
	return fmt.Sprintf("%d..%d", r.b, r.e)
}

// Len is in the iteration interface
func (r Range) Len() int { return r.e + 1 - r.b }

// Index is in the iteration interface
func (r Range) Index(i int) any { return r.b + i }

// AsArray converts the range into an array.
func (r Range) AsArray() []any {
	a := make([]any, 0, r.Len())
	for i := r.b; i <= r.e; i++ {
		a = append(a, i)
	}

	return a
}

// containsInt reports whether n is within the inclusive range [b, e].
func (r Range) containsInt(n int) bool {
	return n >= r.b && n <= r.e
}
