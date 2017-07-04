// Package filters defines the standard Liquid filters.
package filters

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jeffjen/datefmt"
	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/generics"
)

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(settings expressions.Config) { // nolint: gocyclo
	// values
	settings.AddFilter("default", func(value, defaultValue interface{}) interface{} {
		if value == nil || value == false || generics.IsEmpty(value) {
			value = defaultValue
		}
		return value
	})

	// dates
	settings.AddFilter("date", func(t time.Time, format interface{}) (string, error) {
		form, ok := format.(string)
		if !ok {
			form = "%a, %b %d, %y"
		}
		// TODO %\d*N -> truncated fractional seconds, default 9
		form = strings.Replace(form, "%N", "", -1)
		return datefmt.Strftime(form, t)
	})

	// arrays
	settings.AddFilter("compact", func(array []interface{}) interface{} {
		out := []interface{}{}
		for _, item := range array {
			if item != nil {
				out = append(out, item)
			}
		}
		return out
	})
	settings.AddFilter("join", joinFilter)
	settings.AddFilter("map", func(array []map[string]interface{}, key string) interface{} {
		out := []interface{}{}
		for _, obj := range array {
			out = append(out, obj[key])
		}
		return out
	})
	settings.AddFilter("reverse", reverseFilter)
	settings.AddFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	settings.AddFilter("first", func(array []interface{}) interface{} {
		if len(array) == 0 {
			return nil
		}
		return array[0]
	})
	settings.AddFilter("last", func(array []interface{}) interface{} {
		if len(array) == 0 {
			return nil
		}
		return array[len(array)-1]
	})

	// numbers
	settings.AddFilter("abs", math.Abs)
	settings.AddFilter("ceil", math.Ceil)
	settings.AddFilter("floor", math.Floor)
	settings.AddFilter("modulo", math.Mod)
	settings.AddFilter("minus", func(a, b float64) float64 {
		return a - b
	})
	settings.AddFilter("plus", func(a, b float64) float64 {
		return a + b
	})
	settings.AddFilter("times", func(a, b float64) float64 {
		return a * b
	})
	settings.AddFilter("divided_by", func(a float64, b interface{}) interface{} {
		switch bt := b.(type) {
		case int, int16, int32, int64:
			return int(a) / bt.(int)
		case float32, float64:
			return a / b.(float64)
		default:
			return nil
		}
	})
	settings.AddFilter("round", func(n float64, places interface{}) float64 {
		pl, ok := places.(int)
		if !ok {
			pl = 0
		}
		exp := math.Pow10(pl)
		return math.Floor(n*exp+0.5) / exp
	})

	// sequences
	settings.AddFilter("size", generics.Length)

	// strings
	settings.AddFilter("append", func(s, suffix string) string {
		return s + suffix
	})
	settings.AddFilter("capitalize", func(s, suffix string) string {
		if len(s) < 1 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	})
	settings.AddFilter("downcase", func(s, suffix string) string {
		return strings.ToLower(s)
	})
	settings.AddFilter("escape", html.EscapeString)
	settings.AddFilter("escape_once", func(s, suffix string) string {
		return html.EscapeString(html.UnescapeString(s))
	})
	// TODO test case for this
	settings.AddFilter("newline_to_br", func(s string) string {
		return strings.Replace(s, "\n", "<br />", -1)
	})
	settings.AddFilter("prepend", func(s, prefix string) string {
		return prefix + s
	})
	settings.AddFilter("remove", func(s, old string) string {
		return strings.Replace(s, old, "", -1)
	})
	settings.AddFilter("remove_first", func(s, old string) string {
		return strings.Replace(s, old, "", 1)
	})
	settings.AddFilter("replace", func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	})
	settings.AddFilter("replace_first", func(s, old, new string) string {
		return strings.Replace(s, old, new, 1)
	})
	settings.AddFilter("slice", func(s string, start int, length interface{}) string {
		// runes aren't bytes; don't use slice
		n, ok := length.(int)
		if !ok {
			n = 1
		}
		if start < 0 {
			start = utf8.RuneCountInString(s) + start
		}
		p := regexp.MustCompile(fmt.Sprintf(`^.{%d}(.{0,%d}).*$`, start, n))
		return p.ReplaceAllString(s, "$1")
	})
	settings.AddFilter("split", splitFilter)
	settings.AddFilter("strip_html", func(s string) string {
		// TODO this probably isn't sufficient
		return regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	})
	// TODO test case for this
	settings.AddFilter("strip_newlines", func(s string) string {
		return strings.Replace(s, "\n", "", -1)
	})
	settings.AddFilter("strip", strings.TrimSpace)
	settings.AddFilter("lstrip", func(s string) string {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	})
	settings.AddFilter("rstrip", func(s string) string {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	settings.AddFilter("truncate", func(s string, n int, ellipsis interface{}) string {
		// runes aren't bytes; don't use slice
		el, ok := ellipsis.(string)
		if !ok {
			el = "..."
		}
		p := regexp.MustCompile(fmt.Sprintf(`^(.{%d})..{%d,}`, n-len(el), len(el)))
		return p.ReplaceAllString(s, `$1`+el)
	})
	settings.AddFilter("upcase", func(s, suffix string) string {
		return strings.ToUpper(s)
	})

	// debugging extensions
	// inspect is from Jekyll
	settings.AddFilter("inspect", func(value interface{}) string {
		s, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%#v", value)
		}
		return string(s)
	})
	settings.AddFilter("type", func(value interface{}) string {
		return reflect.TypeOf(value).String()
	})
}

func joinFilter(array []interface{}, sep interface{}) interface{} {
	a := make([]string, len(array))
	s := ", "
	if sep != nil {
		s = fmt.Sprint(sep)
	}
	for i, x := range array {
		a[i] = fmt.Sprint(x)
	}
	return strings.Join(a, s)
}

func reverseFilter(array []interface{}) interface{} {
	out := make([]interface{}, len(array))
	for i, x := range array {
		out[len(out)-1-i] = x
	}
	return out
}

func sortFilter(array []interface{}, key interface{}) []interface{} {
	out := make([]interface{}, len(array))
	copy(out, array)
	if key == nil {
		generics.Sort(out)
	} else {
		generics.SortByProperty(out, key.(string), true)
	}
	return out
}

func splitFilter(s, sep string) interface{} {
	out := strings.Split(s, sep)
	// This matches Jekyll's observed behavior.
	// TODO test case
	if len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}
