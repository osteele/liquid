package expressions

type wrapper struct {
	fn func(ctx Context) (interface{}, error)
}

func (w wrapper) Evaluate(ctx Context) (interface{}, error) {
	return w.fn(ctx)
}

// True returns the same value each time.
func Constant(k interface{}) Expression {
	return wrapper{
		func(_ Context) (interface{}, error) {
			return k, nil
		},
	}
}

// Negate negates its argument.
func Negate(e Expression) Expression {
	return wrapper{
		func(ctx Context) (interface{}, error) {
			value, err := e.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			return (value == nil || value == false), nil
		},
	}
}
