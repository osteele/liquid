package evaluator

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
