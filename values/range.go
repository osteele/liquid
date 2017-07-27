package values

// A Range is the range of integers from b to e inclusive.
type Range struct {
	b, e int
}

// NewRange returns a new Range
func NewRange(b, e int) Range {
	return Range{b, e}
}

// Len is in the iteration interface
func (r Range) Len() int { return r.e + 1 - r.b }

// Index is in the iteration interface
func (r Range) Index(i int) interface{} { return r.b + i }
