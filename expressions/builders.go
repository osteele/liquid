package expressions

import (
	"github.com/osteele/liquid/evaluator"
)

func makeRangeExpr(startFn, endFn func(Context) evaluator.Value) func(Context) evaluator.Value {
	return func(ctx Context) evaluator.Value {
		a := startFn(ctx).Int()
		b := endFn(ctx).Int()
		return evaluator.ValueOf(evaluator.NewRange(a, b))
	}
}

func makeContainsExpr(e1, e2 func(Context) evaluator.Value) func(Context) evaluator.Value {
	return func(ctx Context) evaluator.Value {
		return evaluator.ValueOf(e1(ctx).Contains(e2(ctx)))
	}
}

func makeFilter(fn valueFn, name string, args []valueFn) valueFn {
	return func(ctx Context) evaluator.Value {
		result, err := ctx.ApplyFilter(name, fn, args)
		if err != nil {
			panic(FilterError{
				FilterName: name,
				Err:        err,
			})
		}
		return evaluator.ValueOf(result)
	}
}

func makeIndexExpr(sequenceFn, indexFn func(Context) evaluator.Value) func(Context) evaluator.Value {
	return func(ctx Context) evaluator.Value {
		return sequenceFn(ctx).IndexValue(indexFn(ctx))
	}
}

func makeObjectPropertyExpr(objFn func(Context) evaluator.Value, name string) func(Context) evaluator.Value {
	index := evaluator.ValueOf(name)
	return func(ctx Context) evaluator.Value {
		return objFn(ctx).PropertyValue(index)
	}
}
