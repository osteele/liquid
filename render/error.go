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
	// Message returns the error message without the "Liquid error" prefix or
	// location information.
	Message() string
	// MarkupContext returns the source text of the expression/tag that produced
	// this error, e.g. "{{ product.price | divided_by: 0 }}".
	MarkupContext() string
}

// RenderError is a render-time error with source location information.
// Use errors.As to check whether a liquid error originates from rendering
// (as opposed to parsing).
type RenderError struct {
	inner parser.Error
}

// Error builds the error string with the "Liquid error" prefix.  This overrides
// the inner parser.Error's "Liquid syntax error" prefix, since render-time
// failures are not syntax errors.
func (e *RenderError) Error() string {
	line := ""
	if n := e.inner.LineNumber(); n > 0 {
		line = fmt.Sprintf(" (line %d)", n)
	}
	locative := ""
	if p := e.inner.Path(); p != "" {
		locative = " in " + p
	} else if mc := e.inner.MarkupContext(); mc != "" {
		locative = " in " + mc
	}
	return fmt.Sprintf("Liquid error%s: %s%s", line, e.inner.Message(), locative)
}

func (e *RenderError) Cause() error          { return e.inner.Cause() }
func (e *RenderError) Path() string          { return e.inner.Path() }
func (e *RenderError) LineNumber() int       { return e.inner.LineNumber() }
func (e *RenderError) Message() string       { return e.inner.Message() }
func (e *RenderError) MarkupContext() string { return e.inner.MarkupContext() }

// Unwrap returns the inner parse-level error, enabling errors.As to walk the
// chain and find causes such as ZeroDivisionError.
func (e *RenderError) Unwrap() error { return e.inner }

// UndefinedVariableError is returned when StrictVariables is enabled and a
// template variable resolves to nil. The Name field contains the root variable
// name (e.g. "user" for {{ user.name | upcase }}). BlockContext and BlockLine
// are set to the innermost enclosing block tag source and line when the error
// bubbles up through BlockNode.render.
type UndefinedVariableError struct {
	RootName     string // root segment only, e.g. "user" for {{ user.name }}
	FullPath     string // full dotted path, e.g. "user.name"; empty when single-segment
	loc          parser.Error
	BlockContext string // e.g. "{% if cond %}"
	BlockLine    int    // 1-based line of the enclosing block tag
}

func (e *UndefinedVariableError) Error() string {
	line := ""
	if e.loc.LineNumber() > 0 {
		line = fmt.Sprintf(" (line %d)", e.loc.LineNumber())
	}
	// Primary locative: file path, then markup context of the {{ expr }}.
	locative := ""
	if e.loc.Path() != "" {
		locative = " in " + e.loc.Path()
	} else if mc := e.loc.MarkupContext(); mc != "" {
		locative = " in " + mc
	}
	// Secondary context: the innermost enclosing block tag, if available.
	blockCtx := ""
	if e.BlockContext != "" {
		if e.BlockLine > 0 {
			blockCtx = fmt.Sprintf(" (inside %s, line %d)", e.BlockContext, e.BlockLine)
		} else {
			blockCtx = fmt.Sprintf(" (inside %s)", e.BlockContext)
		}
	}
	display := e.RootName
	if e.FullPath != "" {
		display = e.FullPath
	}
	return fmt.Sprintf("Liquid error%s: undefined variable %q%s%s", line, display, locative, blockCtx)
}

func (e *UndefinedVariableError) Cause() error    { return nil }
func (e *UndefinedVariableError) Path() string    { return e.loc.Path() }
func (e *UndefinedVariableError) LineNumber() int { return e.loc.LineNumber() }
func (e *UndefinedVariableError) Message() string {
	display := e.RootName
	if e.FullPath != "" {
		display = e.FullPath
	}
	return fmt.Sprintf("undefined variable %q", display)
}
func (e *UndefinedVariableError) MarkupContext() string { return e.loc.MarkupContext() }

// Unwrap allows errors.As / errors.Is to find this error through a wrapping chain.
func (e *UndefinedVariableError) Unwrap() error { return e.loc }

// ArgumentError is returned by filters or tags that receive invalid arguments.
// Return it from a filter or tag renderer; the render engine will wrap it with
// source-location information so the full Error() string contains "Liquid error (line N): …".
// Use errors.As to detect this in the error chain returned by Engine.ParseAndRender.
type ArgumentError struct {
	msg string
}

// NewArgumentError creates an ArgumentError with the given message.
func NewArgumentError(msg string) *ArgumentError { return &ArgumentError{msg: msg} }

func (e *ArgumentError) Error() string { return e.msg }

// ContextError is returned when a context variable lookup or scope operation fails.
// It surfaces through the render error chain; use errors.As to detect it.
type ContextError struct {
	msg string
}

// NewContextError creates a ContextError with the given message.
func NewContextError(msg string) *ContextError { return &ContextError{msg: msg} }

func (e *ContextError) Error() string { return e.msg }

func renderErrorf(loc parser.Locatable, format string, a ...any) Error {
	return &RenderError{parser.Errorf(loc, format, a...)}
}

func wrapRenderError(err error, loc parser.Locatable) Error {
	if err == nil {
		return nil
	}
	// UndefinedVariableError is already fully formed — preserve it as-is.
	if ue, ok := err.(*UndefinedVariableError); ok {
		return ue
	}
	// If already a RenderError, preserve it when:
	//   - it already has a file path (most specific possible), OR
	//   - it already has a line number (came from a specific inner node such as
	//     an ObjectNode or TagNode; a parent BlockNode must not overwrite it with
	//     a less-specific context such as "{% if … %}"), OR
	//   - the wrapping location itself has no useful information.
	if re, ok := err.(*RenderError); ok {
		if re.Path() != "" || re.LineNumber() > 0 || loc.SourceLocation().IsZero() {
			return re
		}
	}
	return &RenderError{parser.WrapError(err, loc)}
}
