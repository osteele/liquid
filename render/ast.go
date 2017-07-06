package render

import (
	"github.com/osteele/liquid/expression"
)

// ASTNode is a node of an AST.
type ASTNode interface{}

// ASTBlock represents a {% tag %}…{% endtag %}.
type ASTBlock struct {
	Chunk
	syntax   BlockSyntax
	Body     []ASTNode   // Body is the nodes before the first branch
	Branches []*ASTBlock // E.g. else and elseif w/in an if
}

// ASTRaw holds the text between the start and end of a raw tag.
type ASTRaw struct {
	slices []string
}

// ASTTag is a tag.
type ASTTag struct {
	Chunk
}

// ASTText is a text chunk, that is rendered verbatim.
type ASTText struct {
	Chunk
}

// ASTObject is an {{ object }} object.
type ASTObject struct {
	Chunk
	expr expression.Expression
}

// ASTSeq is a sequence of nodes.
type ASTSeq struct {
	Children []ASTNode
}
