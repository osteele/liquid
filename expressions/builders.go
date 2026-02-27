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

// filterArgs holds both positional and keyword arguments for a filter.
type filterArgs struct {
	positional []valueFn
	keyword    []keywordArg
}

// keywordArg represents a named argument (e.g., allow_false: true).
type keywordArg struct {
	name string
	val  valueFn
}

func makeFilter(fn valueFn, name string, args *filterArgs) valueFn {
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
