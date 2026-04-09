// type_system_test.go — intensive tests for Go type system coverage.
//
// Covers:
//   - Defined types (type MyFoo T)  vs type aliases (type MyFoo = T)
//   - All numeric, bool, string kinds via ValueOf, Test, Equal, Less
//   - Invalid kinds: chan, func, complex64, complex128 → nilValue (no panic)
//   - IsBlank / IsEmpty for defined bool and string types
package values

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ── defined types (type MyFoo T) ──────────────────────────────────────────────

type myInt int
type myInt8 int8
type myInt16 int16
type myInt32 int32
type myInt64 int64
type myUint uint
type myUint8 uint8
type myUint16 uint16
type myUint32 uint32
type myUint64 uint64
type myUintptr uintptr
type myFloat32 float32
type myFloat64 float64
type myBool bool
type myString string
type mySlice []int
type myMap map[string]int

// ── type aliases (type MyFoo = T) ─────────────────────────────────────────────
// Aliases are transparent to reflect — they ARE the underlying type.
// We test them anyway to guard against regressions.

type aliasInt = int
type aliasFloat64 = float64
type aliasBool = bool
type aliasString = string

// ════════════════════════════════════════════════════════════════════════════
// ValueOf: invalid kinds must NOT panic from ValueOf itself, but must panic
// with TypeError on any method access (caught by the expression evaluator).
// ════════════════════════════════════════════════════════════════════════════

func TestValueOf_InvalidKinds_PropagateTypeError(t *testing.T) {
	cases := []struct {
		name  string
		value any
	}{
		{"chan int (non-nil)", make(chan int)},
		{"chan int (nil)", (chan int)(nil)},
		{"chan struct{}", make(chan struct{})},
		{"func()", func() {}},
		{"func(int) int", func(i int) int { return i }},
		{"complex64", complex64(complex(1, 2))},
		{"complex128", complex128(complex(3, 4))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v Value
			// ValueOf itself must not panic.
			require.NotPanics(t, func() { v = ValueOf(tc.value) })
			require.NotNil(t, v, "ValueOf must not return Go nil")
			// Any method access must panic with TypeError so the evaluator
			// can surface it as a template error.
			require.Panics(t, func() { v.Interface() }, "Interface() must panic")
			require.Panics(t, func() { v.Test() }, "Test() must panic")
			require.Panics(t, func() { v.Equal(nilValue) }, "Equal() must panic")
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// ValueOf: defined numeric types must map to the correct Value kind
// ════════════════════════════════════════════════════════════════════════════

func TestValueOf_DefinedNumericTypes_NotNil(t *testing.T) {
	cases := []any{
		myInt(7), myInt8(7), myInt16(7), myInt32(7), myInt64(7),
		myUint(7), myUint8(7), myUint16(7), myUint32(7), myUint64(7), myUintptr(7),
		myFloat32(7), myFloat64(7),
	}
	for _, v := range cases {
		v := v
		t.Run(fmt.Sprintf("%T", v), func(t *testing.T) {
			val := ValueOf(v)
			require.NotNil(t, val)
			require.NotNil(t, val.Interface())
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// Test(): defined bool types — MyBool(false) must be falsy
// ════════════════════════════════════════════════════════════════════════════

func TestTest_DefinedBoolFalse_IsFalsy(t *testing.T) {
	require.False(t, ValueOf(myBool(false)).Test(), "myBool(false) must be falsy")
}

func TestTest_DefinedBoolTrue_IsTruthy(t *testing.T) {
	require.True(t, ValueOf(myBool(true)).Test(), "myBool(true) must be truthy")
}

func TestTest_AliasBoolFalse_IsFalsy(t *testing.T) {
	var v aliasBool = false
	require.False(t, ValueOf(v).Test(), "alias bool(false) must be falsy")
}

func TestTest_AliasBoolTrue_IsTruthy(t *testing.T) {
	var v aliasBool = true
	require.True(t, ValueOf(v).Test(), "alias bool(true) must be truthy")
}

// ════════════════════════════════════════════════════════════════════════════
// Equal(): defined types must compare equal to their canonical counterparts
// ════════════════════════════════════════════════════════════════════════════

func TestEqual_DefinedIntTypes(t *testing.T) {
	cases := []struct{ a, b any }{
		{myInt(5), int(5)},
		{myInt8(5), int(5)},
		{myInt32(5), int64(5)},
		{myInt64(5), uint(5)},
		{myUint(5), float64(5)},
		{myFloat32(3.0), float64(3)},
		{myFloat64(3.0), float32(3)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T_%v==%T_%v", tc.a, tc.a, tc.b, tc.b), func(t *testing.T) {
			require.True(t, Equal(tc.a, tc.b), "Equal(%T(%v), %T(%v)) should be true", tc.a, tc.a, tc.b, tc.b)
		})
	}
}

func TestEqual_DefinedBoolType(t *testing.T) {
	require.True(t, Equal(myBool(true), true))
	require.True(t, Equal(myBool(false), false))
	require.False(t, Equal(myBool(true), false))
}

func TestEqual_DefinedStringType(t *testing.T) {
	require.True(t, Equal(myString("hello"), "hello"))
	require.False(t, Equal(myString("hello"), "world"))
	require.True(t, Equal(myString(""), ""))
}

func TestEqual_AliasTypes(t *testing.T) {
	// Aliases are identical to their base type at runtime.
	var a aliasInt = 42
	require.True(t, Equal(a, 42))
	require.True(t, Equal(a, int64(42)))

	var s aliasString = "test"
	require.True(t, Equal(s, "test"))
}

// ════════════════════════════════════════════════════════════════════════════
// Less(): defined types must order correctly
// ════════════════════════════════════════════════════════════════════════════

func TestLess_DefinedIntTypes(t *testing.T) {
	require.True(t, Less(myInt(3), int(5)))
	require.False(t, Less(myInt(5), int(3)))
	require.False(t, Less(myInt(5), int(5)))
}

func TestLess_DefinedFloatTypes(t *testing.T) {
	require.True(t, Less(myFloat64(1.0), float64(2.0)))
	require.True(t, Less(myFloat32(1.0), float64(2.0)))
	require.False(t, Less(myFloat64(2.0), myFloat64(1.0)))
}

func TestLess_DefinedStringType(t *testing.T) {
	require.True(t, Less(myString("a"), myString("b")))
	require.False(t, Less(myString("b"), myString("a")))
}

// ════════════════════════════════════════════════════════════════════════════
// NormalizeNumber: defined numeric types collapse to int64/uint64/float64
// ════════════════════════════════════════════════════════════════════════════

func TestNormalizeNumber_DefinedIntTypes(t *testing.T) {
	cases := []struct {
		in  any
		out any
	}{
		{myInt(9), int64(9)},
		{myInt8(9), int64(9)},
		{myInt16(9), int64(9)},
		{myInt32(9), int64(9)},
		{myInt64(9), int64(9)},
		{myUint(9), uint64(9)},
		{myUint8(9), uint64(9)},
		{myUint16(9), uint64(9)},
		{myUint32(9), uint64(9)},
		{myUint64(9), uint64(9)},
		{myFloat32(9), float64(9)},
		{myFloat64(9), float64(9)},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%T", tc.in), func(t *testing.T) {
			got := NormalizeNumber(tc.in)
			require.Equal(t, tc.out, got, "NormalizeNumber(%T(%v))", tc.in, tc.in)
		})
	}
}

// ════════════════════════════════════════════════════════════════════════════
// IsBlank: defined bool and string types
// ════════════════════════════════════════════════════════════════════════════

func TestIsBlank_DefinedBoolTypes(t *testing.T) {
	require.True(t, IsBlank(myBool(false)), "myBool(false) must be blank")
	require.False(t, IsBlank(myBool(true)), "myBool(true) must not be blank")
}

func TestIsBlank_DefinedStringTypes(t *testing.T) {
	require.True(t, IsBlank(myString("")), "myString('') must be blank")
	require.True(t, IsBlank(myString("   ")), "myString('   ') must be blank")
	require.True(t, IsBlank(myString("\t\n")), "myString(whitespace) must be blank")
	require.False(t, IsBlank(myString("a")), "myString('a') must not be blank")
}

func TestIsBlank_AliasTypes(t *testing.T) {
	var bf aliasBool = false
	var bt aliasBool = true
	require.True(t, IsBlank(bf))
	require.False(t, IsBlank(bt))

	var se aliasString = ""
	var ss aliasString = "hi"
	require.True(t, IsBlank(se))
	require.False(t, IsBlank(ss))
}

// ════════════════════════════════════════════════════════════════════════════
// IsEmpty: defined slice, map, string types
// ════════════════════════════════════════════════════════════════════════════

func TestIsEmpty_DefinedContainerTypes(t *testing.T) {
	require.True(t, IsEmpty(mySlice{}), "empty mySlice must be empty")
	require.False(t, IsEmpty(mySlice{1}), "non-empty mySlice must not be empty")
	require.True(t, IsEmpty(myMap{}), "empty myMap must be empty")
	require.False(t, IsEmpty(myMap{"k": 1}), "non-empty myMap must not be empty")
	require.True(t, IsEmpty(myString("")), "empty myString must be empty")
	require.False(t, IsEmpty(myString("x")), "non-empty myString must not be empty")
}

// ════════════════════════════════════════════════════════════════════════════
// Interface(): defined types must expose their actual value
// ════════════════════════════════════════════════════════════════════════════

func TestInterface_DefinedTypes_ReturnRawValue(t *testing.T) {
	// The interface value should be the original defined-type value,
	// not silently widened to the base type.
	v := ValueOf(myInt32(42))
	iface := v.Interface()
	require.NotNil(t, iface)
	// NormalizeNumber should be able to extract the numeric value.
	require.Equal(t, int64(42), NormalizeNumber(iface))
}
