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
type Engine struct{ cfg render.Config }

// NewEngine returns a new Engine.
func NewEngine() *Engine {
	e := Engine{render.NewConfig()}
	filters.AddStandardFilters(&e.cfg)
	tags.AddStandardTags(e.cfg)
	return &e
}

// RegisterBlock defines a block e.g. {% tag %}…{% endtag %}.
func (e *Engine) RegisterBlock(name string, td Renderer) {
	e.cfg.AddBlock(name).Renderer(func(w io.Writer, ctx render.Context) error {
		s, err := td(ctx)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(s))
		return err
	})
}

// RegisterFilter defines a Liquid filter, for use as `{{ value | my_filter }}` or `{{ value | my_filter: arg }}`.
//
// A filter is a function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must have type error.
//
// Examples:
//
// * https://github.com/osteele/liquid/blob/master/filters/filters.go
//
// * https://github.com/osteele/gojekyll/blob/master/filters/filters.go
//
func (e *Engine) RegisterFilter(name string, fn interface{}) {
	e.cfg.AddFilter(name, fn)
}

// RegisterTag defines a tag e.g. {% tag %}.
//
// RegisterTag defines a tag, for use as `{% tag args %}`.
//
// Examples:
//
// * https://github.com/osteele/gojekyll/blob/master/tags/tags.go
func (e *Engine) RegisterTag(name string, td Renderer) {
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

// ParseTemplate creates a new Template using the engine configuration.
func (e *Engine) ParseTemplate(text []byte) (*Template, error) {
	root, err := e.cfg.Compile(string(text))
	if err != nil {
		return nil, err
	}
	return &Template{root, &e.cfg}, nil
}

// ParseAndRender parses and then renders the template.
func (e *Engine) ParseAndRender(text []byte, b Bindings) ([]byte, error) {
	tpl, err := e.ParseTemplate(text)
	if err != nil {
		return nil, err
	}
	return tpl.Render(b)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that has string input and output.
func (e *Engine) ParseAndRenderString(text string, b Bindings) (string, error) {
	bs, err := e.ParseAndRender([]byte(text), b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
