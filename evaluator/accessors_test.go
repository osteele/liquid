package evaluator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var indexTests = []struct{ array, index, expect interface{} }{
	{[]int{1, 2, 3}, 1, 2},
	{[]int{1, 2, 3}, 1.0, 2},
	{[]int{1, 2, 3}, 1.1, nil},
}

func TestIndex(t *testing.T) {
	for i, test := range indexTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			actual := Index(test.array, test.index)
			require.Equalf(t, test.expect, actual, "%v[%v]", test.array, test.index)
		})
	}
}
