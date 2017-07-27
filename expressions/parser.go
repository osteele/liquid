//go:generate ragel -Z scanner.rl
//go:generate gofmt -w scanner.go
//go:generate goyacc expressions.y

package expressions

import (
	"fmt"

	"github.com/osteele/liquid/values"
)

type parseValue struct {
	Assignment
	Cycle
	Loop
	When
	val func(Context) values.Value
}

// SyntaxError represents a syntax error. The yacc-generated compiler
// doesn't use error returns; this lets us recognize them.
type SyntaxError string

func (e SyntaxError) Error() string { return string(e) }

// Parse parses an expression string into an Expression.
func Parse(source string) (expr Expression, err error) {
	p, err := parse(source)
	if err != nil {
		return nil, err
	}
	return &expression{p.val}, nil
}

func parse(source string) (p *parseValue, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case SyntaxError:
				err = e
			case UndefinedFilter:
				err = e
			default:
				panic(r)
			}
		}
	}()
	// FIXME hack to recognize EOF
	lex := newLexer([]byte(source + ";"))
	n := yyParse(lex)
	if n != 0 {
		return nil, SyntaxError(fmt.Errorf("syntax error in %q", source).Error())
	}
	return &lex.parseValue, nil
}

// EvaluateString is a wrapper for Parse and Evaluate.
func EvaluateString(source string, ctx Context) (interface{}, error) {
	expr, err := Parse(source)
	if err != nil {
		return nil, err
	}
	return expr.Evaluate(ctx)
}
