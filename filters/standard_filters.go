// Package filters is an internal package that defines the standard Liquid filters.
package filters

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/osteele/tuesday"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/values"
)

// ZeroDivisionError is returned by the divided_by and modulo filters when
// the divisor is zero. Use errors.As to detect this specific condition.
type ZeroDivisionError struct{}

func (e *ZeroDivisionError) Error() string { return "divided by 0" }

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, any)
	AddContextFilter(string, expressions.ContextFilterFn)
}

// Helper functions for type-aware arithmetic operations

// isIntegerType checks if a value is an integer type that can be safely
// represented as int64 without overflow
func isIntegerType(v any) bool {
	switch val := v.(type) {
	case int, int8, int16, int32, int64, uint8, uint16, uint32:
		return true
	case uint:
		// Check if uint value fits in int64 range
		return val <= math.MaxInt64
	case uint64:
		// Check if uint64 value fits in int64 range
		return val <= math.MaxInt64
	case uintptr:
		// Check if uintptr value fits in int64 range
		return val <= math.MaxInt64
	default:
		return false
	}
}

// toInt64 converts a value to int64
// Caller must ensure value fits in int64 range by calling isIntegerType first
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
	case uint:
		return int64(val) //nolint:gosec // G115: Safe - isIntegerType verifies val <= math.MaxInt64
	case uint64:
		return int64(val) //nolint:gosec // G115: Safe - isIntegerType verifies val <= math.MaxInt64
	case uintptr:
		return int64(val) //nolint:gosec // G115: Safe - isIntegerType verifies val <= math.MaxInt64
	default:
		return 0
	}
}

// toFloat64 converts a value to float64.
// Strings are parsed as floats, matching Ruby Liquid's String#to_f behavior.
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
	case uintptr:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { //nolint: gocyclo
	// value filters
	fd.AddFilter("default", func(value, defaultValue any, kwargs ...any) any {
		allowFalse := false
		for _, kw := range kwargs {
			if na, ok := kw.(expressions.NamedArg); ok && na.Name == "allow_false" {
				if b, ok := na.Value.(bool); ok {
					allowFalse = b
				}
			}
		}
		isFalsy := value == nil || values.IsEmpty(value)
		if !allowFalse {
			isFalsy = isFalsy || value == false
		}
		if isFalsy {
			value = defaultValue
		}

		return value
	})
	fd.AddFilter("json", func(a any) any {
		result, _ := json.Marshal(a)
		return result
	})
	fd.AddFilter("jsonify", func(a any) any {
		result, _ := json.Marshal(a)
		return result
	})

	// array filters
	fd.AddFilter("compact", func(a []any, property func(string) string) (result []any) {
		prop := property("")
		for _, item := range a {
			if prop == "" {
				if item != nil {
					result = append(result, item)
				}
			} else {
				if getPropertyValue(item, prop) != nil {
					result = append(result, item)
				}
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
	// but https://help.shopify.com/themes/liquid/filters/array-filters does.
	// Ruby and JS also support strings: first/last returns the first/last Unicode character.
	fd.AddFilter("first", func(v any) any {
		if s, ok := v.(string); ok {
			if s == "" {
				return ""
			}
			r, _ := utf8.DecodeRuneInString(s)
			return string(r)
		}
		var slice []any
		a, err := values.Convert(v, reflect.TypeOf(slice))
		if err != nil {
			return nil
		}
		arr, _ := a.([]any)
		if len(arr) == 0 {
			return nil
		}
		return arr[0]
	})
	fd.AddFilter("last", func(v any) any {
		if s, ok := v.(string); ok {
			if s == "" {
				return ""
			}
			runes := []rune(s)
			return string(runes[len(runes)-1])
		}
		var slice []any
		a, err := values.Convert(v, reflect.TypeOf(slice))
		if err != nil {
			return nil
		}
		arr, _ := a.([]any)
		if len(arr) == 0 {
			return nil
		}
		return arr[len(arr)-1]
	})
	fd.AddFilter("uniq", func(a []any, property func(string) string) []any {
		prop := property("")
		if prop == "" {
			return uniqFilter(a)
		}
		// property-based uniq: deduplicate by property value
		seen := map[string]bool{}
		var result []any
		for _, item := range a {
			pv := getPropertyValue(item, prop)
			pvKey := fmt.Sprintf("%T|%v", pv, pv)
			if !seen[pvKey] {
				seen[pvKey] = true
				result = append(result, item)
			}
		}
		return result
	})
	fd.AddFilter("where", whereFilter)
	fd.AddFilter("reject", rejectFilter)
	fd.AddFilter("group_by", groupByFilter)
	fd.AddFilter("find", findFilter)
	fd.AddFilter("find_index", findIndexFilter)
	fd.AddFilter("has", hasFilter)
	fd.AddFilter("sum", sumFilter)
	fd.AddFilter("push", pushFilter)
	fd.AddFilter("unshift", unshiftFilter)
	fd.AddFilter("pop", popFilter)
	fd.AddFilter("shift", shiftFilter)
	fd.AddFilter("sample", sampleFilter)

	fd.AddContextFilter("where_exp", whereExpFilter)
	fd.AddContextFilter("reject_exp", rejectExpFilter)
	fd.AddContextFilter("group_by_exp", groupByExpFilter)
	fd.AddContextFilter("find_exp", findExpFilter)
	fd.AddContextFilter("find_index_exp", findIndexExpFilter)
	fd.AddContextFilter("has_exp", hasExpFilter)

	// date filters
	fd.AddFilter("date", func(v any, format func(string) string) (any, error) {
		if v == nil {
			return nil, nil
		}
		t, ok := parseToTime(v)
		if !ok {
			return "", nil
		}
		f := format("%a, %b %d, %y")
		return tuesday.Strftime(f, t)
	})
	fd.AddFilter("date_to_xmlschema", func(v any) string {
		t, ok := parseToTime(v)
		if !ok {
			return fmt.Sprint(v)
		}
		result, _ := tuesday.Strftime("%Y-%m-%dT%H:%M:%S%:z", t)
		return result
	})
	fd.AddFilter("date_to_rfc822", func(v any) string {
		t, ok := parseToTime(v)
		if !ok {
			return fmt.Sprint(v)
		}
		result, _ := tuesday.Strftime("%a, %d %b %Y %H:%M:%S %z", t)
		return result
	})
	fd.AddFilter("date_to_string", func(v any, typ func(string) string, style func(string) string) string {
		return formatJekyllDate(v, "%b", typ(""), style(""))
	})
	fd.AddFilter("date_to_long_string", func(v any, typ func(string) string, style func(string) string) string {
		return formatJekyllDate(v, "%B", typ(""), style(""))
	})

	// number filters
	fd.AddFilter("abs", math.Abs)
	fd.AddFilter("ceil", func(a float64) int {
		return int(math.Ceil(a))
	})
	fd.AddFilter("floor", func(a float64) int {
		return int(math.Floor(a))
	})
	fd.AddFilter("modulo", func(rawA, b any) (any, error) {
		// modulo semantics (Ruby/Shopify Liquid compatible):
		//   • Both operands are integer types → floored integer modulo.
		//   • Either operand is a float or string → floored float modulo.
		// Ruby's % operator uses floor modulo (result has same sign as divisor).
		// Go's % operator and math.Mod use truncated modulo (same sign as dividend),
		// so we adjust the result when the signs differ.
		modInt := func(a, b int64) (int64, error) {
			if b == 0 {
				return 0, &ZeroDivisionError{}
			}
			result := a % b
			// floor modulo: adjust sign to match divisor
			if result != 0 && (result > 0) != (b > 0) {
				result += b
			}
			return result, nil
		}
		modFloat := func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, &ZeroDivisionError{}
			}
			result := math.Mod(a, b)
			// floor modulo: adjust sign to match divisor
			if result != 0 && math.Signbit(result) != math.Signbit(b) {
				result += b
			}
			return result, nil
		}

		aIsInt := isIntegerType(rawA)

		switch q := b.(type) {
		case int:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case int8:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case int16:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case int32:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case int64:
			if aIsInt {
				return modInt(toInt64(rawA), q)
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uint8:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uint16:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uint32:
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uint: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uint64: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case uintptr: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return modInt(toInt64(rawA), int64(q))
			}
			return modFloat(toFloat64(rawA), float64(q))
		case float32:
			return modFloat(toFloat64(rawA), float64(q))
		case float64:
			return modFloat(toFloat64(rawA), q)
		case string:
			return modFloat(toFloat64(rawA), toFloat64(q))
		default:
			return nil, fmt.Errorf("invalid modulus: '%v'", b)
		}
	})
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
	fd.AddFilter("divided_by", func(rawA, b any) (any, error) {
		// divided_by semantics (Ruby/Shopify Liquid compatible):
		//   • Both operands are integer types → floor (integer) division.
		//   • Either operand is a float → float division.
		// The filter parameter rawA is declared as `any` so that we can
		// distinguish between integer literals (e.g. `2`) and float literals
		// (e.g. `2.0`).  When registered as `func(float64, any)` the original
		// int vs. float distinction was lost because the conversion always
		// produced float64.
		divInt := func(a, b int64) (int64, error) {
			if b == 0 {
				return 0, &ZeroDivisionError{}
			}

			return a / b, nil
		}

		divFloat := func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, &ZeroDivisionError{}
			}

			return a / b, nil
		}

		aIsInt := isIntegerType(rawA)

		switch q := b.(type) {
		case int:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case int8:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case int16:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case int32:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case int64:
			if aIsInt {
				return divInt(toInt64(rawA), q)
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uint8:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uint16:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uint32:
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uint: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uint64: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case uintptr: //nolint:gosec // G115: safe for values <= math.MaxInt64
			if aIsInt {
				return divInt(toInt64(rawA), int64(q))
			}
			return divFloat(toFloat64(rawA), float64(q))
		case float32:
			return divFloat(toFloat64(rawA), float64(q))
		case float64:
			return divFloat(toFloat64(rawA), q)
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
	fd.AddFilter("capitalize", func(s string) string {
		if s == "" {
			return s
		}
		r, size := utf8.DecodeRuneInString(s)
		return string(unicode.ToUpper(r)) + strings.ToLower(s[size:])
	})
	fd.AddFilter("downcase", func(s, suffix string) string {
		return strings.ToLower(s)
	})
	fd.AddFilter("escape", html.EscapeString)
	fd.AddFilter("h", html.EscapeString)
	fd.AddFilter("escape_once", func(s, suffix string) string {
		return html.EscapeString(html.UnescapeString(s))
	})
	fd.AddFilter("newline_to_br", func(s string) string {
		// Normalize Windows line endings (\r\n) to Unix (\n) first,
		// then convert all \n to <br />\n — matching Ruby/JS behaviour.
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		return strings.ReplaceAll(s, "\n", "<br />\n")
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
	fd.AddFilter("slice", func(v interface{}, start int, length func(int) int) interface{} {
		// Are we in the []byte case? Transform []byte to string
		if b, ok := v.([]byte); ok {
			v = string(b)
		}
		// Are we in the string case?
		if s, ok := v.(string); ok {
			// Work on runes, not chars
			runes := []rune(s)
			n := length(1)
			if n < 0 {
				n = 0
			}
			if start < 0 {
				start = len(runes) + start
				if start < 0 {
					start = 0
				}
			}
			if start > len(runes) {
				start = len(runes)
			}
			end := start + n
			if end > len(runes) {
				end = len(runes)
			}
			return string(runes[start:end])
		}
		// Are we in the slice case?
		// A type test cannot suffice because []T and []U are different types, so we must use conversion.
		var slice []interface{}
		if sliceIface, err := values.Convert(v, reflect.TypeOf(slice)); err == nil {
			var ok bool
			if slice, ok = sliceIface.([]interface{}); ok {
				n := length(1)
				if start < 0 {
					start = len(slice) + start
					if start < 0 {
						start = 0
					}
				}
				if start > len(slice) {
					start = len(slice)
				}
				end := start + n
				if end > len(slice) {
					end = len(slice)
				}
				return slice[start:end]
			}
		}
		return nil
	})
	fd.AddFilter("split", splitFilter)
	fd.AddFilter("strip_html", stripHTMLFilter)
	fd.AddFilter("strip_newlines", func(s string) string {
		// Remove \r\n (Windows), \r (old Mac), and \n (Unix) — matching Ruby/JS.
		s = strings.ReplaceAll(s, "\r\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		return strings.ReplaceAll(s, "\n", "")
	})
	fd.AddFilter("strip", func(s string, chars func(string) string) string {
		if c := chars(""); c != "" {
			return strings.Trim(s, c)
		}
		return strings.TrimSpace(s)
	})
	fd.AddFilter("lstrip", func(s string, chars func(string) string) string {
		if c := chars(""); c != "" {
			return strings.TrimLeft(s, c)
		}
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("rstrip", func(s string, chars func(string) string) string {
		if c := chars(""); c != "" {
			return strings.TrimRight(s, c)
		}
		return strings.TrimRightFunc(s, unicode.IsSpace)
	})
	fd.AddFilter("squish", func(s string) string {
		return strings.TrimSpace(wsre.ReplaceAllString(s, " "))
	})
	fd.AddFilter("truncate", func(s string, length func(int) int, ellipsis func(string) string) string {
		n := length(50)
		el := ellipsis("...")
		// Ruby/JS: if n <= len(el), return the full ellipsis (e.g. truncate: 0 => "...").
		erunes := []rune(el)
		if n <= len(erunes) {
			return el
		}
		// If the string already fits within the limit, return it unchanged.
		srunes := []rune(s)
		if len(srunes) <= n {
			return s
		}
		// Take first (n - len(el)) runes, then append ellipsis.
		return string(srunes[:n-len(erunes)]) + el
	})
	fd.AddFilter("truncatewords", func(s string, length func(int) int, ellipsis func(string) string) string {
		el := ellipsis("...")
		n := length(15)
		// n < 1 behaves like n = 1 (Ruby/JS: truncate to 1 word)
		if n < 1 {
			n = 1
		}
		// Count words first: if the string has <= n words, return it unchanged.
		// We cannot rely solely on the regex because Go's RE2 allows backtracking
		// across word boundaries (e.g. {4} on "one two three" still matches via
		// splitting the last word), giving false positives.
		words := strings.Fields(s)
		if len(words) <= n {
			return s
		}
		// There are more than n words: join the first n words with single spaces
		// (matches Ruby behaviour which normalises internal whitespace).
		return strings.Join(words[:n], " ") + el
	})
	fd.AddFilter("upcase", func(s, suffix string) string {
		return strings.ToUpper(s)
	})
	fd.AddFilter("url_encode", url.QueryEscape)
	fd.AddFilter("url_decode", url.QueryUnescape)

	// string filters
	fd.AddFilter("remove_last", func(s, sub string) string {
		idx := strings.LastIndex(s, sub)
		if idx < 0 {
			return s
		}

		return s[:idx] + s[idx+len(sub):]
	})
	fd.AddFilter("replace_last", func(s, old, new string) string {
		idx := strings.LastIndex(s, old)
		if idx < 0 {
			return s
		}

		return s[:idx] + new + s[idx+len(old):]
	})
	fd.AddFilter("normalize_whitespace", func(s string) string {
		return wsre.ReplaceAllString(s, " ")
	})
	fd.AddFilter("number_of_words", func(s string, mode func(string) string) int {
		m := mode("default")
		switch m {
		case "cjk":
			return countWordsWithCJK(s)
		case "auto":
			for _, r := range s {
				if isCJKRune(r) {
					return countWordsWithCJK(s)
				}
			}

			return len(strings.Fields(s))
		default:
			return len(strings.Fields(s))
		}
	})
	fd.AddFilter("array_to_sentence_string", func(a []any, connector func(string) string) string {
		con := connector("and")
		strs := make([]string, len(a))
		for i, v := range a {
			strs[i] = fmt.Sprint(v)
		}

		switch len(strs) {
		case 0:
			return ""
		case 1:
			return strs[0]
		case 2:
			return strs[0] + " " + con + " " + strs[1]
		default:
			return strings.Join(strs[:len(strs)-1], ", ") + ", " + con + " " + strs[len(strs)-1]
		}
	})

	// math filters
	fd.AddFilter("at_least", func(a, b float64) float64 {
		return math.Max(a, b)
	})
	fd.AddFilter("at_most", func(a, b float64) float64 {
		return math.Min(a, b)
	})

	// html/url filters
	// raw marks a value as safe, bypassing autoescape. Mirrors LiquidJS's | raw filter.
	// When autoescape is disabled (the default), raw wraps in SafeValue, which
	// is immediately transparent at render time — effectively a no-op.
	fd.AddFilter("raw", func(v any) values.SafeValue {
		return values.SafeValue{Value: v}
	})
	fd.AddFilter("xml_escape", func(s string) string {
		var buf strings.Builder
		for _, r := range s {
			switch r {
			case '&':
				buf.WriteString("&amp;")
			case '<':
				buf.WriteString("&lt;")
			case '>':
				buf.WriteString("&gt;")
			case '"':
				buf.WriteString("&#34;")
			case '\'':
				buf.WriteString("&#39;")
			default:
				buf.WriteRune(r)
			}
		}

		return buf.String()
	})
	fd.AddFilter("cgi_escape", url.QueryEscape)
	fd.AddFilter("uri_escape", func(s string) string {
		var buf strings.Builder
		for i := 0; i < len(s); {
			r, size := utf8.DecodeRuneInString(s[i:])
			if isURISafe(r) {
				buf.WriteRune(r)
			} else {
				for _, b := range []byte(s[i : i+size]) {
					fmt.Fprintf(&buf, "%%%02X", b)
				}
			}
			i += size
		}

		return buf.String()
	})
	fd.AddFilter("slugify", func(s string, mode func(string) string) string {
		return slugifyString(s, mode("default"))
	})

	// base64 filters
	fd.AddFilter("base64_encode", func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	})
	fd.AddFilter("base64_decode", func(s string) (string, error) {
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}

		return string(b), nil
	})
	fd.AddFilter("base64_url_safe_encode", func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	})
	fd.AddFilter("base64_url_safe_decode", func(s string) (string, error) {
		b, err := base64.URLEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}

		return string(b), nil
	})

	// type conversion filters
	fd.AddFilter("to_integer", func(v any) int {
		switch val := v.(type) {
		case int:
			return val
		case int8:
			return int(val)
		case int16:
			return int(val)
		case int32:
			return int(val)
		case int64:
			return int(val)
		case uint:
			return int(val)
		case uint8:
			return int(val)
		case uint16:
			return int(val)
		case uint32:
			return int(val)
		case uint64:
			return int(val)
		case float32:
			return int(val)
		case float64:
			return int(val)
		case string:
			trimmed := strings.TrimSpace(val)
			if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				return int(i)
			}
			if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
				return int(f)
			}

			return 0
		case bool:
			if val {
				return 1
			}

			return 0
		default:
			return 0
		}
	})

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

var (
	stripHTMLScriptStyleRe = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	stripHTMLCommentRe     = regexp.MustCompile(`(?s)<!--.*?-->`)
	stripHTMLTagRe         = regexp.MustCompile(`<[^>]*>`)
)

func stripHTMLFilter(s string) string {
	s = stripHTMLScriptStyleRe.ReplaceAllString(s, "")
	s = stripHTMLCommentRe.ReplaceAllString(s, "")
	s = stripHTMLTagRe.ReplaceAllString(s, "")
	return s
}

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

// isCJKRune reports whether r is a CJK (Chinese, Japanese, Korean) character.
func isCJKRune(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK Extension B
		(r >= 0xAC00 && r <= 0xD7AF) || // Hangul
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) // Katakana
}

// countWordsWithCJK counts words treating each CJK character as an individual word.
func countWordsWithCJK(s string) int {
	count := 0
	inWord := false

	for _, r := range s {
		if isCJKRune(r) {
			if inWord {
				count++
				inWord = false
			}
			count++
		} else if unicode.IsSpace(r) {
			if inWord {
				count++
				inWord = false
			}
		} else {
			inWord = true
		}
	}

	if inWord {
		count++
	}

	return count
}

// isURISafe reports whether r should not be percent-encoded in a URI.
// Matches the behavior of JavaScript's encodeURI().
func isURISafe(r rune) bool {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
		return true
	}

	switch r {
	case '-', '_', '.', '!', '~', '*', '\'', '(', ')',
		';', ',', '/', '?', ':', '@', '&', '=', '+', '$', '#', '[', ']':
		return true
	}

	return false
}

var (
	slugifyDefaultRe   = regexp.MustCompile(`[^\p{L}\p{N}\-]+`)
	slugifyASCIIRe     = regexp.MustCompile(`[^a-z0-9\-]+`)
	slugifyPrettyRe    = regexp.MustCompile(`[^\p{L}\p{N}._~!$&'()*+,;=:@/\-]+`)
	slugifyMultiHyphRe = regexp.MustCompile(`-{2,}`)
	slugifyTrimHyphRe  = regexp.MustCompile(`^-+|-+$`)
)

// latinAccentReplacer maps common accented latin characters to their ASCII equivalents.
var latinAccentReplacer = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o", "ø", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ý", "y", "ÿ", "y",
	"ñ", "n", "ç", "c", "ß", "ss",
	"À", "a", "Á", "a", "Â", "a", "Ã", "a", "Ä", "a", "Å", "a",
	"È", "e", "É", "e", "Ê", "e", "Ë", "e",
	"Ì", "i", "Í", "i", "Î", "i", "Ï", "i",
	"Ò", "o", "Ó", "o", "Ô", "o", "Õ", "o", "Ö", "o", "Ø", "o",
	"Ù", "u", "Ú", "u", "Û", "u", "Ü", "u",
	"Ý", "y", "Ñ", "n", "Ç", "c",
)

// parseToTime converts a Liquid date value (string, time.Time, or int64 unix
// timestamp) to time.Time. Returns (t, true) on success, (zero, false) on failure.
func parseToTime(v any) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case string:
		parsed, err := values.ParseDate(t)
		if err != nil {
			return time.Time{}, false
		}
		return parsed, true
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return time.Unix(rv.Int(), 0), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return time.Unix(int64(rv.Uint()), 0), true
		case reflect.Float32, reflect.Float64:
			return time.Unix(int64(rv.Float()), 0), true
		default:
			return time.Time{}, false
		}
	}
}

// ordinalSuffix returns the ordinal suffix for a day number (1→"st", 2→"nd", etc.).
func ordinalSuffix(n int) string {
	switch {
	case n >= 11 && n <= 13:
		return "th"
	case n%10 == 1:
		return "st"
	case n%10 == 2:
		return "nd"
	case n%10 == 3:
		return "rd"
	default:
		return "th"
	}
}

// formatJekyllDate formats a date in Jekyll's date_to_string / date_to_long_string style.
// monthFmt is the strftime token for the month (%b abbreviated, %B full).
// typeArg is "" for the default DD Mon YYYY format or "ordinal" for ordinal day.
// styleArg is "" (UK: 7th Nov 2008) or "US" (Nov 7th, 2008).
func formatJekyllDate(v any, monthFmt, typeArg, styleArg string) string {
	t, ok := parseToTime(v)
	if !ok {
		return fmt.Sprint(v)
	}
	if typeArg == "ordinal" {
		day := t.Day()
		suffix := ordinalSuffix(day)
		month, _ := tuesday.Strftime(monthFmt, t)
		year := t.Year()
		if styleArg == "US" {
			return fmt.Sprintf("%s %d%s, %d", month, day, suffix, year)
		}
		return fmt.Sprintf("%d%s %s %d", day, suffix, month, year)
	}
	result, _ := tuesday.Strftime("%d "+monthFmt+" %Y", t)
	return result
}

// slugifyString normalizes a string to a URL slug according to the given mode.
// Modes: "default" (unicode-aware), "ascii", "latin" (transliterate accents),
// "pretty" (preserve common URL chars), "none"/"raw" (lowercase only).
// Unknown modes fall back to lowercase-only, matching LiquidJS behavior.
func slugifyString(s, mode string) string {
	applyHyphens := func(s string, re *regexp.Regexp) string {
		s = re.ReplaceAllString(s, "-")
		s = slugifyTrimHyphRe.ReplaceAllString(s, "")
		s = slugifyMultiHyphRe.ReplaceAllString(s, "-")

		return s
	}

	switch mode {
	case "default":
		return applyHyphens(strings.ToLower(s), slugifyDefaultRe)
	case "ascii":
		return applyHyphens(strings.ToLower(s), slugifyASCIIRe)
	case "latin":
		return applyHyphens(strings.ToLower(latinAccentReplacer.Replace(s)), slugifyASCIIRe)
	case "pretty":
		return applyHyphens(strings.ToLower(s), slugifyPrettyRe)
	default:
		// "none", "raw", and any unknown mode: lowercase only, no char replacement.
		return strings.ToLower(s)
	}
}
