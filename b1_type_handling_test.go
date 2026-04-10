// b1_type_handling_test.go — e2e type system tests: defined types, aliases,
// invalid kinds, and rendering consistency across the full template pipeline.
//
// Covers paths: binding → ValueOf → expression evaluator → writeObject →
// string output, truthiness, comparisons.
package liquid

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ── defined types ─────────────────────────────────────────────────────────────

type e2eMyInt int
type e2eMyInt8 int8
type e2eMyInt16 int16
type e2eMyInt32 int32
type e2eMyInt64 int64
type e2eMyUint uint
type e2eMyUint8 uint8
type e2eMyUint16 uint16
type e2eMyUint32 uint32
type e2eMyUint64 uint64
type e2eMyUintptr uintptr
type e2eMyFloat32 float32
type e2eMyFloat64 float64
type e2eMyBool bool
type e2eMyString string
type e2eMySlice []int

// ── type aliases ──────────────────────────────────────────────────────────────

type e2eAliasInt = int
type e2eAliasFloat64 = float64
type e2eAliasBool = bool
type e2eAliasString = string

// ── helper ────────────────────────────────────────────────────────────────────

func renderType(t *testing.T, tpl string, bindings map[string]any) string {
	t.Helper()
	eng := NewEngine()
	out, err := eng.ParseAndRenderString(tpl, bindings)
	require.NoError(t, err)
	return out
}

// ════════════════════════════════════════════════════════════════════════════
// Output rendering: defined numeric types must produce the same string as the
// corresponding built-in type, including large exponent values (no sci notation)
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedInt_Output(t *testing.T) {
	cases := []any{
		e2eMyInt(42), e2eMyInt8(42), e2eMyInt16(42),
		e2eMyInt32(42), e2eMyInt64(42),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderType(t, `{{ x }}`, map[string]any{"x": v})
			require.Equal(t, "42", out)
		})
	}
}

func TestTypeHandling_DefinedUint_Output(t *testing.T) {
	cases := []any{
		e2eMyUint(99), e2eMyUint8(99), e2eMyUint16(99),
		e2eMyUint32(99), e2eMyUint64(99), e2eMyUintptr(99),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderType(t, `{{ x }}`, map[string]any{"x": v})
			require.Equal(t, "99", out)
		})
	}
}

func TestTypeHandling_DefinedFloat_Output(t *testing.T) {
	t.Run("e2eMyFloat32 simple", func(t *testing.T) {
		out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyFloat32(3.14)})
		require.Equal(t, "3.14", out)
	})
	t.Run("e2eMyFloat64 simple", func(t *testing.T) {
		out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyFloat64(3.14)})
		require.Equal(t, "3.14", out)
	})
	// Large value: must use fixed notation, not scientific notation.
	t.Run("float32 large no scientific notation", func(t *testing.T) {
		out := renderType(t, `{{ x }}`, map[string]any{"x": float32(1e7)})
		require.Equal(t, "10000000", out)
	})
	t.Run("e2eMyFloat64 large no scientific notation", func(t *testing.T) {
		out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyFloat64(1e10)})
		require.Equal(t, "10000000000", out)
	})
	t.Run("e2eMyFloat32 large no scientific notation", func(t *testing.T) {
		out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyFloat32(1e7)})
		require.Equal(t, "10000000", out)
	})
}

func TestTypeHandling_DefinedString_Output(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyString("hello")})
	require.Equal(t, "hello", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Truthiness: defined bool types (MyBool(false) must be falsy)
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedBool_Truthiness(t *testing.T) {
	t.Run("MyBool(false) is falsy", func(t *testing.T) {
		out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`,
			map[string]any{"x": e2eMyBool(false)})
		require.Equal(t, "no", out)
	})
	t.Run("MyBool(true) is truthy", func(t *testing.T) {
		out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`,
			map[string]any{"x": e2eMyBool(true)})
		require.Equal(t, "yes", out)
	})
	t.Run("alias bool(false) is falsy", func(t *testing.T) {
		var v e2eAliasBool = false
		out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`,
			map[string]any{"x": v})
		require.Equal(t, "no", out)
	})
}

// ════════════════════════════════════════════════════════════════════════════
// Comparison: defined types must compare equal to literals and each other
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedInt_Equality(t *testing.T) {
	cases := []any{
		e2eMyInt(10), e2eMyInt8(10), e2eMyInt16(10),
		e2eMyInt32(10), e2eMyInt64(10),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderType(t, `{% if x == 10 %}yes{% else %}no{% endif %}`,
				map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(10) should == literal 10", v)
		})
	}
}

func TestTypeHandling_DefinedUint_Equality(t *testing.T) {
	cases := []any{
		e2eMyUint(10), e2eMyUint8(10), e2eMyUint16(10),
		e2eMyUint32(10), e2eMyUint64(10), e2eMyUintptr(10),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderType(t, `{% if x == 10 %}yes{% else %}no{% endif %}`,
				map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(10) should == literal 10", v)
		})
	}
}

func TestTypeHandling_DefinedFloat_Equality(t *testing.T) {
	cases := []any{
		e2eMyFloat32(10), e2eMyFloat64(10),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			out := renderType(t, `{% if x == 10 %}yes{% else %}no{% endif %}`,
				map[string]any{"x": v})
			require.Equal(t, "yes", out, "%T(10) should == literal 10", v)
		})
	}
}

func TestTypeHandling_DefinedString_Equality(t *testing.T) {
	out := renderType(t, `{% if x == "hello" %}yes{% else %}no{% endif %}`,
		map[string]any{"x": e2eMyString("hello")})
	require.Equal(t, "yes", out)
}

func TestTypeHandling_DefinedTypes_CrossComparison(t *testing.T) {
	// Two defined types with same underlying value must be equal cross-kind.
	out := renderType(t, `{% if a == b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": e2eMyInt32(7), "b": e2eMyUint64(7)})
	require.Equal(t, "yes", out)

	out = renderType(t, `{% if a == b %}yes{% else %}no{% endif %}`,
		map[string]any{"a": e2eMyFloat32(5), "b": e2eMyInt64(5)})
	require.Equal(t, "yes", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Type aliases: transparent at runtime — same behaviour as base types
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_AliasInt_Output(t *testing.T) {
	var v e2eAliasInt = 55
	out := renderType(t, `{{ x }}`, map[string]any{"x": v})
	require.Equal(t, "55", out)
}

func TestTypeHandling_AliasFloat64_Output(t *testing.T) {
	var v e2eAliasFloat64 = 2.5
	out := renderType(t, `{{ x }}`, map[string]any{"x": v})
	require.Equal(t, "2.5", out)
}

func TestTypeHandling_AliasString_Output(t *testing.T) {
	var v e2eAliasString = "world"
	out := renderType(t, `{{ x }}`, map[string]any{"x": v})
	require.Equal(t, "world", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Invalid kinds: chan, func, complex — must render as empty, no panic
// ════════════════════════════════════════════════════════════════════════════

// ════════════════════════════════════════════════════════════════════════════
// Invalid kinds: chan, func, complex — must return TypeError, never panic
// These Go kinds are not representable in Liquid templates.
// ════════════════════════════════════════════════════════════════════════════

func renderTypeError(t *testing.T, tpl string, bindings map[string]any) {
	t.Helper()
	eng := NewEngine()
	require.NotPanics(t, func() {
		_, err := eng.ParseAndRenderString(tpl, bindings)
		require.Error(t, err, "expected error for template %q", tpl)
	})
}

func TestTypeHandling_Chan_Errors_NoPanic(t *testing.T) {
	ch := make(chan int)
	defer close(ch)
	renderTypeError(t, `{{ x }}`, map[string]any{"x": ch})
}

func TestTypeHandling_NilChan_Errors_NoPanic(t *testing.T) {
	var ch chan int // typed nil — still invalid kind
	renderTypeError(t, `{{ x }}`, map[string]any{"x": ch})
}

func TestTypeHandling_Func_Errors_NoPanic(t *testing.T) {
	fn := func() string { return "secret" }
	renderTypeError(t, `{{ x }}`, map[string]any{"x": fn})
}

func TestTypeHandling_Complex64_Errors_NoPanic(t *testing.T) {
	renderTypeError(t, `{{ x }}`, map[string]any{"x": complex64(complex(1, 2))})
}

func TestTypeHandling_Complex128_Errors_NoPanic(t *testing.T) {
	renderTypeError(t, `{{ x }}`, map[string]any{"x": complex128(complex(3, 4))})
}

func TestTypeHandling_Chan_ConditionErrors(t *testing.T) {
	ch := make(chan int)
	defer close(ch)
	renderTypeError(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": ch})
}

func TestTypeHandling_Func_ConditionErrors(t *testing.T) {
	renderTypeError(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": func() {}})
}

func TestTypeHandling_Complex_ConditionErrors(t *testing.T) {
	renderTypeError(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": complex(1.0, 2.0)})
}

// ════════════════════════════════════════════════════════════════════════════
// Defined slice / map types must behave as arrays / hashes
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedSlice_IteratesNormally(t *testing.T) {
	s := e2eMySlice{1, 2, 3}
	out := renderType(t, `{% for v in x %}{{ v }}{% endfor %}`,
		map[string]any{"x": s})
	require.Equal(t, "123", out)
}

func TestTypeHandling_DefinedSlice_SizeFilter(t *testing.T) {
	s := e2eMySlice{10, 20}
	out := renderType(t, `{{ x.size }}`, map[string]any{"x": s})
	require.Equal(t, "2", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Defined numeric types in arithmetic filters (plus, minus, times, divided_by)
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedInt_Filters(t *testing.T) {
	out := renderType(t, `{{ x | plus: 3 }}`, map[string]any{"x": e2eMyInt(7)})
	require.Equal(t, "10", out)

	out = renderType(t, `{{ x | times: 2 }}`, map[string]any{"x": e2eMyInt32(4)})
	require.Equal(t, "8", out)
}

func TestTypeHandling_DefinedFloat_Filters(t *testing.T) {
	out := renderType(t, `{{ x | plus: 1.5 }}`, map[string]any{"x": e2eMyFloat64(1.5)})
	require.Equal(t, "3", out) // 1.5 + 1.5 = 3 (renders without trailing .0)
}

// ════════════════════════════════════════════════════════════════════════════
// Defined bool type in assign / capture
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedBool_InAssign(t *testing.T) {
	out := renderType(t,
		`{% assign v = x %}{% if v %}yes{% else %}no{% endif %}`,
		map[string]any{"x": e2eMyBool(false)})
	require.Equal(t, "no", out)
}

// ════════════════════════════════════════════════════════════════════════════
// NegativeInt defined types: map negative values to under-zero behaviour
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_DefinedNegativeInt_Output(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyInt(-5)})
	require.Equal(t, "-5", out)
}

func TestTypeHandling_DefinedNegativeInt_Comparison(t *testing.T) {
	out := renderType(t, `{% if x < 0 %}negative{% else %}non-negative{% endif %}`,
		map[string]any{"x": e2eMyInt(-1)})
	require.Equal(t, "negative", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Pointer dereferencing: *T pointers must transparently dereference in the
// engine, including through defined pointer types and double pointers.
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_PtrInt_Dereferences(t *testing.T) {
	n := 42
	out := renderType(t, `{{ x }}`, map[string]any{"x": &n})
	require.Equal(t, "42", out)
}

func TestTypeHandling_PtrString_Dereferences(t *testing.T) {
	s := "hello"
	out := renderType(t, `{{ x }}`, map[string]any{"x": &s})
	require.Equal(t, "hello", out)
}

func TestTypeHandling_NilPtr_RendersEmpty(t *testing.T) {
	var p *int
	out := renderType(t, `{{ x }}`, map[string]any{"x": p})
	require.Equal(t, "", out)
}

func TestTypeHandling_DoublePtrInt_Dereferences(t *testing.T) {
	n := 99
	p := &n
	out := renderType(t, `{{ x }}`, map[string]any{"x": &p})
	require.Equal(t, "99", out)
}

func TestTypeHandling_PtrDefinedInt_Dereferences(t *testing.T) {
	v := e2eMyInt(7)
	out := renderType(t, `{{ x | plus: 1 }}`, map[string]any{"x": &v})
	require.Equal(t, "8", out)
}

func TestTypeHandling_NilPtr_IsFalsy(t *testing.T) {
	var p *int
	out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": p})
	require.Equal(t, "no", out)
}

func TestTypeHandling_NonNilPtr_IsTruthy(t *testing.T) {
	n := 1
	out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": &n})
	require.Equal(t, "yes", out)
}

// ════════════════════════════════════════════════════════════════════════════
// []byte and [N]byte: must render as strings, not as numeric arrays.
// ════════════════════════════════════════════════════════════════════════════

type e2eMyBytes []byte

func TestTypeHandling_ByteSlice_RendersAsString(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": []byte("hello")})
	require.Equal(t, "hello", out)
}

func TestTypeHandling_ByteArray_RendersAsString(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": [5]byte{'w', 'o', 'r', 'l', 'd'}})
	require.Equal(t, "world", out)
}

func TestTypeHandling_DefinedByteSlice_RendersAsString(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": e2eMyBytes("liquid")})
	require.Equal(t, "liquid", out)
}

func TestTypeHandling_ByteSlice_IsTruthy(t *testing.T) {
	out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": []byte("hi")})
	require.Equal(t, "yes", out)
}

func TestTypeHandling_EmptyByteSlice_IsTruthy(t *testing.T) {
	// In Liquid spec, only nil and false are falsy; empty string is truthy.
	out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": []byte("")})
	require.Equal(t, "yes", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Numeric filters with defined types: minus, abs, ceil, floor, round,
// divided_by, modulo. The filter system must convert defined types without
// special-casing them.
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_Minus_DefinedInt(t *testing.T) {
	out := renderType(t, `{{ x | minus: 3 }}`, map[string]any{"x": e2eMyInt(10)})
	require.Equal(t, "7", out)
}

func TestTypeHandling_Abs_DefinedNegativeInt(t *testing.T) {
	out := renderType(t, `{{ x | abs }}`, map[string]any{"x": e2eMyInt(-8)})
	require.Equal(t, "8", out)
}

func TestTypeHandling_Abs_DefinedFloat(t *testing.T) {
	out := renderType(t, `{{ x | abs }}`, map[string]any{"x": e2eMyFloat64(-3.5)})
	require.Equal(t, "3.5", out)
}

func TestTypeHandling_Ceil_DefinedFloat(t *testing.T) {
	out := renderType(t, `{{ x | ceil }}`, map[string]any{"x": e2eMyFloat64(1.2)})
	require.Equal(t, "2", out)
}

func TestTypeHandling_Floor_DefinedFloat(t *testing.T) {
	out := renderType(t, `{{ x | floor }}`, map[string]any{"x": e2eMyFloat64(4.9)})
	require.Equal(t, "4", out)
}

func TestTypeHandling_Round_DefinedFloat(t *testing.T) {
	out := renderType(t, `{{ x | round }}`, map[string]any{"x": e2eMyFloat64(2.6)})
	require.Equal(t, "3", out)
}

func TestTypeHandling_DividedBy_DefinedInt(t *testing.T) {
	out := renderType(t, `{{ x | divided_by: 3 }}`, map[string]any{"x": e2eMyInt(9)})
	require.Equal(t, "3", out)
}

func TestTypeHandling_Modulo_DefinedInt(t *testing.T) {
	out := renderType(t, `{{ x | modulo: 4 }}`, map[string]any{"x": e2eMyInt(10)})
	require.Equal(t, "2", out)
}

func TestTypeHandling_Times_DefinedFloat(t *testing.T) {
	out := renderType(t, `{{ x | times: 2 }}`, map[string]any{"x": e2eMyFloat64(1.5)})
	require.Equal(t, "3", out)
}

// Cross-type numeric comparison: defined numeric type compared with a literal.
func TestTypeHandling_DefinedInt_CrossTypeComparison(t *testing.T) {
	out := renderType(t, `{% if x == 5 %}match{% else %}no{% endif %}`,
		map[string]any{"x": e2eMyInt(5)})
	require.Equal(t, "match", out)
}

func TestTypeHandling_DefinedFloat_CrossTypeComparison(t *testing.T) {
	out := renderType(t, `{% if x > 1 %}big{% else %}small{% endif %}`,
		map[string]any{"x": e2eMyFloat64(2.0)})
	require.Equal(t, "big", out)
}

// ════════════════════════════════════════════════════════════════════════════
// Nil binding:
//   - Non-strict mode: renders as "", is falsy, equals blank — no error.
//   - Strict mode: treated as undefined variable — returns UndefinedVariableError,
//     same as a key that doesn't exist in the bindings map at all.
// ════════════════════════════════════════════════════════════════════════════

func TestTypeHandling_NilBinding_NonStrict_RendersEmpty(t *testing.T) {
	out := renderType(t, `{{ x }}`, map[string]any{"x": nil})
	require.Equal(t, "", out)
}

func TestTypeHandling_NilBinding_NonStrict_IsFalsy(t *testing.T) {
	out := renderType(t, `{% if x %}yes{% else %}no{% endif %}`, map[string]any{"x": nil})
	require.Equal(t, "no", out)
}

func TestTypeHandling_NilBinding_NonStrict_IsBlank(t *testing.T) {
	out := renderType(t, `{% if x == blank %}blank{% else %}not-blank{% endif %}`,
		map[string]any{"x": nil})
	require.Equal(t, "blank", out)
}

func TestTypeHandling_NilBinding_NonStrict_NoError(t *testing.T) {
	// In non-strict mode, nil renders as "" without error.
	eng := NewEngine()
	_, err := eng.ParseAndRenderString(`{{ x }}`, map[string]any{"x": nil})
	require.NoError(t, err)
}

func TestTypeHandling_NilBinding_NonStrict_FilterChain(t *testing.T) {
	// nil | default: "fallback" → "fallback", no panic, no error.
	_ = fmt.Sprintf // keep import
	eng := NewEngine()
	require.NotPanics(t, func() {
		out, err := eng.ParseAndRenderString(`{{ x | default: "fallback" }}`, map[string]any{"x": nil})
		require.NoError(t, err)
		require.Equal(t, "fallback", out)
	})
}

func TestTypeHandling_NilBinding_Strict_IsUndefined(t *testing.T) {
	// In strict mode, nil binding must error the same as a missing key.
	eng := NewEngine()
	_, err := eng.ParseAndRenderString(`{{ x }}`,
		map[string]any{"x": nil},
		WithStrictVariables())
	require.Error(t, err, "nil binding in strict mode must produce UndefinedVariableError")
}

func TestTypeHandling_NilBinding_Strict_MissingKeyBehaviourMatches(t *testing.T) {
	// nil binding and missing key must produce the same kind of error in strict mode.
	eng := NewEngine()
	_, errNil := eng.ParseAndRenderString(`{{ x }}`, map[string]any{"x": nil}, WithStrictVariables())
	_, errMissing := eng.ParseAndRenderString(`{{ x }}`, map[string]any{}, WithStrictVariables())
	require.Error(t, errNil)
	require.Error(t, errMissing)
	require.IsType(t, errNil, errMissing, "nil binding and missing key must return the same error type")
}
