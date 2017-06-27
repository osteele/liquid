package expressions

import (
	"fmt"

	"github.com/osteele/liquid/errors"
)

// Loop describes the result of parsing and then evaluating a loop statement.
type Loop struct {
	Name string
	Expr interface{}
	LoopModifiers
}

type LoopModifiers struct {
	Reversed bool
}

type ParseError string

func (e ParseError) Error() string { return string(e) }

// Expression is a parsed expression.
type Expression interface {
	// Evaluate evaluates an expression in a context.
	Evaluate(ctx Context) (interface{}, error)
}

type expression struct {
	evaluator func(Context) interface{}
}

// Parse parses an expression string into an Expression.
func Parse(source string) (expr Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case ParseError:
				err = e
			case errors.UndefinedFilter:
				err = e
			default:
				panic(r)
			}
		}
	}()
	lexer := newLexer([]byte(source + ";"))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %s", source)
	}
	return &expression{lexer.val}, nil
}
