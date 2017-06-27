// Package tags defines the standard Liquid tags.
package tags

import (
	"io"

	"github.com/osteele/liquid/chunks"
)

// DefineStandardTags defines the standard Liquid tags.
func DefineStandardTags() {
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	loopTags := []string{"break", "continue", "cycle"}
	chunks.DefineControlTag("capture")
	chunks.DefineControlTag("case").Branch("when")
	chunks.DefineControlTag("comment")
	chunks.DefineControlTag("for").Governs(loopTags).Action(loopTag)
	chunks.DefineControlTag("if").Branch("else").Branch("elsif").Action(ifTagAction(true))
	chunks.DefineControlTag("raw")
	chunks.DefineControlTag("tablerow").Governs(loopTags)
	chunks.DefineControlTag("unless").SameSyntaxAs("if").Action(ifTagAction(false))
}

func ifTagAction(polarity bool) func(chunks.ASTControlTag) func(io.Writer, chunks.Context) error {
	return func(node chunks.ASTControlTag) func(io.Writer, chunks.Context) error {
		expr, err := chunks.MakeExpressionValueFn(node.Args)
		if err != nil {
			return func(io.Writer, chunks.Context) error { return err }
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
		}
	}
}
