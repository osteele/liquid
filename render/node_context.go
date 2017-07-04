package render

import (
	"github.com/osteele/liquid/expression"
)

// nodeContext is the evaluation context for chunk AST rendering.
type nodeContext struct {
	bindings map[string]interface{}
	config   Config
}

// newNodeContext creates a new evaluation context.
func newNodeContext(scope map[string]interface{}, s Config) nodeContext {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := map[string]interface{}{}
	for k, v := range scope {
		vars[k] = v
	}
	return nodeContext{vars, s}
}

// Clone makes a copy of a context, with copied bindings.
func (c nodeContext) Clone() nodeContext {
	bindings := map[string]interface{}{}
	for k, v := range c.bindings {
		bindings[k] = v
	}
	return nodeContext{bindings, c.config}
}

// Evaluate evaluates an expression within the template context.
func (c nodeContext) Evaluate(expr expression.Expression) (out interface{}, err error) {
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
	return expr.Evaluate(expression.NewContext(c.bindings, c.config.Config))
}
