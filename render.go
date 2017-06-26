package main

import (
	"fmt"
	"io"

	"github.com/osteele/liquid/expressions"
)

type Context struct {
	Variables map[string]interface{}
}

func (c *Context) EvaluateExpr(expr string) (interface{}, error) {
	return expressions.EvaluateExpr(expr, expressions.Context{Variables: c.Variables})
}

func (n *ASTSeq) Render(w io.Writer, ctx Context) error {
	for _, c := range n.Children {
		if err := c.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *ASTChunks) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte("{chunks}"))
	return err
}

func (n *ASTText) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte(n.chunk.Source))
	return err
}

func writeASTs(w io.Writer, seq []AST, ctx Context) error {
	for _, n := range seq {
		if err := n.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

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
			return writeASTs(w, n.body, ctx)
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
					return writeASTs(w, c.body, ctx)
				}
			}
		}
		return nil
	default:
		_, err := w.Write([]byte("{control}"))
		return err
	}
}

func (n *ASTObject) Render(w io.Writer, ctx Context) error {
	val, err := ctx.EvaluateExpr(n.chunk.Tag)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprint(val)))
	return err
}
