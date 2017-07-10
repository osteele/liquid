package render

import (
	"fmt"

	"github.com/osteele/liquid/parser"
)

// TODO DRY with render.Error
// Do there really need to be two kinds of errors?

// A CompilationError is an error in the template source, encountered during template compilation.
type CompilationError interface {
	error
	Cause() error
	Filename() string
	LineNumber() int
}
type compilationError struct {
	parser.SourceLoc
	context string
	message string
	cause   error
}

func (e *compilationError) Cause() error {
	return e.cause
}

func (e *compilationError) Filename() string {
	return e.Pathname
}

func (e *compilationError) LineNumber() int {
	return e.LineNo
}

func (e *compilationError) Error() string {
	locative := "in " + e.context
	if e.Pathname != "" {
		locative = "in " + e.Pathname
	}
	return fmt.Sprintf("Liquid exception: Liquid syntax error (line %d): %s%s", e.LineNo, e.message, locative)
}

func compilationErrorf(loc parser.SourceLoc, context, format string, a ...interface{}) *compilationError {
	return &compilationError{loc, context, fmt.Sprintf(format, a...), nil}
}

func wrapCompilationError(err error, n Node) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	re := compilationErrorf(n.SourceLocation(), n.SourceText(), "%s", err)
	re.cause = err
	return re
}

// Compile parses a source template. It returns an AST root, that can be evaluated.
func (c Config) Compile(source string) (parser.ASTNode, CompilationError) {
	root, err := c.Parse(source)
	if err != nil {
		return nil, err
	}
	return c.compileNode(root)
}

// nolint: gocyclo
func (c Config) compileNode(n parser.ASTNode) (Node, CompilationError) {
	switch n := n.(type) {
	case *parser.ASTBlock:
		body, err := c.compileNodes(n.Body)
		if err != nil {
			return nil, err
		}
		branches, err := c.compileBlocks(n.Branches)
		if err != nil {
			return nil, err
		}

		cd, ok := c.findBlockDef(n.Name)
		if !ok {
			return nil, compilationErrorf(n.SourceLoc, n.Source, "undefined tag %q", n.Name)
		}
		node := BlockNode{
			Token:    n.Token,
			Body:     body,
			Branches: branches,
		}
		if cd.parser != nil {
			r, err := cd.parser(node)
			if err != nil {
				return nil, wrapCompilationError(err, n)
			}
			node.renderer = r
		}
		return &node, nil
	case *parser.ASTRaw:
		return &RawNode{n.Slices, sourcelessNode{}}, nil
	case *parser.ASTSeq:
		children, err := c.compileNodes(n.Children)
		if err != nil {
			return nil, err
		}
		return &SeqNode{children, sourcelessNode{}}, nil
	case *parser.ASTTag:
		if td, ok := c.FindTagDefinition(n.Name); ok {
			f, err := td(n.Args)
			if err != nil {
				return nil, compilationErrorf(n.SourceLoc, n.Source, "%s", err)
			}
			return &TagNode{n.Token, f}, nil
		}
		return nil, compilationErrorf(n.SourceLoc, n.Source, "unknown tag: %s", n.Name)
	case *parser.ASTText:
		return &TextNode{n.Token}, nil
	case *parser.ASTObject:
		return &ObjectNode{n.Token, n.Expr}, nil
	default:
		panic(fmt.Errorf("un-compilable node type %T", n))
	}
}

func (c Config) compileBlocks(blocks []*parser.ASTBlock) ([]*BlockNode, CompilationError) {
	out := make([]*BlockNode, 0, len(blocks))
	for _, child := range blocks {
		compiled, err := c.compileNode(child)
		if err != nil {
			return nil, err
		}
		out = append(out, compiled.(*BlockNode))
	}
	return out, nil
}

func (c Config) compileNodes(nodes []parser.ASTNode) ([]Node, CompilationError) {
	out := make([]Node, 0, len(nodes))
	for _, child := range nodes {
		compiled, err := c.compileNode(child)
		if err != nil {
			return nil, err
		}
		out = append(out, compiled)
	}
	return out, nil
}
