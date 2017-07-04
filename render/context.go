package render

import (
	"github.com/osteele/liquid/expression"
)

// Context is the evaluation context for chunk AST rendering.
type Context struct {
	bindings map[string]interface{}
	settings Config
}

// NewContext creates a new evaluation context.
func NewContext(scope map[string]interface{}, s Config) Context {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := map[string]interface{}{}
	for k, v := range scope {
		vars[k] = v
	}
	return Context{vars, s}
}

// Clone makes a copy of a context, with copied bindings.
func (c Context) Clone() Context {
	bindings := map[string]interface{}{}
	for k, v := range c.bindings {
		bindings[k] = v
	}
	return Context{bindings, c.settings}
}

// Evaluate evaluates an expression within the template context.
func (c Context) Evaluate(expr expression.Expression) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case expression.InterpreterError:
				err = e
			default:
				// fmt.Println(string(debug.Stack()))
				panic(e)
			}
		}
	}()
	return expr.Evaluate(expression.NewContext(c.bindings, c.settings.ExpressionConfig))
}
