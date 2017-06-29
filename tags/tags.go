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
	chunks.DefineControlTag("capture").Parser(captureTagParser)
	chunks.DefineControlTag("case").Branch("when").Parser(caseTagParser)
	chunks.DefineControlTag("comment")
	chunks.DefineControlTag("for").Governs(loopTags).Parser(loopTagParser)
	chunks.DefineControlTag("if").Branch("else").Branch("elsif").Parser(ifTagParser(true))
	chunks.DefineControlTag("raw")
	chunks.DefineControlTag("tablerow").Governs(loopTags)
	chunks.DefineControlTag("unless").SameSyntaxAs("if").Parser(ifTagParser(false))
}

func captureTagParser(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	// TODO verify syntax
	varname := node.Args
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
	expr, err := chunks.MakeExpressionValueFn(node.Args)
	if err != nil {
		return nil, err
	}
	type caseRec struct {
		fn   func(chunks.Context) (interface{}, error)
		node *chunks.ASTControlTag
	}
	cases := []caseRec{}
	for _, branch := range node.Branches {
		bfn, err := chunks.MakeExpressionValueFn(branch.Args)
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

func constTrueExpr(_ chunks.Context) (interface{}, error) { return true, nil }

func negateExpr(f func(chunks.Context) (interface{}, error)) func(chunks.Context) (interface{}, error) {
	return func(ctx chunks.Context) (interface{}, error) {
		value, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return value == nil || value == false, nil
	}
}

func ifTagParser(polarity bool) func(chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	return func(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
		type branchRec struct {
			test func(chunks.Context) (interface{}, error)
			body []chunks.ASTNode
		}
		expr, err := chunks.MakeExpressionValueFn(node.Args)
		if err != nil {
			return nil, err
		}
		if !polarity {
			expr = negateExpr(expr)
		}
		branches := []branchRec{
			{expr, node.Body},
		}
		for _, c := range node.Branches {
			test := constTrueExpr
			switch c.Tag {
			case "else":
			// TODO parse error if this isn't the last branch
			case "elsif":
				t, err := chunks.MakeExpressionValueFn(c.Args)
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
