package values

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEmptyDrop mirrors the LiquidJS integration tests in
// test/integration/drop/empty-drop.spec.ts

func TestEmptyDrop_Interface(t *testing.T) {
	require.Equal(t, "", EmptyDrop.Interface())
}

func TestEmptyDrop_Test(t *testing.T) {
	// Only nil and false are falsy in Liquid; EmptyDrop is truthy.
	require.True(t, EmptyDrop.Test())
}

func TestEmptyDrop_Equal_selfIsFalse(t *testing.T) {
	// empty != empty
	require.False(t, EmptyDrop.Equal(EmptyDrop))
}

func TestEmptyDrop_Equal_nilIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(nilValue))
}

func TestEmptyDrop_Equal_falseIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(falseValue))
}

func TestEmptyDrop_Equal_emptyString(t *testing.T) {
	require.True(t, EmptyDrop.Equal(ValueOf("")))
}

func TestEmptyDrop_Equal_whitespaceStringIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(ValueOf("  ")))
}

func TestEmptyDrop_Equal_emptyMap(t *testing.T) {
	require.True(t, EmptyDrop.Equal(ValueOf(map[string]any{})))
}

func TestEmptyDrop_Equal_nonEmptyMapIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(ValueOf(map[string]any{"foo": 1})))
}

func TestEmptyDrop_Equal_emptySlice(t *testing.T) {
	require.True(t, EmptyDrop.Equal(ValueOf([]any{})))
}

func TestEmptyDrop_Equal_nonEmptySliceIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(ValueOf([]any{1})))
}

func TestEmptyDrop_Equal_intIsFalse(t *testing.T) {
	require.False(t, EmptyDrop.Equal(ValueOf(1)))
}

func TestEmptyDrop_Less(t *testing.T) {
	require.False(t, EmptyDrop.Less(ValueOf(1)))
}

// Symmetric: value == empty should behave the same as empty == value.
func TestEmptyDrop_SymmetricViaEqual(t *testing.T) {
	require.True(t, Equal("", EmptyDrop))
	require.False(t, Equal("  ", EmptyDrop))
	require.False(t, Equal(nil, EmptyDrop))
	require.False(t, Equal(false, EmptyDrop))
	require.True(t, Equal([]any{}, EmptyDrop))
	require.False(t, Equal([]any{1}, EmptyDrop))
}

// -- BlankDrop ----------------------------------------------------------------

// TestBlankDrop_* mirrors test/integration/drop/blank-drop.spec.ts

func TestBlankDrop_Interface(t *testing.T) {
	require.Equal(t, "", BlankDrop.Interface())
}

func TestBlankDrop_Test(t *testing.T) {
	require.True(t, BlankDrop.Test())
}

func TestBlankDrop_Equal_selfIsFalse(t *testing.T) {
	require.False(t, BlankDrop.Equal(BlankDrop))
}

func TestBlankDrop_Equal_nil(t *testing.T) {
	require.True(t, BlankDrop.Equal(nilValue))
}

func TestBlankDrop_Equal_false(t *testing.T) {
	require.True(t, BlankDrop.Equal(falseValue))
}

func TestBlankDrop_Equal_emptyString(t *testing.T) {
	require.True(t, BlankDrop.Equal(ValueOf("")))
}

func TestBlankDrop_Equal_whitespaceString(t *testing.T) {
	require.True(t, BlankDrop.Equal(ValueOf("  ")))
}

func TestBlankDrop_Equal_nonBlankStringIsFalse(t *testing.T) {
	require.False(t, BlankDrop.Equal(ValueOf("hello")))
}

func TestBlankDrop_Equal_emptyMap(t *testing.T) {
	require.True(t, BlankDrop.Equal(ValueOf(map[string]any{})))
}

func TestBlankDrop_Equal_nonEmptyMapIsFalse(t *testing.T) {
	require.False(t, BlankDrop.Equal(ValueOf(map[string]any{"foo": 1})))
}

func TestBlankDrop_Equal_emptySlice(t *testing.T) {
	require.True(t, BlankDrop.Equal(ValueOf([]any{})))
}

func TestBlankDrop_Equal_nonEmptySliceIsFalse(t *testing.T) {
	require.False(t, BlankDrop.Equal(ValueOf([]any{1})))
}

// Symmetric: value == blank
func TestBlankDrop_SymmetricViaEqual(t *testing.T) {
	require.True(t, Equal(nil, BlankDrop))
	require.True(t, Equal(false, BlankDrop))
	require.True(t, Equal("", BlankDrop))
	require.True(t, Equal("  ", BlankDrop))
	require.False(t, Equal("hello", BlankDrop))
	require.True(t, Equal([]any{}, BlankDrop))
	require.False(t, Equal([]any{1}, BlankDrop))
}

// -- blank != empty -----------------------------------------------------------

func TestBlankDrop_NotEqualToEmptyDrop(t *testing.T) {
	require.False(t, Equal(BlankDrop, EmptyDrop))
	require.False(t, Equal(EmptyDrop, BlankDrop))
}
