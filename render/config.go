package render

import "github.com/osteele/liquid/expression"

// Config holds configuration information for parsing and rendering.
type Config struct {
	// ExpressionConfig expression.Config
	expression.Config
	Filename  string
	tags      map[string]TagCompiler
	blockDefs map[string]*blockSyntax
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	s := Config{
		Config:    expression.NewConfig(),
		tags:      map[string]TagCompiler{},
		blockDefs: map[string]*blockSyntax{},
	}
	return s
}

// AddFilter adds a filter to settings.
func (s Config) AddFilter(name string, fn interface{}) {
	s.Config.AddFilter(name, fn)
}
