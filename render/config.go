package render

import "github.com/osteele/liquid/expression"

// Config holds configuration information for parsing and rendering.
type Config struct {
	ExpressionConfig expression.Config
	tags             map[string]TagDefinition
	controlTags      map[string]*blockDef
}

// NewConfig creates a new Settings.
func NewConfig() Config {
	s := Config{
		expression.NewConfig(),
		map[string]TagDefinition{},
		map[string]*blockDef{},
	}
	s.AddTag("assign", assignTagDef)
	return s
}

// AddFilter adds a filter to settings.
func (s Config) AddFilter(name string, fn interface{}) {
	s.ExpressionConfig.AddFilter(name, fn)
}
