package liquid

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type dropTest struct{}

func (d dropTest) ToLiquid() interface{} { return "drop" }

func TestDrops(t *testing.T) {
	require.Equal(t, "drop", FromDrop(dropTest{}))

	require.Equal(t, "not a drop", FromDrop("not a drop"))
}
