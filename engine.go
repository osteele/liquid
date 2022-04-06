package liquid

import (
	"context"
	"fmt"
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

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// An Engine parses template source into renderable text.
//
// An engine can be configured with additional filters and tags.
type Engine struct{ cfg render.Config }

// NewEngine returns a new Engine.
func NewEngine() *Engine {
	engine := &Engine{render.NewConfig()}
	filters.AddStandardFilters(&engine.cfg)
	tags.AddStandardTags(engine.cfg)
	engine.RegisterFilter("hideCountryCodeAndDefault", func(s phone.International, hide bool, defaultValue string) string {
		if s.Number.IsZero() && s.CountryCode.IsZero() {
			return defaultValue
		}
		if hide {
			return s.Number.String()
		}
		return s.String()
	})
	engine.RegisterFilter("rawPhone", func(s phone.International) string {
		return s.CountryCode.String() + s.Number.String()
	})

	engine.RegisterFilter("dateTimeFormatOrDefault", func(s time.Time, format string, defaultValue string) string {
		if s.IsZero() {
			return defaultValue
		}

		switch format {
		case "mdy12":
			return s.Format("Jan 02 2006 3:04 PM")
		case "mdy24":
			return s.Format("Jan 02 2006 15:04")
		case "dmy12":
			return s.Format("02 Jan 2006 3:04 PM")
		case "dmy24":
			return s.Format("02 Jan 2006 15:04")
		default:
			return s.String()
		}
	})

	engine.RegisterFilter("dateFormatOrDefault", func(s date.Date, format string, defaultValue string) string {
		if s == 0 {
			return defaultValue
		}

		switch format {
		case "mdy":
			return fmt.Sprintf("%02d/%02d/%d", s.Month(), s.Day(), s.Year())
		case "dmy":
			return fmt.Sprintf("%02d/%02d/%d", s.Day(), s.Month(), s.Year())
		default:
			return s.String()
		}
	})

	engine.RegisterFilter("decimal", func(s string, format string, currency string) string {
		if s == "" {
			return s
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger.Warnw(context.Background(), fmt.Sprintf("failed to parse field value %s to decimal: %s", s, err.Error()), "lqiuid", "filter")
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

	engine.RegisterFilter("numberWithDelimiter", func(s string, loc string, format string) string {
		if s == "" {
			return s
		}
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			logger.Warnw(context.Background(), fmt.Sprintf("failed to parse field value %s to decimal: %s", s, err.Error()), "lqiuid", "filter")
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
			slice := s.([]interface{})
			for _, match := range matches {
				for _, s := range slice {
					if reflect.DeepEqual(s, match) {
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
			slice := s.([]interface{})
			for _, match := range matches {
				containMatch := false
				for _, s := range slice {
					if reflect.DeepEqual(s, match) {
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
//
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
