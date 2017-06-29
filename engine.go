/*
Package liquid is a pure Go implementation of Shopify Liquid templates, for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.
*/
package liquid

import (
	"io"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/tags"
)

// TODO move the filters and tags from globals to the engine
func init() {
	tags.DefineStandardTags()
	filters.DefineStandardFilters()
}

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
type TagDefinition func(parameters string) (func(io.Writer, chunks.Context) error, error)

type engine struct{}

type template struct {
	ast chunks.ASTNode
}

// NewEngine returns a new template engine.
func NewEngine() Engine {
	return engine{}
}

// DefineFilter is in the Engine interface.
func (e engine) DefineFilter(name string, fn interface{}) {
	// TODO define this on the engine, not globally
	expressions.DefineFilter(name, fn)
}

// ParseAndRenderString is in the Engine interface.
func (e engine) DefineTag(name string, td TagDefinition) {
	// TODO define this on the engine, not globally
	chunks.DefineTag(name, chunks.TagDefinition(td))
}

// ParseTemplate is in the Engine interface.
func (e engine) ParseTemplate(text []byte) (Template, error) {
	tokens := chunks.Scan(string(text), "")
	ast, err := chunks.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return &template{ast}, nil
}

// ParseAndRender is in the Engine interface.
func (e engine) ParseAndRender(text []byte, bindings map[string]interface{}) ([]byte, error) {
	t, err := e.ParseTemplate(text)
	if err != nil {
		return nil, err
	}
	return t.Render(bindings)
}

// ParseAndRenderString is in the Engine interface.
func (e engine) ParseAndRenderString(text string, bindings map[string]interface{}) (string, error) {
	b, err := e.ParseAndRender([]byte(text), bindings)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
