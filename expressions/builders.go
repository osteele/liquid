package expressions

import (
	"reflect"

	"github.com/osteele/liquid/evaluator"
)

func makeRangeExpr(startFn, endFn func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		var proto int
		b := evaluator.MustConvert(startFn(ctx), reflect.TypeOf(proto))
		e := evaluator.MustConvert(endFn(ctx), reflect.TypeOf(proto))
		return evaluator.NewRange(b.(int), e.(int))
	}
}

func makeContainsExpr(e1, e2 func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		s, ok := e2((ctx)).(string)
		if !ok {
			return false
		}
		return evaluator.ContainsString(e1(ctx), s)
	}
}

func makeFilter(fn valueFn, name string, args []valueFn) valueFn {
	return func(ctx Context) interface{} {
		result, err := ctx.ApplyFilter(name, fn, args)
		if err != nil {
			panic(err)
		}
		return result
	}
}

func makeIndexExpr(sequenceFn, indexFn func(Context) interface{}) func(Context) interface{} {
	return func(ctx Context) interface{} {
		return evaluator.Index(sequenceFn(ctx), indexFn(ctx))
	}
}

func makeObjectPropertyExpr(objFn func(Context) interface{}, name string) func(Context) interface{} {
	return func(ctx Context) interface{} {
		return evaluator.ObjectProperty(objFn(ctx), name)
	}
}
