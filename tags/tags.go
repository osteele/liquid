// Package tags defines the standard Liquid tags.
package tags

import (
	"bytes"
	"io"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/generics"
)

// DefineStandardTags defines the standard Liquid tags.
func DefineStandardTags() {
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	loopTags := []string{"break", "continue", "cycle"}
	chunks.DefineTag("break", breakTag)
	chunks.DefineTag("continue", continueTag)
	chunks.DefineStartTag("capture").Parser(captureTagParser)
	chunks.DefineStartTag("case").Branch("when").Parser(caseTagParser)
	chunks.DefineStartTag("comment")
	chunks.DefineStartTag("for").Governs(loopTags).Parser(loopTagParser)
	chunks.DefineStartTag("if").Branch("else").Branch("elsif").Parser(ifTagParser(true))
	chunks.DefineStartTag("raw")
	chunks.DefineStartTag("tablerow").Governs(loopTags)
	chunks.DefineStartTag("unless").SameSyntaxAs("if").Parser(ifTagParser(false))
}

func captureTagParser(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	// TODO verify syntax
	varname := node.Parameters
	return func(w io.Writer, ctx chunks.Context) error {
		buf := new(bytes.Buffer)
		if err := ctx.RenderASTSequence(buf, node.Body); err != nil {
			return err
		}
		ctx.Set(varname, buf.String())
		return nil
	}, nil
}

func caseTagParser(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	// TODO parse error on non-empty node.Body
	// TODO case can include an else
	expr, err := chunks.MakeExpressionValueFn(node.Parameters)
	if err != nil {
		return nil, err
	}
	type caseRec struct {
		fn   func(chunks.Context) (interface{}, error)
		node *chunks.ASTControlTag
	}
	cases := []caseRec{}
	for _, branch := range node.Branches {
		bfn, err := chunks.MakeExpressionValueFn(branch.Parameters)
		if err != nil {
			return nil, err
		}
		cases = append(cases, caseRec{bfn, branch})
	}
	return func(w io.Writer, ctx chunks.Context) error {
		value, err := expr(ctx)
		if err != nil {
			return err
		}
		for _, branch := range cases {
			b, err := branch.fn(ctx)
			if err != nil {
				return err
			}
			if generics.Equal(value, b) {
				return ctx.RenderASTSequence(w, branch.node.Body)
			}
		}
		return nil
	}, nil
}

func ifTagParser(polarity bool) func(chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	return func(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
		type branchRec struct {
			test func(chunks.Context) (interface{}, error)
			body []chunks.ASTNode
		}
		expr, err := chunks.MakeExpressionValueFn(node.Parameters)
		if err != nil {
			return nil, err
		}
		if !polarity {
			expr = chunks.Negate(expr)
		}
		branches := []branchRec{
			{expr, node.Body},
		}
		for _, c := range node.Branches {
			test := chunks.True
			switch c.Name {
			case "else":
			// TODO parse error if this isn't the last branch
			case "elsif":
				t, err := chunks.MakeExpressionValueFn(c.Parameters)
				if err != nil {
					return nil, err
				}
				test = t
			default:
			}
			branches = append(branches, branchRec{test, c.Body})
		}
		return func(w io.Writer, ctx chunks.Context) error {
			for _, b := range branches {
				value, err := b.test(ctx)
				if err != nil {
					return err
				}
				if value != nil && value != false {
					return ctx.RenderASTSequence(w, b.body)
				}
			}
			return nil
		}, nil
	}
}
