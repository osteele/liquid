// Package expressions parses and evaluates the expression language.
//
// This is the language that is used inside Liquid object and tags; e.g. "a.b[c]" in {{ a.b[c] }}, and "pages = site.pages | reverse" in {% assign pages = site.pages | reverse %}.
package expressions

import (
	"github.com/osteele/liquid/evaluator"
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
	evaluator func(Context) evaluator.Value
}

func (e expression) Evaluate(ctx Context) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case evaluator.TypeError:
				err = e
			case InterpreterError:
				err = e
			case UndefinedFilter:
				err = e
			default:
				panic(r)
			}
		}
	}()
	return e.evaluator(ctx).Interface(), nil
}
