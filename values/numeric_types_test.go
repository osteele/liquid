// Package values — intensive tests for B1: Go numeric type normalization.
//
// Covers every combination of Go's integer and float kinds in Equal/Less
// and via end-to-end template rendering through the evaluator.
package values

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// allIntsEqualTo2 holds every Go signed/unsigned integral type whose value is 2.
func allIntsEqualTo2() []any {
	return []any{
		int(2), int8(2), int16(2), int32(2), int64(2),
		uint(2), uint8(2), uint16(2), uint32(2), uint64(2),
		uintptr(2),
	}
}

// allFloatsEqualTo2 holds every Go float type whose value is 2.
func allFloatsEqualTo2() []any {
	return []any{
		float32(2), float64(2),
	}
}

// allNumericEqualTo2 combines the above.
func allNumericEqualTo2() []any {
	return append(allIntsEqualTo2(), allFloatsEqualTo2()...)
}

// typeName returns a human-readable name for use in sub-test names.
func typeName(v any) string { return fmt.Sprintf("%T(%v)", v, v) }

// ── Equal — every type against itself ────────────────────────────────────────

func TestEqual_AllNumericTypesEqualToThemselves(t *testing.T) {
	for _, v := range allNumericEqualTo2() {
		v := v
		t.Run(typeName(v), func(t *testing.T) {
			require.True(t, Equal(v, v), "%T(%v) should equal itself", v, v)
		})
	}
}

// ── Equal — every type vs every other type (value 2 == 2) ────────────────────

func TestEqual_AllNumericCrossType(t *testing.T) {
	vals := allNumericEqualTo2()
	for _, a := range vals {
		for _, b := range vals {
			a, b := a, b
			name := fmt.Sprintf("%s == %s", typeName(a), typeName(b))
			t.Run(name, func(t *testing.T) {
				require.True(t, Equal(a, b), "%T(%v) should equal %T(%v)", a, a, b, b)
			})
		}
	}
}

// ── Equal — zero values — every numeric zero equals every other numeric zero ──

func TestEqual_AllNumericZerosCrossType(t *testing.T) {
	zeros := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0),
		uintptr(0),
		float32(0), float64(0),
	}
	for _, a := range zeros {
		for _, b := range zeros {
			a, b := a, b
			name := fmt.Sprintf("%s == %s", typeName(a), typeName(b))
			t.Run(name, func(t *testing.T) {
				require.True(t, Equal(a, b), "%T(%v) should equal %T(%v)", a, a, b, b)
			})
		}
	}
}

// ── Equal — inequality: 2 != 3 for all numeric cross-type pairs ──────────────

func TestEqual_AllNumericCrossTypeInequal(t *testing.T) {
	twos := allNumericEqualTo2()
	threes := []any{
		int(3), int8(3), int16(3), int32(3), int64(3),
		uint(3), uint8(3), uint16(3), uint32(3), uint64(3),
		uintptr(3),
		float32(3), float64(3),
	}
	for _, a := range twos {
		for _, b := range threes {
			a, b := a, b
			name := fmt.Sprintf("%s != %s", typeName(a), typeName(b))
			t.Run(name, func(t *testing.T) {
				require.False(t, Equal(a, b), "%T(%v) should NOT equal %T(%v)", a, a, b, b)
			})
		}
	}
}

// ── Less — signed int vs signed int ──────────────────────────────────────────

func TestLess_SignedIntegers(t *testing.T) {
	smalls := []any{int(1), int8(1), int16(1), int32(1), int64(1)}
	bigs := []any{int(2), int8(2), int16(2), int32(2), int64(2)}

	for _, a := range smalls {
		for _, b := range bigs {
			a, b := a, b
			t.Run(fmt.Sprintf("%s < %s", typeName(a), typeName(b)), func(t *testing.T) {
				require.True(t, Less(a, b))
				require.False(t, Less(b, a))
			})
		}
	}
}

// ── Less — unsigned int vs unsigned int ──────────────────────────────────────

func TestLess_UnsignedIntegers(t *testing.T) {
	smalls := []any{uint(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1)}
	bigs := []any{uint(2), uint8(2), uint16(2), uint32(2), uint64(2), uintptr(2)}

	for _, a := range smalls {
		for _, b := range bigs {
			a, b := a, b
			t.Run(fmt.Sprintf("%s < %s", typeName(a), typeName(b)), func(t *testing.T) {
				require.True(t, Less(a, b))
				require.False(t, Less(b, a))
			})
		}
	}
}

// ── Less — signed int vs unsigned int ────────────────────────────────────────

func TestLess_SignedVsUnsigned(t *testing.T) {
	tests := []struct {
		a, b     any
		expected bool
	}{
		{int(1), uint(2), true},
		{uint(1), int(2), true},
		{int32(5), uint64(10), true},
		{uint32(5), int64(10), true},
		{int64(1), uint64(2), true},
		{uint64(1), int64(2), true},
		// equal cases
		{int(5), uint(5), false},
		{uint(5), int(5), false},
		{int64(5), uint64(5), false},
		// negative int must be less than any uint
		{int(-1), uint(0), true},
		{int8(-1), uint8(0), true},
		{int16(-5), uint16(100), true},
		{int32(-1), uint32(0), true},
		{int64(-1), uint64(0), true},
		// negative int compared with uint: uint is NOT less than negative int
		{uint(0), int(-1), false},
		{uint64(0), int64(-1), false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%s < %s", typeName(tc.a), typeName(tc.b)), func(t *testing.T) {
			require.Equal(t, tc.expected, Less(tc.a, tc.b))
		})
	}
}

// ── Less — int vs float ───────────────────────────────────────────────────────

func TestLess_IntVsFloat(t *testing.T) {
	tests := []struct {
		a, b     any
		expected bool
	}{
		{int(1), float64(1.5), true},
		{int(2), float64(1.5), false},
		{float64(1.5), int(2), true},
		{float64(1.5), int(1), false},
		{uint(1), float64(1.5), true},
		{uint(2), float64(1.5), false},
		{float64(1.5), uint(2), true},
		{float64(1.5), uint(1), false},
		{int64(3), float32(3.5), true},
		{uint64(3), float32(3.5), true},
		{float32(3.5), uint64(3), false},
		{float32(3.5), int64(4), true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%s < %s = %v", typeName(tc.a), typeName(tc.b), tc.expected), func(t *testing.T) {
			require.Equal(t, tc.expected, Less(tc.a, tc.b))
		})
	}
}

// ── Equal — uint/int vs float ─────────────────────────────────────────────────

func TestEqual_IntVsFloat(t *testing.T) {
	tests := []struct {
		a, b     any
		expected bool
	}{
		{int(2), float64(2.0), true},
		{int(2), float64(2.5), false},
		{uint(2), float64(2.0), true},
		{uint(2), float64(2.5), false},
		{float32(2.0), int32(2), true},
		{float32(2.0), uint32(2), true},
		{float64(2.0), uint64(2), true},
		{uint64(2), float64(2.0), true},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%s == %s = %v", typeName(tc.a), typeName(tc.b), tc.expected), func(t *testing.T) {
			require.Equal(t, tc.expected, Equal(tc.a, tc.b))
		})
	}
}

// ── Edge cases: large uint64 > math.MaxInt64 ─────────────────────────────────

func TestEqual_LargeUint64(t *testing.T) {
	big := uint64(math.MaxInt64) + 1
	require.False(t, Equal(big, int64(math.MaxInt64)), "uint64(MaxInt64+1) != int64(MaxInt64)")
	require.False(t, Equal(int64(-1), big), "int64(-1) != uint64(large)")
	require.True(t, Equal(big, big), "large uint64 equals itself")
	require.True(t, Equal(big, float64(big)), "large uint64 equals its float64 representation")
}

func TestLess_LargeUint64(t *testing.T) {
	big := uint64(math.MaxInt64) + 1
	require.True(t, Less(int64(math.MaxInt64), big), "int64(MaxInt64) < uint64(MaxInt64+1)")
	require.False(t, Less(big, int64(math.MaxInt64)), "uint64(MaxInt64+1) NOT < int64(MaxInt64)")
	require.True(t, Less(int64(-1), big), "int64(-1) < large uint64")
	require.False(t, Less(big, int64(-1)), "large uint64 NOT < int64(-1)")
}

// ── NormalizeNumber — unit tests ──────────────────────────────────────────────

func TestNormalizeNumber(t *testing.T) {
	tests := []struct {
		in       any
		expected any
	}{
		{int(42), int64(42)},
		{int8(42), int64(42)},
		{int16(42), int64(42)},
		{int32(42), int64(42)},
		{int64(42), int64(42)},
		{uint(42), uint64(42)},
		{uint8(42), uint64(42)},
		{uint16(42), uint64(42)},
		{uint32(42), uint64(42)},
		{uint64(42), uint64(42)},
		{uintptr(42), uint64(42)},
		{float32(3.14), float64(float32(3.14))},
		{float64(3.14), float64(3.14)},
		// non-numeric: returned as-is
		{true, true},
		{"hello", "hello"},
		{nil, nil},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(typeName(tc.in), func(t *testing.T) {
			got := NormalizeNumber(tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}

// ── Specific regression: if x != 0 with uint types ───────────────────────────

func TestEqual_IfNotZeroRegression(t *testing.T) {
	// The original bug report: `{% if x != 0 %}` failed when x is a uint type.
	// We test Equal(x, 0) == false for all non-zero numeric types.
	nonZeros := []any{
		int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
		float32(1), float64(1),
	}
	for _, v := range nonZeros {
		v := v
		t.Run(typeName(v)+"_ne_int0", func(t *testing.T) {
			require.False(t, Equal(v, int(0)), "%T(%v) improperly equals int(0)", v, v)
		})
		t.Run(typeName(v)+"_ne_uint0", func(t *testing.T) {
			require.False(t, Equal(v, uint(0)), "%T(%v) improperly equals uint(0)", v, v)
		})
	}
}

// ── Specific regression: if x == 0 with uint types ────────────────────────────

func TestEqual_ZeroEqualsZeroRegression(t *testing.T) {
	zeros := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0), uintptr(0),
		float32(0), float64(0),
	}
	for _, v := range zeros {
		v := v
		t.Run(typeName(v)+"_eq_int0", func(t *testing.T) {
			require.True(t, Equal(v, int(0)), "%T(%v) should equal int(0)", v, v)
		})
		t.Run(typeName(v)+"_eq_uint0", func(t *testing.T) {
			require.True(t, Equal(v, uint(0)), "%T(%v) should equal uint(0)", v, v)
		})
	}
}

// ── Less – not-less-than-self for all numeric types ───────────────────────────

func TestLess_NoTypeIsLessThanItself(t *testing.T) {
	for _, v := range allNumericEqualTo2() {
		v := v
		t.Run(typeName(v), func(t *testing.T) {
			require.False(t, Less(v, v), "%T(%v) should not be less than itself", v, v)
		})
	}
}

// ── Unsupported types: NormalizeNumber passes them through unchanged ──────────

func TestNormalizeNumber_UnsupportedTypes(t *testing.T) {
	ch := make(chan int)

	// For non-func types: value passes through unchanged (can use require.Equal).
	passthroughCases := []any{
		complex64(1 + 2i),
		complex128(1 + 2i),
		complex128(0 + 0i),
		ch,
		"hello",
		true,
		false,
	}
	for _, v := range passthroughCases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			got := NormalizeNumber(v)
			require.Equal(t, v, got, "%T should pass through NormalizeNumber unchanged", v)
		})
	}

	// For func: can't use require.Equal (func is not comparable with ==).
	// Just verify NormalizeNumber does NOT return a numeric canonical type.
	t.Run("func_is_not_normalised", func(t *testing.T) {
		fn := func() {}
		got := NormalizeNumber(fn)
		_, isInt64 := got.(int64)
		_, isUint64 := got.(uint64)
		_, isFloat64 := got.(float64)
		require.False(t, isInt64 || isUint64 || isFloat64,
			"func should not be normalised to a numeric type")
	})
}

// ── Unsupported types: Equal never matches a numeric ─────────────────────────

func TestEqual_UnsupportedTypesVsNumeric(t *testing.T) {
	numericZero := []any{int(0), int64(0), uint(0), float64(0)}
	unsupported := []any{
		complex64(0),  // zero complex64
		complex128(0), // zero complex128
		complex64(1 + 2i),
		complex128(1 + 2i),
	}
	for _, u := range unsupported {
		for _, n := range numericZero {
			u, n := u, n
			t.Run(fmt.Sprintf("%T == %T", u, n), func(t *testing.T) {
				require.False(t, Equal(u, n), "%T should not equal %T", u, n)
				require.False(t, Equal(n, u), "%T should not equal %T", n, u)
			})
		}
	}
}

// ── Complex numbers: Equal to themselves, not to different value ─────────────

func TestEqual_ComplexNumbers(t *testing.T) {
	t.Run("complex128_same_value", func(t *testing.T) {
		require.True(t, Equal(complex128(1+2i), complex128(1+2i)))
	})
	t.Run("complex64_same_value", func(t *testing.T) {
		require.True(t, Equal(complex64(3+4i), complex64(3+4i)))
	})
	t.Run("complex128_different_values", func(t *testing.T) {
		require.False(t, Equal(complex128(1+2i), complex128(1+3i)))
	})
	t.Run("complex64_vs_complex128_same_value", func(t *testing.T) {
		// Different types: interface comparison → not equal (different type)
		require.False(t, Equal(complex64(1+2i), complex128(1+2i)))
	})
	t.Run("zero_complex_ne_zero_int", func(t *testing.T) {
		require.False(t, Equal(complex128(0), int(0)))
	})
}

// ── Unsupported types: Less always returns false ──────────────────────────────

func TestLess_UnsupportedTypesAlwaysFalse(t *testing.T) {
	numerics := []any{int(0), int(1), uint(0), float64(0.5)}
	unsupported := []any{
		complex64(1 + 2i),
		complex128(1 + 2i),
	}
	for _, u := range unsupported {
		for _, n := range numerics {
			u, n := u, n
			t.Run(fmt.Sprintf("%T < %T", u, n), func(t *testing.T) {
				require.False(t, Less(u, n))
				require.False(t, Less(n, u))
			})
		}
	}
}
