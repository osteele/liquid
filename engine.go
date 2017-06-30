package liquid

import (
	"io"

	"github.com/osteele/liquid/chunks"
	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/tags"
)

type engine struct{ settings chunks.Settings }

// NewEngine returns a new template engine.
func NewEngine() Engine {
	e := engine{chunks.NewSettings()}
	filters.AddStandardFilters(e.settings.ExpressionSettings)
	tags.AddStandardTags(e.settings)
	return e
}

// DefineStartTag is in the Engine interface.
func (e engine) DefineStartTag(name string, td func(io.Writer, chunks.RenderContext) error) {
	e.settings.AddStartTag(name).Renderer(td)
}

// DefineFilter is in the Engine interface.
func (e engine) DefineFilter(name string, fn interface{}) {
	// TODO define this on the engine, not globally
	e.settings.AddFilter(name, fn)
}

// ParseAndRenderString is in the Engine interface.
func (e engine) DefineTag(name string, td TagDefinition) {
	// TODO define this on the engine, not globally
	e.settings.AddTag(name, chunks.TagDefinition(td))
}

// ParseTemplate is in the Engine interface.
func (e engine) ParseTemplate(text []byte) (Template, error) {
	tokens := chunks.Scan(string(text), "")
	ast, err := e.settings.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return &template{ast, e.settings}, nil
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
