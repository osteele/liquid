/*
Package liquid is a pure Go implementation of Shopify Liquid templates, developed for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.


The liquid package itself is versioned in gopkg.in. Subpackages have no compatibility guarantees. Except where specifically documented, the “public” entities of subpackages are intended only for use by the liquid package and its subpackages.
*/
package liquid

import (
	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/tags"
)

// Bindings is a map of variable names to values.
//
// Clients need not use this type. It is used solely for documentation. Callers can use instances
// of map[string]interface{} itself as argument values to functions declared with this parameter type.
type Bindings map[string]interface{}

// A Renderer returns the rendered string for a block. This is the type of a tag definition.
//
// See the examples at Engine.RegisterTag and Engine.RegisterBlock.
type Renderer func(render.Context) (string, error)

// SourceError records an error with a source location and optional cause.
//
// SourceError does not depend on, but is compatible with, the causer interface of https://github.com/pkg/errors.
type SourceError interface {
	error
	Cause() error
	Path() string
	LineNumber() int
}

// IterationKeyedMap returns a map whose {% for %} tag iteration values are its keys, instead of [key, value] pairs.
// Use this to create a Go map with the semantics of a Ruby struct drop.
func IterationKeyedMap(m map[string]interface{}) tags.IterationKeyedMap {
	return m
}
