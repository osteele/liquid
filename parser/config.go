package parser

import (
	"context"

	"github.com/autopilot3/liquid/expressions"
)

// A Config holds configuration information for parsing and rendering.
type Config struct {
	expressions.Config
	Grammar Grammar
	Delims  []string
}

// NewConfig creates a parser Config.
func NewConfig(g Grammar, ctx context.Context) Config {
	return Config{
		Grammar: g,
		Config:  expressions.NewConfig(ctx),
	}
}
