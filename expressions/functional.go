package expressions

type expressionWrapper struct {
	fn func(ctx Context) (interface{}, error)
}

func (w expressionWrapper) Evaluate(ctx Context) (interface{}, error) {
	return w.fn(ctx)
}

// Constant creates an expression that returns a constant value.
func Constant(k interface{}) Expression {
	return expressionWrapper{
		func(_ Context) (interface{}, error) {
			return k, nil
		},
	}
}

// Not creates an expression that returns ! of the wrapped expression.
func Not(e Expression) Expression {
	return expressionWrapper{
		func(ctx Context) (interface{}, error) {
			value, err := e.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			return (value == nil || value == false), nil
		},
	}
}
