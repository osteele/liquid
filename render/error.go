package render

import (
	"fmt"

	"github.com/osteele/liquid/parser"
)

// An Error is an error during template rendering.
type Error interface {
	Path() string
	LineNumber() int
	Cause() error
	Error() string
}

// RenderError is a render-time error with source location information.
// Use errors.As to check whether a liquid error originates from rendering
// (as opposed to parsing).
type RenderError struct {
	inner parser.Error
}

func (e *RenderError) Error() string    { return e.inner.Error() }
func (e *RenderError) Cause() error     { return e.inner.Cause() }
func (e *RenderError) Path() string     { return e.inner.Path() }
func (e *RenderError) LineNumber() int  { return e.inner.LineNumber() }

// Unwrap returns the inner parse-level error, enabling errors.As to walk the
// chain and find causes such as ZeroDivisionError.
func (e *RenderError) Unwrap() error { return e.inner }

// UndefinedVariableError is returned when StrictVariables is enabled and a
// template variable resolves to nil. The Name field contains the literal
// expression text from the template source.
type UndefinedVariableError struct {
	Name string
	loc  parser.Error
}

func (e *UndefinedVariableError) Error() string {
	line := ""
	if e.loc.LineNumber() > 0 {
		line = fmt.Sprintf(" (line %d)", e.loc.LineNumber())
	}
	locative := ""
	if e.loc.Path() != "" {
		locative = " in " + e.loc.Path()
	}
	return fmt.Sprintf("Liquid error%s: undefined variable %q%s", line, e.Name, locative)
}

func (e *UndefinedVariableError) Cause() error    { return nil }
func (e *UndefinedVariableError) Path() string    { return e.loc.Path() }
func (e *UndefinedVariableError) LineNumber() int { return e.loc.LineNumber() }

// Unwrap allows errors.As / errors.Is to find this error through a wrapping chain.
func (e *UndefinedVariableError) Unwrap() error { return e.loc }

func renderErrorf(loc parser.Locatable, format string, a ...any) Error {
	return &RenderError{parser.Errorf(loc, format, a...)}
}

func wrapRenderError(err error, loc parser.Locatable) Error {
	if err == nil {
		return nil
	}
	// UndefinedVariableError is already fully formed — preserve it as-is with
	// a RenderError wrapper only when it lacks location information.
	if ue, ok := err.(*UndefinedVariableError); ok {
		return ue
	}
	// If already a RenderError with location, return it unchanged.
	if re, ok := err.(*RenderError); ok {
		if re.Path() != "" || loc.SourceLocation().IsZero() {
			return re
		}
	}
	return &RenderError{parser.WrapError(err, loc)}
}

