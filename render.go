package main

import (
	"fmt"
	"io"
)

type Context struct {
	Variables map[string]interface{}
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

func EvaluateExpr(expr string, ctx Context) (interface{}, error) {
	lexer := newLexer([]byte(expr))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %s", expr)
	}
	return lexer.val(ctx), nil
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
	case "if":
		val, err := EvaluateExpr(n.chunk.Args, ctx)
		if err != nil {
			return err
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
					val, err = EvaluateExpr(c.chunk.Args, ctx)
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
	val, err := EvaluateExpr(n.chunk.Tag, ctx)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(fmt.Sprint(val)))
	return err
}
