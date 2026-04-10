package values

import (
	"sync"
)

type drop interface {
	ToLiquid() any
}

// dropMethodMissing is the internal interface for catch-all property access.
// Types implementing this will have MissingMethod called when a property or
// method is not found via reflection.
type dropMethodMissing interface {
	MissingMethod(key string) any
}

// ContextAccess is the minimal interface injected into drops that implement
// ContextSetter. It provides read/write access to the current rendering scope.
// Mirrors Ruby Liquid's Context object (limited to variable bindings).
type ContextAccess interface {
	Get(name string) any
	Set(name string, value any)
}

// ContextSetter is an optional interface for drop types that want to receive
// the current rendering context when they are looked up from the variable scope.
// SetContext is called by the renderer each time the drop is accessed.
// This mirrors Ruby Liquid's context= setter.
type ContextSetter interface {
	SetContext(ctx ContextAccess)
}

// ToLiquid converts an object to Liquid, if it implements the Drop interface.
func ToLiquid(value any) any {
	switch value := value.(type) {
	case drop:
		return value.ToLiquid()
	default:
		return value
	}
}

type dropWrapper struct {
	sync.Once

	d drop
	v Value
}

func (w *dropWrapper) Resolve() Value {
	w.Do(func() { w.v = ValueOf(w.d.ToLiquid()) })
	return w.v
}

func (w *dropWrapper) Equal(o Value) bool          { return w.Resolve().Equal(o) }
func (w *dropWrapper) Less(o Value) bool           { return w.Resolve().Less(o) }
func (w *dropWrapper) IndexValue(i Value) Value    { return w.Resolve().IndexValue(i) }
func (w *dropWrapper) Contains(o Value) bool       { return w.Resolve().Contains(o) }
func (w *dropWrapper) Int() int                    { return w.Resolve().Int() }
func (w *dropWrapper) Interface() any              { return w.Resolve().Interface() }
func (w *dropWrapper) PropertyValue(k Value) Value { return w.Resolve().PropertyValue(k) }
func (w *dropWrapper) Test() bool                  { return w.Resolve().Test() }
