package filters

// Ported filter tests from:
//   - Ruby Liquid: test/integration/standard_filter_test.rb
//   - LiquidJS:    test/integration/filters/string.spec.ts
//   - LiquidJS:    test/integration/filters/array.spec.ts
//   - LiquidJS:    test/integration/filters/math.spec.ts
//   - LiquidJS:    test/integration/filters/html.spec.ts
//   - LiquidJS:    test/integration/filters/url.spec.ts

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/expressions"
)

// portedFilterHelper evaluates a Liquid filter expression against the given bindings.
func portedFilterHelper(t *testing.T, expr string, bindings map[string]any) (any, error) {
	t.Helper()
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	ctx := expressions.NewContext(bindings, cfg)
	return expressions.EvaluateString(expr, ctx)
}

// portedFilterEval evaluates expression and asserts no error plus expected value.
func portedFilterEval(t *testing.T, expr string, expected any, bindings ...map[string]any) {
	t.Helper()
	b := map[string]any{}
	if len(bindings) > 0 && bindings[0] != nil {
		b = bindings[0]
	}
	actual, err := portedFilterHelper(t, expr, b)
	require.NoErrorf(t, err, "expression: %s", expr)
	require.Equalf(t, expected, actual, "expression: %s", expr)
}

// ── 2.1 String Filters ────────────────────────────────────────────────────────

// test_downcase / test_upcase [ruby: standard_filter_test.rb]
// downcase / upcase [liquidjs: string.spec.ts]
func TestPortedFilters_Downcase(t *testing.T) {
	portedFilterEval(t, `"Testing" | downcase`, "testing")
	portedFilterEval(t, `"Parker Moore" | downcase`, "parker moore")
	portedFilterEval(t, `"apple" | downcase`, "apple")
	// nil → empty string
	portedFilterEval(t, `nil | downcase`, "")
}

func TestPortedFilters_Upcase(t *testing.T) {
	portedFilterEval(t, `"Testing" | upcase`, "TESTING")
	portedFilterEval(t, `"Parker Moore" | upcase`, "PARKER MOORE")
	// nil → empty string
	portedFilterEval(t, `nil | upcase`, "")
}

// test_capitalize [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Capitalize(t *testing.T) {
	// basic
	portedFilterEval(t, `"title" | capitalize`, "Title")
	portedFilterEval(t, `"my great title" | capitalize`, "My great title")
	// rest is lowercased [ruby]
	portedFilterEval(t, `"MY GREAT TITLE" | capitalize`, "My great title")
	// empty
	portedFilterEval(t, `"" | capitalize`, "")
	// nil → empty [liquidjs]
	portedFilterEval(t, `nil | capitalize`, "")
	// lowercase trailing words [liquidjs]
	portedFilterEval(t, `"foo BaR" | capitalize`, "Foo bar")
}

// test_append [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Append(t *testing.T) {
	bindings := map[string]any{"a": "bc", "b": "d"}
	portedFilterEval(t, `a | append: 'd'`, "bcd", bindings)
	portedFilterEval(t, `a | append: b`, "bcd", bindings)
	portedFilterEval(t, `"/my/fancy/url" | append: ".html"`, "/my/fancy/url.html")
	portedFilterEval(t, `"website.com" | append: "/index.html"`, "website.com/index.html")
}

// test_prepend [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Prepend(t *testing.T) {
	bindings := map[string]any{"a": "bc", "b": "a"}
	portedFilterEval(t, `a | prepend: 'a'`, "abc", bindings)
	portedFilterEval(t, `a | prepend: b`, "abc", bindings)
	portedFilterEval(t, `"apples, oranges, and bananas" | prepend: "Some fruit: "`, "Some fruit: apples, oranges, and bananas")
}

// test_remove [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Remove(t *testing.T) {
	portedFilterEval(t, `"a a a a" | remove: 'a'`, "   ")
	portedFilterEval(t, `"I strained to see the train through the rain" | remove: "rain"`, "I sted to see the t through the ")
	// remove with numeric pattern [ruby]
	portedFilterEval(t, `"1 1 1 1" | remove: 1`, "   ")
}

// test_remove_first [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_RemoveFirst(t *testing.T) {
	portedFilterEval(t, `"a b a a" | remove_first: 'a '`, "b a a")
	portedFilterEval(t, `"I strained to see the train through the rain" | remove_first: "rain"`, "I sted to see the train through the rain")
	// numeric pattern [ruby]
	portedFilterEval(t, `"1 1 1 1" | remove_first: 1`, " 1 1 1")
}

// test_remove (remove_last) [ruby: standard_filter_test.rb]
func TestPortedFilters_RemoveLast(t *testing.T) {
	portedFilterEval(t, `"a a b a" | remove_last: ' a'`, "a a b")
	portedFilterEval(t, `"I strained to see the train through the rain" | remove_last: "rain"`, "I strained to see the train through the ")
	// numeric pattern [ruby]
	portedFilterEval(t, `"1 1 1 1" | remove_last: 1`, "1 1 1 ")
}

// test_replace [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Replace(t *testing.T) {
	portedFilterEval(t, `"a a a a" | replace: 'a', 'b'`, "b b b b")
	// numeric pattern and replacement [ruby]
	portedFilterEval(t, `"1 1 1 1" | replace: '1', 2`, "2 2 2 2")
	portedFilterEval(t, `"1 1 1 1" | replace: 2, 3`, "1 1 1 1")
	portedFilterEval(t, `"Take my protein pills and put my helmet on" | replace: "my", "your"`, "Take your protein pills and put your helmet on")
}

// test_replace_first [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_ReplaceFirst(t *testing.T) {
	portedFilterEval(t, `"a a a a" | replace_first: 'a', 'b'`, "b a a a")
	// numeric [ruby]
	portedFilterEval(t, `"1 1 1 1" | replace_first: '1', 2`, "2 1 1 1")
	portedFilterEval(t, `"Take my protein pills and put my helmet on" | replace_first: "my", "your"`, "Take your protein pills and put my helmet on")
}

// test_replace (replace_last) [ruby: standard_filter_test.rb]
func TestPortedFilters_ReplaceLast(t *testing.T) {
	portedFilterEval(t, `"a a a a" | replace_last: 'a', 'b'`, "a a a b")
	// numeric [ruby]
	portedFilterEval(t, `"1 1 1 1" | replace_last: '1', 2`, "1 1 1 2")
	portedFilterEval(t, `"Take my protein pills and put my helmet on" | replace_last: "my", "your"`, "Take my protein pills and put your helmet on")
}

// test_split [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Split(t *testing.T) {
	// split returns []string in Go
	portedFilterEval(t, `"12~34" | split: '~'`, []string{"12", "34"})
	portedFilterEval(t, `"A? ~ ~ ~ ,Z" | split: '~ ~ ~'`, []string{"A? ", " ,Z"})
	portedFilterEval(t, `"A?Z" | split: '~'`, []string{"A?Z"})
	// nil input → empty slice (Go returns [] not nil here)
	portedFilterEval(t, `nil | split: ' '`, []string{})
	// trailing empty strings removed [liquidjs]
	portedFilterEval(t, `"zebra,octopus,,,," | split: "," | size`, 2)
}

// test_strip_newlines [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_StripNewlines(t *testing.T) {
	portedFilterEval(t, `source | strip_newlines`, "abc", map[string]any{"source": "a\nb\nc"})
	// Windows line endings (\r\n) [ruby + liquidjs]
	portedFilterEval(t, `source | strip_newlines`, "abc", map[string]any{"source": "a\r\nb\nc"})
	// standalone \r
	portedFilterEval(t, `source | strip_newlines`, "abc", map[string]any{"source": "a\rb\nc"})
}

// test_newlines_to_br [ruby: standard_filter_test.rb; liquidjs: html.spec.ts]
func TestPortedFilters_NewlineToBr(t *testing.T) {
	portedFilterEval(t, `source | newline_to_br`, "a<br />\nb<br />\nc",
		map[string]any{"source": "a\nb\nc"})
	// Windows line endings should normalize to \n before conversion [ruby + liquidjs]
	portedFilterEval(t, `source | newline_to_br`, "a<br />\nb<br />\nc",
		map[string]any{"source": "a\r\nb\nc"})
}

// test_truncate [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_Truncate(t *testing.T) {
	portedFilterEval(t, `"1234567890" | truncate: 7`, "1234...")
	portedFilterEval(t, `"1234567890" | truncate: 20`, "1234567890")
	// n <= len(ellipsis) → return ellipsis [ruby + liquidjs]
	portedFilterEval(t, `"1234567890" | truncate: 0`, "...")
	portedFilterEval(t, `"12345" | truncate: 2`, "...")
	// string fits exactly in the limit → no truncation [liquidjs]
	portedFilterEval(t, `"12345" | truncate: 5`, "12345")
	portedFilterEval(t, `"测试测试测试测试" | truncate: 5`, "测试...")
	portedFilterEval(t, `"Ground control to Major Tom." | truncate: 20`, "Ground control to...")
	portedFilterEval(t, `"Ground control to Major Tom." | truncate: 25, ", and so on"`, "Ground control, and so on")
	portedFilterEval(t, `"Ground control to Major Tom." | truncate: 20, ""`, "Ground control to Ma")
	portedFilterEval(t, `"Ground" | truncate: 20`, "Ground")
}

// test_truncatewords [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_TruncateWords(t *testing.T) {
	// more words than limit → truncate
	portedFilterEval(t, `"one two three" | truncatewords: 2`, "one two...")
	// fewer words than limit → no truncation [ruby]
	portedFilterEval(t, `"one two three" | truncatewords: 4`, "one two three")
	portedFilterEval(t, `"one two three" | truncatewords: 15`, "one two three")
	portedFilterEval(t, `"测试测试测试测试" | truncatewords: 5`, "测试测试测试测试")
	portedFilterEval(t, `"Ground control to Major Tom." | truncatewords: 3`, "Ground control to...")
	portedFilterEval(t, `"Ground control to Major Tom." | truncatewords: 3, "--"`, "Ground control to--")
	portedFilterEval(t, `"Ground control to Major Tom." | truncatewords: 3, ""`, "Ground control to")
	// n=0 behaves like n=1 [ruby + liquidjs]
	portedFilterEval(t, `"Ground control to Major Tom." | truncatewords: 0`, "Ground...")
	portedFilterEval(t, `"one two three four" | truncatewords: 2`, "one two...")
	// with tabs/newlines in source [ruby]
	portedFilterEval(t, `source | truncatewords: 3`, "one two three...",
		map[string]any{"source": "one  two\tthree\nfour"})
}

// test_slice string [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Slice(t *testing.T) {
	portedFilterEval(t, `"foobar" | slice: 1, 3`, "oob")
	portedFilterEval(t, `"foobar" | slice: 1, 1000`, "oobar")
	portedFilterEval(t, `"foobar" | slice: 1, 0`, "")
	portedFilterEval(t, `"foobar" | slice: 1, 1`, "o")
	portedFilterEval(t, `"foobar" | slice: 3, 3`, "bar")
	portedFilterEval(t, `"foobar" | slice: -2, 2`, "ar")
	portedFilterEval(t, `"foobar" | slice: -2, 1000`, "ar")
	portedFilterEval(t, `"foobar" | slice: -1`, "r")
	// nil input → nil/empty [ruby: returns empty string; current Go: returns nil]
	portedFilterEval(t, `nil | slice: 0`, nil)
	portedFilterEval(t, `"foobar" | slice: 100, 10`, "")
	portedFilterEval(t, `"foobar" | slice: -100`, "f") // start clamps to 0, length=1

	// Unicode [liquidjs]
	portedFilterEval(t, `"白鵬翔" | slice: 0`, "白")
	portedFilterEval(t, `"白鵬翔" | slice: 1`, "鵬")
	portedFilterEval(t, `"白鵬翔" | slice: 0, 2`, "白鵬")
}

// test_size [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Size(t *testing.T) {
	portedFilterEval(t, `"Ground control to Major Tom." | size`, 28)
	portedFilterEval(t, `"apples, oranges, peaches, plums" | split: ", " | size`, 4)
	portedFilterEval(t, `nil | size`, 0)      // nil → 0
	portedFilterEval(t, `false | size`, 0)    // false → 0
	portedFilterEval(t, `"Straße" | size`, 6) // codepoints, not bytes
}

// test_strip / test_lstrip / test_rstrip [ruby: standard_filter_test.rb; liquidjs: string.spec.ts]
func TestPortedFilters_StripBasic(t *testing.T) {
	portedFilterEval(t, `source | strip`, "ab c", map[string]any{"source": " ab c  "})
	portedFilterEval(t, `source | strip`, "ab c", map[string]any{"source": " \tab c  \n \t"})

	portedFilterEval(t, `source | lstrip`, "ab c  ", map[string]any{"source": " ab c  "})
	portedFilterEval(t, `source | lstrip`, "ab c  \n \t", map[string]any{"source": " \tab c  \n \t"})

	portedFilterEval(t, `source | rstrip`, " ab c", map[string]any{"source": " ab c  "})
	portedFilterEval(t, `source | rstrip`, " \tab c", map[string]any{"source": " \tab c  \n \t"})
}

// test_first_last_on_strings [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_FirstLastOnStrings(t *testing.T) {
	// Basic ASCII
	portedFilterEval(t, `name | first`, "f", map[string]any{"name": "foo"})
	portedFilterEval(t, `name | last`, "o", map[string]any{"name": "foo"})
	// Empty string
	portedFilterEval(t, `name | first`, "", map[string]any{"name": ""})
	portedFilterEval(t, `name | last`, "", map[string]any{"name": ""})
	// Unicode [ruby: test_first_last_on_unicode_strings]
	portedFilterEval(t, `name | first`, "고", map[string]any{"name": "고스트빈"})
	portedFilterEval(t, `name | last`, "빈", map[string]any{"name": "고스트빈"})
}

// test_first_last [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_FirstLastOnArrays(t *testing.T) {
	portedFilterEval(t, `arr | first`, 1, map[string]any{"arr": []any{1, 2, 3}})
	portedFilterEval(t, `arr | last`, 3, map[string]any{"arr": []any{1, 2, 3}})
	portedFilterEval(t, `arr | first`, nil, map[string]any{"arr": []any{}})
	portedFilterEval(t, `arr | last`, nil, map[string]any{"arr": []any{}})
	// nil input [liquidjs]
	portedFilterEval(t, `nil | first`, nil)
	portedFilterEval(t, `nil | last`, nil)
}

// ── 2.2 HTML Filters ──────────────────────────────────────────────────────────

// test_escape [ruby: standard_filter_test.rb; liquidjs: html.spec.ts]
func TestPortedFilters_Escape(t *testing.T) {
	portedFilterEval(t, `"<strong>" | escape`, "&lt;strong&gt;")
	portedFilterEval(t, `1 | escape`, "1")
	portedFilterEval(t, `"Have you read 'James & the Giant Peach'?" | escape`,
		"Have you read &#39;James &amp; the Giant Peach&#39;?")
	// nil → nil (renders as empty string)
	portedFilterEval(t, `nil | escape`, "")
	// undefined → empty [liquidjs]
	portedFilterEval(t, `nonExistent | escape`, "")
}

// test_escape_once [ruby: standard_filter_test.rb; liquidjs: html.spec.ts]
func TestPortedFilters_EscapeOnce(t *testing.T) {
	portedFilterEval(t, `"&lt;strong&gt;Hulk</strong>" | escape_once`, "&lt;strong&gt;Hulk&lt;/strong&gt;")
	portedFilterEval(t, `"1 < 2 & 3" | escape_once`, "1 &lt; 2 &amp; 3")
	portedFilterEval(t, `"1 &lt; 2 &amp; 3" | escape_once`, "1 &lt; 2 &amp; 3")
}

// test_strip_html [ruby + liquidjs: html.spec.ts] — specific edge-case tests.
// Note: most strip_html tests are already in filterTests; this adds extra cases.
func TestPortedFilters_StripHTML_Quirk(t *testing.T) {
	// multiline tags [liquidjs]
	portedFilterEval(t, `html | strip_html`, "test",
		map[string]any{"html": "<div\nclass='multiline'>test</div>"})
	portedFilterEval(t, `html | strip_html`, "test",
		map[string]any{"html": "<!-- foo bar \n test -->test"})
	portedFilterEval(t, `nil | strip_html`, "")
}

// ── 2.3 URL / Encoding Filters ────────────────────────────────────────────────

// test_url_encode [ruby: standard_filter_test.rb; liquidjs: url.spec.ts]
func TestPortedFilters_URLEncode(t *testing.T) {
	portedFilterEval(t, `"john@liquid.com" | url_encode`, "john%40liquid.com")
	portedFilterEval(t, `"Tetsuro Takara" | url_encode`, "Tetsuro+Takara")
	portedFilterEval(t, `"foo+1@example.com" | url_encode`, "foo%2B1%40example.com")
	// numeric → string encode [ruby]
	portedFilterEval(t, `1 | url_encode`, "1")
	// nil → empty string in Go (Ruby returns nil, but Go converts nil to "")
	portedFilterEval(t, `nil | url_encode`, "")
}

// test_url_decode [ruby: standard_filter_test.rb; liquidjs: url.spec.ts]
func TestPortedFilters_URLDecode(t *testing.T) {
	portedFilterEval(t, `"%27Stop%21%27+said+Fred" | url_decode`, "'Stop!' said Fred")
	portedFilterEval(t, `"foo+bar" | url_decode`, "foo bar")
	portedFilterEval(t, `"foo%20bar" | url_decode`, "foo bar")
	portedFilterEval(t, `"foo%2B1%40example.com" | url_decode`, "foo+1@example.com")
	// nil → empty string in Go (Ruby returns nil, but Go converts nil to "")
	portedFilterEval(t, `nil | url_decode`, "")
}

// test_base64_encode [ruby: standard_filter_test.rb]
func TestPortedFilters_Base64Encode(t *testing.T) {
	portedFilterEval(t, `"one two three" | base64_encode`, "b25lIHR3byB0aHJlZQ==")
	portedFilterEval(t, `"hello" | base64_encode`, "aGVsbG8=")
	// nil → empty string [ruby]
	portedFilterEval(t, `nil | base64_encode`, "")
}

// test_base64_decode [ruby: standard_filter_test.rb]
func TestPortedFilters_Base64Decode(t *testing.T) {
	portedFilterEval(t, `"b25lIHR3byB0aHJlZQ==" | base64_decode`, "one two three")
	portedFilterEval(t, `"aGVsbG8=" | base64_decode`, "hello")
	// unicode [ruby]
	portedFilterEval(t, `"4pyF" | base64_decode`, "✅")
}

// test_base64_url_safe_encode [ruby: standard_filter_test.rb]
func TestPortedFilters_Base64URLSafeEncode(t *testing.T) {
	portedFilterEval(t, `"hello" | base64_url_safe_encode`, "aGVsbG8=")
	portedFilterEval(t, `"Man" | base64_url_safe_encode`, "TWFu")
	portedFilterEval(t,
		`"abcdefghijklmnopqrstuvwxyz ABCDEFGHIJKLMNOPQRSTUVWXYZ 1234567890 !@#$%^&*()-=_+/?.:;[]{}\\|" | base64_url_safe_encode`,
		"YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXogQUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVogMTIzNDU2Nzg5MCAhQCMkJV4mKigpLT1fKy8_Ljo7W117fVx8")
	// nil → empty string [ruby]
	portedFilterEval(t, `nil | base64_url_safe_encode`, "")
}

// test_base64_url_safe_decode [ruby: standard_filter_test.rb]
func TestPortedFilters_Base64URLSafeDecode(t *testing.T) {
	portedFilterEval(t, `"aGVsbG8=" | base64_url_safe_decode`, "hello")
	portedFilterEval(t, `"TWFu" | base64_url_safe_decode`, "Man")
	portedFilterEval(t, `"4pyF" | base64_url_safe_decode`, "✅")
}

// ── 2.4 Math Filters ──────────────────────────────────────────────────────────

// test_abs [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Abs(t *testing.T) {
	portedFilterEval(t, `17 | abs`, 17.0)
	portedFilterEval(t, `-17 | abs`, 17.0)
	portedFilterEval(t, `"17" | abs`, 17.0)
	portedFilterEval(t, `"-17" | abs`, 17.0)
	portedFilterEval(t, `0 | abs`, 0.0)
	portedFilterEval(t, `"0" | abs`, 0.0)
	portedFilterEval(t, `17.42 | abs`, 17.42)
	portedFilterEval(t, `-17.42 | abs`, 17.42)
	portedFilterEval(t, `"17.42" | abs`, 17.42)
	portedFilterEval(t, `"-17.42" | abs`, 17.42)
}

// test_plus [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Plus(t *testing.T) {
	portedFilterEval(t, `1 | plus: 1`, int64(2))
	portedFilterEval(t, `4 | plus: 2`, int64(6))
	portedFilterEval(t, `16 | plus: 4`, int64(20))
	portedFilterEval(t, `183.357 | plus: 12`, 195.357)
	// string inputs [liquidjs]
	portedFilterEval(t, `"4" | plus: 2`, 6.0)
	portedFilterEval(t, `"4" | plus: "2"`, 6.0)
	portedFilterEval(t, `"1" | plus: "1.0"`, 2.0)
}

// test_minus [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Minus(t *testing.T) {
	portedFilterEval(t, `input | minus: operand`, int64(4),
		map[string]any{"input": 5, "operand": 1})
	portedFilterEval(t, `4 | minus: 2`, int64(2))
	portedFilterEval(t, `16 | minus: 4`, int64(12))
	portedFilterEval(t, `183.357 | minus: 12`, 171.357)
	portedFilterEval(t, `"4.3" | minus: "2"`, 2.3)
}

// test_times [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Times(t *testing.T) {
	portedFilterEval(t, `3 | times: 4`, int64(12))
	portedFilterEval(t, `3 | times: 2`, int64(6))
	portedFilterEval(t, `24 | times: 7`, int64(168))
	portedFilterEval(t, `183.357 | times: 12`, 2200.2840000000001)
	// string → 0 [ruby]
	portedFilterEval(t, `"foo" | times: 4`, 0.0)
	// string number [ruby]
	portedFilterEval(t, `"24" | times: "7"`, 168.0)
}

// test_divided_by [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_DividedBy(t *testing.T) {
	portedFilterEval(t, `12 | divided_by: 3`, int64(4))
	portedFilterEval(t, `14 | divided_by: 3`, int64(4)) // floor division [ruby]
	portedFilterEval(t, `15 | divided_by: 3`, int64(5))
	portedFilterEval(t, `20 | divided_by: 7`, int64(2))
	// float / integer → float [ruby: assert_template_result("0.5", "{{ 2.0 | divided_by:4 }}")]
	portedFilterEval(t, `fs | divided_by: 4`, 0.5, map[string]any{"fs": float64(2.0)})
	portedFilterEval(t, `20 | divided_by: 7.0`, 2.857142857142857)
}

// test_modulo [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Modulo(t *testing.T) {
	portedFilterEval(t, `3 | modulo: 2`, int64(1))
	portedFilterEval(t, `24 | modulo: 7`, int64(3))
	portedFilterEval(t, `"24" | modulo: "7"`, 3.0) // string inputs → float path
	// floor modulo: result has same sign as divisor [ruby]
	portedFilterEval(t, `-10 | modulo: 3`, int64(2))  // truncated=-1; adjusted: -1+3=2
	portedFilterEval(t, `10 | modulo: -3`, int64(-2)) // truncated=1;  adjusted: 1+(-3)=-2
	portedFilterEval(t, `-7.5 | modulo: 3.0`, 1.5)    // truncated=-1.5; adjusted: -1.5+3.0=1.5
}

// test_ceil [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Ceil(t *testing.T) {
	portedFilterEval(t, `4.6 | ceil`, 5)
	// string "4.3" → parsed as 4.3 → ceil(4.3) = 5 [ruby]
	portedFilterEval(t, `"4.3" | ceil`, 5)
	portedFilterEval(t, `1.2 | ceil`, 2)
	portedFilterEval(t, `2.0 | ceil`, 2)
	portedFilterEval(t, `183.357 | ceil`, 184)
	portedFilterEval(t, `"3.5" | ceil`, 4)
}

// test_floor [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Floor(t *testing.T) {
	portedFilterEval(t, `4.6 | floor`, 4)
	portedFilterEval(t, `"4.3" | floor`, 4)
	portedFilterEval(t, `1.2 | floor`, 1)
	portedFilterEval(t, `2.0 | floor`, 2)
	portedFilterEval(t, `183.357 | floor`, 183)
	portedFilterEval(t, `"3.5" | floor`, 3)
}

// test_round [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_Round(t *testing.T) {
	portedFilterEval(t, `input | round`, 5.0, map[string]any{"input": 4.6})
	portedFilterEval(t, `"4.3" | round`, 4.0)
	portedFilterEval(t, `4.5612 | round: 2`, 4.56)
	portedFilterEval(t, `1.2 | round`, 1.0)
	portedFilterEval(t, `2.7 | round`, 3.0)
	portedFilterEval(t, `183.357 | round: 2`, 183.36)
}

// test_at_least / test_at_most [ruby: standard_filter_test.rb; liquidjs: math.spec.ts]
func TestPortedFilters_AtLeast(t *testing.T) {
	portedFilterEval(t, `5 | at_least: 4`, 5.0)
	portedFilterEval(t, `5 | at_least: 5`, 5.0)
	portedFilterEval(t, `5 | at_least: 6`, 6.0)
	portedFilterEval(t, `4.5 | at_least: 5`, 5.0)
	portedFilterEval(t, `4 | at_least: 5`, 5.0)
	portedFilterEval(t, `4 | at_least: 3`, 4.0)
}

func TestPortedFilters_AtMost(t *testing.T) {
	portedFilterEval(t, `5 | at_most: 4`, 4.0)
	portedFilterEval(t, `5 | at_most: 5`, 5.0)
	portedFilterEval(t, `5 | at_most: 6`, 5.0)
	portedFilterEval(t, `4.5 | at_most: 5`, 4.5)
	portedFilterEval(t, `4 | at_most: 5`, 4.0)
	portedFilterEval(t, `4 | at_most: 3`, 3.0)
}

// ── 2.5 Date Filters ─────────────────────────────────────────────────────────

// test_date [ruby: standard_filter_test.rb]
func TestPortedFilters_Date(t *testing.T) {
	t.Setenv("TZ", "UTC")

	portedFilterEval(t, `"2006-05-05 10:00:00" | date: "%B"`, "May")
	portedFilterEval(t, `"2006-06-05 10:00:00" | date: "%B"`, "June")
	portedFilterEval(t, `"2006-07-05 10:00:00" | date: "%B"`, "July")
	portedFilterEval(t, `"2006-07-05 10:00:00" | date: "%m/%d/%Y"`, "07/05/2006")
	portedFilterEval(t, `"Fri Jul 16 01:00:00 2004" | date: "%m/%d/%Y"`, "07/16/2004")
	// nil → nil [ruby]
	portedFilterEval(t, `nil | date: "%B"`, nil)
	// empty string → empty string [ruby]
	portedFilterEval(t, `"" | date: "%B"`, "")
	// Unix timestamp as int [ruby: test_date UTC]
	portedFilterEval(t, `ts | date: "%m/%d/%Y"`, "07/05/2006",
		map[string]any{"ts": int64(1152098955)})
	// Unix timestamp as string [ruby]
	portedFilterEval(t, `"1152098955" | date: "%m/%d/%Y"`, "07/05/2006")
	// time.Time input
	portedFilterEval(t, `ts | date: "%Y"`, "2015",
		map[string]any{"ts": func() time.Time {
			t, _ := time.Parse(time.RFC3339, "2015-07-17T15:04:05Z")
			return t
		}()})
}

// ── 2.6 Array Filters ─────────────────────────────────────────────────────────

// test_join [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Join(t *testing.T) {
	portedFilterEval(t, `arr | join`, "1 2 3 4", map[string]any{"arr": []any{1, 2, 3, 4}})
	portedFilterEval(t, `arr | join: ' - '`, "1 - 2 - 3 - 4", map[string]any{"arr": []any{1, 2, 3, 4}})
	// integer separator converted to string [ruby]
	portedFilterEval(t, `arr | join: 1`, "1121314", map[string]any{"arr": []any{1, 2, 3, 4}})
	portedFilterEval(t, `"John, Paul, George, Ringo" | split: ", " | join: " and "`,
		"John and Paul and George and Ringo")
}

// test_reverse [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Reverse(t *testing.T) {
	portedFilterEval(t, `arr | reverse`, []any{4, 3, 2, 1},
		map[string]any{"arr": []any{1, 2, 3, 4}})
	portedFilterEval(t, `"plums, peaches, oranges, apples" | split: ", " | reverse | join: ", "`,
		"apples, oranges, peaches, plums")
}

// test_sort [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Sort(t *testing.T) {
	// basic sort
	portedFilterEval(t, `arr | sort`, []any{1, 2, 3, 4},
		map[string]any{"arr": []any{4, 3, 2, 1}})
	// sort by key
	portedFilterEval(t, `arr | sort: 'a' | map: 'a'`,
		[]any{1, 2, 3, 4},
		map[string]any{"arr": []any{
			map[string]any{"a": 4},
			map[string]any{"a": 3},
			map[string]any{"a": 1},
			map[string]any{"a": 2},
		}})
	// nils go last [ruby: test_sort_with_nils]
	portedFilterEval(t, `arr | sort | last`, nil,
		map[string]any{"arr": []any{nil, 4, 3, 2, 1}})
	// numerical sort (integers sorted numerically) [ruby]
	portedFilterEval(t, `arr | sort | first`, 2,
		map[string]any{"arr": []any{10, 2}})
	// lexicographical sort for strings [ruby]
	portedFilterEval(t, `arr | sort | first`, "10",
		map[string]any{"arr": []any{"10", "2"}})
}

// test_sort_natural [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_SortNatural(t *testing.T) {
	portedFilterEval(t, `arr | sort_natural`, []any{"a", "B", "c", "D"},
		map[string]any{"arr": []any{"c", "D", "a", "B"}})
	// case-insensitive sort [ruby: test_sort_natural_case_check]
	portedFilterEval(t, `arr | sort_natural`, []any{"a", "b", "c", "X", "Y", "Z"},
		map[string]any{"arr": []any{"X", "Y", "Z", "a", "b", "c"}})
	// nils go last [ruby: test_sort_natural_with_nils]
	portedFilterEval(t, `arr | sort_natural | last`, nil,
		map[string]any{"arr": []any{nil, "c", "D", "a", "B"}})
}

// test_map [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Map(t *testing.T) {
	portedFilterEval(t, `arr | map: 'a'`, []any{1, 2, 3, 4},
		map[string]any{"arr": []any{
			map[string]any{"a": 1},
			map[string]any{"a": 2},
			map[string]any{"a": 3},
			map[string]any{"a": 4},
		}})
	// chained map [ruby]
	portedFilterEval(t, `ary | map: 'foo' | map: 'bar'`, []any{"a", "b", "c"},
		map[string]any{"ary": []any{
			map[string]any{"foo": map[string]any{"bar": "a"}},
			map[string]any{"foo": map[string]any{"bar": "b"}},
			map[string]any{"foo": map[string]any{"bar": "c"}},
		}})
}

// test_uniq — empty array [ruby: standard_filter_test.rb]
func TestPortedFilters_UniqEdgeCases(t *testing.T) {
	portedFilterEval(t, `arr | uniq: 'a'`, ([]any)(nil),
		map[string]any{"arr": []any{}})
	portedFilterEval(t, `arr | uniq`, []any{"foo"},
		map[string]any{"arr": []any{"foo"}})
	portedFilterEval(t, `arr | uniq`, []any{1, 3, 2, 4},
		map[string]any{"arr": []any{1, 1, 3, 2, 3, 1, 4, 3, 2, 1}})
}

// test_compact — empty array [ruby: standard_filter_test.rb]
func TestPortedFilters_CompactEdgeCases(t *testing.T) {
	portedFilterEval(t, `arr | compact: 'a'`, ([]any)(nil),
		map[string]any{"arr": []any{}})
	// also compact removes nils from plain array
	portedFilterEval(t, `arr | compact`, []any{1, 2, 3},
		map[string]any{"arr": []any{1, nil, 2, nil, 3}})
}

// test_concat [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Concat(t *testing.T) {
	portedFilterEval(t, `arr | concat: extra`, []any{1, 2, 3, 4},
		map[string]any{"arr": []any{1, 2}, "extra": []any{3, 4}})
	portedFilterEval(t, `arr | concat: extra`, []any{1, 2, "a"},
		map[string]any{"arr": []any{1, 2}, "extra": []any{"a"}})
	// nil left value → use right value [liquidjs]
	portedFilterEval(t, `nil | concat: arr`, []any{1, 2},
		map[string]any{"arr": []any{1, 2}})
}

// push / pop / unshift / shift [liquidjs: array.spec.ts] — already tested in filterTests,
// but we add the purity tests (no mutation of original).
func TestPortedFilters_PushPureArray(t *testing.T) {
	// push returns new array; original unchanged [liquidjs]
	bindings := map[string]any{"fruits": []any{"apples", "oranges", "peaches", "plums"}}
	portedFilterEval(t, `fruits | push: "grapes" | size`, 5, bindings)
	portedFilterEval(t, `fruits | size`, 4, bindings) // original not mutated
}

func TestPortedFilters_PopPureArray(t *testing.T) {
	bindings := map[string]any{"fruits": []any{"apples", "oranges", "peaches", "plums"}}
	portedFilterEval(t, `fruits | pop | size`, 3, bindings)
	portedFilterEval(t, `fruits | size`, 4, bindings)
}

func TestPortedFilters_UnshiftPureArray(t *testing.T) {
	bindings := map[string]any{"val": []any{"you"}}
	portedFilterEval(t, `val | unshift: "hey" | first`, "hey", bindings)
	portedFilterEval(t, `val | size`, 1, bindings)
}

func TestPortedFilters_ShiftPureArray(t *testing.T) {
	bindings := map[string]any{"val": []any{"hey", "you"}}
	portedFilterEval(t, `val | shift | first`, "you", bindings)
	portedFilterEval(t, `val | size`, 2, bindings)
}

// test_sum [ruby: standard_filter_test.rb; liquidjs: array.spec.ts]
func TestPortedFilters_Sum(t *testing.T) {
	portedFilterEval(t, `arr | sum`, int64(3),
		map[string]any{"arr": []any{1, 2}})
	portedFilterEval(t, `arr | sum`, int64(10),
		map[string]any{"arr": []any{1, 2, "3", "4"}})
	portedFilterEval(t, `arr | sum: "quantity"`, int64(3),
		map[string]any{"arr": []any{
			map[string]any{"quantity": 1},
			map[string]any{"quantity": 2, "weight": 3},
			map[string]any{"weight": 4},
		}})
}

// ── 2.7 Misc Filters ──────────────────────────────────────────────────────────

// json / inspect [liquidjs: misc.spec.ts]
func TestPortedFilters_JSON(t *testing.T) {
	portedFilterEval(t, `"string" | json`, "\"string\"")
	portedFilterEval(t, `true | json`, "true")
	portedFilterEval(t, `1 | json`, "1")
	portedFilterEval(t, `arr | json`, "[1,2,3]",
		map[string]any{"arr": []any{1, 2, 3}})
}

// to_integer [liquidjs / jekyll]
func TestPortedFilters_ToInteger(t *testing.T) {
	portedFilterEval(t, `"3.5" | to_integer`, 3)
	portedFilterEval(t, `3.9 | to_integer`, 3)
	portedFilterEval(t, `"42" | to_integer`, 42)
	portedFilterEval(t, `true | to_integer`, 1)
	portedFilterEval(t, `false | to_integer`, 0)
}

// default filter [ruby: standard_filter_test.rb]
func TestPortedFilters_Default(t *testing.T) {
	portedFilterEval(t, `"foo" | default: "bar"`, "foo")
	portedFilterEval(t, `nil | default: "bar"`, "bar")
	portedFilterEval(t, `"" | default: "bar"`, "bar")
	portedFilterEval(t, `false | default: "bar"`, "bar")
	portedFilterEval(t, `arr | default: "bar"`, "bar",
		map[string]any{"arr": []any{}})
	portedFilterEval(t, `4.99 | default: 2.99`, 4.99)
	// allow_false [ruby + liquidjs]
	portedFilterEval(t, `false | default: 2.99, allow_false: true`, false)
	portedFilterEval(t, `nil | default: 2.99, allow_false: true`, 2.99)
}

// h alias = escape [ruby: standard_filter_test.rb]
func TestPortedFilters_H(t *testing.T) {
	portedFilterEval(t, `"<strong>" | h`, "&lt;strong&gt;")
	portedFilterEval(t, `1 | h`, "1")
}

// normalize_whitespace / squish [ruby: standard_filter_test.rb; jekyll]
func TestPortedFilters_Squish(t *testing.T) {
	portedFilterEval(t, `"  Hello   World  " | squish`, "Hello World")
	portedFilterEval(t, `s | squish`, "foo bar boo",
		map[string]any{"s": " foo   bar\n\t   boo   "})

	// squish of nil is empty [ruby]
	portedFilterEval(t, `nil | squish`, "")
	portedFilterEval(t, `"  " | squish`, "")
}

// sort nil-last specifically [ruby: standard_filter_test.rb]
func TestPortedFilters_SortNilLast(t *testing.T) {
	input := []any{
		map[string]any{"price": 4, "handle": "alpha"},
		map[string]any{"handle": "beta"},
		map[string]any{"price": 1, "handle": "gamma"},
		map[string]any{"handle": "delta"},
		map[string]any{"price": 2, "handle": "epsilon"},
	}
	result, err := portedFilterHelper(t, `arr | sort: "price" | map: "handle"`,
		map[string]any{"arr": input})
	require.NoError(t, err)
	handles, ok := result.([]any)
	require.True(t, ok)
	require.Equal(t, []any{"gamma", "epsilon", "alpha", "beta", "delta"}, handles)
}

// sort_natural when property missing → nils last [ruby]
func TestPortedFilters_SortNaturalNilLast(t *testing.T) {
	input := []any{
		map[string]any{"price": "4", "handle": "alpha"},
		map[string]any{"handle": "beta"},
		map[string]any{"price": "1", "handle": "gamma"},
		map[string]any{"handle": "delta"},
		map[string]any{"price": "2", "handle": "epsilon"},
	}
	result, err := portedFilterHelper(t, `arr | sort_natural: "price" | map: "handle"`,
		map[string]any{"arr": input})
	require.NoError(t, err)
	handles, ok := result.([]any)
	require.True(t, ok)
	require.Equal(t, []any{"gamma", "epsilon", "alpha", "beta", "delta"}, handles)
}

// divided_by by zero returns an error (Go returns ZeroDivisionError) [ruby/liquidjs]
func TestPortedFilters_DividedByZero(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	ctx := expressions.NewContext(nil, cfg)
	_, err := expressions.EvaluateString(`5 | divided_by: 0`, ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "divided by 0")
}

// modulo by zero returns an error [ruby/liquidjs]
func TestPortedFilters_ModuloByZero(t *testing.T) {
	cfg := expressions.NewConfig()
	AddStandardFilters(&cfg)
	ctx := expressions.NewContext(nil, cfg)
	_, err := expressions.EvaluateString(`1 | modulo: 0`, ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "divided by 0")
}

// ── group_by / find / find_index / where / reject / has ──────────────────────
// (already in filterTests with reference annotations — these add edge cases)

// where: empty string key on non-map items — Go returns empty (diverges from Ruby)
// Ruby's item_property("alpha", "") returns "" (truthy), so all items pass.
// Go returns nothing for empty-key access on strings.
func TestPortedFilters_Where_EmptyStringProperty(t *testing.T) {
	portedFilterEval(t, `arr | where: '' | join: ' '`, "",
		map[string]any{"arr": []any{"alpha", "beta", "gamma"}})
}

// where: filter by truthy property [liquidjs: array.spec.ts]
func TestPortedFilters_Where_TruthyProperty(t *testing.T) {
	products := []any{
		map[string]any{"title": "Vacuum", "type": "living room"},
		map[string]any{"title": "Spatula", "type": "kitchen"},
		map[string]any{"title": "Television", "type": "living room"},
		map[string]any{"title": "Garlic press", "type": "kitchen"},
		map[string]any{"title": "Coffee mug", "available": true},
		map[string]any{"title": "Limited sneakers", "available": false},
	}
	portedFilterEval(t,
		`products | where: "type", "kitchen" | map: "title" | join: ", "`,
		"Spatula, Garlic press",
		map[string]any{"products": products})
}

// find: empty array returns nil [ruby: test_find_with_empty_arrays]
func TestPortedFilters_Find_EmptyArray(t *testing.T) {
	portedFilterEval(t, `arr | find: 'title', 'Not found'`, nil,
		map[string]any{"arr": []any{}})
}

// find_index: empty array returns nil [ruby: test_find_index_with_empty_arrays]
func TestPortedFilters_FindIndex_EmptyArray(t *testing.T) {
	portedFilterEval(t, `arr | find_index: 'title', 'Not found'`, nil,
		map[string]any{"arr": []any{}})
}

// has: empty array returns false [ruby: test_has_on_empty_array]
func TestPortedFilters_Has_EmptyArray(t *testing.T) {
	portedFilterEval(t, `arr | has: 'title', 'Not found'`, false,
		map[string]any{"arr": []any{}})
}

// xml_escape [liquidjs: html.spec.ts; ruby: standard_filter_test.rb]
func TestPortedFilters_XMLEscape(t *testing.T) {
	portedFilterEval(t, `"Have you read 'James & the Giant Peach'?" | xml_escape`,
		"Have you read &#39;James &amp; the Giant Peach&#39;?")
	portedFilterEval(t, `"<script>\"alert\"</script>" | xml_escape`,
		"&lt;script&gt;&#34;alert&#34;&lt;/script&gt;")
}

// Verify we don't emit negative test count for generated test names.
func TestPortedFilters_SliceEdgeCases(t *testing.T) {
	cases := []struct {
		expr     string
		expected any
	}{
		// [ruby: test_slice] — more edge cases
		{`"foobar" | slice: 0, -1`, ""}, // negative length → 0 length
		{`"foobar" | slice: -100`, "f"}, // clamps start to 0, length defaults to 1
		{`"foobar" | slice: 100`, ""},   // start beyond end
		{`"foobar" | slice: 100, 200`, ""},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			portedFilterEval(t, tc.expr, tc.expected)
		})
	}
}
