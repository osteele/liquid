package strftime

import (
	"encoding/csv"
	"fmt"
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

var conversionTests = []struct{ format, expect string }{
	{"%1N", "1"},
	{"%3N", "123"},
	{"%6N", "123456"},
	{"%9N", "123456789"},
	{"%12N", "123456789000"},
	{"%v", " 2-Jan-2006"},
	{"%Z", "EST"},
	{"%:z", "-05:00"},
	{"%::z", "-05:00:00"},
	{"%%", "%"},
}

var dayOfWeekTests = []string{
	"%A=Sunday %a=Sun %u=7 %w=0 %d=01 %e= 1 %j=001 %U=01 %V=52 %W=00",
	"%A=Monday %a=Mon %u=1 %w=1 %d=02 %e= 2 %j=002 %U=01 %V=01 %W=01",
	"%A=Tuesday %a=Tue %u=2 %w=2 %d=03 %e= 3 %j=003 %U=01 %V=01 %W=01",
	"%A=Wednesday %a=Wed %u=3 %w=3 %d=04 %e= 4 %j=004 %U=01 %V=01 %W=01",
	"%A=Thursday %a=Thu %u=4 %w=4 %d=05 %e= 5 %j=005 %U=01 %V=01 %W=01",
	"%A=Friday %a=Fri %u=5 %w=5 %d=06 %e= 6 %j=006 %U=01 %V=01 %W=01",
	"%A=Saturday %a=Sat %u=6 %w=6 %d=07 %e= 7 %j=007 %U=01 %V=01 %W=01",
}

var hourTests = []struct {
	hour   int
	expect string
}{
	{0, "%H=00 %k= 0 %I=12 %l=12 %P=am %p=AM"},
	{1, "%H=01 %k= 1 %I=01 %l= 1 %P=am %p=AM"},
	{12, "%H=12 %k=12 %I=12 %l=12 %P=pm %p=PM"},
	{13, "%H=13 %k=13 %I=01 %l= 1 %P=pm %p=PM"},
	{23, "%H=23 %k=23 %I=11 %l=11 %P=pm %p=PM"},
}

func TestStrftime(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))

	dt := timeMustParse(time.RFC3339Nano, "2006-01-02T15:04:05.123456789-05:00")
	for _, test := range conversionTests {
		name := fmt.Sprintf("Strftime %q", test.format)
		actual, err := Strftime(test.format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, test.expect, actual, name)
	}

	skip := map[string]bool{}
	f, err := os.Open("testdata/skip.csv")
	require.NoError(t, err)
	defer f.Close() // nolint: errcheck
	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	require.NoError(t, err)
	for _, row := range rows {
		skip[row[0]] = true
	}

	f, err = os.Open("testdata/data.csv")
	require.NoError(t, err)
	defer f.Close() // nolint: errcheck
	r = csv.NewReader(f)
	rows, err = r.ReadAll()
	require.NoError(t, err)
	for _, row := range rows {
		format, expect := row[0], row[1]
		if skip[format] {
			continue
		}
		name := fmt.Sprintf("Strftime %q", format)
		actual, err := Strftime(format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, expect, actual, name)
	}

	dt = timeMustParse(time.RFC1123Z, "Mon, 02 Jan 2006 15:04:05 -0500")
	tests := []struct{ format, expect string }{
		{"%a, %b %d, %Y", "Mon, Jan 02, 2006"},
		{"%Y/%m/%d", "2006/01/02"},
		{"%Y/%m/%e", "2006/01/ 2"},
		{"%Y/%-m/%-d", "2006/1/2"},
		{"%a, %b %d, %Y %z", "Mon, Jan 02, 2006 -0500"},
		{"%a, %b %d, %Y %Z", "Mon, Jan 02, 2006 EST"},
		// {"", ""}, this errors on Travis
	}
	for _, test := range tests {
		s, err := Strftime(test.format, dt)
		require.NoErrorf(t, err, test.format)
		require.Equalf(t, test.expect, s, test.format)
	}
}

func TestStrftime_dow(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))
	for day, expect := range dayOfWeekTests {
		dt := time.Date(2006, 01, day+1, 15, 4, 5, 0, time.UTC)
		format := "%%A=%A %%a=%a %%u=%u %%w=%w %%d=%d %%e=%e %%j=%j %%U=%U %%V=%V %%W=%W"
		name := fmt.Sprintf("%s.Strftime", dt)
		actual, err := Strftime(format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, expect, actual, name)
	}
}

func TestStrftime_hours(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))
	for _, test := range hourTests {
		dt := time.Date(2006, 01, 2, test.hour, 4, 5, 0, time.UTC)
		format := "%%H=%H %%k=%k %%I=%I %%l=%l %%P=%P %%p=%p"
		name := fmt.Sprintf("%s.Strftime", dt)
		actual, err := Strftime(format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, test.expect, actual, name)
	}
}

func TestStrftime_zones(t *testing.T) {
	require.NoError(t, os.Setenv("TZ", "America/New_York"))
	ins := []struct{ source, expect string }{
		{"02 Jan 06 15:04 UTC", "%z=+0000 %Z=UTC"},
		{"02 Jan 06 15:04 EST", "%z=-0500 %Z=EST"},
		{"02 Jul 06 15:04 EDT", "%z=-0400 %Z=EDT"},
	}
	for _, test := range ins {
		rt := timeMustParse(time.RFC822, test.source)
		actual, err := Strftime("%%z=%z %%Z=%Z", rt)
		require.NoErrorf(t, err, test.source)
		require.Equalf(t, test.expect, actual, test.source)
	}
}
