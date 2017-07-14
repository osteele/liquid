package parser

import (
	"github.com/osteele/liquid/expressions"
)

// ASTNode is a node of an AST.
type ASTNode interface {
	SourceLocation() SourceLoc
	SourceText() string
}

// ASTBlock represents a {% tag %}…{% endtag %}.
type ASTBlock struct {
	Token
	syntax  BlockSyntax
	Body    []ASTNode   // Body is the nodes before the first branch
	Clauses []*ASTBlock // E.g. else and elseif w/in an if
}

// ASTRaw holds the text between the start and end of a raw tag.
type ASTRaw struct {
	Slices []string
	sourcelessNode
}

// ASTTag is a tag {% tag %} that is not a block start or end.
type ASTTag struct {
	Token
}

// ASTText is a text span, that is rendered verbatim.
type ASTText struct {
	Token
}

// ASTObject is an {{ object }} object.
type ASTObject struct {
	Token
	Expr expressions.Expression
}

// ASTSeq is a sequence of nodes.
type ASTSeq struct {
	Children []ASTNode
	sourcelessNode
}

// It shouldn't be possible to get an error from one of these node types.
// If it is, this needs to be re-thought to figure out where the source
// location comes from.
type sourcelessNode struct{}

func (n *sourcelessNode) SourceLocation() SourceLoc {
	panic("unexpected call on sourceless node")
}

func (n *sourcelessNode) SourceText() string {
	panic("unexpected call on sourceless node")
}
