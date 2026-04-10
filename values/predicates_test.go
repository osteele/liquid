package values

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEmpty(t *testing.T) {
	require.False(t, IsEmpty(false))  // false is NOT empty (only strings/arrays/maps can be empty)
	require.False(t, IsEmpty(true))
	require.False(t, IsEmpty(nil))    // nil is NOT empty
	require.True(t, IsEmpty(""))
	require.False(t, IsEmpty("  "))   // whitespace-only is not empty (it's blank)
	require.True(t, IsEmpty([]string{}))
	require.True(t, IsEmpty(map[string]any{}))
	require.False(t, IsEmpty([]string{""}))
	require.False(t, IsEmpty(map[string]any{"k": "v"}))
}

func TestIsBlank(t *testing.T) {
	require.True(t, IsBlank(nil))
	require.True(t, IsBlank(false))
	require.False(t, IsBlank(true))
	require.True(t, IsBlank(""))
	require.True(t, IsBlank("  "))
	require.True(t, IsBlank("\t\n"))
	require.False(t, IsBlank("a"))
	require.True(t, IsBlank([]string{}))
	require.False(t, IsBlank([]string{""}))
	require.True(t, IsBlank(map[string]any{}))
	require.False(t, IsBlank(map[string]any{"k": "v"}))
}

