package expressions

// ContextFilterFn is the signature for context-aware filters.
// Unlike regular filters (dispatched via reflection), context-aware filters
// receive the full expression evaluation context as their first argument,
// which allows them to evaluate arbitrary sub-expressions per item.
//
// value is the LHS of the pipe. args contains all evaluated positional
// arguments (strings, numbers, etc.) — NamedArgs are included as-is.
type ContextFilterFn func(ctx Context, value any, args []any) (any, error)

// Config holds configuration information for expression interpretation.
type Config struct {
	filters        map[string]any
	contextFilters map[string]ContextFilterFn
	LaxFilters     bool
}

// NewConfig creates a new Config.
func NewConfig() Config {
	return Config{}
}

// AddContextFilter registers a context-aware filter.
// The filter function receives the current expression Context as its first
// argument, enabling it to evaluate sub-expressions (e.g., for _exp filters).
func (c *Config) AddContextFilter(name string, fn ContextFilterFn) {
	if c.contextFilters == nil {
		c.contextFilters = make(map[string]ContextFilterFn)
	}
	c.contextFilters[name] = fn
}
