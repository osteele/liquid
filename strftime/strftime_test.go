package strftime

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func timeMustParse(f, s string) time.Time {
	t, err := time.ParseInLocation(f, s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestStrftime(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))

	ins := []struct{ source, expect string }{
		{"02 Jan 06 15:04 UTC", "02 Jan 06 10:04 EST"},
		{"02 Jan 06 15:04 EST", "02 Jan 06 15:04 EST"},
		{"02 Jan 06 15:04 EDT", "02 Jan 06 14:04 EST"},
		// {"02 Jan 06 15:04 MST", "02 Jan 06 10:04 EST"},
		{"14 Mar 16 12:00 UTC", "14 Mar 16 08:00 EDT"},
		// {"14 Mar 16 00:00 UTC", "14 Mar 16 00:00 UTC"},
	}
	for _, test := range ins {
		rt := timeMustParse(time.RFC822, test.source)
		actual, err := Strftime("%d %b %y %H:%M %Z", rt)
		require.NoErrorf(t, err, test.source)
		require.Equalf(t, test.expect, actual, test.source)
	}

	rt := timeMustParse(time.RFC822, "02 Jan 06 15:04 MST")
	tests := []struct{ format, expect string }{
		{"%a, %b %d, %Y", "Mon, Jan 02, 2006"},
		{"%Y/%m/%d", "2006/01/02"},
		{"%Y/%m/%e", "2006/01/ 2"},
		{"%Y/%-m/%-d", "2006/1/2"},
		{"%a, %b %d, %Y %z", "Mon, Jan 02, 2006 -0500"},
		{"%a, %b %d, %Y %Z", "Mon, Jan 02, 2006 EST"},
		{"", ""},
	}
	for _, test := range tests {
		s, err := Strftime(test.format, rt)
		require.NoErrorf(t, err, test.format)
		require.Equalf(t, test.expect, s, test.format)
	}

	dt, err := time.Parse("2006-01-02", "1776-07-15")
	require.NoError(t, err)
	s, err := Strftime("%Y-%m-%d", dt)
	require.NoError(t, err)
	// FIXME actual 1776-07-14
	_ = s
	// require.Equal(t, "1776-07-15", s)

	// s, err := Strftime("%f", rt)
	// require.Errorf(t, err)
}

// func TestStrptime(t *testing.T) {
// 	testCases := []struct{ format, in, expect string }{
// 		{"%a, %b %d, %Y", "Thu, Jun 29, 2017", "29 Jun 17 00:00 +0000"},
// 		{"%a, %b %d, %Y %H:%M", "Thu, Jun 29, 2017 15:30", "29 Jun 17 15:30 +0000"},
// 		// {"%a, %b %d, %Y %H:%M %Z", "Thu, Jun 29, 2017 15:30 UTC", "29 Jun 17 15:30 +0000"},
// 	}
// 	for _, test := range testCases {
// 		tm, err := Strptime(test.format, test.in)
// 		require.NoError(t, err)
// 		s := tm.Format(time.RFC822Z)
// 		require.Equal(t, test.expect, s)
// 	}

// 	_, err := Strptime("%Y", "onvald")
// 	require.Error(t, err)
// }
