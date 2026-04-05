package values

import (
	"reflect"
	"strings"
)

// LiquidSentinel is implemented exclusively by EmptyDrop and BlankDrop.
// Tagging these special singletons with this interface lets the expression
// evaluator preserve their identity through the evaluation pipeline (instead
// of discarding it via .Interface()), enabling correct case/when comparisons.
//
// The unexported method prevents external packages from accidentally
// implementing this interface (sealed interface pattern).
type LiquidSentinel interface {
	liquidSentinel()
}

// emptyDropValue is the singleton type for the `empty` literal in Liquid.
// A value compares equal to empty if it is an empty string, empty array, or
// empty map (but not nil or false).
type emptyDropValue struct{}

// blankDropValue is the singleton type for the `blank` literal in Liquid.
// A value compares equal to blank if it is nil, false, an empty or
// whitespace-only string, an empty array, or an empty map.
type blankDropValue struct{}

func (e *emptyDropValue) liquidSentinel() {}
func (b *blankDropValue) liquidSentinel() {}

// EmptyDrop is the singleton for the `empty` keyword.
// After PRE-A lands, the scanner will resolve the `empty` literal directly to
// this value instead of treating it as an undefined variable.
var EmptyDrop Value = &emptyDropValue{}

// BlankDrop is the singleton for the `blank` keyword.
// After PRE-A lands, the scanner will resolve the `blank` literal directly to
// this value instead of treating it as an undefined variable.
var BlankDrop Value = &blankDropValue{}

var reflectIntType = reflect.TypeOf(int(0))

// -- emptyDropValue -----------------------------------------------------------

func (e *emptyDropValue) Interface() any            { return "" }
func (e *emptyDropValue) Int() int                  { panic(conversionError("", e, reflectIntType)) }
func (e *emptyDropValue) Test() bool                { return true } // only nil and false are falsy
func (e *emptyDropValue) Less(Value) bool           { return false }
func (e *emptyDropValue) Contains(Value) bool       { return false }
func (e *emptyDropValue) IndexValue(Value) Value    { return nilValue }
func (e *emptyDropValue) PropertyValue(Value) Value { return nilValue }

// Equal returns true when the other value satisfies IsEmpty.
// empty does not equal empty itself (matches Ruby and LiquidJS behaviour).
func (e *emptyDropValue) Equal(other Value) bool {
	switch other.(type) {
	case *emptyDropValue, *blankDropValue:
		return false
	}
	return IsEmpty(other.Interface())
}

// -- blankDropValue -----------------------------------------------------------

func (b *blankDropValue) Interface() any            { return "" }
func (b *blankDropValue) Int() int                  { panic(conversionError("", b, reflectIntType)) }
func (b *blankDropValue) Test() bool                { return true } // only nil and false are falsy
func (b *blankDropValue) Less(Value) bool           { return false }
func (b *blankDropValue) Contains(Value) bool       { return false }
func (b *blankDropValue) IndexValue(Value) Value    { return nilValue }
func (b *blankDropValue) PropertyValue(Value) Value { return nilValue }

// Equal returns true when the other value satisfies IsBlank.
// blank does not equal blank itself, and blank does not equal empty
// (consistent with EmptyDrop behaviour).
func (b *blankDropValue) Equal(other Value) bool {
	switch other.(type) {
	case *emptyDropValue, *blankDropValue:
		return false
	}
	return IsBlank(other.Interface())
}

// isWhitespaceOnly reports whether s consists entirely of Unicode whitespace.
func isWhitespaceOnly(s string) bool {
	return strings.TrimSpace(s) == ""
}
