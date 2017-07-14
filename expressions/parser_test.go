package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseTests = []struct {
	in     string
	expect interface{}
}{
	{`a | filter: b`, 3},
	// {`%assign a = 3`, nil},
	// {`{%cycle 'a'`, []interface{}{"a"}},
	// {`{%cycle 'a', 'b'`, []interface{}{"a", "b"}},
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
			_ = expr
			value, err := expr.Evaluate(ctx)
			require.NoError(t, err, test.in)
			require.Equal(t, test.expect, value, test.in)
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
