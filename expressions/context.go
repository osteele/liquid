package expressions

import (
	"reflect"

	"github.com/osteele/liquid/values"
)

// Context is the expression evaluation context. It maps variables names to values.
type Context interface {
	ApplyFilter(string, valueFn, []valueFn) (any, error)
	// Clone returns a copy with a new variable binding map
	// (so that copy.Set does effect the source context.)
	Clone() Context
	Get(string) any
	Set(string, any)
}
type context struct {
	Config

	bindings map[string]any
}

// NewContext makes a new expression evaluation context.
func NewContext(vars map[string]any, cfg Config) Context {
	return &context{cfg, vars}
}

func (ctx *context) Clone() Context {
	bindings := map[string]any{}
	for k, v := range ctx.bindings {
		bindings[k] = v
	}

	return &context{ctx.Config, bindings}
}

// Get looks up a variable value in the expression context.
// If the raw binding implements values.ContextSetter, the expression context is
// injected into it before ToLiquid conversion — mirroring Ruby's context= setter.
func (ctx *context) Get(name string) any {
	raw := ctx.bindings[name]
	if cs, ok := raw.(values.ContextSetter); ok {
		cs.SetContext(ctx)
	}
	v := values.ToLiquid(raw)
	// If ToLiquid returned a different value (wrapper resolved), inject context there too.
	// Only compare when both are comparable to avoid panic on slice/map values.
	if v != nil && !areSamePointer(raw, v) {
		if cs, ok := v.(values.ContextSetter); ok {
			cs.SetContext(ctx)
		}
	}
	return v
}

// areSamePointer reports whether a and b are the same underlying pointer.
// It handles pointer-typed values without triggering panic for uncomparable
// types (slice, map, func).
func areSamePointer(a, b any) bool {
	if a == nil || b == nil {
		return false
	}
	ra, rb := reflect.ValueOf(a), reflect.ValueOf(b)
	if ra.Kind() != reflect.Ptr || rb.Kind() != reflect.Ptr {
		return false
	}
	return ra.Pointer() == rb.Pointer()
}

// callComparisonHook calls the ComparisonHook if set, converting values.Value
// operands to their underlying Go values before passing to the hook.
// It also calls TypeMismatchHook when the operands have incompatible types.
func (ctx *context) callComparisonHook(op string, a, b values.Value, result bool) {
	if ctx.ComparisonHook != nil {
		ctx.ComparisonHook(op, a.Interface(), b.Interface(), result)
	}
	ctx.callTypeMismatchHookIfNeeded(op, a, b)
}

// callGroupBeginHook calls ComparisonGroupBeginHook if set.
// Called before evaluating the operands of an and/or expression.
func (ctx *context) callGroupBeginHook() {
	if ctx.ComparisonGroupBeginHook != nil {
		ctx.ComparisonGroupBeginHook()
	}
}

// callGroupEndHook calls ComparisonGroupEndHook if set.
// Called after evaluating an and/or expression.
func (ctx *context) callGroupEndHook(op string, result bool) {
	if ctx.ComparisonGroupEndHook != nil {
		ctx.ComparisonGroupEndHook(op, result)
	}
}

// isTypeMismatch returns true when a and b have fundamentally incompatible
// types for comparison — e.g. string vs numeric. Nil-nil and nil-nonnil are
// intentional Liquid behaviour and do not count as mismatches.
func isTypeMismatch(a, b values.Value) bool {
	ai, bi := a.Interface(), b.Interface()
	if ai == nil || bi == nil {
		return false
	}
	ra, rb := reflect.TypeOf(ai), reflect.TypeOf(bi)
	if ra == nil || rb == nil {
		return false
	}
	aIsNum := isNumericKind(ra.Kind())
	bIsNum := isNumericKind(rb.Kind())
	aIsStr := ra.Kind() == reflect.String
	bIsStr := rb.Kind() == reflect.String
	// string compared to number is a type mismatch
	return (aIsStr && bIsNum) || (aIsNum && bIsStr)
}

func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// callTypeMismatchHookIfNeeded calls TypeMismatchHook when a and b have
// incompatible types for the given comparison operator.
func (ctx *context) callTypeMismatchHookIfNeeded(op string, a, b values.Value) {
	if ctx.TypeMismatchHook != nil && isTypeMismatch(a, b) {
		ctx.TypeMismatchHook(op, a.Interface(), b.Interface())
	}
}

// callNilDereferenceHook calls NilDereferenceHook if set.
func (ctx *context) callNilDereferenceHook(object, property string) {
	if ctx.NilDereferenceHook != nil {
		ctx.NilDereferenceHook(object, property)
	}
}

// Set sets a variable value in the expression context.
func (ctx *context) Set(name string, value any) {
	ctx.bindings[name] = value
}
