package chunks

import (
	"io"
)

type AST interface {
	Render(io.Writer, Context) error
}

type ASTSeq struct {
	Children []AST
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
	body     []AST
	branches []*ASTControlTag
}
