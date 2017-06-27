/*
Package liquid is a very early-stage pure Go library that implements Shopify Liquid <https://shopify.github.io/liquid> templates.

It's intended for use in for use in https://github.com/osteele/gojekyll.
*/
package liquid

import (
	"io"

	"github.com/osteele/liquid/chunks"
)

// Engine parses template source into renderable text.
//
// In the future, it will be configured with additional tags, filters, and the {%include%} search path.
type Engine interface {
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

func (e engine) DefineTag(name string, td func(form string) (func(io.Writer, chunks.Context) error, error)) {
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
