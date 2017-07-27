package values

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDrop struct{ proxy interface{} }

func (d testDrop) ToLiquid() interface{} { return d.proxy }

func TestToLiquid(t *testing.T) {
	require.Equal(t, 2, ToLiquid(2))
	require.Equal(t, 3, ToLiquid(testDrop{3}))
}

func TestValue_drop(t *testing.T) {
	dv := ValueOf(testDrop{"seafood"})
	require.Equal(t, "seafood", dv.Interface())
	require.Equal(t, true, dv.Contains(ValueOf("foo")))
	require.Equal(t, true, dv.Contains(ValueOf(testDrop{"foo"})))
	require.Equal(t, 7, dv.PropertyValue(ValueOf("size")).Interface())
}
