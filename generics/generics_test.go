package generics

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var comparisonTests = []struct {
	a, b     interface{}
	expected bool
}{
	{0, 1, true},
	{1, 0, false},
	{1, 1, false},
	{1, 2.0, true},
	{1.0, 2, true},
	{2.0, 1, false},
	{"a", "b", true},
	{"b", "a", false},
	{nil, nil, false},
	{nil, 1, true},
	{1, nil, false},
	{false, true, true},
}

func TestLess(t *testing.T) {
	for i, test := range comparisonTests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			value := Less(test.a, test.b)
			require.Equalf(t, test.expected, value, "%v < %v", test.a, test.b)
		})
	}
}
