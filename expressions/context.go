package expressions

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	Get(string) interface{}
	Set(string, interface{})
}

type context struct {
	vars   map[string]interface{}
	copied bool
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}) Context {
	return &context{vars, false}
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return c.vars[name]
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.vars[name] = value
}

// Loop describes the result of parsing and then evaluating a loop statement.
type Loop struct {
	Name string
	Expr interface{}
}
