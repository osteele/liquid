package parser

import "fmt"

// An Error is a syntax error during template parsing.
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

func (e *sourceLocError) Path() string {
	return e.Pathname
}

func (e *sourceLocError) LineNumber() int {
	return e.LineNo
}

func (e *sourceLocError) Error() string {
	line := ""
	if e.LineNo > 0 {
		line = fmt.Sprintf(" (line %d)", e.LineNo)
	}
	locative := " in " + e.context
	if e.Pathname != "" {
		locative = " in " + e.Pathname
	}
	return fmt.Sprintf("Liquid error%s: %s%s", line, e.message, locative)
}
