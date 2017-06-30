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
	//
	// Note: Although this function is defined on the engine, its effect is currently global.
	DefineFilter(name string, fn interface{})
	// DefineTag defines a tag function e.g. {% tag %}.
	//
	// Note: Although this function is defined on the engine, its effect is currently global.
	DefineTag(string, TagDefinition)
	// Note: Although this function is defined on the engine, its effect is currently global.
	DefineStartTag(string, TagDefinition)

	ParseTemplate(b []byte) (Template, error)
	// ParseAndRender parses and then renders the template.
	ParseAndRender(b []byte, bindings map[string]interface{}) ([]byte, error)
	// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
	ParseAndRenderString(s string, bindings map[string]interface{}) (string, error)
}

// Renderer is the type of a function that is evaluated within a context and writes to output.
type Renderer func(io.Writer, chunks.Context) error

// TagDefinition is the type of a function that parses the argument string "args" from a tag "{% tagname args %}",
// and returns a renderer.
type TagDefinition func(parameters string) (func(io.Writer, chunks.RenderContext) error, error)
