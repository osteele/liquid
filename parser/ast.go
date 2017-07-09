package parser

import (
	"github.com/osteele/liquid/expression"
)

// ASTNode is a node of an AST.
type ASTNode interface{}

// ASTBlock represents a {% tag %}…{% endtag %}.
type ASTBlock struct {
	Token
	syntax   BlockSyntax
	Body     []ASTNode   // Body is the nodes before the first branch
	Branches []*ASTBlock // E.g. else and elseif w/in an if
}

// ASTRaw holds the text between the start and end of a raw tag.
type ASTRaw struct {
	Slices []string
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
	Expr expression.Expression
}

// ASTSeq is a sequence of nodes.
type ASTSeq struct {
	Children []ASTNode
}
