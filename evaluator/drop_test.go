package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testDrop struct{}

func (d testDrop) ToLiquid() interface{} { return 3 }

func TestToLiquid(t *testing.T) {
	require.Equal(t, 2, ToLiquid(2))
	require.Equal(t, 3, ToLiquid(testDrop{}))
}
