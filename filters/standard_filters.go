// Package filters is an internal package that defines the standard Liquid filters.
package filters

import (
	"crypto/hmac"
	"crypto/md5"  // #nosec G501
	"crypto/sha1" // #nosec G505
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"html"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/osteele/tuesday"

	"github.com/autopilot3/ap3-types-go/types/date"
	"github.com/autopilot3/liquid/values"
)

// A FilterDictionary holds filters.
type FilterDictionary interface {
	AddFilter(string, interface{})
}

// AddStandardFilters defines the standard Liquid filters.
func AddStandardFilters(fd FilterDictionary) { // nolint: gocyclo
	// value filters
	fd.AddFilter("default", func(value, defaultValue interface{}) interface{} {
		if value == nil || value == false || values.IsEmpty(value) {
			value = defaultValue
		}
		return value
	})

	// array filters
	fd.AddFilter("compact", func(a []interface{}) (result []interface{}) {
		for _, item := range a {
			if item != nil {
				result = append(result, item)
			}
		}
		return
	})
	fd.AddFilter("join", joinFilter)
	fd.AddFilter("map", func(a []map[string]interface{}, key string) (result []interface{}) {
		for _, obj := range a {
			result = append(result, obj[key])
		}
		return result
	})
	fd.AddFilter("reverse", reverseFilter)
	fd.AddFilter("sort", sortFilter)
	// https://shopify.github.io/liquid/ does not demonstrate first and last as filters,
	// but https://help.shopify.com/themes/liquid/filters/array-filters does
	fd.AddFilter("first", func(a []interface{}) interface{} {
		if len(a) == 0 {
			return nil
		}
		return a[0]
	})
	fd.AddFilter("last", func(a []interface{}) interface{} {
		if len(a) == 0 {
			return nil
		}
		return a[len(a)-1]
	})
	fd.AddFilter("uniq", uniqFilter)

	// date filters
	fd.AddFilter("date", func(t interface{}, format func(string) string) (string, error) {
		f := format("%a, %b %d, %y")
		switch tp := t.(type) {
		case date.Date:
			d := t.(date.Date)
			tme, err := d.Time()
			if err != nil {
				return "", err
			}
			return tuesday.Strftime(f, tme)
		case string:
			tme, err := values.ParseDate(t.(string))
			if err != nil {
				return "", err
			}
			return tuesday.Strftime(f, tme)
		case time.Time:
			tme := t.(time.Time)
			return tuesday.Strftime(f, tme)
		case int64:
			unixTime := t.(int64)
			tme := time.Unix(unixTime, 0)
			return tuesday.Strftime(f, tme)
		case float64:
			unixTime := t.(float64)
			tme := time.Unix(int64(unixTime), 0)
			return tuesday.Strftime(f, tme)
		case nil:
			return "", nil
		default:
			return "", fmt.Errorf("date filter: unsupported type %T", tp)
		}
	})

	// number filters
	fd.AddFilter("to_number", func(value interface{}) float64 {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int8:
			return float64(v)
		case int16:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case float32:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
			return 0
		default:
			return 0
		}
	})
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

	fd.AddFilter("at_least", floatFilter(math.Max))
	fd.AddFilter("at_most", floatFilter(math.Min))

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
		runes := []rune(s)
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
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
	fd.AddFilter("replace", func(s, old, new string) string {
		return strings.ReplaceAll(s, old, new)
	})
	fd.AddFilter("replace_first", func(s, old, new string) string {
		return strings.Replace(s, old, new, 1)
	})
	fd.AddFilter("sort_natural", sortNaturalFilter)
	fd.AddFilter("slice", func(s string, start int, length func(int) int) string {
		// runes aren't bytes; don't use slice
		n := length(1)
		runes := []rune(s)
		if start < 0 {
			start = len(runes) + start
		}
		if start < 0 {
			return s
		}
		if start >= len(runes) {
			return ""
		}
		end := start + n
		if end > len(runes) {
			end = len(runes)
		}
		return string(runes[start:end])
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

	// Hash filters

	// #nosec G401
	fd.AddFilter("md5", hashFilter(md5.New))
	fd.AddFilter("sha1", hashFilter(sha1.New))
	fd.AddFilter("sha256", hashFilter(sha256.New))
	// #nosec G401
	fd.AddFilter("hmac", hmacFilter(md5.New))
	fd.AddFilter("hmac_sha1", hmacFilter(sha1.New))
	fd.AddFilter("hmac_sha256", hmacFilter(sha256.New))
}

func hashFilter(hashFn func() hash.Hash) func(value interface{}) string {
	return func(value interface{}) string {
		valueBytes := toBytes(value)
		if len(valueBytes) == 0 {
			return ""
		}
		h := hashFn()
		if _, err := h.Write(valueBytes); err == nil {
			return fmt.Sprintf("%x", h.Sum(nil))
		}
		return ""
	}
}

func hmacFilter(hashFn func() hash.Hash) func(value, key interface{}) string {

	return func(value, key interface{}) string {
		valueBytes := toBytes(value)
		if len(valueBytes) == 0 {
			return ""
		}
		keyBytes := toBytes(key)
		if len(keyBytes) == 0 {
			return ""
		}
		hm := hmac.New(hashFn, keyBytes)
		if _, err := hm.Write(valueBytes); err == nil {
			return fmt.Sprintf("%x", hm.Sum(nil))
		}
		return ""
	}
}

func floatFilter(fn func(v1, v2 float64) float64) func(lhs, rhs interface{}) interface{} {
	return func(lhs, rhs interface{}) interface{} {
		lhsValue, ok := parseAsFloat64(lhs)
		if !ok {
			return ""
		}
		rhsValue, ok := parseAsFloat64(rhs)
		if !ok {
			return ""
		}
		return fn(lhsValue, rhsValue)
	}
}

func parseAsFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func toBytes(value interface{}) []byte {
	switch v := value.(type) {
	case string:
		return []byte(v)
	case int, int8, int16, int32, int64, float32, float64:
		return []byte(fmt.Sprint(v))
	default:
		return nil
	}
}

func joinFilter(a []interface{}, sep func(string) string) interface{} {
	ss := make([]string, 0, len(a))
	s := sep(" ")
	for _, v := range a {
		if v != nil {
			ss = append(ss, fmt.Sprint(v))
		}
	}
	return strings.Join(ss, s)
}

func reverseFilter(a []interface{}) interface{} {
	result := make([]interface{}, len(a))
	for i, x := range a {
		result[len(result)-1-i] = x
	}
	return result
}

var wsre = regexp.MustCompile(`[[:space:]]+`)

func splitFilter(s, sep string) interface{} {
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

func uniqFilter(a []interface{}) (result []interface{}) {
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
	for _, item := range a {
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
