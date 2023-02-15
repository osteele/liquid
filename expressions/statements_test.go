package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStatement(t *testing.T) {
	stmt, err := ParseStatement(AssignStatementSelector, "a = b")
	require.NoError(t, err)
	require.Equal(t, "a", stmt.Assignment.Variable)

	stmt, err = ParseStatement(AssignStatementSelector, "a = 1 == 1")
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
