//go:generate stringer -type=TokenType

package main

import (
	"fmt"
	"reflect"
)

type Token struct {
	t TokenType
	s string
	v interface{}
}

type TokenType int

const (
	IdentifierType TokenType = iota
	KeywordType
	RelationType
	ValueType
)

func (t Token) String() string {
	switch t.v {
	case nil:
		return fmt.Sprintf("%s{%s}", t.t, t.s)
	default:
		return fmt.Sprintf("%s{%v}", reflect.TypeOf(t.v), t.v)
	}
}
