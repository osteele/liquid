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

func makeNamedArgFn(name string, valFn valueFn) valueFn {
	return func(ctx Context) values.Value {
		return values.ValueOf(NamedArg{Name: name, Value: valFn(ctx).Interface()})
	}
}

// makeVariableIndirectionExpr implements the Ruby `{{ [varname] }}` syntax:
// it evaluates the inner expression to a string and uses that string as the
// variable name to look up in the context.
func makeVariableIndirectionExpr(keyFn func(Context) values.Value) func(Context) values.Value {
	return func(ctx Context) values.Value {
		key := keyFn(ctx)
		name, ok := key.Interface().(string)
		if !ok {
			return values.ValueOf(nil)
		}
		return values.ValueOf(ctx.Get(name))
	}
}
