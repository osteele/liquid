//go:generate ragel -Z scanner.rl
//go:generate gofmt -w scanner.go
//go:generate goyacc expressions.y

package expression

import "fmt"

// These strings match lexer tokens.
const (
	AssignStatementSelector = "%assign "
	CycleStatementSelector  = "{%cycle "
	LoopStatementSelector   = "%loop "
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
		return nil, ParseError(fmt.Errorf("parse error in %q", source).Error())
	}
	return &expression{lexer.val}, nil
}

// ParseStatement parses an statement into an Expression that can evaluated to return a
// structure specific to the statement.
func ParseStatement(sel, source string) (expr Expression, err error) {
	return Parse(sel + source)
}

// EvaluateString is a wrapper for Parse and Evaluate.
func EvaluateString(source string, ctx Context) (interface{}, error) {
	expr, err := Parse(source)
	if err != nil {
		return nil, err
	}
	return expr.Evaluate(ctx)
}
