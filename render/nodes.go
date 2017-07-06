package render

import (
	"io"

	"github.com/osteele/liquid/expression"
)

// Node is a node of the render tree.
type Node interface {
}

// BlockNode represents a {% tag %}…{% endtag %}.
type BlockNode struct {
	Chunk
	renderer func(io.Writer, Context) error
	Body     []Node
	Branches []*BlockNode
}

// RawNode holds the text between the start and end of a raw tag.
type RawNode struct {
	slices []string
}

// FunctionalNode renders itself via a render function that is created during parsing.
type FunctionalNode struct {
	Chunk
	render func(io.Writer, Context) error
}

// TextNode is a text chunk, that is rendered verbatim.
type TextNode struct {
	Chunk
}

// ObjectNode is an {{ object }} object.
type ObjectNode struct {
	Chunk
	expr expression.Expression
}

// SeqNode is a sequence of nodes.
type SeqNode struct {
	Children []Node
}
