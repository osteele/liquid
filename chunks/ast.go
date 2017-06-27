package chunks

import (
	"io"
)

// ASTNode is a node of an AST.
type ASTNode interface {
	Render(io.Writer, Context) error
}

// ASTRaw holds the text between the start and end of a raw tag.
type ASTRaw struct {
	slices []string
}

// ASTSeq is a sequence of nodes.
type ASTSeq struct {
	Children []ASTNode
}

// ASTChunks is a sequence of chunks.
// TODO probably safe to remove this type and method, once the test suite is larger
type ASTChunks struct {
	chunks []Chunk
}

// ASTGenericTag renders itself via a render function that is created during parsing.
type ASTGenericTag struct {
	render func(io.Writer, Context) error
}

// ASTText is a text chunk, that is rendered verbatim.
type ASTText struct {
	Chunk
}

// ASTObject is an {{ object }} object.
type ASTObject struct {
	Chunk
}

// ASTControlTag is a control tag.
type ASTControlTag struct {
	Chunk
	cd       *controlTagDefinition
	Body     []ASTNode
	Branches []*ASTControlTag
}
