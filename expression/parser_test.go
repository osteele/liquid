package expression

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseTests = []struct{ in, expected string }{
	{"a | filter: b", "parse error"},
}

var parseErrorTests = []struct{ in, expected string }{
	{"a syntax error", "parse error"},
}

func TestParse(t *testing.T) {
	cfg := NewConfig()
	cfg.AddFilter("filter", func(a, b int) int { return a + b })
	ctx := NewContext(map[string]interface{}{"a": 1, "b": 2}, cfg)
	for i, test := range parseTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			expr, err := Parse(test.in)
			require.NoError(t, err, test.in)
			value, err := expr.Evaluate(ctx)
			require.NoError(t, err, test.in)
			require.Equal(t, 3, value, test.in)
		})
	}
}

func TestParse_errors(t *testing.T) {
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			expr, err := Parse(test.in)
			require.Nilf(t, expr, test.in)
			require.Errorf(t, err, test.in, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
