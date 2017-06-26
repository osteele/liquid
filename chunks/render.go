package chunks

import (
	"fmt"
	"io"
)

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTSeq) Render(w io.Writer, ctx Context) error {
	for _, c := range n.Children {
		if err := c.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTChunks) Render(w io.Writer, _ Context) error {
	fmt.Println(MustYAML(n))
	return fmt.Errorf("unimplemented: ASTChunks.Render")
}

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTGenericTag) Render(w io.Writer, ctx Context) error {
	return n.render(w, ctx)
}

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTText) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte(n.chunk.Source))
	return err
}

// Render evaluates an AST node and writes the result to an io.Writer.
func renderASTSequence(w io.Writer, seq []ASTNode, ctx Context) error {
	for _, n := range seq {
		if err := n.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTControlTag) Render(w io.Writer, ctx Context) error {
	switch n.chunk.Tag {
	case "if", "unless":
		val, err := ctx.EvaluateExpr(n.chunk.Args)
		if err != nil {
			return err
		}
		if n.chunk.Tag == "unless" {
			val = (val == nil || val == false)
		}
		switch val {
		default:
			return renderASTSequence(w, n.body, ctx)
		case nil, false:
			for _, c := range n.branches {
				switch c.chunk.Tag {
				case "else":
					val = true
				case "elsif":
					val, err = ctx.EvaluateExpr(c.chunk.Args)
					if err != nil {
						return err
					}
				}
				if val != nil && val != false {
					return renderASTSequence(w, c.body, ctx)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("unimplemented tag: %s", n.chunk.Tag)
	}
}

// Render evaluates an AST node and writes the result to an io.Writer.
func (n *ASTObject) Render(w io.Writer, ctx Context) error {
	val, err := ctx.EvaluateExpr(n.chunk.Args)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprint(val)))
	return err
}
