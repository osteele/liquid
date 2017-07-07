package render

import (
	"github.com/osteele/liquid/parser"
)

// Config holds configuration information for parsing and rendering.
type Config struct {
	parser.Config
	Filename  string
	tags      map[string]TagCompiler
	blockDefs map[string]*blockSyntax
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	c := Config{
		tags:      map[string]TagCompiler{},
		blockDefs: map[string]*blockSyntax{},
	}
	c.Grammar = c
	return c
}
