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

	"github.com/leekchan/timeutil"
	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/generics"
)

// DefineStandardFilters defines the standard Liquid filters.
func DefineStandardFilters() {
	// values
	expressions.DefineFilter("default", func(in, defaultValue interface{}) interface{} {
		if in == nil || in == false || generics.IsEmpty(in) {
			in = defaultValue
		}
		return in
	})

	// dates
	expressions.DefineFilter("date", func(value time.Time, format interface{}) interface{} {
		form, ok := format.(string)
		if !ok {
			form = "%a, %b %d, %y"
		}
		return timeutil.Strftime(&value, form)
	})

	// lists
	expressions.DefineFilter("compact", func(values []interface{}) interface{} {
		out := []interface{}{}
		for _, value := range values {
			if value != nil {
				out = append(out, value)
			}
		}
		return out
	})
	expressions.DefineFilter("join", joinFilter)
	expressions.DefineFilter("map", func(values []map[string]interface{}, key string) interface{} {
		out := []interface{}{}
		for _, obj := range values {
			out = append(out, obj[key])
		}
		return out
	})
	expressions.DefineFilter("reverse", reverseFilter)
	expressions.DefineFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	expressions.DefineFilter("first", func(values []interface{}) interface{} {
		if len(values) == 0 {
			return nil
		}
		return values[0]
	})
	expressions.DefineFilter("last", func(values []interface{}) interface{} {
		if len(values) == 0 {
			return nil
		}
		return values[len(values)-1]
	})

	// numbers
	expressions.DefineFilter("abs", math.Abs)
	expressions.DefineFilter("ceil", math.Ceil)
	expressions.DefineFilter("floor", math.Floor)
	expressions.DefineFilter("modulo", math.Mod)
	expressions.DefineFilter("minus", func(a, b float64) float64 {
		return a - b
	})
	expressions.DefineFilter("plus", func(a, b float64) float64 {
		return a + b
	})
	expressions.DefineFilter("times", func(a, b float64) float64 {
		return a * b
	})
	expressions.DefineFilter("divided_by", func(a float64, b interface{}) interface{} {
		switch bt := b.(type) {
		case int, int16, int32, int64:
			return int(a) / bt.(int)
		case float32, float64:
			return a / float64(b.(float64))
		default:
			return nil
		}
	})
	expressions.DefineFilter("round", func(n float64, places interface{}) float64 {
		pl, ok := places.(int)
		if !ok {
			pl = 0
		}
		exp := math.Pow10(pl)
		return math.Floor(n*exp+0.5) / exp
	})

	// sequences
	expressions.DefineFilter("size", generics.Length)

	// strings
	expressions.DefineFilter("append", func(s, suffix string) string {
		return s + suffix
	})
	expressions.DefineFilter("capitalize", func(s, suffix string) string {
		if len(s) < 1 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	})
	expressions.DefineFilter("downcase", func(s, suffix string) string {
		return strings.ToLower(s)
	})
	expressions.DefineFilter("escape", html.EscapeString)
	expressions.DefineFilter("escape_once", func(s, suffix string) string {
		return html.EscapeString(html.UnescapeString(s))
	})
	// TODO test case for this
	expressions.DefineFilter("newline_to_br", func(s string) string {
		return strings.Replace(s, "\n", "<br />", -1)
	})
	expressions.DefineFilter("prepend", func(s, prefix string) string {
		return prefix + s
	})
	expressions.DefineFilter("remove", func(s, old string) string {
		return strings.Replace(s, old, "", -1)
	})
	expressions.DefineFilter("remove_first", func(s, old string) string {
		return strings.Replace(s, old, "", 1)
	})
	expressions.DefineFilter("replace", func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	})
	expressions.DefineFilter("replace_first", func(s, old, new string) string {
		return strings.Replace(s, old, new, 1)
	})
	expressions.DefineFilter("slice", func(s string, start int, length interface{}) string {
		n, ok := length.(int)
		if !ok {
			n = 1
		}
		if start < 0 {
			start = len(s) + start
		}
		if start >= len(s) {
			return ""
		}
		if start+n > len(s) {
			return s[start:]
		}
		return s[start : start+n]
	})
	expressions.DefineFilter("split", splitFilter)
	expressions.DefineFilter("strip_html", func(s string) string {
		// TODO this probably isn't sufficient
		return regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	})
	// TODO test case for this
	expressions.DefineFilter("strip_newlines", func(s string) string {
		return strings.Replace(s, "\n", "", -1)
	})
	expressions.DefineFilter("strip", strings.TrimSpace)
	expressions.DefineFilter("lstrip", func(s string) string {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	})
	expressions.DefineFilter("rstrip", func(s string) string {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	expressions.DefineFilter("truncate", func(s string, n int, ellipsis interface{}) string {
		el, ok := ellipsis.(string)
		if !ok {
			el = "..."
		}
		if len(s) > n {
			s = s[:n-len(el)] + el
		}
		return s
	})
	expressions.DefineFilter("upcase", func(s, suffix string) string {
		return strings.ToUpper(s)
	})

	// debugging extensions
	// inspect is from Jekyll
	expressions.DefineFilter("inspect", func(value interface{}) string {
		s, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%#v", value)
		}
		return string(s)
	})
	expressions.DefineFilter("type", func(value interface{}) string {
		return reflect.TypeOf(value).String()
	})
}

func joinFilter(in []interface{}, sep interface{}) interface{} {
	a := make([]string, len(in))
	s := ", "
	if sep != nil {
		s = fmt.Sprint(sep)
	}
	for i, x := range in {
		a[i] = fmt.Sprint(x)
	}
	return strings.Join(a, s)
}

func reverseFilter(in []interface{}) interface{} {
	out := make([]interface{}, len(in))
	for i, x := range in {
		out[len(out)-1-i] = x
	}
	return out
}

func sortFilter(in []interface{}, key interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	if key == nil {
		generics.Sort(out)
	} else {
		generics.SortByProperty(out, key.(string))
	}
	return out
}

func splitFilter(in, sep string) interface{} {
	return strings.Split(in, sep)
}
