package expressions

import "github.com/osteele/liquid/values"

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, *filterArgs) (any, error)
	// Clone returns a copy with a new variable binding map
	// (so that copy.Set does effect the source context.)
	Clone() Context
	Get(string) any
	Set(string, any)
}

type context struct {
	Config

	bindings map[string]any
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]any, cfg Config) Context {
	return &context{cfg, vars}
}

func (ctx *context) Clone() Context {
	bindings := map[string]any{}
	for k, v := range ctx.bindings {
		bindings[k] = v
	}

	return &context{ctx.Config, bindings}
}

// Get looks up a variable value in the expression context.
func (ctx *context) Get(name string) any {
	value, ok := ctx.bindings[name]
	if !ok && ctx.Config.StrictVariables {
		panic(InterpreterError("undefined variable"))
	}

	return values.ToLiquid(value)
}

// Set sets a variable value in the expression context.
func (ctx *context) Set(name string, value any) {
	ctx.bindings[name] = value
}

func (ctx *context) StrictVariables() bool {
	return ctx.Config.StrictVariables
}

func strictVariables(ctx Context) bool {
	configured, ok := ctx.(interface{ StrictVariables() bool })
	return ok && configured.StrictVariables()
}
