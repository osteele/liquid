// Package filters is an internal package that defines the standard Liquid filters.
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

	"github.com/osteele/liquid/values"
	"github.com/osteele/tuesday"
)

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, any)
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { // nolint: gocyclo
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
	fd.AddFilter("abs", func(a any) any {
		if ia, ok := values.ToInt64(a); ok {
			if ia < 0 {
				return -ia
			} else {
				return ia
			}
		}
		if fa, ok := values.ToFloat64(a); ok {
			return math.Abs(fa)
		}
		return math.NaN()
	})
	fd.AddFilter("ceil", func(a any) any {
		if ia, ok := values.ToInt64(a); ok {
			return ia
		}
		if fa, ok := values.ToFloat64(a); ok {
			return int64(math.Ceil(fa))
		}
		return math.NaN()
	})
	fd.AddFilter("floor", func(a any) any {
		if ia, ok := values.ToInt64(a); ok {
			return ia
		}
		if fa, ok := values.ToFloat64(a); ok {
			return int64(math.Floor(fa))
		}
		return math.NaN()
	})
	fd.AddFilter("modulo", func(a, b any) any {
		if fa, ok := values.ToFloat64(a); ok {
			if fb, ok := values.ToFloat64(b); ok {
				return math.Mod(fa, fb)
			}
		}
		return math.NaN()
	})
	fd.AddFilter("minus", func(a, b any) any {
		if ia, ok := values.ToInt64(a); ok {
			if ib, ok := values.ToInt64(b); ok {
				return ia - ib
			}
		}
		if fa, ok := values.ToFloat64(a); ok {
			if fb, ok := values.ToFloat64(b); ok {
				return fa - fb
			}
		}
		return math.NaN()
	})
	fd.AddFilter("plus", func(a, b any) any {
		if ia, ok := values.ToInt64(a); ok {
			if ib, ok := values.ToInt64(b); ok {
				return ia + ib
			}
		}
		if fa, ok := values.ToFloat64(a); ok {
			if fb, ok := values.ToFloat64(b); ok {
				return fa + fb
			}
		}
		return math.NaN()
	})
	fd.AddFilter("times", func(a, b any) any {
		if ia, ok := values.ToInt64(a); ok {
			if ib, ok := values.ToInt64(b); ok {
				return ia * ib
			}
		}
		if fa, ok := values.ToFloat64(a); ok {
			if fb, ok := values.ToFloat64(b); ok {
				return fa * fb
			}
		}
		return math.NaN()
	})
	fd.AddFilter("divided_by", func(a any, b any) any {
		if ia, ok := values.ToInt64(a); ok {
			if ib, ok := values.ToInt64(b); ok {
				if ib == 0 {
					if ia == 0 {
						return math.NaN()
					}
					return math.Inf(int(ia))
				}
				return ia / ib
			}
		}
		if fa, ok := values.ToFloat64(a); ok {
			if fb, ok := values.ToFloat64(b); ok {
				if fb == 0 {
					if fa == 0 {
						return math.NaN()
					}
					return math.Inf(sign(fa))
				}
				return fa / fb
			}
		}
		return math.NaN()
	})
	//fd.AddFilter("round", func(a any, places func(int) int) float64 {
	//	if ia, ok := values.ToInt64(a); ok {
	//		return float64(ia)
	//	}
	//	if fa, ok := values.ToFloat64(a); ok {
	//		pl := places(0)
	//		exp := math.Pow10(pl)
	//		return math.Floor(fa*exp+0.5) / exp
	//	}
	//	return math.NaN()
	//})
	fd.AddFilter("round", func(a any, places any) float64 {
		pl, ok := values.ToInt64(places)
		if !ok {
			return math.NaN()
		}
		if ia, ok := values.ToInt64(a); ok {
			return float64(ia)
		}
		if fa, ok := values.ToFloat64(a); ok {
			exp := math.Pow10(int(pl))
			return math.Floor(fa*exp+0.5) / exp
		}
		return math.NaN()
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
		ss := []rune(s)
		n := length(1)
		if start < 0 {
			start = len(ss) + start
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

func sign(a float64) int {
	if a > 0 {
		return 1
	} else if a < 0 {
		return -1
	} else {
		return 0
	}
}
