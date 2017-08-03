package values

import (
	"sync"
)

type drop interface {
	ToLiquid() interface{}
}

// ToLiquid converts an object to Liquid, if it implements the Drop interface.
func ToLiquid(value interface{}) interface{} {
	switch value := value.(type) {
	case drop:
		return value.ToLiquid()
	default:
		return value
	}
}

type dropWrapper struct {
	d drop
	v Value
	sync.Once
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
func (w *dropWrapper) Interface() interface{}      { return w.Resolve().Interface() }
func (w *dropWrapper) PropertyValue(k Value) Value { return w.Resolve().PropertyValue(k) }
func (w *dropWrapper) Test() bool                  { return w.Resolve().Test() }
