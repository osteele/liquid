package tags

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/expression"
	"github.com/osteele/liquid/render"
)

const forloopVarName = "forloop"

var errLoopContinueLoop = fmt.Errorf("continue outside a loop")
var errLoopBreak = fmt.Errorf("break outside a loop")

func breakTag(string) (func(io.Writer, render.Context) error, error) {
	return func(_ io.Writer, ctx render.Context) error {
		return ctx.WrapError(errLoopBreak)
	}, nil
}

func continueTag(string) (func(io.Writer, render.Context) error, error) {
	return func(_ io.Writer, ctx render.Context) error {
		return ctx.WrapError(errLoopContinueLoop)
	}, nil
}

func cycleTag(args string) (func(io.Writer, render.Context) error, error) {
	stmt, err := expression.ParseStatement(expression.CycleStatementSelector, args)
	if err != nil {
		return nil, err
	}
	expr := stmt.Expression()
	return func(w io.Writer, ctx render.Context) error {
		value, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		array := value.([]interface{})
		if len(array) == 0 {
			return nil
		}
		loopVar := ctx.Get(forloopVarName)
		if loopVar == nil {
			return ctx.Errorf("cycle must be within a forloop")
		}
		// the next few lines could panic if the user spoofs us by creating their own loop object
		// “C++ protects against accident, not against fraud.” – Bjarne Stroustrup
		loopRec := loopVar.(map[string]interface{})
		cycleMap := loopRec[".cycles"].(map[string]int)
		group := ""
		n := cycleMap[group]
		cycleMap[group] = n + 1
		fmt.Println(cycleMap)
		_, err = w.Write([]byte(fmt.Sprint(array[n%len(array)])))
		return err
	}, nil
}

func parseLoopExpression(source string) (expression.Expression, error) {
	stmt, err := expression.ParseStatement(expression.LoopStatementSelector, source)
	if err != nil {
		return nil, err
	}
	return stmt.Expression(), nil
}

func loopTagParser(node render.BlockNode) (func(io.Writer, render.Context) error, error) { // nolint: gocyclo
	expr, err := parseLoopExpression(node.Args)
	if err != nil {
		return nil, err
	}
	return func(w io.Writer, ctx render.Context) error {
		val, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		loop := val.(*expression.Loop)
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
		defer func(index, forloop interface{}) {
			ctx.Set(forloopVarName, index)
			ctx.Set(loop.Variable, forloop)
		}(ctx.Get(forloopVarName), ctx.Get(loop.Variable))
		cycleMap := map[string]int{}
	loop:
		for i := 0; i < length; i++ {
			j := i
			if loop.Reversed {
				j = rt.Len() - 1 - i
			}
			ctx.Set(loop.Variable, rt.Index(j).Interface())
			ctx.Set(forloopVarName, map[string]interface{}{
				"first":   i == 0,
				"last":    i == length-1,
				"index":   i + 1,
				"index0":  i,
				"rindex":  length - i,
				"rindex0": length - i - 1,
				"length":  length,
				".cycles": cycleMap,
			})
			err := ctx.RenderChildren(w)
			switch {
			case err == nil:
			// fall through
			case err.Cause() == errLoopBreak:
				break loop
			case err.Cause() == errLoopContinueLoop:
				continue loop
			default:
				return err
			}
		}
		return nil
	}, nil
}
