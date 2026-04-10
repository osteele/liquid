package liquid

import "github.com/osteele/liquid/values"

// Drop indicates that the object will present to templates as its ToLiquid value.
type Drop interface {
	ToLiquid() any
}

// DropMethodMissing is an optional interface that custom Drop types may implement
// to handle property accesses for keys that are not defined as struct fields or
// methods. When Liquid looks up a property on a struct and finds nothing, it
// checks whether the struct implements DropMethodMissing and calls MissingMethod
// with the missing key name.
//
// This mirrors Ruby Liquid's liquid_method_missing and LiquidJS's liquidMethodMissing.
//
// Example:
//
//	type MyDrop struct{ Data map[string]any }
//
//	func (d MyDrop) MissingMethod(key string) any {
//	    return d.Data[key]
//	}
type DropMethodMissing interface {
	MissingMethod(key string) any
}

// DropRenderContext is the rendering context injected into drops that implement
// ContextDrop. It provides read/write access to the current rendering scope.
//
// This mirrors Ruby Liquid's Context object (scoped to variable bindings).
type DropRenderContext = values.ContextAccess

// ContextDrop is an optional interface for Drop types that need access to the
// current rendering context. When Liquid resolves a variable and the value
// implements ContextDrop, it calls SetContext with the current render context.
//
// This mirrors Ruby Liquid's context= setter and LiquidJS's contextDrop.
//
// Example:
//
//	type RegistersDrop struct {
//	    ctx liquid.DropRenderContext
//	}
//
//	func (d *RegistersDrop) ToLiquid() any { return d }
//
//	func (d *RegistersDrop) SetContext(ctx liquid.DropRenderContext) {
//	    d.ctx = ctx
//	}
//
//	func (d *RegistersDrop) CurrentUser() any {
//	    return d.ctx.Get("current_user")
//	}
type ContextDrop = values.ContextSetter

// FromDrop returns object.ToLiquid() if object's type implements this function;
// else the object itself.
func FromDrop(object any) any {
	switch object := object.(type) {
	case Drop:
		return object.ToLiquid()
	default:
		return object
	}
}
