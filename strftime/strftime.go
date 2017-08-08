// Package strftime implements a Strftime function that is compatible with Ruby's Time.strftime.
package strftime

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Strftime is compatible with Ruby's Time.strftime.
func Strftime(format string, t time.Time) (string, error) {
	return re.ReplaceAllStringFunc(format, func(directive string) string {
		var (
			m             = re.FindAllStringSubmatch(directive, 1)[0]
			flags         = m[1]
			width         = m[2]
			conversion, _ = utf8.DecodeRuneInString(m[3])
			c             = replaceComponent(t, conversion, flags, width)
			pad, w        = '0', 2
		)
		if s, ok := c.(string); ok {
			return s
		}
		if f, ok := padding[conversion]; ok {
			pad, w = f.c, f.w
		}
		switch flags {
		case "-":
			w = 0
		case "_":
			pad = ' '
		case "0":
			pad = '0'
		}
		if len(width) > 0 {
			w, _ = strconv.Atoi(width) // nolint: gas
		}
		fm := fmt.Sprintf("%%%c%dd", pad, w)
		if pad == '-' {
			fm = fmt.Sprintf("%%%dd", w)
		}
		s := fmt.Sprintf(fm, c)
		switch flags {
		case "^":
			return strings.ToUpper(s)
		// case "#":
		default:
			return s
		}
	}), nil
}

var re = regexp.MustCompile(`%([-_0]|::?)?(\d+)?[EO]?([a-zA-Z\+nt%])`)

var amPmTable = map[bool]string{true: "AM", false: "PM"}
var amPmLowerTable = map[bool]string{true: "am", false: "pm"}

var padding = map[rune]struct {
	c rune
	w int
}{
	'e': {'-', 2},
	'f': {'0', 6},
	'j': {'0', 3},
	'k': {'-', 2},
	'L': {'0', 3},
	'l': {'-', 2},
	'N': {'0', 9},
	'u': {'-', 0},
	'w': {'-', 0},
	'Y': {'0', 4},
}

func replaceComponent(t time.Time, c rune, flags, width string) interface{} { // nolint: gocyclo
	switch c {

	// Date
	case 'Y':
		return t.Year()
	case 'y':
		return t.Year() % 100
	case 'C':
		return t.Year() / 100

	case 'm':
		return t.Month()
	case 'B':
		return t.Month().String()
	case 'b', 'h':
		return t.Month().String()[:3]

	case 'd', 'e':
		return t.Day()

	case 'j':
		return t.YearDay()

	// Time
	case 'H', 'k':
		return t.Hour()
	case 'I', 'l':
		return (t.Hour()+11)%12 + 1
	case 'M':
		return t.Minute()
	case 'S':
		return t.Second()
	case 'L':
		return t.Nanosecond() / 1e6
	case 'N':
		ns := t.Nanosecond()
		if len(width) > 0 {
			w, _ := strconv.Atoi(width) // nolint: gas
			if w <= 9 {
				return fmt.Sprintf("%09d", ns)[:w]
			}
			return fmt.Sprintf(fmt.Sprintf("%%09d%%0%dd", w-9), ns, 0)
		}
		return ns

	case 'P':
		return amPmLowerTable[t.Hour() < 12]
	case 'p':
		return amPmTable[t.Hour() < 12]

	// Time zone
	case 'z':
		_, offset := t.Zone()
		var (
			h = offset / 3600
			m = (offset / 60) % 60
		)
		switch flags {
		case ":":
			return fmt.Sprintf("%+03d:%02d", h, m)
		case "::":
			return fmt.Sprintf("%+03d:%02d:%02d", h, m, offset%60)
		default:
			return fmt.Sprintf("%+03d%02d", h, m)
		}
	case 'Z':
		z, _ := t.Zone()
		return z

	// Weekday
	case 'A':
		return t.Weekday().String()
	case 'a':
		return t.Weekday().String()[:3]
	case 'u':
		return (t.Weekday()+6)%7 + 1
	case 'w':
		return t.Weekday()

	// ISO Year
	case 'G':
		y, _ := t.ISOWeek()
		return y
	case 'g':
		y, _ := t.ISOWeek()
		return y % 100
	case 'V':
		_, wn := t.ISOWeek()
		return wn

	// ISO Week
	case 'U':
		t = t.Add(24 * time.Hour)
		y, wn := t.ISOWeek()
		if y < t.Year() {
			wn = 0
		}
		return wn
	case 'W':
		y, wn := t.ISOWeek()
		if y < t.Year() {
			wn = 0
		}
		return wn

	// Epoch seconds
	case 's':
		return t.Unix()
	case 'Q':
		return t.UnixNano() / 1000

	// Literals
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case '%':
		return "%"

	// Combinations
	case 'c':
		// date and time (%a %b %e %T %Y)
		h, m, s := t.Clock()
		return fmt.Sprintf("%s %s %2d %02d:%02d:%02d %04d", t.Weekday().String()[:3], t.Month().String()[:3], t.Day(), h, m, s, t.Year())
	case 'D', 'x':
		// Date (%m/%d/%y)
		y, m, d := t.Date()
		return fmt.Sprintf("%02d/%02d/%02d", m, d, y%100)
	case 'F':
		// The ISO 8601 date format (%Y-%m-%d)
		y, m, d := t.Date()
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	case 'v':
		// VMS date (%e-%b-%Y)
		return fmt.Sprintf("%2d-%s-%04d", t.Day(), t.Month().String()[:3], t.Year())
	case 'f':
		return t.Nanosecond() / 1e3
	case 'r':
		// 12-hour time (%I:%M:%S %p)
		h, m, s := t.Clock()
		h12 := (h+11)%12 + 1
		return fmt.Sprintf("%02d:%02d:%02d %s", h12, m, s, amPmTable[h < 12])
	case 'R':
		// 24-hour time (%H:%M)
		h, m, _ := t.Clock()
		return fmt.Sprintf("%02d:%02d", h, m)
	case 'T', 'X':
		// 24-hour time (%H:%M:%S)
		h, m, s := t.Clock()
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	case '+':
		// date(1) (%a %b %e %H:%M:%S %Z %Y)
		s, err := Strftime("%a %b %e %H:%M:%S %Z %Y", t)
		if err != nil {
			panic(err)
		}
		return s
	default:
		return fmt.Sprintf("%%%c", c)
	}
}
