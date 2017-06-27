package expressions

import "github.com/osteele/liquid/errors"

func (e expression) Evaluate(ctx Context) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case InterpreterError:
				err = e
			case UnimplementedError:
				err = e
			case errors.UndefinedFilter:
				err = e
			default:
				panic(r)
			}
		}
	}()
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
