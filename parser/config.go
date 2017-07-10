package parser

import "github.com/osteele/liquid/expression"

// A Config holds configuration information for parsing and rendering.
type Config struct {
	expression.Config
	Grammar  Grammar
	Filename string
	LineNo   int
}

// NewConfig creates a parser Config.
func NewConfig(g Grammar) Config {
	return Config{Grammar: g, LineNo: 1}
}
