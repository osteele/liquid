package expressions

import (
	"github.com/osteele/liquid/values"
)

func makeRangeExpr(startFn, endFn func(Context) values.Value) func(Context) values.Value {
	return func(ctx Context) values.Value {
		a := startFn(ctx).Int()
		b := endFn(ctx).Int()
		return values.ValueOf(values.NewRange(a, b))
	}
}

func makeContainsExpr(e1, e2 func(Context) values.Value) func(Context) values.Value {
	return func(ctx Context) values.Value {
		return values.ValueOf(e1(ctx).Contains(e2(ctx)))
	}
}

func makeFilter(fn valueFn, name string, args []valueFn) valueFn {
	return func(ctx Context) values.Value {
		result, err := ctx.ApplyFilter(name, fn, args)
		if err != nil {
			panic(FilterError{
				FilterName: name,
				Err:        err,
			})
		}
		return values.ValueOf(result)
	}
}

func makeIndexExpr(sequenceFn, indexFn func(Context) values.Value) func(Context) values.Value {
	return func(ctx Context) values.Value {
		return sequenceFn(ctx).IndexValue(indexFn(ctx))
	}
}

func makeObjectPropertyExpr(objFn func(Context) values.Value, name string) func(Context) values.Value {
	index := values.ValueOf(name)
	return func(ctx Context) values.Value {
		return objFn(ctx).PropertyValue(index)
	}
}
