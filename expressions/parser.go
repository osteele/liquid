package expressions

import (
	"fmt"
)

// Loop describes the result of parsing and then evaluating a loop statement.
type Loop struct {
	Variable string
	Expr     interface{}
	loopModifiers
}

type loopModifiers struct {
	Limit    *int
	Offset   int
	Reversed bool
}

// ParseError represents a parse error.
type ParseError string

func (e ParseError) Error() string { return string(e) }

// Parse parses an expression string into an Expression.
func Parse(source string) (expr Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case ParseError:
				err = e
			case UndefinedFilter:
				err = e
			default:
				panic(r)
			}
		}
	}()
	lexer := newLexer([]byte(source + ";"))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %q", source)
	}
	return &expression{lexer.val}, nil
}

// EvaluateString is a wrapper for Parse and Evaluate.
func EvaluateString(source string, ctx Context) (interface{}, error) {
	expr, err := Parse(source)
	if err != nil {
		return nil, err
	}
	return expr.Evaluate(ctx)
}
