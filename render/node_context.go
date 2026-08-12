package render

import (
	"bytes"
	"maps"

	"github.com/osteele/liquid/expressions"
)

// nodeContext provides the evaluation context for rendering the AST.
//
// This type has a clumsy name so that render.Context, in the public API, can
// have a clean name that doesn't stutter.
type nodeContext struct {
	bindings     map[string]any
	config       Config
	exprCtx      expressions.Context
	partialCache map[string]cachedPartial
}

type cachedPartial struct {
	source []byte
	root   Node
}

// newNodeContext creates a new evaluation context.
func newNodeContext(scope map[string]any, c Config) *nodeContext {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := make(map[string]any, len(scope))
	maps.Copy(vars, scope)

	c.Config.StrictVariables = c.StrictVariables
	ctx := nodeContext{
		bindings: vars,
		config:   c,
	}
	ctx.exprCtx = expressions.NewContext(vars, c.Config.Config)
	return &ctx
}

func (c *nodeContext) child(scope map[string]any) *nodeContext {
	child := newNodeContext(scope, c.config)
	child.partialCache = c.partialCache
	return child
}

func (c *nodeContext) cachedPartial(filename string, source []byte) (Node, bool) {
	entry, ok := c.partialCache[filename]
	return entry.root, ok && bytes.Equal(entry.source, source)
}

func (c *nodeContext) cachePartial(filename string, source []byte, root Node) {
	if c.partialCache == nil {
		c.partialCache = make(map[string]cachedPartial)
	}
	c.partialCache[filename] = cachedPartial{
		source: bytes.Clone(source),
		root:   root,
	}
}

// Evaluate evaluates an expression within the template context.
func (c *nodeContext) Evaluate(expr expressions.Expression) (out any, err error) {
	return expr.Evaluate(c.exprCtx)
}
