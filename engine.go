/*
Package liquid is a pure Go implementation of Shopify Liquid templates.

It's intended for use in for use in https://github.com/osteele/gojekyll.

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
type Engine interface {
	DefineFilter(name string, fn interface{})
	DefineTag(string, func(form string) (func(io.Writer, chunks.Context) error, error))

	ParseTemplate(text []byte) (Template, error)
	ParseAndRender(text []byte, bindings map[string]interface{}) ([]byte, error)
	ParseAndRenderString(text string, bindings map[string]interface{}) (string, error)
}

type TagDefinition func(expr string) (func(io.Writer, chunks.Context) error, error)

type engine struct{}

type template struct {
	ast chunks.ASTNode
}

// NewEngine returns a new engine.
func NewEngine() Engine {
	return engine{}
}

// DefineFilter defines a Liquid filter.
//
// A filter is any function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must be an error.
//
// Note: Although this function is defined on the engine, its effect is currently global.
func (e engine) DefineFilter(name string, fn interface{}) {
	// TODO define this on the engine, not globally
	expressions.DefineFilter(name, fn)
}

// DefineTag defines a Liquid filter.
//
// A tag is any function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must be an error.
//
// Note: This interface is likely to change.
//
// Note: Although this function is defined on the engine, its effect is currently global.
func (e engine) DefineTag(name string, td func(form string) (func(io.Writer, chunks.Context) error, error)) {
	// TODO define this on the engine, not globally
	chunks.DefineTag(name, chunks.TagDefinition(td))
}

func (e engine) ParseTemplate(text []byte) (Template, error) {
	tokens := chunks.Scan(string(text), "")
	ast, err := chunks.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return &template{ast}, nil
}

// ParseAndRender parses and then renders the template.
func (e engine) ParseAndRender(text []byte, bindings map[string]interface{}) ([]byte, error) {
	t, err := e.ParseTemplate(text)
	if err != nil {
		return nil, err
	}
	return t.Render(bindings)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
func (e engine) ParseAndRenderString(text string, bindings map[string]interface{}) (string, error) {
	b, err := e.ParseAndRender([]byte(text), bindings)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
