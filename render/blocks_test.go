package render

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
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

func TestBlockParentState(t *testing.T) {
	cfg := NewConfig()
	cfg.AddBlock("if").Clause("else").Clause("elsif")
	cfg.AddBlock("case").Clause("when").Clause("else")
	cfg.AddBlock("unless")
	cfg.AddBlock("for")

	g := cfg.grammar

	mustFind := func(name string) parser.BlockSyntax {
		def, ok := g.findBlockDef(name)
		require.True(t, ok, "block %q should be defined", name)
		return def
	}

	ifBlock := mustFind("if")
	caseBlock := mustFind("case")
	unlessBlock := mustFind("unless")
	forBlock := mustFind("for")

	// blockParentStateTests models the block parent rules as a state table.
	// Each row is (child tag, parent tag) -> allowed by CanHaveParent.
	var blockParentStateTests = []struct {
		name    string
		child   string
		parent  parser.BlockSyntax
		allowed bool
	}{
		// Clauses are allowed only in their declared parents.
		{"else inside if", "else", ifBlock, true},
		{"else inside case", "else", caseBlock, true},
		{"else inside unless", "else", unlessBlock, false},
		{"elsif inside if", "elsif", ifBlock, true},
		{"elsif inside case", "elsif", caseBlock, false},
		{"when inside case", "when", caseBlock, true},
		{"when inside if", "when", ifBlock, false},

		// End tags must match the immediate parent.
		{"endif inside if", "endif", ifBlock, true},
		{"endfor inside for", "endfor", forBlock, true},
		{"endfor inside if", "endfor", ifBlock, false},
		{"endif outside block", "endif", nil, false},

		// Block starts can appear anywhere.
		{"if start at top level", "if", nil, true},
		{"for start inside if", "for", ifBlock, true},
	}

	for _, test := range blockParentStateTests {
		t.Run(test.name, func(t *testing.T) {
			child, ok := g.findBlockDef(test.child)
			require.True(t, ok, "child tag %q should be defined", test.child)
			require.Equalf(t, test.allowed, child.CanHaveParent(test.parent),
				"CanHaveParent(%q, %q)", test.child, test.parent)
		})
	}
}
