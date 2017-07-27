package tags

import (
	"io"

	e "github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/values"
)

type caseInterpreter interface {
	body() *render.BlockNode
	test(interface{}, render.Context) (bool, error)
}
type exprCase struct {
	e.When
	b *render.BlockNode
}

func (c exprCase) body() *render.BlockNode { return c.b }

func (c exprCase) test(caseValue interface{}, ctx render.Context) (bool, error) {
	for _, expr := range c.Exprs {
		whenValue, err := ctx.Evaluate(expr)
		if err != nil {
			return false, err
		}
		if values.Equal(caseValue, whenValue) {
			return true, nil
		}
	}
	return false, nil
}

type elseCase struct{ b *render.BlockNode }

func (c elseCase) body() *render.BlockNode { return c.b }

func (c elseCase) test(interface{}, render.Context) (bool, error) { return true, nil }

func caseTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// TODO syntax error on non-empty node.Body
	expr, err := e.Parse(node.Args)
	if err != nil {
		return nil, err
	}
	cases := []caseInterpreter{}
	for _, clause := range node.Clauses {
		switch clause.Token.Name {
		case "when":
			stmt, err := e.ParseStatement(e.WhenStatementSelector, clause.Args)
			if err != nil {
				return nil, err
			}
			cases = append(cases, exprCase{stmt.When, clause})
		default: // should be a check for "else", but I like the metacircularity
			cases = append(cases, elseCase{clause})
		}
	}
	return func(w io.Writer, ctx render.Context) error {
		sel, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		for _, clause := range cases {
			b, err := clause.test(sel, ctx)
			if err != nil {
				return err
			}
			if b {
				return ctx.RenderBlock(w, clause.body())
			}
		}
		return nil
	}, nil
}

func ifTagCompiler(polarity bool) func(render.BlockNode) (func(io.Writer, render.Context) error, error) { // nolint: gocyclo
	return func(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
		type branchRec struct {
			test e.Expression
			body *render.BlockNode
		}
		expr, err := e.Parse(node.Args)
		if err != nil {
			return nil, err
		}
		if !polarity {
			expr = e.Not(expr)
		}
		branches := []branchRec{
			{expr, &node},
		}
		for _, c := range node.Clauses {
			test := e.Constant(true)
			switch c.Name {
			case "else":
			// TODO syntax error if this isn't the last branch
			case "elsif":
				t, err := e.Parse(c.Args)
				if err != nil {
					return nil, err
				}
				test = t
			}
			branches = append(branches, branchRec{test, c})
		}
		return func(w io.Writer, ctx render.Context) error {
			for _, b := range branches {
				value, err := ctx.Evaluate(b.test)
				if err != nil {
					return err
				}
				if value != nil && value != false {
					return ctx.RenderBlock(w, b.body)
				}
			}
			return nil
		}, nil
	}
}
