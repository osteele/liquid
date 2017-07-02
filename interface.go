/*
Package liquid is a pure Go implementation of Shopify Liquid templates, for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.
*/
package liquid

import (
	"io"

	"github.com/osteele/liquid/chunks"
)

// Engine parses template source into renderable text.
//
// An engine can be configured with additional filters and tags.
//
// Filters
//
// DefineFilter defines a Liquid filter.
//
// A filter is any function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must be an error.
type Engine interface {
	// DefineFilter defines a filter function e.g. {{ value | filter: arg }}.
	DefineFilter(name string, fn interface{})
	// DefineTag defines a tag function e.g. {% tag %}.
	DefineTag(string, TagDefinition)
	DefineStartTag(string, func(io.Writer, chunks.RenderContext) error)

	ParseTemplate([]byte) (Template, error)
	// ParseAndRender parses and then renders the template.
	ParseAndRender([]byte, Context) ([]byte, error)
	// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
	ParseAndRenderString(string, Context) (string, error)
}

// Template renders a template according to scope.
//
// Bindings is a map of liquid variable names to objects.
type Template interface {
	// Render executes the template with the specified bindings.
	Render(Context) ([]byte, error)
	// RenderString is a convenience wrapper for Render, that has string input and output.
	RenderString(Context) (string, error)
}

// Context supplies variable bindings and other information to a
// Render.
//
// In the future, it will hold methods to get and set the current
// filename.
type Context interface {
	Bindings() map[string]interface{}
}

// Renderer is the type of a function that is evaluated within a context and writes to output.
// type Renderer func(io.Writer, chunks.Context) error

// TagDefinition is the type of a function that parses the argument string "args" from a tag "{% tagname args %}",
// and returns a renderer.
type TagDefinition func(io.Writer, chunks.RenderContext) error
