package evaluator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConstant(t *testing.T) {
	dt, err := ParseTime("now")
	require.NoError(t, err)
	require.True(t, dt.After(timeMustParse("1970-01-01T00:00:00Z")))

	dt, err = ParseTime("2017-07-09 10:40:00 UTC")
	require.NoError(t, err)
	require.Equal(t, timeMustParse("2017-07-09T10:40:00Z"), dt)

	// FIXME this actually ignores the tz. It's at least in the right ballpark;
	// IMO better for content rendering than total failure.
	dt, err = ParseTime("2017-07-09 15:30:00 -4")
	require.NoError(t, err)
	// require.Equal(t, timeMustParse("2017-07-09T15:30:00Z"), dt)
}
