package generics

import (
	"reflect"
	"time"

	"github.com/jeffjen/datefmt"
)

var zeroTime time.Time

var dateFormats = []string{
	"%Y-%m-%d %H:%M:%S %Z",
}

var dateLayouts = []string{
	"2006-01-02 15:04:05 -07:00",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05 -7",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"January 2, 2006",
	"January 2 2006",
	"Jan 2, 2006",
	"Jan 2 2006",
}

// ParseTime tries a few heuristics to parse a date from a string
func ParseTime(s string) (time.Time, error) {
	if s == "now" {
		return time.Now(), nil
	}
	for _, layout := range dateLayouts {
		// fmt.Printf("%s\n%s\n\n", time.Now().Format(layout), s)
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	for _, format := range dateFormats {
		// xx, _ := datefmt.Strftime(format, time.Now())
		// fmt.Printf("%s\n%s\n\n", xx, s)
		t, err := datefmt.Strptime(format, s)
		if err == nil {
			return t, nil
		}
	}
	return zeroTime, conversionError("", s, reflect.TypeOf(zeroTime))
}
