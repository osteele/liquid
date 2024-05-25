package parser

import (
	"github.com/osteele/liquid/expressions"
	"regexp"
)

// A Config holds configuration information for parsing and rendering.
type Config struct {
	expressions.Config
	Grammar      Grammar
	delims       []string
	tokenMatcher *regexp.Regexp
}

// NewConfig creates a parser Config.
func NewConfig(g Grammar) Config {
	c := Config{Grammar: g}
	// Apply defaults
	c.delims = []string{"{{", "}}", "{%", "%}"}
	c.tokenMatcher = formTokenMatcher(c.delims)
	return c
}

func (c *Config) Delims(objectLeft, objectRight, tagLeft, tagRight string) {
	c.delims = []string{objectLeft, objectRight, tagLeft, tagRight}
	c.tokenMatcher = formTokenMatcher(c.delims)
}
