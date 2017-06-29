package chunks

// True returns true.
func True(_ Context) (interface{}, error) {
	return true, nil
}

// Negate negates its argument.
func Negate(f func(Context) (interface{}, error)) func(Context) (interface{}, error) {
	return func(ctx Context) (interface{}, error) {
		value, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return value == nil || value == false, nil
	}
}
