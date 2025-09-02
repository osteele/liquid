package expressions

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseTests = []struct {
	in     string
	expect interface{}
}{
	{`true`, true},
	{`false`, false},
	{`nil`, nil},
	{`2`, 2},
	{`"s"`, "s"},
	{`a`, 1},
	{`obj.prop`, 2},
	{`a | add: b`, 3},
	{`1 == 1`, true},
	{`1 != 1`, false},
	{`true and true`, true},
}

var parseErrorTests = []struct{ in, expected string }{
	{"a syntax error", "syntax error"},
	{`%assign a`, "syntax error"},
	{`%assign a 3`, "syntax error"},
	{`%cycle 'a' 'b'`, "syntax error"},
	{`%loop a in in`, "syntax error"},
	{`%when a b`, "syntax error"},
}

// Since the parser returns funcs, there's no easy way to test them except evaluation
func TestParse(t *testing.T) {
	cfg := NewConfig(gocontext.Background())
	cfg.AddFilter("add", func(a, b int) int { return a + b })
	ctx := NewContext(map[string]interface{}{
		"a":   1,
		"b":   2,
		"obj": map[string]int{"prop": 2},
	}, cfg)
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
