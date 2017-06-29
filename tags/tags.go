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
	type branchRec struct {
		fn   func(chunks.Context) (interface{}, error)
		node *chunks.ASTControlTag
	}
	cases := []branchRec{}
	for _, branch := range node.Branches {
		bfn, err := chunks.MakeExpressionValueFn(branch.Args)
		if err != nil {
			return nil, err
		}
		cases = append(cases, branchRec{bfn, branch})
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
	// TODO parse error if the order of branches is other than ifelse*else?
	// TODO parse the tests into a table evaluator -> []AST
	return func(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
		expr, err := chunks.MakeExpressionValueFn(node.Args)
		if err != nil {
			return nil, err
		}
		return func(w io.Writer, ctx chunks.Context) error {
			val, err := expr(ctx)
			if err != nil {
				return err
			}
			if !polarity {
				val = (val == nil || val == false)
			}
			switch val {
			default:
				return ctx.RenderASTSequence(w, node.Body)
			case nil, false:
				for _, c := range node.Branches {
					switch c.Tag {
					case "else":
						val = true
					case "elsif":
						val, err = ctx.EvaluateExpr(c.Args)
						if err != nil {
							return err
						}
					}
					if val != nil && val != false {
						return ctx.RenderASTSequence(w, c.Body)
					}
				}
			}
			return nil
		}, nil
	}
}
