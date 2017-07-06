package render

import (
	"fmt"
	"io"
)

// A CompilationError is a parse error during template compilation.
type CompilationError string

func (e CompilationError) Error() string { return string(e) }

func compilationErrorf(format string, a ...interface{}) CompilationError {
	return CompilationError(fmt.Sprintf(format, a...))
}

// Compile parses a source template. It returns an AST root, that can be evaluated.
func (c Config) Compile(source string) (ASTNode, error) {
	root, err := c.Parse(source)
	if err != nil {
		return nil, err
	}
	return c.compileNode(root)
}

// nolint: gocyclo
func (c Config) compileNode(n ASTNode) (ASTNode, error) {
	switch n := n.(type) {
	case *ASTBlock:
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
			return nil, compilationErrorf("undefined tag %q", n.Name)
		}
		var renderer func(io.Writer, Context) error
		if cd.parser != nil {
			r, err := cd.parser(*n)
			if err != nil {
				return nil, err
			}
			renderer = r
		}
		return &ASTBlock{
			Chunk:    n.Chunk,
			renderer: renderer,
			syntax:   n.syntax,
			Body:     body,
			Branches: branches,
		}, nil
	case *ASTFunctional:
		return &ASTFunctional{n.Chunk, n.render}, nil
	case *ASTRaw:
		return &ASTRaw{n.slices}, nil
	case *ASTSeq:
		children, err := c.compileNodes(n.Children)
		if err != nil {
			return nil, err
		}
		return &ASTSeq{children}, nil
	case *ASTText:
		return &ASTText{n.Chunk}, nil
	case *ASTObject:
		return &ASTObject{n.Chunk, n.expr}, nil
	default:
		panic(fmt.Errorf("un-compilable node type %T", n))
	}
}

func (c Config) compileBlocks(blocks []*ASTBlock) ([]*ASTBlock, error) {
	out := make([]*ASTBlock, 0, len(blocks))
	for _, child := range blocks {
		compiled, err := c.compileNode(child)
		if err != nil {
			return nil, err
		}
		out = append(out, compiled.(*ASTBlock))
	}
	return out, nil
}

func (c Config) compileNodes(nodes []ASTNode) ([]ASTNode, error) {
	out := make([]ASTNode, 0, len(nodes))
	for _, child := range nodes {
		compiled, err := c.compileNode(child)
		if err != nil {
			return nil, err
		}
		out = append(out, compiled)
	}
	return out, nil
}
