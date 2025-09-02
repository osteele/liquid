package render

import (
	"context"

	"github.com/autopilot3/liquid/parser"
)

// Config holds configuration information for parsing and rendering.
type Config struct {
	parser.Config
	grammar
	AllowedTags          map[string]struct{}
	AllowTagsWithDefault bool
}

type grammar struct {
	tags      map[string]TagCompiler
	blockDefs map[string]*blockSyntax
}

func NewConfig() Config {
	return NewConfigWitchContext(context.Background())
}

// NewConfig creates a new Settings.
func NewConfigWitchContext(ctx context.Context) Config {
	g := grammar{
		tags:      map[string]TagCompiler{},
		blockDefs: map[string]*blockSyntax{},
	}
	return Config{
		parser.NewConfig(g, ctx),
		g,
		nil,
		false,
	}
}
