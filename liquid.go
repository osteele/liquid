/*
Package liquid is a pure Go implementation of Shopify Liquid templates, developed for use in https://github.com/osteele/gojekyll.

See the project README https://github.com/osteele/liquid for additional information and implementation status.

The liquid package itself is versioned in gopkg.in. Subpackages have no compatibility guarantees. Except where specifically documented, the “public” entities of subpackages are intended only for use by the liquid package and its subpackages.
*/
package liquid

import (
	"context"
	"maps"

	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/tags"
)

// Bindings is a map of variable names to values.
//
// Clients need not use this type. It is used solely for documentation. Callers can use instances
// of map[string]any itself as argument values to functions declared with this parameter type.
type Bindings map[string]any

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
func IterationKeyedMap(m map[string]any) tags.IterationKeyedMap {
	return m
}

// RenderOption is a functional option that overrides engine-level configuration
// for a single Render or FRender call.
//
// Create options with WithStrictVariables, WithLaxFilters, or WithGlobals.
type RenderOption func(*render.Config)

// WithStrictVariables causes this render call to error when the template
// references an undefined variable, regardless of the engine-level setting.
func WithStrictVariables() RenderOption {
	return func(c *render.Config) {
		c.StrictVariables = true
	}
}

// WithLaxFilters causes this render call to silently pass the input value
// through when the template references an undefined filter, regardless of
// the engine-level setting.
func WithLaxFilters() RenderOption {
	return func(c *render.Config) {
		c.LaxFilters = true
	}
}

// WithGlobals merges the provided map into the globals for this render call.
// Per-call globals are merged on top of any engine-level globals set via
// Engine.SetGlobals; both are superseded by the scope bindings passed to Render.
//
// This mirrors the `globals` render option in LiquidJS.
func WithGlobals(globals map[string]any) RenderOption {
	return func(c *render.Config) {
		if len(globals) == 0 {
			return
		}
		merged := make(map[string]any, len(c.Globals)+len(globals))
		maps.Copy(merged, c.Globals)
		maps.Copy(merged, globals)
		c.Globals = merged
	}
}

// WithErrorHandler registers a function that is called when a render-time error
// occurs instead of stopping the render. The handler receives the error and
// returns a string that is written to the output in place of the failing node.
// Rendering continues with the next node after the handler returns.
//
// This mirrors Ruby Liquid's exception_renderer option.
//
// To collect errors without stopping render:
//
//	var errs []error
//	out, _ := tpl.RenderString(vars, WithErrorHandler(func(err error) string {
//	    errs = append(errs, err)
//	    return "" // or some placeholder
//	}))
func WithErrorHandler(fn func(error) string) RenderOption {
	return func(c *render.Config) {
		c.ExceptionHandler = fn
	}
}

// WithContext sets the context for this render call. When the context is
// cancelled or its deadline is exceeded, rendering stops and the context
// error is returned. Use this for time-based render limits.
//
//	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	defer cancel()
//	out, err := tpl.RenderString(vars, WithContext(ctx))
func WithContext(ctx context.Context) RenderOption {
	return func(c *render.Config) {
		c.Context = ctx
	}
}

// WithSizeLimit limits the total number of bytes written to the output during
// this render call. Rendering is aborted with an error when the limit is exceeded.
func WithSizeLimit(n int64) RenderOption {
	return func(c *render.Config) {
		c.SizeLimit = n
	}
}

// WithGlobalFilter registers a function that is applied to the evaluated value of
// every {{ expression }} for this render call, overriding any engine-level global
// filter set via Engine.SetGlobalFilter.
//
// This mirrors Ruby Liquid's global_filter: render option.
//
// Example:
//
//	out, err := tpl.RenderString(vars, WithGlobalFilter(func(v any) (any, error) {
//	    if s, ok := v.(string); ok {
//	        return strings.ToUpper(s), nil
//	    }
//	    return v, nil
//	}))
func WithGlobalFilter(fn func(any) (any, error)) RenderOption {
	return func(c *render.Config) {
		c.SetGlobalFilter(fn)
	}
}
