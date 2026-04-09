package render

import (
	"fmt"
	"io"

	"github.com/osteele/liquid/parser"
)

// CompileAuditResult is the result of CompileAudit.
// It separates the successfully-compiled tree from the parse-time diagnostics
// and the optional fatal error that prevented a full AST.
type CompileAuditResult struct {
	// Node is the compiled render tree. Non-nil when FatalError == nil.
	Node Node
	// Diags are non-fatal parse diagnostics (syntax-error in {{ }} objects).
	// Present even when FatalError != nil if they occurred before the fatal error.
	Diags []parser.ParseDiag
	// FatalError is the structural parse error (unclosed-tag or unexpected-tag),
	// when the AST could not be completed. Node is nil in this case.
	FatalError parser.Error
}

// CompileAudit parses source in error-recovering mode and compiles the result.
//
// Syntax errors in {{ expr }} objects are collected as non-fatal ParseDiags and
// replaced with BrokenNode in the render tree. Tag/block compile errors are
// also collected as non-fatal diagnostics.
//
// Only two structural errors remain fatal and set FatalError:
//   - unclosed-tag ({% if %} without {% endif %})
//   - unexpected-tag ({% endif %} without an opening {% if %})
func (c *Config) CompileAudit(source string, loc parser.SourceLoc) CompileAuditResult {
	root, diags, fatalErr := c.Config.ParseAudit(source, loc)
	if fatalErr != nil {
		return CompileAuditResult{Diags: diags, FatalError: fatalErr}
	}

	node, compileErr := c.compileNodeAudit(root, &diags)
	if compileErr != nil {
		// Structural compile error (should not normally happen after audit parse).
		return CompileAuditResult{Diags: diags, FatalError: compileErr}
	}

	return CompileAuditResult{Node: node, Diags: diags}
}

// compileNodeAudit is like compileNode but catches non-fatal compile errors
// (e.g. tag argument parse failures) and converts them to ParseDiags + BrokenNode.
func (c *Config) compileNodeAudit(n parser.ASTNode, diags *[]parser.ParseDiag) (Node, parser.Error) { //nolint: gocyclo
	switch n := n.(type) {
	case *parser.ASTBlock:
		body, err := c.compileNodesAudit(n.Body, diags)
		if err != nil {
			return nil, err
		}

		branches, err := c.compileBlocksAudit(n.Clauses, diags)
		if err != nil {
			return nil, err
		}

		cd, ok := c.findBlockDef(n.Name)
		if !ok {
			// Non-fatal: unknown block becomes BrokenNode.
			*diags = append(*diags, parser.ParseDiag{
				Code:    "syntax-error",
				Message: fmt.Sprintf("undefined tag %q", n.Name),
				Tok:     n.Token,
			})
			return &BrokenNode{n.Token}, nil
		}

		node := BlockNode{
			Token:   n.Token,
			Body:    body,
			Clauses: branches,
		}
		if cd.parser != nil {
			r, err := cd.parser(node)
			if err != nil {
				// Non-fatal: block arg parse failure → BrokenNode.
				*diags = append(*diags, parser.ParseDiag{
					Code:    "syntax-error",
					Message: err.Error(),
					Tok:     n.Token,
				})
				return &BrokenNode{n.Token}, nil
			}
			node.renderer = r
		}
		if analyzer, ok := c.findBlockAnalyzer(n.Name); ok {
			node.Analysis = analyzer(node)
		}

		return &node, nil

	case *parser.ASTRaw:
		return &RawNode{sourcelessNode{}, n.Slices}, nil

	case *parser.ASTSeq:
		children, err := c.compileNodesAudit(n.Children, diags)
		if err != nil {
			return nil, err
		}
		return &SeqNode{sourcelessNode{}, children}, nil

	case *parser.ASTTag:
		if td, ok := c.FindTagDefinition(n.Name); ok {
			f, err := td(n.Args)
			if err != nil {
				// Non-fatal: tag arg parse failure → BrokenNode.
				*diags = append(*diags, parser.ParseDiag{
					Code:    "syntax-error",
					Message: err.Error(),
					Tok:     n.Token,
				})
				return &BrokenNode{n.Token}, nil
			}

			var analysis NodeAnalysis
			if analyzer, ok := c.findTagAnalyzer(n.Name); ok {
				analysis = analyzer(n.Args)
			}
			return &TagNode{n.Token, f, analysis}, nil
		}

		if c.LaxTags {
			noopFn := func(io.Writer, Context) error { return nil }
			return &TagNode{n.Token, noopFn, NodeAnalysis{}}, nil
		}

		// Non-fatal: unknown tag → BrokenNode.
		*diags = append(*diags, parser.ParseDiag{
			Code:    "syntax-error",
			Message: fmt.Sprintf("undefined tag %q", n.Name),
			Tok:     n.Token,
		})
		return &BrokenNode{n.Token}, nil

	case *parser.ASTText:
		return &TextNode{n.Token}, nil

	case *parser.ASTObject:
		return &ObjectNode{n.Token, n.Expr}, nil

	case *parser.ASTBroken:
		// Already recorded as a diagnostic during parsing; just create BrokenNode.
		return &BrokenNode{n.Token}, nil

	case *parser.ASTTrim:
		return &TrimNode{TrimDirection: n.TrimDirection, Greedy: c.Greedy}, nil

	default:
		panic(fmt.Errorf("un-compilable node type %T", n))
	}
}

func (c *Config) compileBlocksAudit(blocks []*parser.ASTBlock, diags *[]parser.ParseDiag) ([]*BlockNode, parser.Error) {
	out := make([]*BlockNode, 0, len(blocks))
	for _, child := range blocks {
		compiled, err := c.compileNodeAudit(child, diags)
		if err != nil {
			return nil, err
		}
		// compileNodeAudit never returns BrokenNode for a block that has a
		// matching blockDef, but if it does (e.g. unknown block), skip casting.
		if bn, ok := compiled.(*BlockNode); ok {
			out = append(out, bn)
		}
		// BrokenNode for an unknown block clause: skip it in the clauses list.
	}
	return out, nil
}

func (c *Config) compileNodesAudit(nodes []parser.ASTNode, diags *[]parser.ParseDiag) ([]Node, parser.Error) {
	out := make([]Node, 0, len(nodes))
	for _, child := range nodes {
		compiled, err := c.compileNodeAudit(child, diags)
		if err != nil {
			return nil, err
		}

		var trimLeft, trimRight bool
		switch compiled.(type) {
		case *TagNode, *BlockNode:
			trimLeft = c.TrimTagLeft
			trimRight = c.TrimTagRight
		case *ObjectNode:
			trimLeft = c.TrimOutputLeft
			trimRight = c.TrimOutputRight
		}

		if trimLeft {
			out = append(out, &TrimNode{TrimDirection: parser.Left, Greedy: c.Greedy})
		}
		out = append(out, compiled)
		if trimRight {
			out = append(out, &TrimNode{TrimDirection: parser.Right, Greedy: c.Greedy})
		}
	}
	return out, nil
}
