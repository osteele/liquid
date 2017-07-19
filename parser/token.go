package parser

import "fmt"

// A Token is an object {{ a.b }}, a tag {% if a>b %}, or a text chunk (anything outside of {{}} and {%%}.)
type Token struct {
	Type                TokenType
	SourceLoc           SourceLoc
	Name                string // Name is the tag name of a tag Chunk. E.g. the tag name of "{% if 1 %}" is "if".
	Args                string // Parameters is the tag arguments of a tag Chunk. E.g. the tag arguments of "{% if 1 %}" is "1".
	Source              string // Source is the entirety of the token, including the "{{", "{%", etc. markers.
	TrimLeft, TrimRight bool   // Trim whitespace left or right of this token; from {{- tag -}} and {%- expr -%}
}

// TokenType is the type of a Chunk
type TokenType int

////go:generate stringer -type=TokenType

const (
	// TextTokenType is the type of a text Chunk
	TextTokenType TokenType = iota
	// TagTokenType is the type of a tag Chunk "{%…%}"
	TagTokenType
	// ObjTokenType is the type of an object Chunk "{{…}}"
	ObjTokenType
)

// SourceLoc contains a Token's source location.
type SourceLoc struct {
	Pathname string
	LineNo   int
}

// SourceLocation returns the token's source location, for use in error reporting.
func (c Token) SourceLocation() SourceLoc { return c.SourceLoc }

// SourceText returns the token's source text, for use in error reporting.
func (c Token) SourceText() string { return c.Source }

// IsZero returns a boolean indicating whether the location doesn't have a set path.
func (s SourceLoc) IsZero() bool {
	return s.Pathname == "" && s.LineNo == 0
}

func (c Token) String() string {
	switch c.Type {
	case TextTokenType:
		return fmt.Sprintf("%v{%#v}", c.Type, c.Source)
	case TagTokenType:
		return fmt.Sprintf("%v{Tag:%#v, Args:%#v}", c.Type, c.Name, c.Args)
	case ObjTokenType:
		return fmt.Sprintf("%v{%#v}", c.Type, c.Args)
	default:
		return fmt.Sprintf("%v{%#v}", c.Type, c.Source)
	}
}

func (s SourceLoc) String() string {
	if s.Pathname != "" {
		return fmt.Sprintf("%s:%d", s.Pathname, s.LineNo)
	}
	return fmt.Sprintf("line %d", s.LineNo)
}
