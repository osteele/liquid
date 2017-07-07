package parser

import "github.com/osteele/liquid/expression"

// A Config holds configuration information for parsing and rendering.
type Config struct {
	expression.Config
	Grammar Grammar
}

// NewConfig creates a parser Config.
func NewConfig() Config {
	return Config{}
}
