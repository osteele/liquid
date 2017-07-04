package tags

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

var errLoopContinueLoop = fmt.Errorf("continue outside a loop")
var errLoopBreak = fmt.Errorf("break outside a loop")

func breakTag(parameters string) (func(io.Writer, render.RenderContext) error, error) {
	return func(io.Writer, render.RenderContext) error {
		return errLoopBreak
	}, nil
}

func continueTag(parameters string) (func(io.Writer, render.RenderContext) error, error) {
	return func(io.Writer, render.RenderContext) error {
		return errLoopContinueLoop
	}, nil
}

func parseLoopExpression(source string) (expressions.Expression, error) {
	expr, err := expressions.Parse("%loop " + source)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func loopTagParser(node render.ASTBlock) (func(io.Writer, render.RenderContext) error, error) { // nolint: gocyclo
	expr, err := parseLoopExpression(node.Args)
	if err != nil {
		return nil, err
	}
	return func(w io.Writer, ctx render.RenderContext) error {
		val, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		loop := val.(*expressions.Loop)
		rt := reflect.ValueOf(loop.Expr)
		switch rt.Kind() {
		case reflect.Map:
			array := make([]interface{}, 0, rt.Len())
			for _, k := range rt.MapKeys() {
				array = append(array, k.Interface())
			}
			rt = reflect.ValueOf(array)
		case reflect.Array, reflect.Slice:
		// proceed
		default:
			return nil
		}
		if loop.Offset > 0 {
			if loop.Offset > rt.Len() {
				return nil
			}
			rt = rt.Slice(loop.Offset, rt.Len())
		}
		length := rt.Len()
		if loop.Limit != nil {
			length = *loop.Limit
			if length > rt.Len() {
				length = rt.Len()
			}
		}
		const forloopName = "forloop"
		defer func(index, forloop interface{}) {
			ctx.Set(forloopName, index)
			ctx.Set(loop.Variable, forloop)
		}(ctx.Get(forloopName), ctx.Get(loop.Variable))
		// for forloop variable
		for i := 0; i < length; i++ {
			j := i
			if loop.Reversed {
				j = rt.Len() - 1 - i
			}
			ctx.Set(loop.Variable, rt.Index(j).Interface())
			ctx.Set(forloopName, map[string]interface{}{
				"first":   i == 0,
				"last":    i == length-1,
				"index":   i + 1,
				"index0":  i,
				"rindex":  length - i,
				"rindex0": length - i - 1,
				"length":  length,
			})
			err := ctx.RenderChildren(w)
			if err == errLoopBreak {
				break
			}
			if err == errLoopContinueLoop {
				continue
			}
			if err != nil {
				return err
			}
		}
		return nil
	}, nil
}
