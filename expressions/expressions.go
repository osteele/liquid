// Package expressions is an internal package that parses and evaluates the expression language.
//
// This is the language that is used inside Liquid object and tags; e.g. "a.b[c]" in {{ a.b[c] }}, and "pages = site.pages | reverse" in {% assign pages = site.pages | reverse %}.
package expressions

import (
	"fmt"
	"runtime/debug"

	"github.com/osteele/liquid/values"
)

// TODO Expression and Closure are confusing names.

// An Expression is a compiled expression.
type Expression interface {
	// Evaluate evaluates an expression in a context.
	Evaluate(ctx Context) (interface{}, error)
}

// A Closure is an expression within a lexical environment.
// A closure may refer to variables that are not defined in the
// environment. (Therefore it's not a technically a closure.)
type Closure interface {
	// Bind creates a new closure with a new binding.
	Bind(name string, value interface{}) Closure
	Evaluate() (interface{}, error)
}

type closure struct {
	expr    Expression
	context Context
}

func (c closure) Bind(name string, value interface{}) Closure {
	ctx := c.context.Clone()
	ctx.Set(name, value)
	return closure{c.expr, ctx}
}

func (c closure) Evaluate() (interface{}, error) {
	return c.expr.Evaluate(c.context)
}

type expression struct {
	evaluator func(Context) values.Value
}

func (e expression) Evaluate(ctx Context) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case values.TypeError:
				err = e
			case InterpreterError:
				err = e
			case UndefinedFilter:
				err = e
			case FilterError:
				err = e
			case error:
				panic(&rethrownError{e, debug.Stack()})
			default:
				panic(r)
			}
		}
	}()
	return e.evaluator(ctx).Interface(), nil
}

// rethrownError is for use in a re-thrown error from panic recovery.
// When printed, it prints the original stacktrace.
// This works around a frequent problem, that it's difficult to debug an error inside a filter
// or ToLiquid implementation because Evaluate's recover replaces the stacktrace.
type rethrownError struct {
	cause error
	stack []byte
}

func (e *rethrownError) Error() string {
	return fmt.Sprintf("%s\nOriginal stacktrace:\n%s\n", e.cause, string(e.stack))
}

func (e *rethrownError) Cause() error {
	return e.cause
}
