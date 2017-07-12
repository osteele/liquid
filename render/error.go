package render

import (
	"github.com/osteele/liquid/parser"
)

// An Error is an error during template rendering.
type Error interface {
	Path() string
	LineNumber() int
	Cause() error
	Error() string
}

func renderErrorf(loc parser.Locatable, format string, a ...interface{}) Error {
	return parser.Errorf(loc, format, a...)
}

func wrapRenderError(err error, loc parser.Locatable) Error {
	return parser.WrapError(err, loc)
}
