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

func NewContext(vars map[string]interface{}) Context {
	return &context{vars, false}
}

func (c *context) Get(name string) interface{} {
	return c.vars[name]
}

func (c *context) Set(name string, value interface{}) {
	// if !c.copied {
	// 	vs := map[string]interface{}{}
	// 	for k, v := range c.vars {
	// 		vs[k] = v
	// 	}
	// 	c.vars, c.copied = vs, true
	// }
	c.vars[name] = value
}

type Loop struct {
	Name string
	Expr interface{}
}
