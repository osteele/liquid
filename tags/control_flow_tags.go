package tags

import (
	"io"

	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/expression"
	"github.com/osteele/liquid/render"
)

func caseTagParser(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// TODO parse error on non-empty node.Body
	// TODO case can include an else
	expr, err := expression.Parse(node.Args)
	if err != nil {
		return nil, err
	}
	type caseRec struct {
		// expr expression.Expression
		test func(interface{}, render.Context) (bool, error)
		node *render.BlockNode
	}
	cases := []caseRec{}
	for _, clause := range node.Clauses {
		testFn := func(interface{}, render.Context) (bool, error) { return true, nil }
		if clause.Token.Name == "when" {
			clauseExpr, err := expression.Parse(clause.Args)
			if err != nil {
				return nil, err
			}
			testFn = func(sel interface{}, ctx render.Context) (bool, error) {
				value, err := ctx.Evaluate(clauseExpr)
				if err != nil {
					return false, err
				}
				return evaluator.Equal(sel, value), nil
			}
		}
		cases = append(cases, caseRec{testFn, clause})
	}
	return func(w io.Writer, ctx render.Context) error {
		sel, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		for _, clause := range cases {
			f, err := clause.test(sel, ctx)
			if err != nil {
				return err
			}
			if f {
				return ctx.RenderChild(w, clause.node)
			}
		}
		return nil
	}, nil
}

func ifTagParser(polarity bool) func(render.BlockNode) (func(io.Writer, render.Context) error, error) { // nolint: gocyclo
	return func(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
		type branchRec struct {
			test expression.Expression
			body *render.BlockNode
		}
		expr, err := expression.Parse(node.Args)
		if err != nil {
			return nil, err
		}
		if !polarity {
			expr = expression.Not(expr)
		}
		branches := []branchRec{
			{expr, &node},
		}
		for _, c := range node.Clauses {
			test := expression.Constant(true)
			switch c.Name {
			case "else":
			// TODO parse error if this isn't the last branch
			case "elsif":
				t, err := expression.Parse(c.Args)
				if err != nil {
					return nil, err
				}
				test = t
			default:
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
					return ctx.RenderChild(w, b.body)
				}
			}
			return nil
		}, nil
	}
}
