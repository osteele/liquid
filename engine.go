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
//
// In the future, it will be configured with additional tags, filters, and the {%include%} search path.
type Engine interface {
	DefineFilter(name string, fn interface{})
	DefineTag(string, func(form string) (func(io.Writer, chunks.Context) error, error))

	ParseTemplate(text []byte) (Template, error)
	ParseAndRender(text []byte, scope map[string]interface{}) ([]byte, error)
	ParseAndRenderString(text string, scope map[string]interface{}) (string, error)
}

type TagDefinition func(expr string) (func(io.Writer, chunks.Context) error, error)

type engine struct{}

type template struct {
	ast chunks.ASTNode
}

// NewEngine makes a new engine.
func NewEngine() Engine {
	return engine{}
}

func (e engine) DefineFilter(name string, fn interface{}) {
	// TODO define this on the engine, not globally
	expressions.DefineFilter(name, fn)
}

func (e engine) DefineTag(name string, td func(form string) (func(io.Writer, chunks.Context) error, error)) {
	// TODO define this on the engine, not globally
	chunks.DefineTag(name, chunks.TagDefinition(td))
}

func (e engine) ParseTemplate(text []byte) (Template, error) {
	tokens := chunks.Scan(string(text), "")
	ast, err := chunks.Parse(tokens)
	// fmt.Println(chunks.MustYAML(ast))
	if err != nil {
		return nil, err
	}
	return &template{ast}, nil
}

// ParseAndRender parses and then renders the template.
func (e engine) ParseAndRender(text []byte, scope map[string]interface{}) ([]byte, error) {
	t, err := e.ParseTemplate(text)
	if err != nil {
		return nil, err
	}
	return t.Render(scope)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
func (e engine) ParseAndRenderString(text string, scope map[string]interface{}) (string, error) {
	b, err := e.ParseAndRender([]byte(text), scope)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
