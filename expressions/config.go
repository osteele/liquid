package expressions

import gocontext "context"

// Config holds configuration information for expression interpretation.
type Config struct {
	filters map[string]interface{}
	ctx     gocontext.Context
}

func (c *Config) Context() gocontext.Context {
	return c.ctx
}

// NewConfig creates a new Config.
func NewConfig(ctx gocontext.Context) Config {
	return Config{ctx: ctx}
}
