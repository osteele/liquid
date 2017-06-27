package tags

import (
	"io"

	"github.com/osteele/liquid/chunks"
)

func DefineStandardTags() {
	loopTags := []string{"break", "continue", "cycle"}
	chunks.DefineControlTag("comment")
	chunks.DefineControlTag("if").Branch("else").Branch("elsif").Action(ifTagAction(true))
	chunks.DefineControlTag("unless").SameSyntaxAs("if").Action(ifTagAction(false))
	chunks.DefineControlTag("case").Branch("when")
	chunks.DefineControlTag("for").Governs(loopTags)
	chunks.DefineControlTag("tablerow").Governs(loopTags)
	chunks.DefineControlTag("capture")
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
