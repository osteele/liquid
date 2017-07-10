package render

import (
	"fmt"

	"github.com/osteele/liquid/parser"
)

// An Error is a rendering error during template rendering.
type Error interface {
	Filename() string
	LineNumber() int
	Cause() error
	Error() string
}

type renderError struct {
	parser.SourceLoc
	context string
	message string
	cause   error
}

func (e *renderError) Cause() error {
	return e.cause
}

func (e *renderError) Filename() string {
	return e.Pathname
}

func (e *renderError) LineNumber() int {
	return e.LineNo
}

func (e *renderError) Error() string {
	locative := "in " + e.context
	if e.Pathname != "" {
		locative = "in " + e.Pathname
	}
	return fmt.Sprintf("Liquid exception: Liquid syntax error (line %d): %s%s", e.LineNo, e.message, locative)
}

func renderErrorf(loc parser.SourceLoc, context, format string, a ...interface{}) *renderError {
	return &renderError{loc, context, fmt.Sprintf(format, a...), nil}
}

func wrapRenderError(err error, n Node) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	re := renderErrorf(n.SourceLocation(), n.SourceText(), "%s", err)
	re.cause = err
	return re
}
