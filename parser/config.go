package parser

import "github.com/osteele/liquid/expressions"

// A Config holds configuration information for parsing and rendering.
type Config struct {
	expressions.Config
	Grammar Grammar
	Delims  []string
}

// NewConfig creates a parser Config.
func NewConfig(g Grammar) Config {
	return Config{Grammar: g}
}
