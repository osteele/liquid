package expressions

func (e expression) Evaluate(ctx Context) (interface{}, error) {
	return e.evaluator(ctx), nil
}

// EvaluateExpr is a wrapper for Parse and Evaluate.
func EvaluateExpr(source string, ctx Context) (interface{}, error) {
	expr, err := Parse(source)
	if err != nil {
		return nil, err
	}
	return expr.Evaluate(ctx)
}
