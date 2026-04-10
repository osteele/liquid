package expressions

// ContextFilterFn is the signature for context-aware filters.
// Unlike regular filters (dispatched via reflection), context-aware filters
// receive the full expression evaluation context as their first argument,
// which allows them to evaluate arbitrary sub-expressions per item.
//
// value is the LHS of the pipe. args contains all evaluated positional
// arguments (strings, numbers, etc.) — NamedArgs are included as-is.
type ContextFilterFn func(ctx Context, value any, args []any) (any, error)

// FilterHook is called after each filter is successfully applied during expression evaluation.
// It receives the filter name, input value before the filter, positional arguments, and output value.
type FilterHook func(name string, input any, args []any, output any)

// ComparisonHook is called each time a leaf binary comparison operator is evaluated
// inside an expression (==, !=, >, <, >=, <=, contains).
// op is the operator string; left and right are the evaluated operand values;
// result is the boolean outcome of the comparison.
// Used by the audit/trace subsystem to record comparison details in condition branches.
type ComparisonHook func(op string, left, right any, result bool)

// Config holds configuration information for expression interpretation.
type Config struct {
	filters        map[string]any
	contextFilters map[string]ContextFilterFn
	LaxFilters     bool
	// FilterHook is called after each filter step, when non-nil.
	// It is used by the audit/trace subsystem to record the filter pipeline.
	FilterHook FilterHook
	// ComparisonHook is called after each leaf comparison evaluation, when non-nil.
	// It is used by the audit/trace subsystem to record comparison details.
	ComparisonHook ComparisonHook
	// ComparisonGroupBeginHook is called before evaluating the operands of an
	// and/or logical group expression. Used to build GroupTrace tree nodes.
	ComparisonGroupBeginHook func()
	// ComparisonGroupEndHook is called after evaluating an and/or logical group.
	// op is "and" or "or"; result is the boolean outcome.
	ComparisonGroupEndHook func(op string, result bool)
	// TypeMismatchHook is called when a comparison is made between values of
	// incompatible types (e.g. string vs int). op is the operator, left/right
	// are the raw interface values.
	TypeMismatchHook func(op string, left, right any)
	// NilDereferenceHook is called when a property access on a nil/non-existing
	// intermediate node is detected in a chained path like a.b.c.
	// object is the value that was nil, property is the name that was accessed.
	NilDereferenceHook func(object, property string)
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

// HasFilter reports whether the named filter is registered.
func (c *Config) HasFilter(name string) bool {
	if _, ok := c.filters[name]; ok {
		return true
	}
	_, ok := c.contextFilters[name]
	return ok
}
