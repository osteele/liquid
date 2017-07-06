package render

import (
	"io"
)

// TagCompiler is a function that parses the tag arguments, and returns a renderer.
// TODO instead of using the bare function definition, use a structure that defines how to parse
type TagCompiler func(expr string) (func(io.Writer, Context) error, error)

// AddTag creates a tag definition.
func (c *Config) AddTag(name string, td TagCompiler) {
	c.tags[name] = td
}

// FindTagDefinition looks up a tag definition.
func (c *Config) FindTagDefinition(name string) (TagCompiler, bool) {
	td, ok := c.tags[name]
	return td, ok
}
