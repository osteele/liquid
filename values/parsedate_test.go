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

	dt, err = ParseDate("1730040524")
	require.NoError(t, err)
	require.Equal(t, timeMustParse("2024-10-27T14:48:44Z"), dt.In(time.UTC))
}
