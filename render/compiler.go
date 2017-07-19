package render

import (
	"fmt"

	"github.com/osteele/liquid/parser"
)

// Compile parses a source template. It returns an AST root, that can be evaluated.
func (c Config) Compile(source string, loc parser.SourceLoc) (Node, parser.Error) {
	root, err := c.Parse(source, loc)
	if err != nil {
		return nil, err
	}
	return c.compileNode(root)
}

// nolint: gocyclo
func (c Config) compileNode(n parser.ASTNode) (Node, parser.Error) {
	switch n := n.(type) {
	case *parser.ASTBlock:
		body, err := c.compileNodes(n.Body)
		if err != nil {
			return nil, err
		}
		branches, err := c.compileBlocks(n.Clauses)
		if err != nil {
			return nil, err
		}

		cd, ok := c.findBlockDef(n.Name)
		if !ok {
			return nil, parser.Errorf(n, "undefined tag %q", n.Name)
		}
		node := BlockNode{
			Token:   n.Token,
			Body:    body,
			Clauses: branches,
		}
		if cd.parser != nil {
			r, err := cd.parser(node)
			if err != nil {
				return nil, parser.WrapError(err, n)
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
				return nil, parser.Errorf(n, "%s", err)
			}
			return &TagNode{n.Token, f}, nil
		}
		return nil, parser.Errorf(n, "undefined tag %q", n.Name)
	case *parser.ASTText:
		return &TextNode{n.Token}, nil
	case *parser.ASTObject:
		return &ObjectNode{n.Token, n.Expr}, nil
	default:
		panic(fmt.Errorf("un-compilable node type %T", n))
	}
}

func (c Config) compileBlocks(blocks []*parser.ASTBlock) ([]*BlockNode, parser.Error) {
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

func (c Config) compileNodes(nodes []parser.ASTNode) ([]Node, parser.Error) {
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
