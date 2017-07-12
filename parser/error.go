package parser

import "fmt"

// An Error is a parse error during template parsing.
type Error interface {
	error
	Cause() error
	Path() string
	LineNumber() int
}

// A Locatable provides source location information for error reporting.
type Locatable interface {
	SourceLocation() SourceLoc
	SourceText() string
}

// Errorf creates a parser.Error.
func Errorf(loc Locatable, format string, a ...interface{}) *sourceLocError { // nolint: golint
	return &sourceLocError{loc.SourceLocation(), loc.SourceText(), fmt.Sprintf(format, a...), nil}
}

// WrapError wraps its argument in a parser.Error if this argument is not already a parser.Error and is not locatable.
func WrapError(err error, loc Locatable) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
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

func (e *sourceLocError) Path() string {
	return e.Pathname
}

func (e *sourceLocError) LineNumber() int {
	return e.LineNo
}

func (e *sourceLocError) Error() string {
	locative := " in " + e.context
	if e.Pathname != "" {
		locative = " in " + e.Pathname
	}
	return fmt.Sprintf("Liquid error (line %d): %s%s", e.LineNo, e.message, locative)
}
