package render

import (
	"io"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
)

// Node is a node of the render tree.
type Node interface {
	SourceLocation() parser.SourceLoc // for error reporting
	SourceText() string               // for error reporting
	render(*trimWriter, nodeContext) Error
}

// BlockNode represents a {% tag %}…{% endtag %}.
type BlockNode struct {
	parser.Token

	renderer func(io.Writer, Context) error
	Body     []Node
	Clauses  []*BlockNode
	Analysis NodeAnalysis
}

// RawNode holds the text between the start and end of a raw tag.
type RawNode struct {
	sourcelessNode

	slices []string
}

// TagNode renders itself via a render function that is created during parsing.
type TagNode struct {
	parser.Token

	renderer func(io.Writer, Context) error
	Analysis NodeAnalysis
}

// TextNode is a text chunk, that is rendered verbatim.
type TextNode struct {
	parser.Token
}

// ObjectNode is an {{ object }} object.
type ObjectNode struct {
	parser.Token

	expr expressions.Expression
}

// GetExpr returns the expression associated with this object node.
// Used for static analysis.
func (n *ObjectNode) GetExpr() expressions.Expression { return n.expr }

// SeqNode is a sequence of nodes.
type SeqNode struct {
	sourcelessNode

	Children []Node
}

// TrimNode is a trim object.
type TrimNode struct {
	sourcelessNode
	parser.TrimDirection
	Greedy bool // true = trim all whitespace; false = inline blanks + at most one newline
}

// FIXME requiring this is a bad design
type sourcelessNode struct{}

func (n *sourcelessNode) SourceLocation() parser.SourceLoc {
	panic("unexpected call on sourceless node")
}

func (n *sourcelessNode) SourceText() string {
	panic("unexpected call on sourceless node")
}
