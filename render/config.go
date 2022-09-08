package render

import (
	"github.com/autopilot3/liquid/parser"
)

// Config holds configuration information for parsing and rendering.
type Config struct {
	parser.Config
	grammar
	AllowedTags map[string]struct{}
}

type grammar struct {
	tags      map[string]TagCompiler
	blockDefs map[string]*blockSyntax
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	g := grammar{
		tags:      map[string]TagCompiler{},
		blockDefs: map[string]*blockSyntax{},
	}
	return Config{
		parser.NewConfig(g),
		g,
		nil,
	}
}
