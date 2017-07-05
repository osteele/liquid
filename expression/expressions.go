// Package expression parses and evaluates the expression language that is used in objects {{a.b[c]}} and tag parameters {%assign pages = site.pages | reverse%}.
package expression

import (
	"github.com/osteele/liquid/evaluator"
)

// Expression is a parsed expression.
type Expression interface {
	// Evaluate evaluates an expression in a context.
	Evaluate(ctx Context) (interface{}, error)
}

// Closure binds an environment.
type Closure interface {
	Bind(name string, value interface{}) Closure
	Evaluate() (interface{}, error)
}

type closure struct {
	expr    Expression
	context Context
}

func (c closure) Bind(name string, value interface{}) Closure {
	// TODO create a new context
	c.context.Set(name, value)
	return c
}

func (c closure) Evaluate() (interface{}, error) {
	return c.expr.Evaluate(c.context)
}

type expression struct {
	evaluator func(Context) interface{}
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
	return e.evaluator(ctx), nil
}
