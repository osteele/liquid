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
// Further examples are in https://github.com/osteele/gojekyll/blob/master/tags/tags.go
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
//
// The template is initialized from the engine configuration. It currently
// contains a copy. Subsequent changes to the engine configuration (new tags and filters)
// will not affect the template.
func (e *Engine) ParseTemplate(source []byte) (*Template, SourceError) {
	return newTemplate(&e.cfg, source, "", 0)
}

// ParseTemplateLocation is the same as ParseTemplate followed by SetSourceLocation,
// except that the source location is available during template compilation.
func (e *Engine) ParseTemplateLocation(source []byte, path string, line int) (*Template, SourceError) {
	return newTemplate(&e.cfg, source, path, line)
}

// ParseAndRender parses and then renders the template.
func (e *Engine) ParseAndRender(source []byte, b Bindings) ([]byte, SourceError) {
	tpl, err := e.ParseTemplate(source)
	if err != nil {
		return nil, err
	}
	return tpl.Render(b)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that takes string input and returns a string.
func (e *Engine) ParseAndRenderString(source string, b Bindings) (string, SourceError) {
	bs, err := e.ParseAndRender([]byte(source), b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
