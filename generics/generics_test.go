package generics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
	{[]string{"a"}, []string{"a"}, true},
	{[]string{"a"}, []string{"a", "b"}, false},
	{[]string{"a", "b"}, []string{"a"}, false},
	{[]string{"a", "b"}, []string{"a", "b"}, true},
	{[]string{"a", "b"}, []string{"a", "c"}, false},
	{[]interface{}{1.0, 2}, []interface{}{1, 2.0}, true},
}

var lessTests = []struct {
	a, b     interface{}
	expected bool
}{
	{nil, nil, false},
	{false, true, true},
	{false, false, false},
	{false, nil, false},
	{nil, false, false},
	{0, 1, true},
	{1, 0, false},
	{1, 1, false},
	{1, 2.1, true},
	{1.1, 2, true},
	{2.1, 1, false},
	{"a", "b", true},
	{"b", "a", false},
	{[]string{"a"}, []string{"a"}, false},
}

func TestEqual(t *testing.T) {
	for i, test := range eqTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value := Equal(test.a, test.b)
			require.Equalf(t, test.expected, value, "%#v == %#v", test.a, test.b)
		})
	}
}

func TestLess(t *testing.T) {
	for i, test := range lessTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			value := Less(test.a, test.b)
			require.Equalf(t, test.expected, value, "%#v < %#v", test.a, test.b)
		})
	}
}
