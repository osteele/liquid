package render

import (
	"github.com/osteele/liquid/expressions"
)

// nodeContext provides the evaluation context for rendering the AST.
//
// This type has a clumsy name so that render.Context, in the public API, can
// have a clean name that doesn't stutter.
type nodeContext struct {
	bindings map[string]interface{}
	config   Config
}

// newNodeContext creates a new evaluation context.
func newNodeContext(scope map[string]interface{}, c Config) nodeContext {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := map[string]interface{}{}
	for k, v := range scope {
		vars[k] = v
	}
	return nodeContext{vars, c}
}

// Evaluate evaluates an expression within the template context.
func (c nodeContext) Evaluate(expr expressions.Expression) (out interface{}, err error) {
	return expr.Evaluate(expressions.NewContext(c.bindings, c.config.Config.Config))
}
