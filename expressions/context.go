package expressions

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	Get(string) interface{}
	Set(string, interface{})
	Filters() *filterDictionary
}

type context struct {
	bindings map[string]interface{}
	Config
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}, s Config) Context {
	return &context{vars, s}
}

func (c *context) Filters() *filterDictionary {
	return c.filters
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return ToLiquid(c.bindings[name])
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.bindings[name] = value
}
