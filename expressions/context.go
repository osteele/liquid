package expressions

import "github.com/osteele/liquid/values"

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, []valueFn) (interface{}, error)
	// Clone returns a copy with a new variable binding map
	// (so that copy.Set does effect the source context.)
	Clone() Context
	Get(string) interface{}
	Set(string, interface{})
}

type context struct {
	Config
	bindings map[string]interface{}
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}, cfg Config) Context {
	return &context{cfg, vars}
}

func (c *context) Clone() Context {
	bindings := map[string]interface{}{}
	for k, v := range c.bindings {
		bindings[k] = v
	}
	return &context{c.Config, bindings}
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return values.ToLiquid(c.bindings[name])
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.bindings[name] = value
}
