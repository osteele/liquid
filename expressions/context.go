package expressions

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	Get(string) interface{}
	Set(string, interface{})
	Filters() *FilterDictionary
}

type context struct {
	bindings map[string]interface{}
	Settings
}

type Settings struct {
	filters *FilterDictionary
}

func NewSettings() Settings {
	return Settings{NewFilterDictionary()}
}

func (s Settings) AddFilter(name string, fn interface{}) {
	s.filters.AddFilter(name, fn)
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]interface{}, s Settings) Context {
	return &context{vars, s}
}

func (c *context) Filters() *FilterDictionary {
	return c.filters
}

// Get looks up a variable value in the expression context.
func (c *context) Get(name string) interface{} {
	return c.bindings[name]
}

// Set sets a variable value in the expression context.
func (c *context) Set(name string, value interface{}) {
	c.bindings[name] = value
}
