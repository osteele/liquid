package expressions

import (
	"fmt"

	"github.com/osteele/liquid/errors"
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

type ParseError string

func (e ParseError) Error() string { return string(e) }

type UnimplementedError string

func (e UnimplementedError) Error() string {
	return fmt.Sprintf("unimplemented %s", string(e))
}

// Parse parses an expression string into an Expression.
func Parse(source string) (expr Expression, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case ParseError:
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
	lexer := newLexer([]byte(source + ";"))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %s", source)
	}
	return &expression{lexer.val}, nil
}
