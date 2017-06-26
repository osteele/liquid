package expressions

import (
	"fmt"
)

// Expression is a parsed expression.
type Expression interface {
	// Evaluate evaluates an expression in a context.
	Evaluate(ctx Context) (interface{}, error)
}

type expression struct {
	evaluator func(Context) interface{}
}

// Parse parses an expression string into an Expression.
func Parse(source string) (Expression, error) {
	lexer := newLexer([]byte(source + ";"))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %s", source)
	}
	return &expression{lexer.val}, nil
}

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
