package values

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var eqTestObj = struct{ a, b int }{1, 2}
var eqArrayTestObj = [2]int{1, 2}

var eqTests = []struct {
	a, b     interface{}
	expected bool
}{
	{nil, nil, true},
	{nil, 1, false},
	{1, nil, false},
	{false, false, true},
	{false, true, false},
	{0, 1, false},
	{1, 1, true},
	{1.0, 1.0, true},
	{1, 1.0, true},
	{1, 2.0, false},
	{1.0, 1, true},
	{"a", "b", false},
	{"a", "a", true},
	{int8(2), int16(2), true}, // TODO
	// {uint8(2), int8(2), true}, // FIXME
	{eqArrayTestObj, eqArrayTestObj[:], true},
	{[]string{"a"}, []string{"a"}, true},
	{[]string{"a"}, []string{"a", "b"}, false},
	{[]string{"a", "b"}, []string{"a"}, false},
	{[]string{"a", "b"}, []string{"a", "b"}, true},
	{[]string{"a", "b"}, []string{"a", "c"}, false},
	{[]interface{}{1.0, 2}, []interface{}{1, 2.0}, true},
	{eqTestObj, eqTestObj, true},
}

func TestEqual(t *testing.T) {
	for i, test := range eqTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value := Equal(test.a, test.b)
			require.Equalf(t, test.expected, value, "%#v == %#v", test.a, test.b)
		})
	}
}

func TestEqual_ptr(t *testing.T) {
	var (
		n  int
		f  float64
		pn *int
		pf *float64
		s  struct{}
	)
	require.True(t, Equal(&s, &s))
	require.True(t, Equal(&n, &n))
	require.False(t, Equal(&n, &f))

	// // null pointers
	require.True(t, Equal(pn, pn))
	require.False(t, Equal(pn, &n))
	// null pointers should compare equal, even if they're different types
	require.True(t, Equal(pn, pf))
	// require.True(t, Equal(pn, nil)) // TODO
	// require.True(t, Equal(nil, pn)) // TODO
}
