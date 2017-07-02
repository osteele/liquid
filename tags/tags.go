// Package tags defines the standard Liquid tags.
package tags

import (
	"io"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/generics"
)

// AddStandardTags defines the standard Liquid tags.
func AddStandardTags(settings chunks.Settings) {
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	loopTags := []string{"break", "continue", "cycle"}
	settings.AddTag("break", breakTag)
	settings.AddTag("continue", continueTag)
	settings.AddBlock("capture").Parser(captureTagParser)
	settings.AddBlock("case").Branch("when").Parser(caseTagParser)
	settings.AddBlock("comment")
	settings.AddBlock("for").Governs(loopTags).Parser(loopTagParser)
	settings.AddBlock("if").Branch("else").Branch("elsif").Parser(ifTagParser(true))
	settings.AddBlock("raw")
	settings.AddBlock("tablerow").Governs(loopTags)
	settings.AddBlock("unless").SameSyntaxAs("if").Parser(ifTagParser(false))
}

func captureTagParser(node chunks.ASTBlockNode) (func(io.Writer, chunks.RenderContext) error, error) {
	// TODO verify syntax
	varname := node.Args
	return func(w io.Writer, ctx chunks.RenderContext) error {
		s, err := ctx.InnerString()
		if err != nil {
			return err
		}
		ctx.Set(varname, s)
		return nil
	}, nil
}

func caseTagParser(node chunks.ASTBlockNode) (func(io.Writer, chunks.RenderContext) error, error) {
	// TODO parse error on non-empty node.Body
	// TODO case can include an else
	expr, err := expressions.Parse(node.Args)
	if err != nil {
		return nil, err
	}
	type caseRec struct {
		expr expressions.Expression
		node *chunks.ASTBlockNode
	}
	cases := []caseRec{}
	for _, branch := range node.Branches {
		bfn, err := expressions.Parse(branch.Args)
		if err != nil {
			return nil, err
		}
		cases = append(cases, caseRec{bfn, branch})
	}
	return func(w io.Writer, ctx chunks.RenderContext) error {
		value, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		for _, branch := range cases {
			b, err := ctx.Evaluate(branch.expr)
			if err != nil {
				return err
			}
			if generics.Equal(value, b) {
				return ctx.RenderChild(w, branch.node)
			}
		}
		return nil
	}, nil
}

func ifTagParser(polarity bool) func(chunks.ASTBlockNode) (func(io.Writer, chunks.RenderContext) error, error) { // nolint: gocyclo
	return func(node chunks.ASTBlockNode) (func(io.Writer, chunks.RenderContext) error, error) {
		type branchRec struct {
			test expressions.Expression
			body *chunks.ASTBlockNode
		}
		expr, err := expressions.Parse(node.Args)
		if err != nil {
			return nil, err
		}
		if !polarity {
			expr = expressions.Not(expr)
		}
		branches := []branchRec{
			{expr, &node},
		}
		for _, c := range node.Branches {
			test := expressions.Constant(true)
			switch c.Name {
			case "else":
			// TODO parse error if this isn't the last branch
			case "elsif":
				t, err := expressions.Parse(c.Args)
				if err != nil {
					return nil, err
				}
				test = t
			default:
			}
			branches = append(branches, branchRec{test, c})
		}
		return func(w io.Writer, ctx chunks.RenderContext) error {
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
