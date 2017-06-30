package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
// {"a | unknown_filter", "undefined filter: unknown_filter"},
}

func TestParseErrors(t *testing.T) {
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			expr, err := Parse(test.in)
			require.Nilf(t, expr, test.in)
			require.Errorf(t, err, test.in, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
