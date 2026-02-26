package values

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConstant(t *testing.T) {
	dt, err := ParseDate("now")
	require.NoError(t, err)
	require.True(t, dt.After(timeMustParse("1970-01-01T00:00:00Z")))

	dt, err = ParseDate("2017-07-09 10:40:00 UTC")
	require.NoError(t, err)
	require.Equal(t, timeMustParse("2017-07-09T10:40:00Z"), dt)
}

func TestParseDate_numericString(t *testing.T) {
	// Unix timestamp as string
	dt, err := ParseDate("1152098955")
	require.NoError(t, err)
	require.Equal(t, time.Unix(1152098955, 0), dt)

	// Zero timestamp
	dt, err = ParseDate("0")
	require.NoError(t, err)
	require.Equal(t, time.Unix(0, 0), dt)
}
