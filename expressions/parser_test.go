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
}

var parseErrorTests = []struct{ in, expected string }{
	{"a syntax error", "parse error"},
	{`%assign a`, "parse error"},
	{`%assign a 3`, "parse error"},
	{`%cycle 'a' 'b'`, "parse error"},
	{`%loop a in in`, "parse error"},
	{`%when a b`, "parse error"},
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

func TestParseStatement(t *testing.T) {
	stmt, err := ParseStatement(AssignStatementSelector, "a = b")
	require.NoError(t, err)
	require.Equal(t, "a", stmt.Assignment.Variable)

	stmt, err = ParseStatement(CycleStatementSelector, "'a', 'b'")
	require.NoError(t, err)
	require.Equal(t, "", stmt.Cycle.Group)
	require.Len(t, stmt.Cycle.Values, 2)
	require.Equal(t, []string{"a", "b"}, stmt.Cycle.Values)

	stmt, err = ParseStatement(CycleStatementSelector, "'g': 'a', 'b'")
	require.NoError(t, err)
	require.Equal(t, "g", stmt.Cycle.Group)
	require.Len(t, stmt.Cycle.Values, 2)
	require.Equal(t, []string{"a", "b"}, stmt.Cycle.Values)

	stmt, err = ParseStatement(LoopStatementSelector, "x in array reversed offset: 2 limit: 3")
	require.NoError(t, err)
	require.Equal(t, "x", stmt.Loop.Variable)
	require.True(t, stmt.Loop.Reversed)
	require.Equal(t, 2, stmt.Loop.Offset)
	require.NotNil(t, stmt.Loop.Limit)
	require.Equal(t, 3, *stmt.Loop.Limit)

	stmt, err = ParseStatement(WhenStatementSelector, "a, b")
	require.NoError(t, err)
	require.Len(t, stmt.When.Exprs, 2)
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
