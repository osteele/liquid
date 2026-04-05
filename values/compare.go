package values

import (
	"reflect"
)

var (
	int64Type   = reflect.TypeOf(int64(0))
	float64Type = reflect.TypeOf(float64(0))
)

// NormalizeNumber converts any Go numeric scalar to one of three canonical types:
// int64 (signed integers), uint64 (unsigned integers), or float64.
// Non-numeric values are returned unchanged.
func NormalizeNumber(v any) any {
	if v == nil {
		return v
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() // always int64
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() // always uint64
	case reflect.Float32, reflect.Float64:
		return rv.Float() // always float64
	default:
		return v
	}
}

// isNormalizedNumeric reports whether v is one of the three canonical numeric
// types produced by NormalizeNumber.
func isNormalizedNumeric(v any) bool {
	switch v.(type) {
	case int64, uint64, float64:
		return true
	default:
		return false
	}
}

// numericCompare compares two normalized numeric values (int64, uint64, float64).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Both a and b must be the output of NormalizeNumber.
func numericCompare(a, b any) int {
	switch av := a.(type) {
	case int64:
		switch bv := b.(type) {
		case int64:
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		case uint64:
			// negative int is always less than any uint
			if av < 0 {
				return -1
			}
			u := uint64(av)
			if u < bv {
				return -1
			} else if u > bv {
				return 1
			}
			return 0
		case float64:
			f := float64(av)
			if f < bv {
				return -1
			} else if f > bv {
				return 1
			}
			return 0
		}
	case uint64:
		switch bv := b.(type) {
		case int64:
			// any uint is greater than any negative int
			if bv < 0 {
				return 1
			}
			ub := uint64(bv)
			if av < ub {
				return -1
			} else if av > ub {
				return 1
			}
			return 0
		case uint64:
			if av < bv {
				return -1
			} else if av > bv {
				return 1
			}
			return 0
		case float64:
			f := float64(av)
			if f < bv {
				return -1
			} else if f > bv {
				return 1
			}
			return 0
		}
	case float64:
		var fb float64
		switch bv := b.(type) {
		case int64:
			fb = float64(bv)
		case uint64:
			fb = float64(bv)
		case float64:
			fb = bv
		}
		if av < fb {
			return -1
		} else if av > fb {
			return 1
		}
		return 0
	}

	return 0
}

// Equal returns a bool indicating whether a == b after conversion.
func Equal(a, b any) bool { //nolint: gocyclo
	a, b = ToLiquid(a), ToLiquid(b)

	// EmptyDrop / BlankDrop: delegate to the drop's symmetric Equal logic.
	switch av := a.(type) {
	case *emptyDropValue:
		return av.Equal(ValueOf(b))
	case *blankDropValue:
		return av.Equal(ValueOf(b))
	}
	switch bv := b.(type) {
	case *emptyDropValue:
		return bv.Equal(ValueOf(a))
	case *blankDropValue:
		return bv.Equal(ValueOf(a))
	}

	if a == nil || b == nil {
		return a == b
	}

	// Normalize all Go numeric types to int64, uint64, or float64 before
	// comparing so that cross-type comparisons (e.g. uint32 == int64) work
	// correctly without losing precision for large uint64 values.
	a, b = NormalizeNumber(a), NormalizeNumber(b)
	if isNormalizedNumeric(a) && isNormalizedNumeric(b) {
		return numericCompare(a, b) == 0
	}

	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	switch joinKind(ra.Kind(), rb.Kind()) {
	case reflect.Array, reflect.Slice:
		if ra.Len() != rb.Len() {
			return false
		}

		for i := range ra.Len() {
			if !Equal(ra.Index(i).Interface(), rb.Index(i).Interface()) {
				return false
			}
		}

		return true
	case reflect.Bool:
		return ra.Bool() == rb.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ra.Convert(int64Type).Int() == rb.Convert(int64Type).Int()
	case reflect.Float32, reflect.Float64:
		return ra.Convert(float64Type).Float() == rb.Convert(float64Type).Float()
	case reflect.String:
		return ra.String() == rb.String()
	case reflect.Ptr:
		if rb.Kind() == reflect.Ptr && (ra.IsNil() || rb.IsNil()) {
			return ra.IsNil() == rb.IsNil()
		}

		return a == b
	default:
		return a == b
	}
}

// Less returns a bool indicating whether a < b.
func Less(a, b any) bool {
	a, b = ToLiquid(a), ToLiquid(b)
	if a == nil || b == nil {
		return false
	}

	// Normalize all Go numeric types to int64, uint64, or float64 before
	// comparing so that cross-type comparisons work correctly.
	a, b = NormalizeNumber(a), NormalizeNumber(b)
	if isNormalizedNumeric(a) && isNormalizedNumeric(b) {
		return numericCompare(a, b) < 0
	}

	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	switch joinKind(ra.Kind(), rb.Kind()) {
	case reflect.Bool:
		return !ra.Bool() && rb.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ra.Convert(int64Type).Int() < rb.Convert(int64Type).Int()
	case reflect.Float32, reflect.Float64:
		return ra.Convert(float64Type).Float() < rb.Convert(float64Type).Float()
	case reflect.String:
		return ra.String() < rb.String()
	default:
		return false
	}
}

func joinKind(a, b reflect.Kind) reflect.Kind { //nolint: gocyclo
	if a == b {
		return a
	}

	switch a {
	case reflect.Array, reflect.Slice:
		if b == reflect.Array || b == reflect.Slice {
			return reflect.Slice
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if isIntKind(b) {
			return reflect.Int64
		}

		if isFloatKind(b) {
			return reflect.Float64
		}
	case reflect.Float32, reflect.Float64:
		if isIntKind(b) || isFloatKind(b) {
			return reflect.Float64
		}
	}

	return reflect.Invalid
}

func isIntKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isFloatKind(k reflect.Kind) bool {
	switch k {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
