// Package tags defines the standard Liquid tags.
package tags

import (
	"io"

	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/expression"
	"github.com/osteele/liquid/render"
)

// AddStandardTags defines the standard Liquid tags.
func AddStandardTags(c render.Config) {
	c.AddTag("assign", assignTag)
	c.AddTag("include", includeTag)

	// blocks
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	c.AddTag("break", breakTag)
	c.AddTag("continue", continueTag)
	c.AddTag("cycle", cycleTag)
	c.AddBlock("capture").Compiler(captureTagParser)
	c.AddBlock("case").Clause("when").Clause("else").Compiler(caseTagParser)
	c.AddBlock("comment")
	c.AddBlock("for").Compiler(loopTagParser)
	c.AddBlock("if").Clause("else").Clause("elsif").Compiler(ifTagParser(true))
	c.AddBlock("raw")
	c.AddBlock("tablerow")
	c.AddBlock("unless").Compiler(ifTagParser(false))
}

func assignTag(source string) (func(io.Writer, render.Context) error, error) {
	return func(w io.Writer, ctx render.Context) error {
		_, err := ctx.EvaluateStatement("assign", source)
		return err
	}, nil
}

func captureTagParser(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// TODO verify syntax
	varname := node.Args
	return func(w io.Writer, ctx render.Context) error {
		s, err := ctx.InnerString()
		if err != nil {
			return err
		}
		ctx.Set(varname, s)
		return nil
	}, nil
}

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
