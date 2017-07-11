package liquid

// Drop indicates that the object will present to templates as its ToLiquid value.
type Drop interface {
	ToLiquid() interface{}
}

// FromDrop returns returns object.ToLiquid() if object's type implement this function;
// else the object itself.
func FromDrop(object interface{}) interface{} {
	switch object := object.(type) {
	case Drop:
		return object.ToLiquid()
	default:
		return object
	}
}
