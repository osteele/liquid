package expression

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, []valueFn) interface{}
	Get(string) interface{}
	Set(string, interface{})
}

type context struct {
	Config
	bindings map[string]interface{}
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}, s Config) Context {
	return &context{s, vars}
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return ToLiquid(c.bindings[name])
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.bindings[name] = value
}
