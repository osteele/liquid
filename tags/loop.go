package tags

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/expressions"
)

var errLoopContinueLoop = fmt.Errorf("continue outside a loop")
var errLoopBreak = fmt.Errorf("break outside a loop")

func breakTag(parameters string) (func(io.Writer, chunks.Context) error, error) {
	return func(io.Writer, chunks.Context) error {
		return errLoopBreak
	}, nil
}

func continueTag(parameters string) (func(io.Writer, chunks.Context) error, error) {
	return func(io.Writer, chunks.Context) error {
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

func loopTagParser(node chunks.ASTControlTag) (func(io.Writer, chunks.Context) error, error) {
	expr, err := parseLoopExpression(node.Parameters)
	if err != nil {
		return nil, err
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
		start := loop.Offset
		limit := rt.Len()
		if loop.Limit != nil {
			limit = *loop.Limit
		}
		const forloopName = "forloop"
		defer func(index, forloop interface{}) {
			ctx.Set(forloopName, index)
			ctx.Set(loop.Variable, forloop)
		}(ctx.Get(forloopName), ctx.Get(loop.Variable))
		// for forloop variable
		var (
			first  = true
			index  = 1
			length = limit
		)
		for i := start; i < rt.Len(); i++ {
			if limit == 0 {
				break
			}
			limit--
			j := i
			if loop.Reversed {
				j = rt.Len() - 1 - i
			}
			ctx.Set(loop.Variable, rt.Index(j).Interface())
			ctx.Set(forloopName, map[string]interface{}{
				"first":   first,
				"last":    limit == 0,
				"index":   index,
				"index0":  index - 1,
				"rindex":  length + 1 - index,
				"rindex0": length - index,
				"length":  length,
			})
			first, index = false, index+1
			err := ctx.RenderASTSequence(w, node.Body)
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
