package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsEmpty(t *testing.T) {
	require.True(t, IsEmpty(false))
	require.False(t, IsEmpty(true))
	require.True(t, IsEmpty([]string{}))
	require.True(t, IsEmpty(map[string]interface{}{}))
	require.False(t, IsEmpty([]string{""}))
	require.False(t, IsEmpty(map[string]interface{}{"k": "v"}))
}
