// Package filters defines the standard Liquid filters.
package filters

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/osteele/liquid/evaluator"
	"github.com/osteele/liquid/strftime"
)

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, interface{})
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { // nolint: gocyclo
	// values
	fd.AddFilter("default", func(value, defaultValue interface{}) interface{} {
		if value == nil || value == false || evaluator.IsEmpty(value) {
			value = defaultValue
		}
		return value
	})

	// arrays
	fd.AddFilter("compact", func(array []interface{}) interface{} {
		out := []interface{}{}
		for _, item := range array {
			if item != nil {
				out = append(out, item)
			}
		}
		return out
	})
	fd.AddFilter("join", joinFilter)
	fd.AddFilter("map", func(array []map[string]interface{}, key string) interface{} {
		out := []interface{}{}
		for _, obj := range array {
			out = append(out, obj[key])
		}
		return out
	})
	fd.AddFilter("reverse", reverseFilter)
	fd.AddFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	fd.AddFilter("first", func(array []interface{}) interface{} {
		if len(array) == 0 {
			return nil
		}
		return array[0]
	})
	fd.AddFilter("last", func(array []interface{}) interface{} {
		if len(array) == 0 {
			return nil
		}
		return array[len(array)-1]
	})
	fd.AddFilter("uniq", uniqFilter)

	// dates
	fd.AddFilter("date", func(t time.Time, format func(string) string) (string, error) {
		f := format("%a, %b %d, %y")
		// TODO %\d*N -> truncated fractional seconds, default 9
		f = strings.Replace(f, "%N", "", -1)
		return strftime.Strftime(f, t)
	})

	// numbers
	fd.AddFilter("abs", math.Abs)
	fd.AddFilter("ceil", math.Ceil)
	fd.AddFilter("floor", math.Floor)
	fd.AddFilter("modulo", math.Mod)
	fd.AddFilter("minus", func(a, b float64) float64 {
		return a - b
	})
	fd.AddFilter("plus", func(a, b float64) float64 {
		return a + b
	})
	fd.AddFilter("times", func(a, b float64) float64 {
		return a * b
	})
	fd.AddFilter("divided_by", func(a float64, b interface{}) interface{} {
		switch q := b.(type) {
		case int, int16, int32, int64:
			return int(a) / q.(int)
		case float32, float64:
			return a / b.(float64)
		default:
			return nil
		}
	})
	fd.AddFilter("round", func(n float64, places func(int) int) float64 {
		pl := places(0)
		exp := math.Pow10(pl)
		return math.Floor(n*exp+0.5) / exp
	})

	// sequences
	fd.AddFilter("size", evaluator.Length)

	// strings
	fd.AddFilter("append", func(s, suffix string) string {
		return s + suffix
	})
	fd.AddFilter("capitalize", func(s, suffix string) string {
		if len(s) < 1 {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	})
	fd.AddFilter("downcase", func(s, suffix string) string {
		return strings.ToLower(s)
	})
	fd.AddFilter("escape", html.EscapeString)
	fd.AddFilter("escape_once", func(s, suffix string) string {
		return html.EscapeString(html.UnescapeString(s))
	})
	fd.AddFilter("newline_to_br", func(s string) string {
		return strings.Replace(s, "\n", "<br />", -1)
	})
	fd.AddFilter("prepend", func(s, prefix string) string {
		return prefix + s
	})
	fd.AddFilter("remove", func(s, old string) string {
		return strings.Replace(s, old, "", -1)
	})
	fd.AddFilter("remove_first", func(s, old string) string {
		return strings.Replace(s, old, "", 1)
	})
	fd.AddFilter("replace", func(s, old, new string) string {
		return strings.Replace(s, old, new, -1)
	})
	fd.AddFilter("replace_first", func(s, old, new string) string {
		return strings.Replace(s, old, new, 1)
	})
	fd.AddFilter("sort_natural", sortNaturalFilter)
	fd.AddFilter("slice", func(s string, start int, length func(int) int) string {
		// runes aren't bytes; don't use slice
		n := length(1)
		if start < 0 {
			start = utf8.RuneCountInString(s) + start
		}
		p := regexp.MustCompile(fmt.Sprintf(`^.{%d}(.{0,%d}).*$`, start, n))
		return p.ReplaceAllString(s, "$1")
	})
	fd.AddFilter("split", splitFilter)
	fd.AddFilter("strip_html", func(s string) string {
		// TODO this probably isn't sufficient
		return regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	})
	// TODO test case for this
	fd.AddFilter("strip_newlines", func(s string) string {
		return strings.Replace(s, "\n", "", -1)
	})
	fd.AddFilter("strip", strings.TrimSpace)
	fd.AddFilter("lstrip", func(s string) string {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("rstrip", func(s string) string {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("truncate", func(s string, length func(int) int, ellipsis func(string) string) string {
		n := length(50)
		el := ellipsis("...")
		// runes aren't bytes; don't use slice
		re := regexp.MustCompile(fmt.Sprintf(`^(.{%d})..{%d,}`, n-len(el), len(el)))
		return re.ReplaceAllString(s, `$1`+el)
	})
	fd.AddFilter("truncatewords", func(s string, length func(int) int, ellipsis func(string) string) string {
		el := ellipsis("...")
		n := length(15)
		// re := regexp.MustCompile(fmt.Sprintf(`^\s*(?:\S+\s+){%d}`, n))
		re := regexp.MustCompile(fmt.Sprintf(`^(?:\s*\S+){%d}`, n))
		m := re.FindString(s)
		if m == "" {
			return s
		}
		return m + el
	})
	fd.AddFilter("upcase", func(s, suffix string) string {
		return strings.ToUpper(s)
	})
	fd.AddFilter("url_encode", url.QueryEscape)
	fd.AddFilter("url_decode", url.QueryUnescape)

	// debugging extensions
	// inspect is from Jekyll
	fd.AddFilter("inspect", func(value interface{}) string {
		s, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%#v", value)
		}
		return string(s)
	})
	fd.AddFilter("type", func(value interface{}) string {
		return fmt.Sprintf("%T", value)
	})
}

func joinFilter(array []interface{}, sep func(string) string) interface{} {
	a := make([]string, len(array))
	s := sep(", ")
	for i, x := range array {
		a[i] = fmt.Sprint(x)
	}
	return strings.Join(a, s)
}

func reverseFilter(array []interface{}) interface{} {
	result := make([]interface{}, len(array))
	for i, x := range array {
		result[len(result)-1-i] = x
	}
	return result
}

func splitFilter(s, sep string) interface{} {
	result := strings.Split(s, sep)
	// This matches Jekyll's observed behavior.
	// TODO test case
	if len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return result
}

func uniqFilter(array []interface{}) (result []interface{}) {
	seenMap := map[interface{}]bool{}
	seen := func(item interface{}) bool {
		if k := reflect.TypeOf(item).Kind(); k < reflect.Array || k == reflect.Ptr || k == reflect.UnsafePointer {
			if seenMap[item] {
				return true
			}
			seenMap[item] = true
			return false
		}
		// the O(n^2) case:
		for _, other := range result {
			if eqItems(item, other) {
				return true
			}
		}
		return false
	}
	for _, item := range array {
		if !seen(item) {
			result = append(result, item)
		}
	}
	return
}

func eqItems(a, b interface{}) bool {
	if reflect.TypeOf(a).Comparable() && reflect.TypeOf(b).Comparable() {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}
