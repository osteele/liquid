package chunks

import (
	"io"
)

// ASTNode is a node of an AST.
type ASTNode interface {
	Render(io.Writer, Context) error
}

type ASTSeq struct {
	Children []ASTNode
}

type ASTChunks struct {
	chunks []Chunk
}

type ASTText struct {
	chunk Chunk
}

type ASTObject struct {
	chunk Chunk
}

type ASTControlTag struct {
	chunk    Chunk
	cd       *ControlTagDefinition
	body     []ASTNode
	branches []*ASTControlTag
}
