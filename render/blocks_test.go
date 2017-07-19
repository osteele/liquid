package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockSyntax(t *testing.T) {
	cfg := NewConfig()
	cfg.AddBlock("if").Clause("else")
	cfg.AddBlock("case").Clause("else")
	cfg.AddBlock("unless")

	require.Panics(t, func() { cfg.AddBlock("if") })

	g := cfg.grammar
	ifBlock, _ := g.findBlockDef("if")
	elseBlock, _ := g.findBlockDef("else")
	unlessBlock, _ := g.findBlockDef("unless")
	require.True(t, elseBlock.CanHaveParent(ifBlock))
	require.False(t, elseBlock.CanHaveParent(unlessBlock))
	require.Equal(t, []string{"case", "if"}, elseBlock.ParentTags())
}
