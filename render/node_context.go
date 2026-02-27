package render

import (
	"maps"

	"github.com/osteele/liquid/expressions"
)

// nodeContext provides the evaluation context for rendering the AST.
//
// This type has a clumsy name so that render.Context, in the public API, can
// have a clean name that doesn't stutter.
type nodeContext struct {
	bindings map[string]any
	config   Config
	exprCtx  expressions.Context
}

// newNodeContext creates a new evaluation context.
func newNodeContext(scope map[string]any, c Config) nodeContext {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := make(map[string]any, len(scope))
	maps.Copy(vars, scope)

	ctx := nodeContext{bindings: vars, config: c}
	ctx.exprCtx = expressions.NewContext(vars, c.Config.Config)
	return ctx
}

// Evaluate evaluates an expression within the template context.
func (c nodeContext) Evaluate(expr expressions.Expression) (out any, err error) {
	return expr.Evaluate(c.exprCtx)
}
