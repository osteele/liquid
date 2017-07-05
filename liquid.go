/*
Package liquid is a pure Go implementation of Shopify Liquid templates, for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.

Note that the API for this package is not frozen. It is *especiallY* likely that subpackage APIs will
change drastically. Don't use anything except from a subpackage except render.Context.
*/
package liquid

import (
	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/expression"
	"github.com/osteele/liquid/render"
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

// A Template renders a template according to scope.
type Template interface {
	// Render executes the template with the specified bindings.
	Render(Bindings) ([]byte, error)
	// RenderString is a convenience wrapper for Render, that has string input and output.
	RenderString(Bindings) (string, error)
	SetSourcePath(string)
}

// Bindings is a map of variable names to values.
type Bindings map[string]interface{}

// TagParser parses the argument string "args" from a tag "{% tagname args %}",
// and returns a renderer.
// type TagParser func(chunks.RenderContext) (string, error)

// A Renderer returns the rendered string for a block.
type Renderer func(render.Context) (string, error)

// IsTemplateError returns a boolean indicating whether the error indicates
// an error in the template. All other errors are either errors in added
// tags or filters, or errors the implementation of this package.
//
// Use this to avoid coding the specific types of subpackage errors, which
// are likely to change.
func IsTemplateError(err error) bool {
	switch err.(type) {
	case evaluator.TypeError:
		return true
	case expression.InterpreterError:
		return true
	case expression.ParseError:
		return true
	case render.ParseError:
		return true
	default:
		return render.IsRenderError(err)
	}
}
