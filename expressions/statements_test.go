package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseStatement(t *testing.T) {
	stmt, err := ParseStatement(AssignStatementSelector, "a = b")
	require.NoError(t, err)
	require.Equal(t, "a", stmt.Assignment.Variable)
	require.Implements(t, (*Expression)(nil), stmt.ValueFn)

	stmt, err = ParseStatement(AssignStatementSelector, "a = 1 == 1")
	require.NoError(t, err)
	require.Equal(t, "a", stmt.Assignment.Variable)

	stmt, err = ParseStatement(CycleStatementSelector, "'a', 'b'")
	require.NoError(t, err)
	require.Empty(t, stmt.Group)
	require.Len(t, stmt.Values, 2)
	require.Equal(t, []string{"a", "b"}, stmt.Values)

	stmt, err = ParseStatement(CycleStatementSelector, "'g': 'a', 'b'")
	require.NoError(t, err)
	require.Equal(t, "g", stmt.Group)
	require.Len(t, stmt.Values, 2)
	require.Equal(t, []string{"a", "b"}, stmt.Values)

	stmt, err = ParseStatement(LoopStatementSelector, "x in array reversed offset: 2 limit: 3")
	require.NoError(t, err)
	require.Equal(t, "x", stmt.Loop.Variable)
	require.True(t, stmt.Reversed)

	require.Nil(t, stmt.Cols)
	require.NotNil(t, stmt.Limit)
	require.Implements(t, (*Expression)(nil), stmt.Limit)
	require.NotNil(t, stmt.Offset)
	require.Implements(t, (*Expression)(nil), stmt.Offset)

	stmt, err = ParseStatement(WhenStatementSelector, "a, b")
	require.NoError(t, err)
	require.Len(t, stmt.Exprs, 2)
}
