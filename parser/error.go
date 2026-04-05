package parser

import "fmt"

// An Error is a syntax error during template parsing.
type Error interface {
	error
	Cause() error
	Path() string
	LineNumber() int
	// Message returns the error message without the "Liquid error" prefix or
	// location information. Useful for re-formatting errors with a different prefix.
	Message() string
	// MarkupContext returns the source text of the token/node that produced the
	// error. For example, for a {{ expr }} node it returns the full "{{ expr }}"
	// string. Returns an empty string when no source text is available.
	MarkupContext() string
}

// A Locatable provides source location information for error reporting.
type Locatable interface {
	SourceLocation() SourceLoc
	SourceText() string
}

// ParseError is a parse-time syntax error with source location information.
// The Error() string uses the "Liquid syntax error" prefix, matching Ruby Liquid.
// Use errors.As to check whether a liquid error originates from parsing.
//
// SyntaxError is provided as a type alias so callers can use the more
// semantically precise name: errors.As(err, new(*parser.SyntaxError)).
type ParseError struct {
	*sourceLocError
}

// SyntaxError is an alias for ParseError.  Both names refer to the same type;
// errors.As patterns using either *ParseError or *SyntaxError are equivalent.
type SyntaxError = ParseError

// Error overrides sourceLocError.Error to use the "Liquid syntax error" prefix.
// This matches Ruby Liquid, where parse-time errors are "Liquid syntax error: …".
func (e *ParseError) Error() string {
	return e.sourceLocError.errorWithPrefix("Liquid syntax error")
}

// Errorf creates a parser.Error at the given source location.
func Errorf(loc Locatable, format string, a ...any) *ParseError { //nolint: golint
	return &ParseError{&sourceLocError{loc.SourceLocation(), loc.SourceText(), fmt.Sprintf(format, a...), nil}}
}

// WrapError wraps its argument in a parser.Error if this argument is not already a parser.Error and is not locatable.
func WrapError(err error, loc Locatable) Error {
	if err == nil {
		return nil
	}

	if e, ok := err.(Error); ok {
		// re-wrap the error, if the inner layer implemented the locatable interface
		// but didn't actually provide any information
		if e.Path() != "" || loc.SourceLocation().IsZero() {
			return e
		}

		if e.Cause() != nil {
			err = e.Cause()
		}
	}

	re := Errorf(loc, "%s", err)
	re.cause = err

	return re
}

type sourceLocError struct {
	SourceLoc

	context string
	message string
	cause   error
}

func (e *sourceLocError) Cause() error {
	return e.cause
}

// Unwrap returns the underlying cause of this error, enabling errors.As and errors.Is
// to walk the error chain (e.g. to find a ZeroDivisionError or UndefinedVariableError).
func (e *sourceLocError) Unwrap() error {
	return e.cause
}

func (e *sourceLocError) Path() string {
	return e.Pathname
}

func (e *sourceLocError) LineNumber() int {
	return e.LineNo
}

func (e *sourceLocError) Message() string {
	return e.message
}

func (e *sourceLocError) MarkupContext() string {
	return e.context
}

// errorWithPrefix formats the error message with the given prefix string.
// This exists so ParseError can override the default "Liquid error" prefix
// with "Liquid syntax error" without duplicating the formatting logic.
func (e *sourceLocError) errorWithPrefix(prefix string) string {
	line := ""
	if e.LineNo > 0 {
		line = fmt.Sprintf(" (line %d)", e.LineNo)
	}

	locative := " in " + e.context
	if e.Pathname != "" {
		locative = " in " + e.Pathname
	}

	return fmt.Sprintf("%s%s: %s%s", prefix, line, e.message, locative)
}

func (e *sourceLocError) Error() string {
	return e.errorWithPrefix("Liquid error")
}
