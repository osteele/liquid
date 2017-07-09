/*
Package liquid is a pure Go implementation of Shopify Liquid templates, for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.

Note that the API for this package is not frozen. It is *especially* likely that subpackage APIs will
change drastically. Don't use anything except from a subpackage except render.Context.
*/
package liquid

import (
	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/expression"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// Bindings is a map of variable names to values.
type Bindings map[string]interface{}

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
	case parser.ParseError:
		return true
	case render.CompilationError:
		return true
	case render.Error:
		return true
	default:
		return false
	}
}
