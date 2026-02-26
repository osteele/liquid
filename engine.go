package liquid

import (
	"context"
	"fmt"
	"html"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/autopilot3/ap3-helpers-go/logger"
	"github.com/autopilot3/ap3-types-go/types/date"
	"github.com/autopilot3/ap3-types-go/types/phone"
	"github.com/autopilot3/liquid/filters"
	"github.com/autopilot3/liquid/render"
	"github.com/autopilot3/liquid/tags"

	"github.com/bojanz/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// An Engine parses template source into renderable text.
//
// An engine can be configured with additional filters and tags.
type Engine struct{ cfg render.Config }

func formatDate(d date.Date, format string) string {
	// Convert to time.Time for standard formatting
	t, err := d.Time()
	if err != nil {
		return d.String()
	}

	switch format {
	case "mdy":
		return t.Format("01/02/2006")
	case "dmy":
		return t.Format("02/01/2006")
	case "ymd":
		return t.Format("2006/01/02")
	case "ydm":
		return t.Format("2006/02/01")
	// US formats
	case "mdyaw":
		return t.Format("Monday, January 2, 2006")
	case "mdya":
		return t.Format("January 2, 2006")
	case "mdys":
		return t.Format("1/2/06")
	// Everyone else formats
	case "dmyaw":
		return t.Format("Monday, 2 January, 2006")
	case "dmya":
		return t.Format("2 January, 2006")
	case "dmys":
		return t.Format("2/1/06")
	// Individual pieces
	case "d":
		return t.Format("2")
	case "dd":
		return t.Format("02")
	case "m":
		return t.Format("1")
	case "mm":
		return t.Format("01")
	case "yy":
		return t.Format("06")
	case "yyyy":
		return t.Format("2006")
	default:
		return d.String()
	}
}

func lowerMeridiem(value string) string {
	value = strings.ReplaceAll(value, "AM", "am")
	return strings.ReplaceAll(value, "PM", "pm")
}

func formatDateTime(t time.Time, format string) string {
	switch format {
	case "mdy12":
		return t.Format("Jan 02 2006 3:04 PM")
	case "mdy24":
		return t.Format("Jan 02 2006 15:04")
	case "dmy12":
		return t.Format("02 Jan 2006 3:04 PM")
	case "dmy24":
		return t.Format("02 Jan 2006 15:04")
	case "ymd12":
		return t.Format("2006 Jan 02 3:04 PM")
	case "ymd24":
		return t.Format("2006 Jan 02 15:04")
	case "ydm12":
		return t.Format("2006 02 Jan 3:04 PM")
	case "ydm24":
		return t.Format("2006 02 Jan 15:04")
	// US formats
	case "mdy24aw":
		return t.Format("15:04 Monday, January 2, 2006")
	case "mdy12aw":
		return lowerMeridiem(t.Format("3:04PM Monday, January 2, 2006"))
	case "mdyaw":
		return t.Format("Monday, January 2, 2006")
	case "mdy24a":
		return t.Format("15:04 January 2, 2006")
	case "mdy12a":
		return lowerMeridiem(t.Format("3:04PM January 2, 2006"))
	case "mdya":
		return t.Format("January 2, 2006")
	case "mdy24n":
		return t.Format("15:04 01/02/2006")
	case "mdy12n":
		return lowerMeridiem(t.Format("3:04PM 01/02/2006"))
	case "mdy24nd":
		return t.Format("01/02/2006 15:04")
	case "mdy12nd":
		return lowerMeridiem(t.Format("01/02/2006 3:04PM"))
	case "mdy":
		return t.Format("01/02/2006")
	case "mdys24":
		return t.Format("15:04 1/2/06")
	case "mdys12":
		return lowerMeridiem(t.Format("3:04PM 1/2/06"))
	case "mdys24d":
		return t.Format("1/2/06 15:04")
	case "mdys12d":
		return lowerMeridiem(t.Format("1/2/06 3:04PM"))
	case "mdys":
		return t.Format("1/2/06")
	// Everyone else formats
	case "dmy24aw":
		return t.Format("15:04 Monday, 2 January, 2006")
	case "dmy12aw":
		return lowerMeridiem(t.Format("3:04PM Monday, 2 January, 2006"))
	case "dmyaw":
		return t.Format("Monday, 2 January, 2006")
	case "dmy24a":
		return t.Format("15:04 2 January, 2006")
	case "dmy12a":
		return lowerMeridiem(t.Format("3:04PM 2 January, 2006"))
	case "dmya":
		return t.Format("2 January, 2006")
	case "dmy24n":
		return t.Format("15:04 02/01/2006")
	case "dmy12n":
		return lowerMeridiem(t.Format("3:04PM 02/01/2006"))
	case "dmy24nd":
		return t.Format("02/01/2006 15:04")
	case "dmy12nd":
		return lowerMeridiem(t.Format("02/01/2006 3:04PM"))
	case "dmy":
		return t.Format("02/01/2006")
	case "dmys24":
		return t.Format("15:04 2/1/06")
	case "dmys12":
		return lowerMeridiem(t.Format("3:04PM 2/1/06"))
	case "dmys24d":
		return t.Format("2/1/06 15:04")
	case "dmys12d":
		return lowerMeridiem(t.Format("2/1/06 3:04PM"))
	case "dmys":
		return t.Format("2/1/06")
	// Individual pieces
	case "h24":
		return t.Format("15")
	case "h12":
		return t.Format("3")
	case "min":
		return t.Format("04")
	case "p":
		return lowerMeridiem(t.Format("PM"))
	case "d":
		return t.Format("2")
	case "dd":
		return t.Format("02")
	case "dow":
		return t.Format("Monday")
	case "m":
		return t.Format("1")
	case "mm":
		return t.Format("01")
	case "mon":
		return t.Format("January")
	case "yy":
		return t.Format("06")
	case "yyyy":
		return t.Format("2006")
	default:
		return t.String()
	}
}

func (e *Engine) SetAllowedTags(allowedTags map[string]struct{}) *Engine {
	e.cfg.AllowedTags = allowedTags
	return e
}
func (e *Engine) AllowedTagsWithDefault() *Engine {
	e.cfg.AllowTagsWithDefault = true
	return e
}

func NewEngine() *Engine {
	return NewEngineWithContext(context.Background())
}

// NewEngine returns a new Engine.
func NewEngineWithContext(ctx context.Context) *Engine {
	engine := &Engine{render.NewConfigWitchContext(ctx)}
	filters.AddStandardFilters(&engine.cfg)
	tags.AddStandardTags(engine.cfg)
	engine.RegisterFilter("hideCountryCodeAndDefault", func(v interface{}, hide bool, defaultValue string) string {
		s, ok := v.(phone.International)
		if !ok {
			return defaultValue
		}
		if s.Number.IsZero() && s.CountryCode.IsZero() {
			return defaultValue
		}
		if hide {
			return s.Number.String()
		}
		return s.String()
	})

	engine.RegisterFilter("timeInTimezone", func(s time.Time, timezone string, format string) string {
		tz, err := time.LoadLocation(timezone)
		if err != nil {
			return ""
		}
		st := s.In(tz)
		return formatDateTime(st, format)
	})

	engine.RegisterFilter("rawPhone", func(s phone.International) string {
		return s.CountryCode.String() + s.Number.String()
	})

	engine.RegisterFilter("dateTimeFormatOrDefault", func(s time.Time, format string, defaultValue string) string {
		if s.IsZero() {
			return defaultValue
		}
		return formatDateTime(s, format)
	})

	engine.RegisterFilter("dateFormatOrDefault", func(s interface{}, format string, defaultValue string) string {
		var (
			d   date.Date
			err error
		)
		switch s := s.(type) {
		case date.Date:
			d = s
		case time.Time:
			d, err = date.NewFromTime(s)
			if err != nil {
				return defaultValue
			}
		}
		if d.IsZero() {
			return defaultValue
		}
		return formatDate(d, format)
	})

	engine.RegisterFilter("decimal", func(s string, format string, currency string) string {
		if s == "" {
			return s
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger.Warnw(engine.cfg.Context(), fmt.Sprintf("failed to parse field value %s to decimal: %s", s, err.Error()), "lqiuid", "filter")
			return s
		}
		var formatTemplate string
		switch format {
		case "whole":
			formatTemplate = "%.0f"
		case "one":
			formatTemplate = "%.1f"
		case "two":
			formatTemplate = "%.2f"
		default:
			formatTemplate = "%.2f"
		}

		p := message.NewPrinter(language.English)
		value := p.Sprintf(formatTemplate, float64(num)/1000)
		if currency != "" {
			return currency + value
		}

		return value
	})

	engine.RegisterFilter("decimalWithDelimiter", func(s string, format string, currencyCode string, loc string) string {
		if s == "" {
			return s
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger.Warnw(engine.cfg.Context(), fmt.Sprintf("failed to parse field value %s to decimal: %s", s, err.Error()), "lqiuid", "filter")
			return s
		}

		if isoCode := strings.TrimSpace(currencyCode); len(isoCode) == 3 { // iso code
			amount, err := currency.NewAmount(fmt.Sprintf("%.3f", num/1000), isoCode)
			if err == nil {
				locale := currency.NewLocale(loc)
				formatter := currency.NewFormatter(locale)

				switch format {
				case "whole":
					formatter.MaxDigits = 0
				case "one":
					formatter.MaxDigits = 1
				case "two":
					formatter.MaxDigits = 2
				default:
					formatter.MaxDigits = 2
				}
				return formatter.Format(amount)
			} else {
				// log and fallback to the previous logic
				logger.Warnw(engine.cfg.Context(), fmt.Sprintf("failed to parse field value %s with currency code %s to decimal: %s", s, isoCode, err.Error()), "lqiuid", "filter")
			}
		}
		var formatTemplate string
		switch format {
		case "whole":
			formatTemplate = "%.0f"
		case "one":
			formatTemplate = "%.1f"
		case "two":
			formatTemplate = "%.2f"
		default:
			formatTemplate = "%.2f"
		}

		tag, err := language.Parse(loc)
		if err != nil {
			tag = language.English
		}
		p := message.NewPrinter(tag)
		val := p.Sprintf(formatTemplate, num/1000)
		if currencyCode != "" {
			return currencyCode + val
		}
		return val
	})

	engine.RegisterFilter("numberWithDelimiter", func(s string, loc string, format string) string {
		if s == "" {
			return s
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger.Warnw(engine.cfg.Context(), fmt.Sprintf("failed to parse field value %s to decimal: %s", s, err.Error()), "lqiuid", "filter")
			return s
		}
		var formatTemplate string
		switch format {
		case "whole":
			formatTemplate = "%.0f"
		case "one":
			formatTemplate = "%.1f"
		case "two":
			formatTemplate = "%.2f"
		default:
			formatTemplate = "%.2f"
		}

		tag, err := language.Parse(loc)
		if err != nil {
			tag = language.English
		}
		p := message.NewPrinter(tag)
		value := p.Sprintf(formatTemplate, float64(num))

		return value
	})

	engine.RegisterFilter("booleanFormat", func(s string, format string) string {
		if s == "" {
			return ""
		}
		var b bool
		if s == "true" {
			b = true
		}
		if format == "yesNo" {
			if b {
				return "Yes"
			}
			return "No"
		}
		if format == "onOff" {
			if b {
				return "On"
			}
			return "Off"
		}
		if b {
			return "True"
		}
		return "False"
	})

	// a set [a,b,c] contains at least one of matches, [a,d] will return true in this case
	engine.RegisterFilter("setContains", func(s interface{}, matches ...interface{}) bool {
		if s == nil {
			return false
		}
		switch k := reflect.TypeOf(s).Kind(); k {
		case reflect.String:
			str := s.(string)
			splits := strings.Split(str, ",")
			for _, match := range matches {
				for _, s := range splits {
					if s == match {
						return true
					}
				}
			}
			return false
		case reflect.Slice:
			val := reflect.ValueOf(s)
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				for _, match := range matches {
					if reflect.DeepEqual(elem, match) {
						return true
					}
				}
			}
			return false
		}
		return false
	})

	// a set [a,b,c] contains all matches, [a,d] will return false in this case, [a,c] will return true
	engine.RegisterFilter("setContainsAll", func(s interface{}, matches ...interface{}) bool {
		if s == nil {
			return false
		}
		switch k := reflect.TypeOf(s).Kind(); k {
		case reflect.String:
			str := s.(string)
			splits := strings.Split(str, ",")
			for _, match := range matches {
				containMatch := false
				for _, s := range splits {
					if s == match {
						containMatch = true
						break
					}
				}
				if !containMatch {
					return false
				}
			}
			return true
		case reflect.Slice:
			val := reflect.ValueOf(s)
			for i := 0; i < val.Len(); i++ {
				elem := val.Index(i).Interface()
				containMatch := false
				for _, match := range matches {
					if reflect.DeepEqual(elem, match) {
						containMatch = true
						break
					}
				}
				if !containMatch {
					return false
				}
			}
			return true
		}
		return false
	})

	// trackURL here is a dummy filter, it is used to avoid error when parsing liquid template. Services support this filter will replace it with real filter
	engine.RegisterFilter("trackURL", func(s string) string {
		return "$$TRACK_ME:" + html.UnescapeString(s) + "$$"
	})

	engine.RegisterFilter("startsWith", func(s string, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})

	engine.RegisterFilter("endsWith", func(s string, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	})

	return engine
}

// RegisterBlock defines a block e.g. {% tag %}…{% endtag %}.
func (e *Engine) RegisterBlock(name string, td Renderer) {
	e.cfg.AddBlock(name).Renderer(func(w io.Writer, ctx render.Context) error {
		s, err := td(ctx)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, s)
		return err
	})
}

// RegisterFilter defines a Liquid filter, for use as `{{ value | my_filter }}` or `{{ value | my_filter: arg }}`.
//
// A filter is a function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must have type error.
//
// Examples:
//
// * https://github.com/autopilot3/liquid/blob/master/filters/filters.go
//
// * https://github.com/osteele/gojekyll/blob/master/filters/filters.go
func (e *Engine) RegisterFilter(name string, fn interface{}) {
	e.cfg.AddFilter(name, fn)
}

// RegisterTag defines a tag e.g. {% tag %}.
//
// Further examples are in https://github.com/osteele/gojekyll/blob/master/tags/tags.go
func (e *Engine) RegisterTag(name string, td Renderer) {
	// For simplicity, don't expose the two stage parsing/rendering process to clients.
	// Client tags do everything at runtime.
	e.cfg.AddTag(name, func(_ string) (func(io.Writer, render.Context) error, error) {
		return func(w io.Writer, ctx render.Context) error {
			s, err := td(ctx)
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, s)
			return err
		}, nil
	})
}

// ParseTemplate creates a new Template using the engine configuration.
func (e *Engine) ParseTemplate(source []byte) (*Template, SourceError) {
	return newTemplate(&e.cfg, source, "", 0)
}

// ParseString creates a new Template using the engine configuration.
func (e *Engine) ParseString(source string) (*Template, SourceError) {
	return e.ParseTemplate([]byte(source))
}

// ParseTemplateLocation is the same as ParseTemplate, except that the source location is used
// for error reporting and for the {% include %} tag.
//
// The path and line number are used for error reporting.
// The path is also the reference for relative pathnames in the {% include %} tag.
func (e *Engine) ParseTemplateLocation(source []byte, path string, line int) (*Template, SourceError) {
	return newTemplate(&e.cfg, source, path, line)
}

// ParseAndRender parses and then renders the template.
func (e *Engine) ParseAndRender(source []byte, b Bindings) ([]byte, SourceError) {
	tpl, err := e.ParseTemplate(source)
	if err != nil {
		return nil, err
	}
	return tpl.Render(b)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that takes string input and returns a string.
func (e *Engine) ParseAndRenderString(source string, b Bindings) (string, SourceError) {
	bs, err := e.ParseAndRender([]byte(source), b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// Delims sets the action delimiters to the specified strings, to be used in subsequent calls to
// ParseTemplate, ParseTemplateLocation, ParseAndRender, or ParseAndRenderString. An empty delimiter
// stands for the corresponding default: objectLeft = {{, objectRight = }}, tagLeft = {% , tagRight = %}
func (e *Engine) Delims(objectLeft, objectRight, tagLeft, tagRight string) *Engine {
	e.cfg.Delims = []string{objectLeft, objectRight, tagLeft, tagRight}
	return e
}
