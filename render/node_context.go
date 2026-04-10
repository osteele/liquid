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
	// Globals have the lowest priority: scope bindings win over globals.
	vars := make(map[string]any, len(c.Globals)+len(scope))
	maps.Copy(vars, c.Globals)
	maps.Copy(vars, scope)

	ctx := nodeContext{bindings: vars, config: c}
	ctx.exprCtx = expressions.NewContext(vars, c.Config.Config)
	return ctx
}

// SpawnIsolated creates a new node context that inherits the config but NOT
// the parent bindings. Only the explicitly provided bindings are visible,
// plus any globals defined on the engine config (which always propagate).
// This is used by the {% render %} tag and layout/block inheritance, which
// must not see variables from the calling scope.
func (c nodeContext) SpawnIsolated(bindings map[string]any) nodeContext {
	// Globals have the lowest priority; explicit bindings win.
	vars := make(map[string]any, len(c.config.Globals)+len(bindings))
	maps.Copy(vars, c.config.Globals)
	maps.Copy(vars, bindings)
	child := nodeContext{bindings: vars, config: c.config}
	child.exprCtx = expressions.NewContext(vars, c.config.Config.Config)
	return child
}

// Evaluate evaluates an expression within the template context.
func (c nodeContext) Evaluate(expr expressions.Expression) (out any, err error) {
	return expr.Evaluate(c.exprCtx)
}
