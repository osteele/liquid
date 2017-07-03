package liquid

// Drop indicates that the object will present to templates as its ToLiquid value.
type Drop interface {
	ToLiquid() interface{}
}
