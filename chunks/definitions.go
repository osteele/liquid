package chunks

import (
	"io"
)

// TagDefinition is a function that parses the tag arguments, and returns a renderer.
// TODO instead of using the bare function definition, use a structure that defines how to parse
type TagDefinition func(expr string) (func(io.Writer, RenderContext) error, error)

// TODO parse during definition stage, not rendering stage
func assignTagDef(source string) (func(io.Writer, RenderContext) error, error) {
	return func(w io.Writer, ctx RenderContext) error {
		_, err := ctx.EvaluateStatement("assign", source)
		return err
	}, nil
}

var tagDefinitions = map[string]TagDefinition{
	"assign": assignTagDef,
}

// DefineTag creates a tag definition.
func DefineTag(name string, td TagDefinition) {
	tagDefinitions[name] = td
}

// FindTagDefinition looks up a tag definition.
func FindTagDefinition(name string) (TagDefinition, bool) {
	td, ok := tagDefinitions[name]
	return td, ok
}
