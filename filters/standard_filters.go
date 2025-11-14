// Package filters is an internal package that defines the standard Liquid filters.
package filters

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/osteele/liquid/values"
	"github.com/osteele/tuesday"
)

var errDivisionByZero = errors.New("division by zero")

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, any)
}

// Helper functions for type-aware arithmetic operations

// isIntegerType checks if a value is an integer type
// Note: uint and uint64 are not included to avoid potential overflow issues
// when converting to int64. These types will fall back to float arithmetic.
func isIntegerType(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32:
		return true
	default:
		return false
	}
}

// toInt64 converts a value to int64
func toInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case int64:
		return val
	case uint8:
		return int64(val)
	case uint16:
		return int64(val)
	case uint32:
		return int64(val)
	default:
		return 0
	}
}

// toFloat64 converts a value to float64
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { //nolint: gocyclo
	// value filters
	fd.AddFilter("default", func(value, defaultValue any) any {
		if value == nil || value == false || values.IsEmpty(value) {
			value = defaultValue
		}

		return value
	})
	fd.AddFilter("json", func(a any) any {
		result, _ := json.Marshal(a)
		return result
	})

	// array filters
	fd.AddFilter("compact", func(a []any) (result []any) {
		for _, item := range a {
			if item != nil {
				result = append(result, item)
			}
		}

		return
	})
	fd.AddFilter("concat", func(a, b []any) (result []any) {
		result = make([]any, 0, len(a)+len(b))
		return append(append(result, a...), b...)
	})
	fd.AddFilter("join", joinFilter)
	fd.AddFilter("map", func(a []any, key string) (result []any) {
		keyValue := values.ValueOf(key)
		for _, obj := range a {
			value := values.ValueOf(obj)
			result = append(result, value.PropertyValue(keyValue).Interface())
		}

		return result
	})
	fd.AddFilter("reverse", reverseFilter)
	fd.AddFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	fd.AddFilter("first", func(a []any) any {
		if len(a) == 0 {
			return nil
		}

		return a[0]
	})
	fd.AddFilter("last", func(a []any) any {
		if len(a) == 0 {
			return nil
		}

		return a[len(a)-1]
	})
	fd.AddFilter("uniq", uniqFilter)

	// date filters
	fd.AddFilter("date", func(t time.Time, format func(string) string) (string, error) {
		f := format("%a, %b %d, %y")
		return tuesday.Strftime(f, t)
	})

	// number filters
	fd.AddFilter("abs", math.Abs)
	fd.AddFilter("ceil", func(a float64) int {
		return int(math.Ceil(a))
	})
	fd.AddFilter("floor", func(a float64) int {
		return int(math.Floor(a))
	})
	fd.AddFilter("modulo", math.Mod)
	fd.AddFilter("minus", func(a, b any) any {
		// If both operands are integers, perform integer arithmetic
		if isIntegerType(a) && isIntegerType(b) {
			return toInt64(a) - toInt64(b)
		}
		// Otherwise, perform float arithmetic
		return toFloat64(a) - toFloat64(b)
	})
	fd.AddFilter("plus", func(a, b any) any {
		// If both operands are integers, perform integer arithmetic
		if isIntegerType(a) && isIntegerType(b) {
			return toInt64(a) + toInt64(b)
		}
		// Otherwise, perform float arithmetic
		return toFloat64(a) + toFloat64(b)
	})
	fd.AddFilter("times", func(a, b any) any {
		// If both operands are integers, perform integer arithmetic
		if isIntegerType(a) && isIntegerType(b) {
			return toInt64(a) * toInt64(b)
		}
		// Otherwise, perform float arithmetic
		return toFloat64(a) * toFloat64(b)
	})
	fd.AddFilter("divided_by", func(a float64, b any) (any, error) {
		divInt := func(a, b int64) (int64, error) {
			if b == 0 {
				return 0, errDivisionByZero
			}

			return a / b, nil
		}

		divFloat := func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, errDivisionByZero
			}

			return a / b, nil
		}
		switch q := b.(type) {
		case int:
			return divInt(int64(a), int64(q))
		case int8:
			return divInt(int64(a), int64(q))
		case int16:
			return divInt(int64(a), int64(q))
		case int32:
			return divInt(int64(a), int64(q))
		case int64:
			return divInt(int64(a), q)
		case uint8:
			return divInt(int64(a), int64(q))
		case uint16:
			return divInt(int64(a), int64(q))
		case uint32:
			return divInt(int64(a), int64(q))
		case float32:
			return divFloat(a, float64(q))
		case float64:
			return divFloat(a, q)
		default:
			return nil, fmt.Errorf("invalid divisor: '%v'", b)
		}
	})
	fd.AddFilter("round", func(n float64, places func(int) int) float64 {
		pl := places(0)
		exp := math.Pow10(pl)

		return math.Floor(n*exp+0.5) / exp
	})

	// sequence filters
	fd.AddFilter("size", values.Length)

	// string filters
	fd.AddFilter("append", func(s, suffix string) string {
		return s + suffix
	})
	fd.AddFilter("capitalize", func(s, suffix string) string {
		if len(s) == 0 {
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
		return strings.ReplaceAll(s, "\n", "<br />")
	})
	fd.AddFilter("prepend", func(s, prefix string) string {
		return prefix + s
	})
	fd.AddFilter("remove", func(s, old string) string {
		return strings.ReplaceAll(s, old, "")
	})
	fd.AddFilter("remove_first", func(s, old string) string {
		return strings.Replace(s, old, "", 1)
	})
	fd.AddFilter("replace", strings.ReplaceAll)
	fd.AddFilter("replace_first", func(s, old, n string) string {
		return strings.Replace(s, old, n, 1)
	})
	fd.AddFilter("sort_natural", sortNaturalFilter)
	fd.AddFilter("slice", func(s string, start int, length func(int) int) string {
		if len(s) == 0 {
			return ""
		}

		ss := []rune(s)
		n := length(1)

		if start < 0 {
			start = len(ss) + start
		}

		if start < 0 {
			return ""
		}

		end := start + n
		if end > len(ss) {
			end = len(ss)
		}

		return string(ss[start:end])
	})
	fd.AddFilter("split", splitFilter)
	fd.AddFilter("strip_html", func(s string) string {
		// TODO this probably isn't sufficient
		return regexp.MustCompile(`<.*?>`).ReplaceAllString(s, "")
	})
	fd.AddFilter("strip_newlines", func(s string) string {
		return strings.ReplaceAll(s, "\n", "")
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

	// debugging filters
	// inspect is from Jekyll
	fd.AddFilter("inspect", func(value any) string {
		s, err := json.Marshal(value)
		if err != nil {
			return fmt.Sprintf("%#v", value)
		}

		return string(s)
	})
	fd.AddFilter("type", func(value any) string {
		return fmt.Sprintf("%T", value)
	})
}

func joinFilter(a []any, sep func(string) string) any {
	ss := make([]string, 0, len(a))
	s := sep(" ")

	for _, v := range a {
		if v != nil {
			ss = append(ss, fmt.Sprint(v))
		}
	}

	return strings.Join(ss, s)
}

func reverseFilter(a []any) any {
	result := make([]any, len(a))
	for i, x := range a {
		result[len(result)-1-i] = x
	}

	return result
}

var wsre = regexp.MustCompile(`[[:space:]]+`)

func splitFilter(s, sep string) any {
	result := strings.Split(s, sep)
	if sep == " " {
		// Special case for Ruby, therefore Liquid
		result = wsre.Split(s, -1)
	}
	// This matches Ruby / Liquid / Jekyll's observed behavior.
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}

	return result
}

func uniqFilter(a []any) (result []any) {
	seenMap := map[any]bool{}

	seen := func(item any) bool {
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
	for _, item := range a {
		if !seen(item) {
			result = append(result, item)
		}
	}

	return
}

func eqItems(a, b any) bool {
	if reflect.TypeOf(a).Comparable() && reflect.TypeOf(b).Comparable() {
		return a == b
	}

	return reflect.DeepEqual(a, b)
}
