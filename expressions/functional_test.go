package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstant(t *testing.T) {
	ctx := NewContext(map[string]interface{}{}, NewConfig())
	k := Constant(10)
	v, err := k.Evaluate(ctx)
	require.NoError(t, err)
	require.Equal(t, 10, v)
}
func TestNot(t *testing.T) {
	ctx := NewContext(map[string]interface{}{}, NewConfig())
	k := Constant(10)
	v, err := Not(k).Evaluate(ctx)
	require.NoError(t, err)
	require.Equal(t, false, v)
}
