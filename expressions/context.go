package expressions

import (
	"reflect"

	"github.com/osteele/liquid/values"
)

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, []valueFn) (any, error)
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
// If the raw binding implements values.ContextSetter, the expression context is
// injected into it before ToLiquid conversion — mirroring Ruby's context= setter.
func (ctx *context) Get(name string) any {
	raw := ctx.bindings[name]
	if cs, ok := raw.(values.ContextSetter); ok {
		cs.SetContext(ctx)
	}
	v := values.ToLiquid(raw)
	// If ToLiquid returned a different value (wrapper resolved), inject context there too.
	// Only compare when both are comparable to avoid panic on slice/map values.
	if v != nil && !areSamePointer(raw, v) {
		if cs, ok := v.(values.ContextSetter); ok {
			cs.SetContext(ctx)
		}
	}
	return v
}

// areSamePointer reports whether a and b are the same underlying pointer.
// It handles pointer-typed values without triggering panic for uncomparable
// types (slice, map, func).
func areSamePointer(a, b any) bool {
	if a == nil || b == nil {
		return false
	}
	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	if ra.Kind() != reflect.Ptr || rb.Kind() != reflect.Ptr {
		return false
	}
	return ra.Pointer() == rb.Pointer()
}

// Set sets a variable value in the expression context.
func (ctx *context) Set(name string, value any) {
	ctx.bindings[name] = value
}
