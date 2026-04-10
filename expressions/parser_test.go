package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseTests = []struct {
	in     string
	expect any
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
	{`1 <> 1`, false},
	{`not false`, true},
	{`not true`, false},
	{`true and true`, true},
	// keyword arg: add_kw is registered with allow_false kwarg below
	{`a | add_kw: b, flag: true`, 3},
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
	cfg := NewConfig()
	cfg.AddFilter("add", func(a, b int) int { return a + b })
	// add_kw strips NamedArg kwargs and adds a+b, to test keyword arg parsing
	cfg.AddFilter("add_kw", func(a int, args ...any) int {
		var b int
		for _, arg := range args {
			if na, ok := arg.(NamedArg); ok {
				_ = na // ignore keyword args in this test filter
				continue
			}
			if n, ok := arg.(int); ok {
				b = n
			}
		}
		return a + b
	})
	ctx := NewContext(map[string]any{
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
