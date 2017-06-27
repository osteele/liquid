package tags

import (
	"io"
	"reflect"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/expressions"
)

func parseLoop(source string) (expressions.Expression, error) {
	expr, err := expressions.Parse("%loop " + source)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func loopTag(node chunks.ASTControlTag) func(io.Writer, chunks.Context) error {
	expr, err := parseLoop(node.Args)
	if err != nil {
		return func(io.Writer, chunks.Context) error { return err }
	}
	return func(w io.Writer, ctx chunks.Context) error {
		val, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		loop := val.(*expressions.Loop)
		rt := reflect.ValueOf(loop.Expr)
		if rt.Kind() != reflect.Array && rt.Kind() != reflect.Slice {
			return nil
		}
		for i := 0; i < rt.Len(); i++ {
			ctx.Set(loop.Name, rt.Index(i).Interface())
			err := ctx.RenderASTSequence(w, node.Body)
			if err != nil {
				return err
			}
		}
		ctx.Set(loop.Name, nil)
		return nil
	}
}
