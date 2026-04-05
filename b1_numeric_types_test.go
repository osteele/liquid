// B1 integration tests — Go numeric type normalization via template rendering.
//
// All tests use ParseAndRenderString so they exercise the full path:
// variable binding → expression evaluator → comparison → output.
package liquid

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func renderNumeric(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	engine := NewEngine()
	out, err := engine.ParseAndRenderString(tpl, bindings)
	require.NoErrorf(t, err, "template: %s | bindings: %v", tpl, bindings)
	return out
}

// ── B1: truthiness — every non-zero numeric is truthy ────────────────────────

func TestB1_AllNumericNonZeroAreTruthy(t *testing.T) {
	nonZeros := []any{
		int(1), int8(1), int16(1), int32(1), int64(1),
		uint(1), uint8(1), uint16(1), uint32(1), uint64(1), uintptr(1),
		float32(1), float64(1),
	}
	for _, v := range nonZeros {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderNumeric(t, `{% if x %}yes{% endif %}`, map[string]any{"x": v})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: zero values — zero should still be truthy (Liquid: only nil/false are falsy) ──

func TestB1_AllNumericZerosAreTruthy(t *testing.T) {
	zeros := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0), uintptr(0),
		float32(0), float64(0),
	}
	for _, v := range zeros {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderNumeric(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(0) should be truthy in Liquid", v)
		})
	}
}

// ── B1: equality — x == 0 for all zero-valued numeric types ──────────────────

func TestB1_NumericEquality_ZeroComparisons(t *testing.T) {
	zeros := []any{
		int(0), int8(0), int16(0), int32(0), int64(0),
		uint(0), uint8(0), uint16(0), uint32(0), uint64(0), uintptr(0),
		float32(0), float64(0),
	}
	for _, v := range zeros {
		v := v
		t.Run(fmt.Sprintf("%T_eq_intlit", v), func(t *testing.T) {
			out := renderNumeric(t, `{% if x == 0 %}yes{% else %}no{% endif %}`, map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(0) should == literal 0", v)
		})
	}
}

// ── B1: inequality — x != 0 for all non-zero types ───────────────────────────

func TestB1_NumericInequality_NonZeroVsZero(t *testing.T) {
	nonZeros := []any{
		int(5), int8(5), int16(5), int32(5), int64(5),
		uint(5), uint8(5), uint16(5), uint32(5), uint64(5), uintptr(5),
		float32(5), float64(5),
	}
	for _, v := range nonZeros {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderNumeric(t, `{% if x != 0 %}yes{% else %}no{% endif %}`, map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(5) should satisfy != 0", v)
		})
	}
}

// ── B1: cross-type equality — two variables of different numeric types ────────

func TestB1_CrossTypeEquality(t *testing.T) {
	type pair struct{ a, b any }
	equalPairs := []pair{
		{int(10), uint(10)},
		{int8(10), uint8(10)},
		{int16(10), uint16(10)},
		{int32(10), uint32(10)},
		{int64(10), uint64(10)},
		{int(10), float64(10)},
		{uint(10), float64(10)},
		{uint64(10), float64(10)},
		{float32(10), int64(10)},
		{float32(10), uint64(10)},
		{uint(0), int(0)},
		{uint64(0), float64(0)},
	}
	for _, p := range equalPairs {
		p := p
		name := fmt.Sprintf("%T(%v) == %T(%v)", p.a, p.a, p.b, p.b)
		t.Run(name, func(t *testing.T) {
			out := renderNumeric(t, `{% if a == b %}yes{% else %}no{% endif %}`,
				map[string]any{"a": p.a, "b": p.b})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: ordering — less-than across type combinations ────────────────────────

func TestB1_CrossTypeLessThan(t *testing.T) {
	type testCase struct {
		desc string
		a, b any
	}
	cases := []testCase{
		{"int < uint", int(3), uint(5)},
		{"uint < int", uint(3), int(5)},
		{"int8 < uint16", int8(3), uint16(5)},
		{"uint32 < int64", uint32(3), int64(5)},
		{"uint64 < float64", uint64(3), float64(5.5)},
		{"float64 < uint64", float64(3.5), uint64(5)},
		{"int < float32", int(3), float32(5.5)},
		{"uint < float32", uint(3), float32(5.5)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			out := renderNumeric(t, `{% if a < b %}yes{% else %}no{% endif %}`,
				map[string]any{"a": tc.a, "b": tc.b})
			require.Equal(t, "yes", out)
		})
	}
}

func TestB1_CrossTypeGreaterThan(t *testing.T) {
	type testCase struct {
		desc string
		a, b any
	}
	cases := []testCase{
		{"uint > int", uint(5), int(3)},
		{"int > uint", int(5), uint(3)},
		{"uint64 > int32", uint64(5), int32(3)},
		{"float64 > uint64", float64(5.5), uint64(3)},
		{"uint64 > float64", uint64(5), float64(3.5)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			out := renderNumeric(t, `{% if a > b %}yes{% else %}no{% endif %}`,
				map[string]any{"a": tc.a, "b": tc.b})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: ordering — less-or-equal, greater-or-equal ───────────────────────────

func TestB1_CrossTypeLessOrEqual(t *testing.T) {
	type testCase struct{ a, b any }
	cases := []testCase{
		{uint(5), int(5)},
		{int(5), uint(5)},
		{uint64(5), float64(5.0)},
		{uint(3), int(5)},
		{int(3), uint(5)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T(%v) <= %T(%v)", tc.a, tc.a, tc.b, tc.b), func(t *testing.T) {
			out := renderNumeric(t, `{% if a <= b %}yes{% else %}no{% endif %}`,
				map[string]any{"a": tc.a, "b": tc.b})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: negative int is less than any uint ───────────────────────────────────

func TestB1_NegativeIntLessThanUint(t *testing.T) {
	// Negative integers should always compare as less than any unsigned value.
	negVsUints := []struct{ neg, pos any }{
		{int(-1), uint(0)},
		{int8(-1), uint8(0)},
		{int16(-1), uint16(0)},
		{int32(-1), uint32(0)},
		{int64(-1), uint64(0)},
		{int(-5), uint(100)},
		{int64(-1), uint64(math.MaxUint64 / 2)},
	}
	for _, tc := range negVsUints {
		tc := tc
		t.Run(fmt.Sprintf("%T(%v) < %T(%v)", tc.neg, tc.neg, tc.pos, tc.pos), func(t *testing.T) {
			out := renderNumeric(t, `{% if neg < pos %}yes{% else %}no{% endif %}`,
				map[string]any{"neg": tc.neg, "pos": tc.pos})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: large uint64 past int64 boundary ─────────────────────────────────────

func TestB1_LargeUint64Comparisons(t *testing.T) {
	big := uint64(math.MaxInt64) + 1

	t.Run("large_uint64_ne_maxint64", func(t *testing.T) {
		out := renderNumeric(t, `{% if a != b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": big, "b": int64(math.MaxInt64)})
		require.Equal(t, "yes", out)
	})

	t.Run("large_uint64_gt_maxint64", func(t *testing.T) {
		out := renderNumeric(t, `{% if a > b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": big, "b": int64(math.MaxInt64)})
		require.Equal(t, "yes", out)
	})

	t.Run("maxint64_lt_large_uint64", func(t *testing.T) {
		out := renderNumeric(t, `{% if b < a %}yes{% else %}no{% endif %}`,
			map[string]any{"a": big, "b": int64(math.MaxInt64)})
		require.Equal(t, "yes", out)
	})
}

// ── B1: filter arithmetic on uint types ──────────────────────────────────────

func TestB1_FilterArithmeticOnUintTypes(t *testing.T) {
	type testCase struct {
		tpl      string
		x        any
		expected string
	}
	cases := []testCase{
		{`{{ x | plus: 1 }}`, uint(5), "6"},
		{`{{ x | minus: 1 }}`, uint(5), "4"},
		{`{{ x | times: 2 }}`, uint(5), "10"},
		{`{{ x | divided_by: 2 }}`, uint(10), "5"},
		{`{{ x | modulo: 3 }}`, uint(10), "1"},
		{`{{ x | plus: 1 }}`, uint8(5), "6"},
		{`{{ x | plus: 1 }}`, uint16(5), "6"},
		{`{{ x | plus: 1 }}`, uint32(5), "6"},
		{`{{ x | plus: 1 }}`, uint64(5), "6"},
		{`{{ x | plus: 1 }}`, uintptr(5), "6"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T_%s", tc.x, tc.tpl), func(t *testing.T) {
			out := renderNumeric(t, tc.tpl, map[string]any{"x": tc.x})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: output — all numeric types render as their value ─────────────────────

func TestB1_AllNumericTypesRenderCorrectly(t *testing.T) {
	cases := []struct {
		v        any
		expected string
	}{
		{int(42), "42"},
		{int8(42), "42"},
		{int16(42), "42"},
		{int32(42), "42"},
		{int64(42), "42"},
		{uint(42), "42"},
		{uint8(42), "42"},
		{uint16(42), "42"},
		{uint32(42), "42"},
		{uint64(42), "42"},
		{uintptr(42), "42"},
		{float32(3.5), "3.5"},
		{float64(3.5), "3.5"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T", tc.v), func(t *testing.T) {
			out := renderNumeric(t, `{{ x }}`, map[string]any{"x": tc.v})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: assign preserves numeric type used in comparison ─────────────────────

func TestB1_AssignAndCompareUintVariable(t *testing.T) {
	cases := []struct {
		x        any
		expected string
	}{
		{uint(5), "greater"},
		{uint64(5), "greater"},
		{uint32(5), "greater"},
		{int(-1), "less"},
		{uint(0), "equal"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T(%v)", tc.x, tc.x), func(t *testing.T) {
			tpl := `{% if x > 3 %}greater{% elsif x == 0 %}equal{% else %}less{% endif %}`
			out := renderNumeric(t, tpl, map[string]any{"x": tc.x})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: >= operator cross types ───────────────────────────────────────────────

func TestB1_CrossTypeGreaterOrEqual(t *testing.T) {
	type testCase struct{ a, b any }
	// a >= b must be true in all cases
	cases := []testCase{
		{uint(5), int(5)}, // equal, different sign
		{int(5), uint(5)}, // symmetric
		{uint(5), int(3)}, // strictly greater
		{int(5), uint(3)},
		{uint64(5), float64(5.0)},
		{float64(5.0), uint64(5)},
		{uint64(10), int64(10)},
		{uint32(10), int32(10)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T(%v) >= %T(%v)", tc.a, tc.a, tc.b, tc.b), func(t *testing.T) {
			out := renderNumeric(t, `{% if a >= b %}yes{% else %}no{% endif %}`,
				map[string]any{"a": tc.a, "b": tc.b})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: unless tag with uint condition ───────────────────────────────────────

func TestB1_UnlessWithUintTypes(t *testing.T) {
	// unless x == 0 → body executes when x != 0
	nonZeros := []any{
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5), uintptr(6),
	}
	for _, v := range nonZeros {
		v := v
		t.Run(fmt.Sprintf("%T(%v)", v, v), func(t *testing.T) {
			out := renderNumeric(t,
				`{% unless x == 0 %}yes{% else %}no{% endunless %}`,
				map[string]any{"x": v})
			require.Equal(t, "yes", out)
		})
	}
	// unless x != 0 → body executes when x == 0 (all zero uints)
	zeros := []any{uint(0), uint8(0), uint16(0), uint32(0), uint64(0), uintptr(0)}
	for _, v := range zeros {
		v := v
		t.Run(fmt.Sprintf("%T(0)_is_zero", v), func(t *testing.T) {
			out := renderNumeric(t,
				`{% unless x != 0 %}zero{% else %}nonzero{% endunless %}`,
				map[string]any{"x": v})
			require.Equal(t, "zero", out)
		})
	}
}

// ── B1: case/when with uint types ────────────────────────────────────────────

func TestB1_CaseWhenWithUintTypes(t *testing.T) {
	cases := []struct {
		x        any
		expected string
	}{
		{uint(1), "one"},
		{uint8(2), "two"},
		{uint16(3), "three"},
		{uint32(1), "one"},
		{uint64(2), "two"},
		{uintptr(3), "three"},
		// cross-type: int literal in when, uint variable
		{uint(0), "other"},
	}
	tpl := `{% case x %}{% when 1 %}one{% when 2 %}two{% when 3 %}three{% else %}other{% endcase %}`
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T(%v)", tc.x, tc.x), func(t *testing.T) {
			out := renderNumeric(t, tpl, map[string]any{"x": tc.x})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: compound and/or conditions with mixed numeric types ──────────────────

func TestB1_CompoundConditionsWithMixedTypes(t *testing.T) {
	tests := []struct {
		desc     string
		tpl      string
		bindings map[string]any
		expected string
	}{
		{
			"uint and uint both true",
			`{% if a > 0 and b > 0 %}yes{% else %}no{% endif %}`,
			map[string]any{"a": uint(3), "b": uint(4)},
			"yes",
		},
		{
			"uint and int mixed",
			`{% if a > 0 and b < 10 %}yes{% else %}no{% endif %}`,
			map[string]any{"a": uint(3), "b": int(5)},
			"yes",
		},
		{
			"uint or int, first false",
			`{% if a == 0 or b == 5 %}yes{% else %}no{% endif %}`,
			map[string]any{"a": uint(3), "b": int(5)},
			"yes",
		},
		{
			"int < 0 and uint > 0",
			`{% if a < 0 and b > 0 %}yes{% else %}no{% endif %}`,
			map[string]any{"a": int(-1), "b": uint(1)},
			"yes",
		},
		{
			"float and uint comparison",
			`{% if a >= 3.0 and b <= 10 %}yes{% else %}no{% endif %}`,
			map[string]any{"a": float64(3.5), "b": uint(8)},
			"yes",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			out := renderNumeric(t, tc.tpl, tc.bindings)
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: struct fields with uint types ────────────────────────────────────────

type numericStruct struct {
	Count    uint
	Count8   uint8
	Count16  uint16
	Count32  uint32
	Count64  uint64
	UintPtr  uintptr
	Signed   int32
	BigFloat float32
}

func TestB1_StructFieldsWithUintTypes(t *testing.T) {
	data := numericStruct{
		Count:    10,
		Count8:   10,
		Count16:  10,
		Count32:  10,
		Count64:  10,
		UintPtr:  10,
		Signed:   10,
		BigFloat: 10.0,
	}

	fieldTests := []struct {
		field string
		tplEq string
		tplGt string
	}{
		{"Count", `{% if obj.Count == 10 %}yes{% else %}no{% endif %}`, `{% if obj.Count > 5 %}yes{% else %}no{% endif %}`},
		{"Count8", `{% if obj.Count8 == 10 %}yes{% else %}no{% endif %}`, `{% if obj.Count8 > 5 %}yes{% else %}no{% endif %}`},
		{"Count16", `{% if obj.Count16 == 10 %}yes{% else %}no{% endif %}`, `{% if obj.Count16 > 5 %}yes{% else %}no{% endif %}`},
		{"Count32", `{% if obj.Count32 == 10 %}yes{% else %}no{% endif %}`, `{% if obj.Count32 > 5 %}yes{% else %}no{% endif %}`},
		{"Count64", `{% if obj.Count64 == 10 %}yes{% else %}no{% endif %}`, `{% if obj.Count64 > 5 %}yes{% else %}no{% endif %}`},
		{"UintPtr", `{% if obj.UintPtr == 10 %}yes{% else %}no{% endif %}`, `{% if obj.UintPtr > 5 %}yes{% else %}no{% endif %}`},
	}
	for _, tc := range fieldTests {
		tc := tc
		t.Run(tc.field+"_eq", func(t *testing.T) {
			out := renderNumeric(t, tc.tplEq, map[string]any{"obj": data})
			require.Equal(t, "yes", out)
		})
		t.Run(tc.field+"_gt", func(t *testing.T) {
			out := renderNumeric(t, tc.tplGt, map[string]any{"obj": data})
			require.Equal(t, "yes", out)
		})
	}
}

// ── B1: struct fields cross-type comparison ──────────────────────────────────

func TestB1_StructFieldCrossTypeComparison(t *testing.T) {
	data := numericStruct{Count: 10, Signed: 10, BigFloat: 10.0}
	// uint struct field == int struct field (both 10)
	t.Run("uint_Count_eq_int_Signed", func(t *testing.T) {
		out := renderNumeric(t,
			`{% if obj.Count == obj.Signed %}yes{% else %}no{% endif %}`,
			map[string]any{"obj": data})
		require.Equal(t, "yes", out)
	})
	// uint struct field == float struct field
	t.Run("uint_Count_eq_float_BigFloat", func(t *testing.T) {
		out := renderNumeric(t,
			`{% if obj.Count == obj.BigFloat %}yes{% else %}no{% endif %}`,
			map[string]any{"obj": data})
		require.Equal(t, "yes", out)
	})
}

// ── B1: math filters on all integer types ────────────────────────────────────

func TestB1_MathFiltersOnAllIntTypes(t *testing.T) {
	type filterCase struct {
		tpl      string
		x        any
		expected string
	}
	cases := []filterCase{
		// abs — negative int
		{`{{ x | abs }}`, int8(-5), "5"},
		{`{{ x | abs }}`, int16(-100), "100"},
		{`{{ x | abs }}`, int32(-1), "1"},
		{`{{ x | abs }}`, int64(-42), "42"},
		// abs — unsigned (no-op)
		{`{{ x | abs }}`, uint(5), "5"},
		{`{{ x | abs }}`, uint32(5), "5"},
		// at_least
		{`{{ x | at_least: 3 }}`, uint(1), "3"},
		{`{{ x | at_least: 3 }}`, uint8(1), "3"},
		{`{{ x | at_least: 3 }}`, uint32(10), "10"},
		{`{{ x | at_least: 3 }}`, int(1), "3"},
		{`{{ x | at_least: 3 }}`, int8(10), "10"},
		// at_most
		{`{{ x | at_most: 7 }}`, uint(10), "7"},
		{`{{ x | at_most: 7 }}`, uint16(10), "7"},
		{`{{ x | at_most: 7 }}`, uint32(3), "3"},
		{`{{ x | at_most: 7 }}`, int(10), "7"},
		// ceil / floor / round on uint (should pass through as-is)
		{`{{ x | ceil }}`, uint(5), "5"},
		{`{{ x | floor }}`, uint(5), "5"},
		{`{{ x | round }}`, uint(5), "5"},
		// ceil / floor on float holding integer value
		{`{{ x | ceil }}`, float64(5.0), "5"},
		{`{{ x | floor }}`, float64(5.0), "5"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T_%s", tc.x, tc.tpl), func(t *testing.T) {
			out := renderNumeric(t, tc.tpl, map[string]any{"x": tc.x})
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: chained arithmetic preserves value across mixed types ────────────────

func TestB1_ChainedFiltersMixedTypes(t *testing.T) {
	type testCase struct {
		desc     string
		tpl      string
		bindings map[string]any
		expected string
	}
	cases := []testCase{
		{
			"uint plus int literal chain",
			`{{ x | plus: 2 | plus: 3 }}`,
			map[string]any{"x": uint(10)},
			"15",
		},
		{
			"uint times then minus",
			`{{ x | times: 3 | minus: 1 }}`,
			map[string]any{"x": uint(5)},
			"14",
		},
		{
			"int32 chain through at_least and at_most",
			`{{ x | at_least: 5 | at_most: 20 }}`,
			map[string]any{"x": int32(3)},
			"5",
		},
		{
			"uint64 chain at_least at_most",
			`{{ x | at_least: 5 | at_most: 20 }}`,
			map[string]any{"x": uint64(30)},
			"20",
		},
		{
			"uint plus float variable",
			`{{ x | plus: y }}`,
			map[string]any{"x": uint(3), "y": float64(2.5)},
			"5.5",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			out := renderNumeric(t, tc.tpl, tc.bindings)
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── B1: sort filter on mixed-type numeric array ───────────────────────────────

func TestB1_SortFilterMixedNumericArray(t *testing.T) {
	// All-uint array sorts correctly
	t.Run("all_uint", func(t *testing.T) {
		out := renderNumeric(t,
			`{% assign sorted = arr | sort %}{{ sorted | join: "," }}`,
			map[string]any{"arr": []any{uint(3), uint(1), uint(2)}})
		require.Equal(t, "1,2,3", out)
	})
	// Mixed int/uint sorts correctly
	t.Run("mixed_int_uint", func(t *testing.T) {
		out := renderNumeric(t,
			`{% assign sorted = arr | sort %}{{ sorted | join: "," }}`,
			map[string]any{"arr": []any{int(3), uint(1), int32(2)}})
		require.Equal(t, "1,2,3", out)
	})
}

// ── B1: where filter matching uint values ─────────────────────────────────────

func TestB1_WhereFilterWithUintValues(t *testing.T) {
	products := []map[string]any{
		{"name": "a", "qty": uint(0)},
		{"name": "b", "qty": uint(5)},
		{"name": "c", "qty": uint32(5)},
		{"name": "d", "qty": int(5)},
	}
	// where qty == 5 should match b, c, d regardless of uint/int type
	out := renderNumeric(t,
		`{% assign found = products | where: "qty", 5 %}{{ found | map: "name" | join: "," }}`,
		map[string]any{"products": products})
	require.Equal(t, "b,c,d", out)
}

// ── B1: array indexed by uint variable ───────────────────────────────────────

func TestB1_ArrayIndexWithUintVariable(t *testing.T) {
	arr := []string{"zero", "one", "two", "three"}
	idxTypes := []any{uint(2), uint8(2), uint16(2), uint32(2), uint64(2), uintptr(2)}
	for _, idx := range idxTypes {
		idx := idx
		t.Run(fmt.Sprintf("%T", idx), func(t *testing.T) {
			out := renderNumeric(t,
				`{{ arr[i] }}`,
				map[string]any{"arr": arr, "i": idx})
			require.Equal(t, "two", out)
		})
	}
}

// ── B1: assign + re-compare numeric literals vs variables ────────────────────

func TestB1_AssignThenCompare(t *testing.T) {
	// assign from a uint binding, then compare the assigned var to a literal
	t.Run("uint_assign_eq_literal", func(t *testing.T) {
		out := renderNumeric(t,
			`{% assign n = x %}{% if n == 42 %}yes{% else %}no{% endif %}`,
			map[string]any{"x": uint(42)})
		require.Equal(t, "yes", out)
	})
	t.Run("uint_assign_gt_literal", func(t *testing.T) {
		out := renderNumeric(t,
			`{% assign n = x %}{% if n > 10 %}yes{% else %}no{% endif %}`,
			map[string]any{"x": uint32(100)})
		require.Equal(t, "yes", out)
	})
	t.Run("int_assign_compare_uint_var", func(t *testing.T) {
		out := renderNumeric(t,
			`{% assign a = x %}{% assign b = y %}{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"x": int(7), "y": uint(7)})
		require.Equal(t, "yes", out)
	})
}

// ── B1: for loop with uint limit/offset ──────────────────────────────────────

func TestB1_ForLoopWithUintLimitOffset(t *testing.T) {
	arr := []int{10, 20, 30, 40, 50}

	t.Run("uint_limit", func(t *testing.T) {
		out := renderNumeric(t,
			`{% for item in arr limit: lim %}{{ item }},{% endfor %}`,
			map[string]any{"arr": arr, "lim": uint(3)})
		require.Equal(t, "10,20,30,", out)
	})
	t.Run("uint_offset", func(t *testing.T) {
		out := renderNumeric(t,
			`{% for item in arr offset: off %}{{ item }},{% endfor %}`,
			map[string]any{"arr": arr, "off": uint(2)})
		require.Equal(t, "30,40,50,", out)
	})
	t.Run("uint_limit_and_offset", func(t *testing.T) {
		out := renderNumeric(t,
			`{% for item in arr limit: lim offset: off %}{{ item }},{% endfor %}`,
			map[string]any{"arr": arr, "lim": uint(2), "off": uint(1)})
		require.Equal(t, "20,30,", out)
	})
}

// ── B1: float precision in comparisons ───────────────────────────────────────

func TestB1_FloatPrecisionComparisons(t *testing.T) {
	tests := []struct {
		desc     string
		tpl      string
		bindings map[string]any
		expected string
	}{
		{
			"float32 exact vs float64 same value",
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": float32(1.5), "b": float64(float32(1.5))},
			"yes",
		},
		{
			"float64 1.0 == uint 1",
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": float64(1.0), "b": uint(1)},
			"yes",
		},
		{
			"uint 1 == float64 1.0",
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": uint(1), "b": float64(1.0)},
			"yes",
		},
		{
			"float 1.5 != uint 1",
			`{% if a != b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": float64(1.5), "b": uint(1)},
			"yes",
		},
		{
			"float result of divided_by compared to uint",
			`{% assign r = x | divided_by: 2.0 %}{% if r == y %}yes{% else %}no{% endif %}`,
			map[string]any{"x": uint(4), "y": float64(2.0)},
			"yes",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			out := renderNumeric(t, tc.tpl, tc.bindings)
			require.Equal(t, tc.expected, out)
		})
	}
}

// ── helpers for unsupported-type tests ───────────────────────────────────────

// renderUnsafeNoError renders a template that must not return an error and must
// not panic. It returns the output string for optional assertions.
func renderUnsafeNoError(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	engine := NewEngine()
	var out string
	var err error
	require.NotPanics(t, func() {
		out, err = engine.ParseAndRenderString(tpl, bindings)
	}, "template %q with bindings %v panicked", tpl, bindings)
	require.NoError(t, err, "template %q should not error", tpl)
	return out
}

// ── B1-unsupported: chan, func, complex do not panic in templates ─────────────
//
// Liquid does not support these Go types, but when passed as bindings they
// must not cause panics or unexpected crashes. The behaviour is best-effort:
//   - rendering {{ x }}: some string representation is written (no panic)
//   - truthiness: non-nil chan/func/complex are truthy (Liquid: only nil and
//     false are falsy)
//   - comparisons with numerics: no panic

func TestB1_UnsupportedTypes_RenderDoesNotPanic(t *testing.T) {
	ch := make(chan int)
	fn := func() int { return 42 }

	cases := []struct {
		name string
		x    any
	}{
		{"complex64", complex64(1 + 2i)},
		{"complex128", complex128(3 + 4i)},
		{"zero_complex128", complex128(0)},
		{"chan_int", ch},
		{"func", fn},
	}
	templates := []string{
		`{{ x }}`,
		`{% if x %}yes{% else %}no{% endif %}`,
		`{% assign v = x %}{{ v }}`,
	}
	for _, tc := range cases {
		for _, tpl := range templates {
			tc, tpl := tc, tpl
			t.Run(tc.name+"/"+tpl, func(t *testing.T) {
				renderUnsafeNoError(t, tpl, map[string]any{"x": tc.x})
			})
		}
	}
}

// ── B1-unsupported: non-nil chan / func / complex are truthy ──────────────────

func TestB1_UnsupportedTypes_AreTruthy(t *testing.T) {
	ch := make(chan int)
	fn := func() {}

	truthy := []struct {
		name string
		x    any
	}{
		{"complex64_nonzero", complex64(1 + 2i)},
		{"complex128_nonzero", complex128(1 + 2i)},
		// zero complex is still truthy — Liquid only treats nil and false as falsy
		{"complex128_zero", complex128(0)},
		{"chan_int", ch},
		{"func", fn},
	}
	for _, tc := range truthy {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out := renderUnsafeNoError(t,
				`{% if x %}yes{% else %}no{% endif %}`,
				map[string]any{"x": tc.x})
			require.Equal(t, "yes", out,
				"%T(%v) should be truthy in Liquid (only nil and false are falsy)", tc.x, tc.x)
		})
	}
}

// ── B1-unsupported: comparisons with numeric literals do not panic ────────────

func TestB1_UnsupportedTypes_ComparisonWithNumericDoesNotPanic(t *testing.T) {
	ch := make(chan int)
	fn := func() {}

	cases := []struct {
		name string
		x    any
	}{
		{"complex64", complex64(1 + 2i)},
		{"complex128", complex128(1 + 2i)},
		{"zero_complex128", complex128(0)},
		{"chan_int", ch},
		{"func", fn},
	}
	// All these must not panic; result is "yes" or "no" depending on semantics.
	compareTemplates := []string{
		`{% if x == 0 %}yes{% else %}no{% endif %}`,
		`{% if x != 0 %}yes{% else %}no{% endif %}`,
		`{% if x < 1 %}yes{% else %}no{% endif %}`,
		`{% if x > 0 %}yes{% else %}no{% endif %}`,
	}
	for _, tc := range cases {
		for _, tpl := range compareTemplates {
			tc, tpl := tc, tpl
			t.Run(tc.name+"/"+tpl, func(t *testing.T) {
				require.NotPanics(t, func() {
					engine := NewEngine()
					_, _ = engine.ParseAndRenderString(tpl, map[string]any{"x": tc.x})
				})
			})
		}
	}
}

// ── B1-unsupported: complex == complex comparison is self-consistent ──────────

func TestB1_ComplexSelfEquality(t *testing.T) {
	// complex numbers are Go-comparable — two interface values of the same
	// complex type with the same value should be equal.
	t.Run("complex128_eq_itself_via_var", func(t *testing.T) {
		v := complex128(1 + 2i)
		out := renderUnsafeNoError(t,
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": v, "b": v})
		require.Equal(t, "yes", out)
	})
	t.Run("complex128_ne_different_value", func(t *testing.T) {
		out := renderUnsafeNoError(t,
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": complex128(1 + 2i), "b": complex128(1 + 3i)})
		require.Equal(t, "no", out)
	})
	// zero complex is NOT equal to integer 0 (different types in interface comparison)
	t.Run("zero_complex_ne_int_zero", func(t *testing.T) {
		out := renderUnsafeNoError(t,
			`{% if a == b %}yes{% else %}no{% endif %}`,
			map[string]any{"a": complex128(0), "b": int(0)})
		require.Equal(t, "no", out)
	})
}

// ── B1-unsupported: rendering complex numbers produces expected format ────────

func TestB1_ComplexRendering(t *testing.T) {
	t.Run("complex128_output", func(t *testing.T) {
		out := renderUnsafeNoError(t, `{{ x }}`, map[string]any{"x": complex128(1 + 2i)})
		// fmt.Sprint(complex128(1+2i)) == "(1+2i)"
		require.Equal(t, "(1+2i)", out)
	})
	t.Run("complex64_output", func(t *testing.T) {
		out := renderUnsafeNoError(t, `{{ x }}`, map[string]any{"x": complex64(3 + 4i)})
		require.Equal(t, "(3+4i)", out)
	})
	t.Run("zero_complex128_output", func(t *testing.T) {
		out := renderUnsafeNoError(t, `{{ x }}`, map[string]any{"x": complex128(0)})
		require.Equal(t, "(0+0i)", out)
	})
}

// ── B1-unsupported: func rendering produces non-empty output ─────────────────

func TestB1_FuncRendering_NonEmpty(t *testing.T) {
	fn := func() {}
	out := renderUnsafeNoError(t, `{{ x }}`, map[string]any{"x": fn})
	// fmt.Sprint on a func prints its address (e.g. "0x12345678")
	require.NotEmpty(t, out, "rendering a func should produce a non-empty string")
}

// ── B1-unsupported: chan rendering produces non-empty output ──────────────────

func TestB1_ChanRendering_NonEmpty(t *testing.T) {
	ch := make(chan int)
	out := renderUnsafeNoError(t, `{{ x }}`, map[string]any{"x": ch})
	require.NotEmpty(t, out, "rendering a chan should produce a non-empty string")
}
