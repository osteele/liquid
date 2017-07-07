package liquid

import (
	"io"

	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/tags"
)

// An Engine parses template source into renderable text.
//
// An engine can be configured with additional filters and tags.
//
// Filters
//
// RegisterFilter defines a Liquid filter.
//
// A filter is any function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must be an error.
type Engine interface {
	// RegisterFilter defines a filter function e.g. {{ value | filter: arg }}.
	RegisterFilter(name string, fn interface{})
	// RegisterTag defines a tag function e.g. {% tag %}.
	RegisterTag(string, Renderer)
	RegisterBlock(string, Renderer)

	ParseTemplate([]byte) (Template, error)
	// ParseAndRender parses and then renders the template.
	ParseAndRender([]byte, Bindings) ([]byte, error)
	// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
	ParseAndRenderString(string, Bindings) (string, error)
}

type engine struct{ cfg render.Config }

// NewEngine returns a new template engine.
func NewEngine() Engine {
	e := engine{render.NewConfig()}
	filters.AddStandardFilters(&e.cfg.Config.Config)
	tags.AddStandardTags(e.cfg)
	return e
}

// RegisterBlock is in the Engine interface.
func (e engine) RegisterBlock(name string, td Renderer) {
	e.cfg.AddBlock(name).Renderer(func(w io.Writer, ctx render.Context) error {
		s, err := td(ctx)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(s))
		return err
	})
}

// RegisterFilter is in the Engine interface.
func (e engine) RegisterFilter(name string, fn interface{}) {
	e.cfg.AddFilter(name, fn)
}

// RegisterTag is in the Engine interface.
func (e engine) RegisterTag(name string, td Renderer) {
	// For simplicity, don't expose the two stage parsing/rendering process to clients.
	// Client tags do everything at runtime.
	e.cfg.AddTag(name, func(_ string) (func(io.Writer, render.Context) error, error) {
		return func(w io.Writer, ctx render.Context) error {
			s, err := td(ctx)
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(s))
			return err
		}, nil
	})
}

// ParseTemplate is in the Engine interface.
func (e engine) ParseTemplate(text []byte) (Template, error) {
	ast, err := e.cfg.Compile(string(text))
	if err != nil {
		return nil, err
	}
	return &template{ast, &e.cfg}, nil
}

// ParseAndRender is in the Engine interface.
func (e engine) ParseAndRender(text []byte, b Bindings) ([]byte, error) {
	t, err := e.ParseTemplate(text)
	if err != nil {
		return nil, err
	}
	return t.Render(b)
}

// ParseAndRenderString is in the Engine interface.
func (e engine) ParseAndRenderString(text string, b Bindings) (string, error) {
	bs, err := e.ParseAndRender([]byte(text), b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
