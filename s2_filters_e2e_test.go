package liquid_test

// s2_filters_e2e_test.go — Intensive E2E tests for Section 2: Filtros / Filters
//
// Coverage matrix:
//   A. String filters — downcase, upcase, capitalize, append, prepend,
//      remove/remove_first/remove_last, replace/replace_first/replace_last,
//      split, strip/lstrip/rstrip, strip_html, strip_newlines, newline_to_br,
//      truncate, truncatewords, size, slice, squish, h, xml_escape
//   B. HTML filters — escape, escape_once
//   C. URL/Encoding filters — url_encode/decode, base64_encode/decode,
//      base64_url_safe_encode/decode
//   D. Math filters — abs, plus, minus, times, divided_by, modulo,
//      ceil, floor, round, at_least, at_most
//   E. Date filters — date (string/int/time.Time/nil), date_to_string,
//      date_to_long_string, date_to_xmlschema, date_to_rfc822
//   F. Array filters — join, first/last, reverse, sort, sort_natural, map,
//      sum, compact, uniq, concat, push/pop/unshift/shift, where, reject,
//      find, find_index, has, group_by
//   G. Misc filters — default, json, to_integer, jsonify
//   H. Filter chaining — multiple filters in a single pipeline
//   I. Nil safety — critical filters with nil input
//   J. Regression guard — exact behaviors of the bugs we fixed in this session
//
// Every test creates its own Engine instance for full isolation.
// Reference: Ruby Liquid standard_filter_test.rb + LiquidJS *.spec.ts

import (
	"testing"
	"time"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func s2eng() *liquid.Engine { return liquid.NewEngine() }

func s2render(t *testing.T, tpl string, binds map[string]any) string {
	t.Helper()
	out, err := s2eng().ParseAndRenderString(tpl, binds)
	require.NoError(t, err, "template: %q", tpl)
	return out
}

func s2renderErr(t *testing.T, tpl string, binds map[string]any) (string, error) {
	t.Helper()
	return s2eng().ParseAndRenderString(tpl, binds)
}

// s2eq is a one-liner assertion helper.
func s2eq(t *testing.T, want, tpl string, binds map[string]any) {
	t.Helper()
	assert.Equal(t, want, s2render(t, tpl, binds), "template: %q", tpl)
}

// s2plain renders with no bindings.
func s2plain(t *testing.T, tpl string) string {
	t.Helper()
	return s2render(t, tpl, nil)
}

// ═════════════════════════════════════════════════════════════════════════════
// A. String Filters
// ═════════════════════════════════════════════════════════════════════════════

// ── A1. downcase / upcase ────────────────────────────────────────────────────

func TestS2_Downcase_Basic(t *testing.T) {
	assert.Equal(t, "hello world", s2plain(t, `{{ "Hello World" | downcase }}`))
}

func TestS2_Downcase_AlreadyLower(t *testing.T) {
	assert.Equal(t, "abc", s2plain(t, `{{ "abc" | downcase }}`))
}

func TestS2_Downcase_Mixed(t *testing.T) {
	assert.Equal(t, "abc 123 xyz", s2plain(t, `{{ "ABC 123 XYZ" | downcase }}`))
}

func TestS2_Upcase_Basic(t *testing.T) {
	assert.Equal(t, "HELLO WORLD", s2plain(t, `{{ "Hello World" | upcase }}`))
}

func TestS2_Upcase_AlreadyUpper(t *testing.T) {
	assert.Equal(t, "ABC", s2plain(t, `{{ "ABC" | upcase }}`))
}

func TestS2_Downcase_ViaBinding(t *testing.T) {
	s2eq(t, "liquid", `{{ s | downcase }}`, map[string]any{"s": "LIQUID"})
}

func TestS2_Upcase_ViaBinding(t *testing.T) {
	s2eq(t, "LIQUID", `{{ s | upcase }}`, map[string]any{"s": "liquid"})
}

// ── A2. capitalize ───────────────────────────────────────────────────────────

func TestS2_Capitalize_Basic(t *testing.T) {
	assert.Equal(t, "Hello", s2plain(t, `{{ "hello" | capitalize }}`))
}

func TestS2_Capitalize_RestBecomesLower(t *testing.T) {
	// Ruby/LiquidJS: capitalize uppercases first char, lowercases the rest
	assert.Equal(t, "My great title", s2plain(t, `{{ "MY GREAT TITLE" | capitalize }}`))
}

func TestS2_Capitalize_SingleChar(t *testing.T) {
	assert.Equal(t, "A", s2plain(t, `{{ "a" | capitalize }}`))
}

func TestS2_Capitalize_AlreadyCapitalized(t *testing.T) {
	assert.Equal(t, "Hello world", s2plain(t, `{{ "Hello World" | capitalize }}`))
}

func TestS2_Capitalize_EmptyString(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | capitalize }}`))
}

// ── A3. append / prepend ─────────────────────────────────────────────────────

func TestS2_Append_Basic(t *testing.T) {
	assert.Equal(t, "hello world", s2plain(t, `{{ "hello" | append: " world" }}`))
}

func TestS2_Append_Chained(t *testing.T) {
	assert.Equal(t, "hello world!", s2plain(t, `{{ "hello" | append: " world" | append: "!" }}`))
}

func TestS2_Append_WithBinding(t *testing.T) {
	s2eq(t, "/index.html", `{{ url | append: ".html" }}`, map[string]any{"url": "/index"})
}

func TestS2_Prepend_Basic(t *testing.T) {
	assert.Equal(t, "world hello", s2plain(t, `{{ "hello" | prepend: "world " }}`))
}

func TestS2_Prepend_Chained(t *testing.T) {
	assert.Equal(t, "!world hello", s2plain(t, `{{ "hello" | prepend: "world " | prepend: "!" }}`))
}

func TestS2_AppendPrepend_Combined(t *testing.T) {
	s2eq(t, "<em>text</em>", `{{ word | prepend: "<em>" | append: "</em>" }}`,
		map[string]any{"word": "text"})
}

// ── A4. remove / remove_first / remove_last ──────────────────────────────────

func TestS2_Remove_AllOccurrences(t *testing.T) {
	assert.Equal(t, "The cat sat on the mat", s2plain(t,
		`{{ "The r cat r sat on r the mat" | remove: "r " }}`))
}

func TestS2_Remove_NotFound(t *testing.T) {
	assert.Equal(t, "abc", s2plain(t, `{{ "abc" | remove: "z" }}`))
}

func TestS2_RemoveFirst_OnlyFirst(t *testing.T) {
	assert.Equal(t, "The cat sat on the rat mat", s2plain(t,
		`{{ "The rat cat sat on the rat mat" | remove_first: "rat " }}`))
}

func TestS2_RemoveLast_OnlyLast(t *testing.T) {
	assert.Equal(t, "The rat cat sat on the mat", s2plain(t,
		`{{ "The rat cat sat on the rat mat" | remove_last: " rat" }}`))
}

func TestS2_Remove_InTemplate(t *testing.T) {
	s2eq(t, "hello", `{{ s | remove: " world" }}`, map[string]any{"s": "hello world"})
}

// ── A5. replace / replace_first / replace_last ──────────────────────────────

func TestS2_Replace_AllOccurrences(t *testing.T) {
	assert.Equal(t, "1, 1, 1", s2plain(t, `{{ "1, 2, 3" | replace: "2", "1" | replace: "3", "1" }}`))
}

func TestS2_Replace_NotFound(t *testing.T) {
	assert.Equal(t, "abc", s2plain(t, `{{ "abc" | replace: "z", "x" }}`))
}

func TestS2_ReplaceFirst_OnlyFirst(t *testing.T) {
	// replace_first replaces the very first occurrence
	assert.Equal(t, "2, 1, 3", s2plain(t, `{{ "1, 1, 3" | replace_first: "1", "2" }}`))
}

func TestS2_ReplaceLast_OnlyLast(t *testing.T) {
	assert.Equal(t, "1, 1, 2", s2plain(t, `{{ "1, 1, 1" | replace_last: "1", "2" }}`))
}

func TestS2_Replace_InAssign(t *testing.T) {
	assert.Equal(t, "Hello Liquid", s2plain(t,
		`{% assign s = "Hello World" | replace: "World", "Liquid" %}{{ s }}`))
}

// ── A6. split ────────────────────────────────────────────────────────────────

func TestS2_Split_Basic(t *testing.T) {
	assert.Equal(t, "foo bar baz", s2plain(t, `{{ "foo,bar,baz" | split: "," | join: " " }}`))
}

func TestS2_Split_MultiCharSep(t *testing.T) {
	// "A? ~ ~ ~ ,Z".split("~ ~ ~") = ["A? ", " ,Z"]; join(" ") = "A?   ,Z" (3 spaces: trailing + sep + leading)
	assert.Equal(t, "A?   ,Z", s2plain(t, `{{ "A? ~ ~ ~ ,Z" | split: "~ ~ ~" | join: " " }}`))
}

func TestS2_Split_NoSepFound(t *testing.T) {
	// When separator not found, returns array with the whole string
	assert.Equal(t, "1", s2plain(t, `{{ "abc" | split: "~" | size }}`))
}

func TestS2_Split_TrailingEmptyStringsRemoved(t *testing.T) {
	// Ruby removes trailing empty strings after split
	assert.Equal(t, "2", s2plain(t, `{{ "zebra,octopus,,,," | split: "," | size }}`))
}

func TestS2_Split_ThenFirst(t *testing.T) {
	assert.Equal(t, "one", s2plain(t, `{{ "one two three" | split: " " | first }}`))
}

func TestS2_Split_ThenLast(t *testing.T) {
	assert.Equal(t, "three", s2plain(t, `{{ "one two three" | split: " " | last }}`))
}

func TestS2_Split_ThenReverse(t *testing.T) {
	assert.Equal(t, "c b a", s2plain(t, `{{ "a b c" | split: " " | reverse | join: " " }}`))
}

func TestS2_Split_InForLoop(t *testing.T) {
	// split result can be iterated in a for loop
	out := s2plain(t, `{% for w in "one,two,three" | split: "," %}<{{ w }}>{% endfor %}`)
	assert.Equal(t, "<one><two><three>", out)
}

// ── A7. strip / lstrip / rstrip ──────────────────────────────────────────────

func TestS2_Strip_RemovesBothSides(t *testing.T) {
	assert.Equal(t, "hello", s2plain(t, `{{ "  hello   " | strip }}`))
}

func TestS2_Lstrip_RemovesLeft(t *testing.T) {
	assert.Equal(t, "hello   ", s2plain(t, `{{ "  hello   " | lstrip }}`))
}

func TestS2_Rstrip_RemovesRight(t *testing.T) {
	assert.Equal(t, "  hello", s2plain(t, `{{ "  hello   " | rstrip }}`))
}

func TestS2_Strip_Tabs(t *testing.T) {
	assert.Equal(t, "x", s2plain(t, `{{ "\t x \t" | strip }}`))
}

func TestS2_Strip_Newlines(t *testing.T) {
	assert.Equal(t, "x", s2plain(t, `{{ "\nx\n" | strip }}`))
}

func TestS2_Strip_AlreadyClean(t *testing.T) {
	assert.Equal(t, "clean", s2plain(t, `{{ "clean" | strip }}`))
}

func TestS2_Strip_Empty(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | strip }}`))
}

// ── A8. strip_html ───────────────────────────────────────────────────────────

func TestS2_StripHtml_BasicTags(t *testing.T) {
	assert.Equal(t, "Hello World", s2plain(t, `{{ "<p>Hello <b>World</b></p>" | strip_html }}`))
}

func TestS2_StripHtml_ScriptTagWithContent(t *testing.T) {
	// <script>…</script> is removed including content; surrounding spaces preserved
	assert.Equal(t, "before  after", s2plain(t,
		`{{ "before <script>alert('xss')</script> after" | strip_html }}`))
}

func TestS2_StripHtml_StyleTagWithContent(t *testing.T) {
	assert.Equal(t, "visible", s2plain(t,
		`{{ "<style>body{color:red}</style>visible" | strip_html }}`))
}

func TestS2_StripHtml_HtmlComment(t *testing.T) {
	assert.Equal(t, "visible", s2plain(t,
		`{{ "<!-- hidden -->visible" | strip_html }}`))
}

func TestS2_StripHtml_CaseInsensitiveScript(t *testing.T) {
	assert.Equal(t, "clean", s2plain(t,
		`{{ "<SCRIPT>bad()</SCRIPT>clean" | strip_html }}`))
}

func TestS2_StripHtml_EmptyString(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | strip_html }}`))
}

func TestS2_StripHtml_NoTags(t *testing.T) {
	assert.Equal(t, "plain text", s2plain(t, `{{ "plain text" | strip_html }}`))
}

// ── A9. strip_newlines ───────────────────────────────────────────────────────

func TestS2_StripNewlines_UnixLineEndings(t *testing.T) {
	s2eq(t, "abc", `{{ s | strip_newlines }}`, map[string]any{"s": "a\nb\nc"})
}

func TestS2_StripNewlines_WindowsLineEndings(t *testing.T) {
	// \r\n must also be stripped — regression guard for the fix
	s2eq(t, "abc", `{{ s | strip_newlines }}`, map[string]any{"s": "a\r\nb\r\nc"})
}

func TestS2_StripNewlines_StandaloneCarriageReturn(t *testing.T) {
	s2eq(t, "abc", `{{ s | strip_newlines }}`, map[string]any{"s": "a\rb\rc"})
}

func TestS2_StripNewlines_Mixed(t *testing.T) {
	s2eq(t, "abcd", `{{ s | strip_newlines }}`, map[string]any{"s": "a\nb\r\nc\rd"})
}

func TestS2_StripNewlines_NoNewlines(t *testing.T) {
	assert.Equal(t, "hello", s2plain(t, `{{ "hello" | strip_newlines }}`))
}

func TestS2_StripNewlines_EmptyResult(t *testing.T) {
	s2eq(t, "", `{{ s | strip_newlines }}`, map[string]any{"s": "\n\r\n\r"})
}

// ── A10. newline_to_br ───────────────────────────────────────────────────────

func TestS2_NewlineToBr_Basic(t *testing.T) {
	s2eq(t, "a<br />\nb<br />\nc", `{{ s | newline_to_br }}`,
		map[string]any{"s": "a\nb\nc"})
}

func TestS2_NewlineToBr_WindowsLineEndings(t *testing.T) {
	// \r\n → <br />\n (not double <br />) — regression guard
	s2eq(t, "a<br />\nb<br />\nc", `{{ s | newline_to_br }}`,
		map[string]any{"s": "a\r\nb\r\nc"})
}

func TestS2_NewlineToBr_PreservesNewlineAfterBr(t *testing.T) {
	// The newline after <br /> must exist for HTML block formatting
	s2eq(t, "line1<br />\nline2", `{{ s | newline_to_br }}`,
		map[string]any{"s": "line1\nline2"})
}

func TestS2_NewlineToBr_EmptyString(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | newline_to_br }}`))
}

func TestS2_NewlineToBr_NoNewlines(t *testing.T) {
	assert.Equal(t, "hello", s2plain(t, `{{ "hello" | newline_to_br }}`))
}

// ── A11. truncate ────────────────────────────────────────────────────────────

func TestS2_Truncate_Basic(t *testing.T) {
	assert.Equal(t, "1234...", s2plain(t, `{{ "1234567890" | truncate: 7 }}`))
}

func TestS2_Truncate_StringShorterThanLimit(t *testing.T) {
	// String that fits entirely — no truncation
	assert.Equal(t, "1234567890", s2plain(t, `{{ "1234567890" | truncate: 20 }}`))
}

func TestS2_Truncate_ExactFit(t *testing.T) {
	// String whose length == n — must NOT be truncated — regression guard
	assert.Equal(t, "12345", s2plain(t, `{{ "12345" | truncate: 5 }}`))
}

func TestS2_Truncate_LimitSmallerThanEllipsis(t *testing.T) {
	// n < len(ellipsis="...") → return just the ellipsis — regression guard
	assert.Equal(t, "...", s2plain(t, `{{ "1234567890" | truncate: 0 }}`))
	assert.Equal(t, "...", s2plain(t, `{{ "1234567890" | truncate: 2 }}`))
}

func TestS2_Truncate_CustomEllipsis(t *testing.T) {
	assert.Equal(t, "Ground control, and so on", s2plain(t,
		`{{ "Ground control to Major Tom." | truncate: 25, ", and so on" }}`))
}

func TestS2_Truncate_EmptyEllipsis(t *testing.T) {
	assert.Equal(t, "Ground control to Ma", s2plain(t,
		`{{ "Ground control to Major Tom." | truncate: 20, "" }}`))
}

func TestS2_Truncate_Unicode(t *testing.T) {
	// truncate counts Unicode runes, not bytes
	assert.Equal(t, "测试...", s2plain(t, `{{ "测试测试测试测试" | truncate: 5 }}`))
}

func TestS2_Truncate_InAssign(t *testing.T) {
	assert.Equal(t, "Ground control to...", s2plain(t,
		`{% assign s = "Ground control to Major Tom." | truncate: 20 %}{{ s }}`))
}

// ── A12. truncatewords ───────────────────────────────────────────────────────

func TestS2_TruncateWords_MoreWordsThanLimit(t *testing.T) {
	assert.Equal(t, "one two...", s2plain(t, `{{ "one two three" | truncatewords: 2 }}`))
}

func TestS2_TruncateWords_FewerWordsThanLimit(t *testing.T) {
	// String has fewer words than n → return unchanged — no ellipsis
	assert.Equal(t, "one two three", s2plain(t, `{{ "one two three" | truncatewords: 4 }}`))
}

func TestS2_TruncateWords_ExactWordCount(t *testing.T) {
	assert.Equal(t, "one two three", s2plain(t, `{{ "one two three" | truncatewords: 3 }}`))
}

func TestS2_TruncateWords_NIsZero(t *testing.T) {
	// n=0 → behaves like n=1 (keeps first word) — regression guard
	assert.Equal(t, "Ground...", s2plain(t,
		`{{ "Ground control to Major Tom." | truncatewords: 0 }}`))
}

func TestS2_TruncateWords_N1(t *testing.T) {
	assert.Equal(t, "Ground...", s2plain(t,
		`{{ "Ground control to Major Tom." | truncatewords: 1 }}`))
}

func TestS2_TruncateWords_BasicThree(t *testing.T) {
	assert.Equal(t, "Ground control to...", s2plain(t,
		`{{ "Ground control to Major Tom." | truncatewords: 3 }}`))
}

func TestS2_TruncateWords_CustomEllipsis(t *testing.T) {
	assert.Equal(t, "Ground control to--", s2plain(t,
		`{{ "Ground control to Major Tom." | truncatewords: 3, "--" }}`))
}

func TestS2_TruncateWords_EmptyEllipsis(t *testing.T) {
	assert.Equal(t, "Ground control to", s2plain(t,
		`{{ "Ground control to Major Tom." | truncatewords: 3, "" }}`))
}

func TestS2_TruncateWords_WhitespaceNormalized(t *testing.T) {
	// tabs and newlines in source: words are joined with single spaces — regression guard
	s2eq(t, "one two three...", `{{ s | truncatewords: 3 }}`,
		map[string]any{"s": "one  two\tthree\nfour five"})
}

// ── A13. size ────────────────────────────────────────────────────────────────

func TestS2_Size_String(t *testing.T) {
	assert.Equal(t, "6", s2plain(t, `{{ "foobar" | size }}`))
}

func TestS2_Size_Array(t *testing.T) {
	s2eq(t, "3", `{{ arr | size }}`, map[string]any{"arr": []any{1, 2, 3}})
}

func TestS2_Size_EmptyString(t *testing.T) {
	assert.Equal(t, "0", s2plain(t, `{{ "" | size }}`))
}

func TestS2_Size_EmptyArray(t *testing.T) {
	s2eq(t, "0", `{{ arr | size }}`, map[string]any{"arr": []any{}})
}

func TestS2_Size_Unicode(t *testing.T) {
	// Size counts characters (runes), not bytes
	assert.Equal(t, "3", s2plain(t, `{{ "日本語" | size }}`))
}

func TestS2_Size_InCondition(t *testing.T) {
	// Filter chains are not valid directly in {% if %} — must assign first
	s2eq(t, "long", `{% assign n = s | size %}{% if n > 5 %}long{% else %}short{% endif %}`,
		map[string]any{"s": "foobar"})
}

// ── A14. slice ───────────────────────────────────────────────────────────────

func TestS2_Slice_String_Basic(t *testing.T) {
	assert.Equal(t, "oob", s2plain(t, `{{ "foobar" | slice: 1, 3 }}`))
}

func TestS2_Slice_String_SingleChar_Default(t *testing.T) {
	assert.Equal(t, "o", s2plain(t, `{{ "foobar" | slice: 1 }}`))
}

func TestS2_Slice_String_NegativeStart(t *testing.T) {
	assert.Equal(t, "ar", s2plain(t, `{{ "foobar" | slice: -2, 2 }}`))
}

func TestS2_Slice_String_StartBeyondEnd(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "foobar" | slice: 100 }}`))
}

func TestS2_Slice_String_LengthBeyondEnd(t *testing.T) {
	assert.Equal(t, "oobar", s2plain(t, `{{ "foobar" | slice: 1, 1000 }}`))
}

func TestS2_Slice_String_NegativeLength_NoOutput(t *testing.T) {
	// slice with negative length is clamped to zero — regression guard (no panic)
	assert.Equal(t, "", s2plain(t, `{{ "foobar" | slice: 0, -1 }}`))
}

func TestS2_Slice_String_NegativeStartClampedToZero(t *testing.T) {
	// start -100 on 6-char string → clamped to 0; length=1 → "f"
	assert.Equal(t, "f", s2plain(t, `{{ "foobar" | slice: -100 }}`))
}

func TestS2_Slice_Array_Basic(t *testing.T) {
	s2eq(t, "b c", `{{ arr | slice: 1, 2 | join: " " }}`,
		map[string]any{"arr": []any{"a", "b", "c", "d"}})
}

func TestS2_Slice_Array_NegativeStart(t *testing.T) {
	s2eq(t, "d", `{{ arr | slice: -1 | join: "" }}`,
		map[string]any{"arr": []any{"a", "b", "c", "d"}})
}

func TestS2_Slice_Unicode_Runes(t *testing.T) {
	// Slice works on Unicode code points, not bytes
	assert.Equal(t, "本語", s2plain(t, `{{ "日本語" | slice: 1, 2 }}`))
}

// ── A15. squish ───────────────────────────────────────────────────────────────

func TestS2_Squish_CollapseSpaces(t *testing.T) {
	assert.Equal(t, "Hello World", s2plain(t, `{{ "  Hello   World  " | squish }}`))
}

func TestS2_Squish_CollapseTabsAndNewlines(t *testing.T) {
	s2eq(t, "foo bar boo", `{{ s | squish }}`,
		map[string]any{"s": " foo   bar\n\t   boo   "})
}

func TestS2_Squish_WhitespaceOnly(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "   " | squish }}`))
}

func TestS2_Squish_EmptyString(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | squish }}`))
}

func TestS2_Squish_AlreadyClean(t *testing.T) {
	assert.Equal(t, "clean string", s2plain(t, `{{ "clean string" | squish }}`))
}

// ── A16. h (alias for escape) ─────────────────────────────────────────────────

func TestS2_H_EscapesHtml(t *testing.T) {
	assert.Equal(t, "&lt;strong&gt;", s2plain(t, `{{ "<strong>" | h }}`))
}

func TestS2_H_AllSpecialChars(t *testing.T) {
	// Go's html.EscapeString encodes " as &#34; (not &quot;)
	s2eq(t, "&lt;p class=&#34;x&#34;&gt;&amp;hello&lt;/p&gt;", `{{ s | h }}`,
		map[string]any{"s": `<p class="x">&hello</p>`})
}

func TestS2_H_Number(t *testing.T) {
	assert.Equal(t, "42", s2plain(t, `{{ 42 | h }}`))
}

// ── A17. xml_escape ───────────────────────────────────────────────────────────

func TestS2_XmlEscape_Basic(t *testing.T) {
	// Go's html.EscapeString encodes " as &#34;
	s2eq(t, "&lt;tag&gt;&amp;&#34;hello&#34;", `{{ s | xml_escape }}`,
		map[string]any{"s": `<tag>&"hello"`})
}

func TestS2_XmlEscape_Apos(t *testing.T) {
	s2eq(t, "it&#39;s", `{{ s | xml_escape }}`,
		map[string]any{"s": "it's"})
}

func TestS2_XmlEscape_EmptyString(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "" | xml_escape }}`))
}

// ═════════════════════════════════════════════════════════════════════════════
// B. HTML Filters
// ═════════════════════════════════════════════════════════════════════════════

func TestS2_Escape_AllSpecialChars(t *testing.T) {
	// Go's html.EscapeString encodes " as &#34; (not &quot;)
	s2eq(t, "&lt;p&gt;&amp;&#34;test&#34;&lt;/p&gt;", `{{ s | escape }}`,
		map[string]any{"s": `<p>&"test"</p>`})
}

func TestS2_Escape_Idempotent(t *testing.T) {
	// escape of an already-escaped string — does NOT double-escape
	// This is the raw escape filter (not escape_once)
	s2eq(t, "&amp;lt;p&amp;gt;", `{{ s | escape }}`,
		map[string]any{"s": "&lt;p&gt;"})
}

func TestS2_Escape_CleanString(t *testing.T) {
	assert.Equal(t, "hello world", s2plain(t, `{{ "hello world" | escape }}`))
}

func TestS2_EscapeOnce_DoesNotDoubleEscape(t *testing.T) {
	// escape_once leaves already-escaped sequences alone
	s2eq(t, "&lt;p&gt;&amp;already&lt;/p&gt;", `{{ s | escape_once }}`,
		map[string]any{"s": "&lt;p&gt;&already</p>"})
}

func TestS2_EscapeOnce_EscapesUnescaped(t *testing.T) {
	s2eq(t, "&lt;b&gt;bold&lt;/b&gt;", `{{ s | escape_once }}`,
		map[string]any{"s": "<b>bold</b>"})
}

// ═════════════════════════════════════════════════════════════════════════════
// C. URL / Encoding Filters
// ═════════════════════════════════════════════════════════════════════════════

func TestS2_UrlEncode_SpacesAndSpecialChars(t *testing.T) {
	assert.Equal(t, "foo+bar+baz", s2plain(t, `{{ "foo bar baz" | url_encode }}`))
}

func TestS2_UrlEncode_SpecialSymbols(t *testing.T) {
	assert.Equal(t, "foo%40bar.com", s2plain(t, `{{ "foo@bar.com" | url_encode }}`))
}

func TestS2_UrlDecode_Basic(t *testing.T) {
	assert.Equal(t, "foo bar baz", s2plain(t, `{{ "foo+bar+baz" | url_decode }}`))
}

func TestS2_UrlEncodeDecode_RoundTrip(t *testing.T) {
	s2eq(t, "foo@bar.com", `{{ s | url_encode | url_decode }}`,
		map[string]any{"s": "foo@bar.com"})
}

func TestS2_Base64Encode_Basic(t *testing.T) {
	assert.Equal(t, "aGVsbG8=", s2plain(t, `{{ "hello" | base64_encode }}`))
}

func TestS2_Base64Decode_Basic(t *testing.T) {
	assert.Equal(t, "hello", s2plain(t, `{{ "aGVsbG8=" | base64_decode }}`))
}

func TestS2_Base64EncodeDecode_RoundTrip(t *testing.T) {
	s2eq(t, "hello world", `{{ s | base64_encode | base64_decode }}`,
		map[string]any{"s": "hello world"})
}

func TestS2_Base64UrlSafeEncode_NoPlusOrSlash(t *testing.T) {
	// URL-safe base64 uses - and _ instead of + and /
	out := s2plain(t, `{{ "hello world+/" | base64_url_safe_encode }}`)
	assert.NotContains(t, out, "+")
	assert.NotContains(t, out, "/")
}

func TestS2_Base64UrlSafe_RoundTrip(t *testing.T) {
	s2eq(t, "hello world+/!", `{{ s | base64_url_safe_encode | base64_url_safe_decode }}`,
		map[string]any{"s": "hello world+/!"})
}

// ═════════════════════════════════════════════════════════════════════════════
// D. Math Filters
// ═════════════════════════════════════════════════════════════════════════════

// ── D1. abs ──────────────────────────────────────────────────────────────────

func TestS2_Abs_Positive(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 5 | abs }}`))
}

func TestS2_Abs_Negative(t *testing.T) {
	assert.Equal(t, "17", s2plain(t, `{{ -17 | abs }}`))
}

func TestS2_Abs_NegativeFloat(t *testing.T) {
	// abs returns float64 but printed as "4" (trailing .0 stripped)
	assert.Equal(t, "4", s2plain(t, `{{ -4.0 | abs }}`))
}

func TestS2_Abs_StringNumber(t *testing.T) {
	assert.Equal(t, "19.86", s2plain(t, `{{ "-19.86" | abs }}`))
}

func TestS2_Abs_Zero(t *testing.T) {
	assert.Equal(t, "0", s2plain(t, `{{ 0 | abs }}`))
}

// ── D2. plus / minus / times ─────────────────────────────────────────────────

func TestS2_Plus_Integers(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 2 | plus: 3 }}`))
}

func TestS2_Plus_Floats(t *testing.T) {
	assert.Equal(t, "5.5", s2plain(t, `{{ 3.5 | plus: 2.0 }}`))
}

func TestS2_Plus_IntFloat(t *testing.T) {
	assert.Equal(t, "5.5", s2plain(t, `{{ 3 | plus: 2.5 }}`))
}

func TestS2_Plus_StringCoercion(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ "2" | plus: 3 }}`))
}

func TestS2_Minus_Basic(t *testing.T) {
	assert.Equal(t, "2", s2plain(t, `{{ 5 | minus: 3 }}`))
}

func TestS2_Minus_Floats(t *testing.T) {
	assert.Equal(t, "1.5", s2plain(t, `{{ 4.5 | minus: 3.0 }}`))
}

func TestS2_Times_BasicInt(t *testing.T) {
	assert.Equal(t, "12", s2plain(t, `{{ 3 | times: 4 }}`))
}

func TestS2_Times_Float(t *testing.T) {
	assert.Equal(t, "7.5", s2plain(t, `{{ 3 | times: 2.5 }}`))
}

func TestS2_Times_StringCoercion(t *testing.T) {
	assert.Equal(t, "6", s2plain(t, `{{ "3" | times: 2 }}`))
}

// ── D3. divided_by ───────────────────────────────────────────────────────────

func TestS2_DividedBy_IntInt_FloorDivision(t *testing.T) {
	// int / int → integer floor division
	assert.Equal(t, "3", s2plain(t, `{{ 10 | divided_by: 3 }}`))
}

func TestS2_DividedBy_IntFloat_FloatResult(t *testing.T) {
	// int / float → float
	assert.Equal(t, "3.3333333333333335", s2plain(t, `{{ 10 | divided_by: 3.0 }}`))
}

func TestS2_DividedBy_FloatInt_FloatResult(t *testing.T) {
	// float input / int → float result — regression guard
	s2eq(t, "0.5", `{{ n | divided_by: 4 }}`, map[string]any{"n": float64(2.0)})
}

func TestS2_DividedBy_FloatFloat(t *testing.T) {
	assert.Equal(t, "2.5", s2plain(t, `{{ 5.0 | divided_by: 2.0 }}`))
}

func TestS2_DividedBy_NegativeFloor(t *testing.T) {
	// Go integer division truncates toward zero: -10 / 3 = -3 (not -4)
	assert.Equal(t, "-3", s2plain(t, `{{ -10 | divided_by: 3 }}`))
}

func TestS2_DividedBy_ZeroReturnsError(t *testing.T) {
	_, err := s2renderErr(t, `{{ 5 | divided_by: 0 }}`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "divided by 0")
}

func TestS2_DividedBy_FloatZeroReturnsError(t *testing.T) {
	_, err := s2renderErr(t, `{{ 5 | divided_by: 0.0 }}`, nil)
	require.Error(t, err)
}

// ── D4. modulo ───────────────────────────────────────────────────────────────

func TestS2_Modulo_Basic(t *testing.T) {
	assert.Equal(t, "1", s2plain(t, `{{ 10 | modulo: 3 }}`))
}

func TestS2_Modulo_Float(t *testing.T) {
	assert.Equal(t, "1.5", s2plain(t, `{{ 7.5 | modulo: 3.0 }}`))
}

func TestS2_Modulo_ZeroReturnsError(t *testing.T) {
	_, err := s2renderErr(t, `{{ 1 | modulo: 0 }}`, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "divided by 0")
}

func TestS2_Modulo_NegativeFloor(t *testing.T) {
	// Ruby uses floor modulo: result has same sign as divisor.
	// truncated would give -1; floor gives 2 (-10 = (-4)*3 + 2)
	assert.Equal(t, "2", s2plain(t, `{{ -10 | modulo: 3 }}`))
	// truncated would give 1; floor gives -2 (10 = (-4)*(-3) + (-2))
	assert.Equal(t, "-2", s2plain(t, `{{ 10 | modulo: -3 }}`))
	// float: truncated would give -1.5; floor gives 1.5
	assert.Equal(t, "1.5", s2plain(t, `{{ -7.5 | modulo: 3.0 }}`))
}

// ── D5. ceil / floor / round ──────────────────────────────────────────────────

func TestS2_Ceil_Float(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 4.3 | ceil }}`))
}

func TestS2_Ceil_Negative(t *testing.T) {
	assert.Equal(t, "-4", s2plain(t, `{{ -4.6 | ceil }}`))
}

func TestS2_Ceil_StringNumber(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ "4.3" | ceil }}`))
}

func TestS2_Ceil_AlreadyInteger(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 5.0 | ceil }}`))
}

func TestS2_Floor_Float(t *testing.T) {
	assert.Equal(t, "4", s2plain(t, `{{ 4.9 | floor }}`))
}

func TestS2_Floor_Negative(t *testing.T) {
	assert.Equal(t, "-5", s2plain(t, `{{ -4.1 | floor }}`))
}

func TestS2_Floor_StringNumber(t *testing.T) {
	assert.Equal(t, "3", s2plain(t, `{{ "3.7" | floor }}`))
}

func TestS2_Round_HalfUp(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 4.6 | round }}`))
}

func TestS2_Round_HalfDown(t *testing.T) {
	assert.Equal(t, "4", s2plain(t, `{{ 4.4 | round }}`))
}

func TestS2_Round_WithPrecision(t *testing.T) {
	assert.Equal(t, "3.14", s2plain(t, `{{ 3.14159 | round: 2 }}`))
}

func TestS2_Round_Negative(t *testing.T) {
	assert.Equal(t, "-5", s2plain(t, `{{ -4.6 | round }}`))
}

// ── D6. at_least / at_most ───────────────────────────────────────────────────

func TestS2_AtLeast_BelowFloor(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 3 | at_least: 5 }}`))
}

func TestS2_AtLeast_AboveFloor(t *testing.T) {
	assert.Equal(t, "8", s2plain(t, `{{ 8 | at_least: 5 }}`))
}

func TestS2_AtLeast_Equal(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 5 | at_least: 5 }}`))
}

func TestS2_AtMost_AboveCeil(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 8 | at_most: 5 }}`))
}

func TestS2_AtMost_BelowCeil(t *testing.T) {
	assert.Equal(t, "3", s2plain(t, `{{ 3 | at_most: 5 }}`))
}

func TestS2_AtMost_Equal(t *testing.T) {
	assert.Equal(t, "5", s2plain(t, `{{ 5 | at_most: 5 }}`))
}

func TestS2_AtLeast_Float(t *testing.T) {
	assert.Equal(t, "3.5", s2plain(t, `{{ 2.5 | at_least: 3.5 }}`))
}

// ── D7. Math in real templates ────────────────────────────────────────────────

func TestS2_Math_InConditional(t *testing.T) {
	// Use math filter result in condition
	s2eq(t, "big", `{% assign v = n | times: 2 %}{% if v > 10 %}big{% else %}small{% endif %}`,
		map[string]any{"n": 6})
}

func TestS2_Math_PriceCalculation(t *testing.T) {
	// Realistic e-commerce calculation: trailing .0 is stripped
	s2eq(t, "90", `{{ price | times: qty | times: discount }}`,
		map[string]any{"price": 10.0, "qty": 10, "discount": 0.9})
}

func TestS2_Math_InForLoop(t *testing.T) {
	// Sum a computed value over loop iterations
	out := s2plain(t, `{% assign total = 0 %}{% for i in (1..5) %}{% assign total = total | plus: i %}{% endfor %}{{ total }}`)
	assert.Equal(t, "15", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// E. Date Filters
// ═════════════════════════════════════════════════════════════════════════════

func TestS2_Date_FromString(t *testing.T) {
	t.Setenv("TZ", "UTC")
	assert.Equal(t, "May", s2plain(t, `{{ "2006-05-05 10:00:00" | date: "%B" }}`))
}

func TestS2_Date_FromUnixTimestamp(t *testing.T) {
	t.Setenv("TZ", "UTC")
	s2eq(t, "07/05/2006", `{{ ts | date: "%m/%d/%Y" }}`,
		map[string]any{"ts": int64(1152098955)})
}

func TestS2_Date_FromTimeTime(t *testing.T) {
	t.Setenv("TZ", "UTC")
	tm, _ := time.Parse(time.RFC3339, "2015-07-17T15:04:05Z")
	s2eq(t, "2015", `{{ ts | date: "%Y" }}`, map[string]any{"ts": tm})
}

func TestS2_Date_NilInputReturnsNil(t *testing.T) {
	// nil | date: fmt → nil → renders as "" — regression guard
	s2eq(t, "", `{{ v | date: "%B" }}`, map[string]any{"v": nil})
}

func TestS2_Date_FormatMonthAndDay(t *testing.T) {
	t.Setenv("TZ", "UTC")
	assert.Equal(t, "07/16/2004", s2plain(t,
		`{{ "Fri Jul 16 01:00:00 2004" | date: "%m/%d/%Y" }}`))
}

func TestS2_Date_FormatYear(t *testing.T) {
	t.Setenv("TZ", "UTC")
	assert.Equal(t, "2006", s2plain(t, `{{ "2006-05-05 10:00:00" | date: "%Y" }}`))
}

func TestS2_DateToString_US(t *testing.T) {
	assert.Equal(t, "07 Nov 2008", s2plain(t,
		`{{ "2008-11-07T13:07:54-08:00" | date_to_string }}`))
}

func TestS2_DateToString_OrdinalUS(t *testing.T) {
	assert.Equal(t, "Nov 7th, 2008", s2plain(t,
		`{{ "2008-11-07T13:07:54-08:00" | date_to_string: "ordinal", "US" }}`))
}

func TestS2_DateToLongString_Basic(t *testing.T) {
	assert.Equal(t, "07 November 2008", s2plain(t,
		`{{ "2008-11-07T13:07:54-08:00" | date_to_long_string }}`))
}

func TestS2_DateToXmlSchema_Basic(t *testing.T) {
	out := s2plain(t, `{{ "2008-11-07T13:07:54-08:00" | date_to_xmlschema }}`)
	assert.Contains(t, out, "2008-11-07")
}

func TestS2_DateToRfc822_Basic(t *testing.T) {
	out := s2plain(t, `{{ "2008-11-07T13:07:54-08:00" | date_to_rfc822 }}`)
	assert.Contains(t, out, "Nov")
	assert.Contains(t, out, "2008")
}

// ═════════════════════════════════════════════════════════════════════════════
// F. Array Filters
// ═════════════════════════════════════════════════════════════════════════════

// ── F1. join ─────────────────────────────────────────────────────────────────

func TestS2_Join_Basic(t *testing.T) {
	s2eq(t, "one two three", `{{ arr | join: " " }}`,
		map[string]any{"arr": []any{"one", "two", "three"}})
}

func TestS2_Join_CommaSep(t *testing.T) {
	s2eq(t, "a,b,c", `{{ arr | join: "," }}`,
		map[string]any{"arr": []any{"a", "b", "c"}})
}

func TestS2_Join_EmptyArray(t *testing.T) {
	s2eq(t, "", `{{ arr | join: " " }}`, map[string]any{"arr": []any{}})
}

func TestS2_Join_SingleElement(t *testing.T) {
	s2eq(t, "solo", `{{ arr | join: "," }}`, map[string]any{"arr": []any{"solo"}})
}

// ── F2. first / last ─────────────────────────────────────────────────────────

func TestS2_First_OnArray(t *testing.T) {
	s2eq(t, "a", `{{ arr | first }}`, map[string]any{"arr": []any{"a", "b", "c"}})
}

func TestS2_Last_OnArray(t *testing.T) {
	s2eq(t, "c", `{{ arr | last }}`, map[string]any{"arr": []any{"a", "b", "c"}})
}

func TestS2_First_OnString(t *testing.T) {
	// first on a string returns the first character — regression guard
	assert.Equal(t, "f", s2plain(t, `{{ "foobar" | first }}`))
}

func TestS2_Last_OnString(t *testing.T) {
	// last on a string returns the last character — regression guard
	assert.Equal(t, "r", s2plain(t, `{{ "foobar" | last }}`))
}

func TestS2_First_Unicode(t *testing.T) {
	assert.Equal(t, "日", s2plain(t, `{{ "日本語" | first }}`))
}

func TestS2_Last_Unicode(t *testing.T) {
	assert.Equal(t, "語", s2plain(t, `{{ "日本語" | last }}`))
}

func TestS2_First_AfterSplit(t *testing.T) {
	assert.Equal(t, "one", s2plain(t, `{{ "one two three" | split: " " | first }}`))
}

func TestS2_Last_AfterSplit(t *testing.T) {
	assert.Equal(t, "three", s2plain(t, `{{ "one two three" | split: " " | last }}`))
}

// ── F3. reverse ───────────────────────────────────────────────────────────────

func TestS2_Reverse_Basic(t *testing.T) {
	s2eq(t, "c b a", `{{ arr | reverse | join: " " }}`,
		map[string]any{"arr": []any{"a", "b", "c"}})
}

func TestS2_Reverse_SingleElement(t *testing.T) {
	s2eq(t, "x", `{{ arr | reverse | join: "" }}`, map[string]any{"arr": []any{"x"}})
}

func TestS2_Reverse_OddLength(t *testing.T) {
	s2eq(t, "5 4 3 2 1", `{{ arr | reverse | join: " " }}`,
		map[string]any{"arr": []any{1, 2, 3, 4, 5}})
}

// ── F4. sort / sort_natural ───────────────────────────────────────────────────

func TestS2_Sort_Strings(t *testing.T) {
	s2eq(t, "apple banana cherry", `{{ arr | sort | join: " " }}`,
		map[string]any{"arr": []any{"cherry", "apple", "banana"}})
}

func TestS2_Sort_Numbers(t *testing.T) {
	s2eq(t, "1 2 3 5", `{{ arr | sort | join: " " }}`,
		map[string]any{"arr": []any{3, 1, 5, 2}})
}

func TestS2_Sort_NilLast(t *testing.T) {
	// nil values go to the end — regression guard
	input := []any{
		map[string]any{"price": 4, "handle": "alpha"},
		map[string]any{"handle": "beta"},
		map[string]any{"price": 1, "handle": "gamma"},
		map[string]any{"handle": "delta"},
		map[string]any{"price": 2, "handle": "epsilon"},
	}
	s2eq(t, "gamma epsilon alpha beta delta",
		`{{ arr | sort: "price" | map: "handle" | join: " " }}`,
		map[string]any{"arr": input})
}

func TestS2_Sort_ByProperty(t *testing.T) {
	input := []any{
		map[string]any{"name": "Zebra"},
		map[string]any{"name": "Apple"},
		map[string]any{"name": "Mango"},
	}
	s2eq(t, "Apple Mango Zebra",
		`{{ arr | sort: "name" | map: "name" | join: " " }}`,
		map[string]any{"arr": input})
}

func TestS2_SortNatural_CaseInsensitive(t *testing.T) {
	s2eq(t, "Apple banana Cherry",
		`{{ arr | sort_natural | join: " " }}`,
		map[string]any{"arr": []any{"Cherry", "Apple", "banana"}})
}

func TestS2_SortNatural_NilLast(t *testing.T) {
	// nil property values must go last — regression guard (no panic)
	input := []any{
		map[string]any{"price": "4", "handle": "alpha"},
		map[string]any{"handle": "beta"},
		map[string]any{"price": "1", "handle": "gamma"},
		map[string]any{"handle": "delta"},
		map[string]any{"price": "2", "handle": "epsilon"},
	}
	s2eq(t, "gamma epsilon alpha beta delta",
		`{{ arr | sort_natural: "price" | map: "handle" | join: " " }}`,
		map[string]any{"arr": input})
}

func TestS2_SortNatural_NilElementsNoParanic(t *testing.T) {
	// Arrays with nil entries must not panic — regression guard
	s2eq(t, "apple cherry",
		`{{ arr | sort_natural | compact | join: " " }}`,
		map[string]any{"arr": []any{nil, "cherry", nil, "apple"}})
}

// ── F5. map ───────────────────────────────────────────────────────────────────

func TestS2_Map_BasicProperty(t *testing.T) {
	input := []any{
		map[string]any{"name": "Alice"},
		map[string]any{"name": "Bob"},
	}
	s2eq(t, "Alice Bob", `{{ arr | map: "name" | join: " " }}`,
		map[string]any{"arr": input})
}

func TestS2_Map_ThenFilter(t *testing.T) {
	input := []any{
		map[string]any{"title": "One", "published": true},
		map[string]any{"title": "Two", "published": false},
		map[string]any{"title": "Three", "published": true},
	}
	s2eq(t, "One Three",
		`{{ posts | where: "published", true | map: "title" | join: " " }}`,
		map[string]any{"posts": input})
}

// ── F6. sum ───────────────────────────────────────────────────────────────────

func TestS2_Sum_Integers(t *testing.T) {
	s2eq(t, "10", `{{ arr | sum }}`, map[string]any{"arr": []any{1, 2, 3, 4}})
}

func TestS2_Sum_ByProperty(t *testing.T) {
	input := []any{
		map[string]any{"qty": 3},
		map[string]any{"qty": 7},
	}
	s2eq(t, "10", `{{ arr | sum: "qty" }}`, map[string]any{"arr": input})
}

func TestS2_Sum_MixedStringNumbers(t *testing.T) {
	s2eq(t, "10", `{{ arr | sum }}`, map[string]any{"arr": []any{1, 2, "3", "4"}})
}

func TestS2_Sum_EmptyArray(t *testing.T) {
	s2eq(t, "0", `{{ arr | sum }}`, map[string]any{"arr": []any{}})
}

// ── F7. compact / uniq ────────────────────────────────────────────────────────

func TestS2_Compact_RemovesNils(t *testing.T) {
	s2eq(t, "1 2 3", `{{ arr | compact | join: " " }}`,
		map[string]any{"arr": []any{1, nil, 2, nil, 3}})
}

func TestS2_Compact_EmptyArray(t *testing.T) {
	s2eq(t, "", `{{ arr | compact | join: " " }}`, map[string]any{"arr": []any{}})
}

func TestS2_Compact_NoNils(t *testing.T) {
	s2eq(t, "a b c", `{{ arr | compact | join: " " }}`,
		map[string]any{"arr": []any{"a", "b", "c"}})
}

func TestS2_Uniq_Basic(t *testing.T) {
	s2eq(t, "1 2 3", `{{ arr | uniq | join: " " }}`,
		map[string]any{"arr": []any{1, 1, 2, 3, 2, 1}})
}

func TestS2_Uniq_PreservesOrder(t *testing.T) {
	s2eq(t, "c a b", `{{ arr | uniq | join: " " }}`,
		map[string]any{"arr": []any{"c", "a", "c", "b", "a"}})
}

func TestS2_Uniq_EmptyArray(t *testing.T) {
	s2eq(t, "0", `{{ arr | uniq | size }}`, map[string]any{"arr": []any{}})
}

// ── F8. concat ────────────────────────────────────────────────────────────────

func TestS2_Concat_Basic(t *testing.T) {
	s2eq(t, "1 2 3 4", `{{ a | concat: b | join: " " }}`,
		map[string]any{"a": []any{1, 2}, "b": []any{3, 4}})
}

func TestS2_Concat_OriginalUnchanged(t *testing.T) {
	// concat is pure — original array not mutated
	s2eq(t, "2", `{{ a | size }}`,
		map[string]any{"a": []any{1, 2}, "b": []any{3, 4}})
}

func TestS2_Concat_EmptyLeft(t *testing.T) {
	s2eq(t, "3 4", `{{ arr | concat: extra | join: " " }}`,
		map[string]any{"arr": []any{}, "extra": []any{3, 4}})
}

// ── F9. push / pop / unshift / shift ─────────────────────────────────────────

func TestS2_Push_ReturnsNewArray(t *testing.T) {
	s2eq(t, "5", `{{ arr | push: "new" | size }}`,
		map[string]any{"arr": []any{"a", "b", "c", "d"}})
}

func TestS2_Push_OriginalUnchanged(t *testing.T) {
	s2eq(t, "4", `{{ arr | size }}`,
		map[string]any{"arr": []any{"a", "b", "c", "d"}})
}

func TestS2_Pop_ReturnsNewArray(t *testing.T) {
	s2eq(t, "3", `{{ arr | pop | size }}`,
		map[string]any{"arr": []any{"a", "b", "c", "d"}})
}

func TestS2_Unshift_PrependElement(t *testing.T) {
	s2eq(t, "new", `{{ arr | unshift: "new" | first }}`,
		map[string]any{"arr": []any{"a", "b"}})
}

func TestS2_Shift_RemovesFirst(t *testing.T) {
	s2eq(t, "b", `{{ arr | shift | first }}`,
		map[string]any{"arr": []any{"a", "b", "c"}})
}

// ── F10. where / reject ────────────────────────────────────────────────────────

func TestS2_Where_ByValue(t *testing.T) {
	products := []any{
		map[string]any{"title": "A", "type": "kitchen"},
		map[string]any{"title": "B", "type": "living"},
		map[string]any{"title": "C", "type": "kitchen"},
	}
	s2eq(t, "A C",
		`{{ products | where: "type", "kitchen" | map: "title" | join: " " }}`,
		map[string]any{"products": products})
}

func TestS2_Where_TruthyProperty(t *testing.T) {
	items := []any{
		map[string]any{"name": "A", "available": true},
		map[string]any{"name": "B"},
		map[string]any{"name": "C", "available": true},
	}
	s2eq(t, "A C",
		`{{ items | where: "available" | map: "name" | join: " " }}`,
		map[string]any{"items": items})
}

func TestS2_Reject_Basic(t *testing.T) {
	products := []any{
		map[string]any{"title": "A", "type": "kitchen"},
		map[string]any{"title": "B", "type": "living"},
		map[string]any{"title": "C", "type": "kitchen"},
	}
	s2eq(t, "B",
		`{{ products | reject: "type", "kitchen" | map: "title" | join: " " }}`,
		map[string]any{"products": products})
}

// ── F11. find / find_index / has ─────────────────────────────────────────────

func TestS2_Find_Basic(t *testing.T) {
	items := []any{
		map[string]any{"id": 1, "name": "foo"},
		map[string]any{"id": 2, "name": "bar"},
	}
	// find returns the matching map; access its field via assign + map access
	s2eq(t, "bar", `{% assign f = items | find: "id", 2 %}{{ f.name }}`,
		map[string]any{"items": items})
}

func TestS2_Find_NotFound(t *testing.T) {
	items := []any{map[string]any{"id": 1}}
	s2eq(t, "", `{% assign f = items | find: "id", 99 %}{{ f }}`,
		map[string]any{"items": items})
}

func TestS2_FindIndex_Basic(t *testing.T) {
	items := []any{
		map[string]any{"id": 1},
		map[string]any{"id": 2},
		map[string]any{"id": 3},
	}
	s2eq(t, "1", `{{ items | find_index: "id", 2 }}`,
		map[string]any{"items": items})
}

func TestS2_FindIndex_NotFound(t *testing.T) {
	// find_index returns nil when not found; nil renders as ""
	items := []any{map[string]any{"id": 1}}
	s2eq(t, "", `{{ items | find_index: "id", 99 }}`,
		map[string]any{"items": items})
}

func TestS2_Has_True(t *testing.T) {
	items := []any{
		map[string]any{"id": 1},
		map[string]any{"id": 2},
	}
	s2eq(t, "true", `{{ items | has: "id", 2 }}`, map[string]any{"items": items})
}

func TestS2_Has_False(t *testing.T) {
	items := []any{map[string]any{"id": 1}}
	s2eq(t, "false", `{{ items | has: "id", 99 }}`, map[string]any{"items": items})
}

// ── F12. group_by ─────────────────────────────────────────────────────────────

func TestS2_GroupBy_Basic(t *testing.T) {
	items := []any{
		map[string]any{"name": "A", "type": "x"},
		map[string]any{"name": "B", "type": "y"},
		map[string]any{"name": "C", "type": "x"},
	}
	// group_by returns array of {name, items} maps
	out := s2render(t, `{% assign g = items | group_by: "type" %}{{ g | size }}`,
		map[string]any{"items": items})
	assert.Equal(t, "2", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// G. Misc Filters
// ═════════════════════════════════════════════════════════════════════════════

// ── G1. default ──────────────────────────────────────────────────────────────

func TestS2_Default_Nil(t *testing.T) {
	s2eq(t, "fallback", `{{ v | default: "fallback" }}`, map[string]any{"v": nil})
}

func TestS2_Default_False(t *testing.T) {
	s2eq(t, "fallback", `{{ v | default: "fallback" }}`, map[string]any{"v": false})
}

func TestS2_Default_EmptyString(t *testing.T) {
	s2eq(t, "fallback", `{{ v | default: "fallback" }}`, map[string]any{"v": ""})
}

func TestS2_Default_EmptyArray(t *testing.T) {
	s2eq(t, "fallback", `{{ v | default: "fallback" }}`, map[string]any{"v": []any{}})
}

func TestS2_Default_TruthyString(t *testing.T) {
	s2eq(t, "hello", `{{ v | default: "fallback" }}`, map[string]any{"v": "hello"})
}

func TestS2_Default_ZeroIsNotDefault(t *testing.T) {
	// 0 is truthy in Liquid — default must NOT activate
	s2eq(t, "0", `{{ v | default: "fallback" }}`, map[string]any{"v": 0})
}

func TestS2_Default_AllowFalse(t *testing.T) {
	// allow_false: true → false does NOT trigger default
	s2eq(t, "false", `{{ v | default: "fallback", allow_false: true }}`,
		map[string]any{"v": false})
}

func TestS2_Default_AllowFalse_NilStillTriggersDefault(t *testing.T) {
	// allow_false: true → nil still triggers default
	s2eq(t, "fallback", `{{ v | default: "fallback", allow_false: true }}`,
		map[string]any{"v": nil})
}

func TestS2_Default_Float(t *testing.T) {
	s2eq(t, "4.99", `{{ v | default: 2.99 }}`, map[string]any{"v": 4.99})
}

// ── G2. json / jsonify / to_integer ──────────────────────────────────────────

func TestS2_JSON_String(t *testing.T) {
	assert.Equal(t, `"hello"`, s2plain(t, `{{ "hello" | json }}`))
}

func TestS2_JSON_Integer(t *testing.T) {
	assert.Equal(t, "42", s2plain(t, `{{ 42 | json }}`))
}

func TestS2_JSON_Bool(t *testing.T) {
	assert.Equal(t, "true", s2plain(t, `{{ true | json }}`))
}

func TestS2_JSON_Array(t *testing.T) {
	s2eq(t, `[1,2,3]`, `{{ arr | json }}`, map[string]any{"arr": []any{1, 2, 3}})
}

func TestS2_ToInteger_FloatString(t *testing.T) {
	assert.Equal(t, "3", s2plain(t, `{{ "3.9" | to_integer }}`))
}

func TestS2_ToInteger_IntString(t *testing.T) {
	assert.Equal(t, "42", s2plain(t, `{{ "42" | to_integer }}`))
}

func TestS2_ToInteger_TrueIsOne(t *testing.T) {
	assert.Equal(t, "1", s2plain(t, `{{ true | to_integer }}`))
}

func TestS2_ToInteger_FalseIsZero(t *testing.T) {
	assert.Equal(t, "0", s2plain(t, `{{ false | to_integer }}`))
}

// ═════════════════════════════════════════════════════════════════════════════
// H. Filter Chaining
// ═════════════════════════════════════════════════════════════════════════════

func TestS2_Chain_SplitReverseJoin(t *testing.T) {
	assert.Equal(t, "c,b,a", s2plain(t, `{{ "a,b,c" | split: "," | reverse | join: "," }}`))
}

func TestS2_Chain_DowncaseTruncate(t *testing.T) {
	assert.Equal(t, "hello...", s2plain(t, `{{ "HELLO WORLD" | downcase | truncate: 8 }}`))
}

func TestS2_Chain_StripHtmlDowncaseStrip(t *testing.T) {
	assert.Equal(t, "hello world", s2plain(t,
		`{{ "  <b>Hello</b> World  " | strip_html | downcase | strip }}`))
}

func TestS2_Chain_MathChain(t *testing.T) {
	// 3 | times: 4 | minus: 2 | divided_by: 2 = (12 - 2) / 2 = 5
	assert.Equal(t, "5", s2plain(t, `{{ 3 | times: 4 | minus: 2 | divided_by: 2 }}`))
}

func TestS2_Chain_ArrayChain(t *testing.T) {
	products := []any{
		map[string]any{"title": "Tomato", "type": "fruit"},
		map[string]any{"title": "Banana", "type": "fruit"},
		map[string]any{"title": "Carrot", "type": "vegetable"},
		map[string]any{"title": "Apple", "type": "fruit"},
	}
	s2eq(t, "Apple Banana Tomato",
		`{{ products | where: "type", "fruit" | map: "title" | sort | join: " " }}`,
		map[string]any{"products": products})
}

func TestS2_Chain_InForLoop(t *testing.T) {
	s2eq(t, "<A><B><C>",
		`{% for w in s | split: "," %}<{{ w | upcase }}>{% endfor %}`,
		map[string]any{"s": "a,b,c"})
}

// ═════════════════════════════════════════════════════════════════════════════
// I. Nil Safety
// ═════════════════════════════════════════════════════════════════════════════

func TestS2_Nil_DowncaseEmpty(t *testing.T) {
	s2eq(t, "", `{{ v | downcase }}`, map[string]any{"v": nil})
}

func TestS2_Nil_AppendEmpty(t *testing.T) {
	s2eq(t, "!!", `{{ v | append: "!!" }}`, map[string]any{"v": nil})
}

func TestS2_Nil_SizeZero(t *testing.T) {
	s2eq(t, "0", `{{ v | size }}`, map[string]any{"v": nil})
}

func TestS2_Nil_StripEmpty(t *testing.T) {
	s2eq(t, "", `{{ v | strip }}`, map[string]any{"v": nil})
}

func TestS2_Nil_ReverseEmpty(t *testing.T) {
	s2eq(t, "", `{{ v | reverse | join: "" }}`, map[string]any{"v": nil})
}

func TestS2_Nil_URLEncodeEmpty(t *testing.T) {
	s2eq(t, "", `{{ v | url_encode }}`, map[string]any{"v": nil})
}

func TestS2_Nil_DateNil(t *testing.T) {
	// nil date filter → output empty string — regression guard
	s2eq(t, "", `{{ v | date: "%B" }}`, map[string]any{"v": nil})
}

func TestS2_Nil_JoinEmpty(t *testing.T) {
	s2eq(t, "", `{{ v | join: "," }}`, map[string]any{"v": nil})
}

// ═════════════════════════════════════════════════════════════════════════════
// J. Regression Guards — exact behaviors of bugs fixed in this session
// ═════════════════════════════════════════════════════════════════════════════

// J1. truncate: n <= len(ellipsis) returns the ellipsis
func TestS2_Regression_Truncate_ZeroN_ReturnsEllipsis(t *testing.T) {
	assert.Equal(t, "...", s2plain(t, `{{ "1234567890" | truncate: 0 }}`))
}

func TestS2_Regression_Truncate_SmallN_ReturnsEllipsis(t *testing.T) {
	assert.Equal(t, "...", s2plain(t, `{{ "1234567890" | truncate: 2 }}`))
}

// J2. truncate: exact-fit string is NOT truncated
func TestS2_Regression_Truncate_ExactFit_NoEllipsis(t *testing.T) {
	assert.Equal(t, "hello", s2plain(t, `{{ "hello" | truncate: 5 }}`))
}

// J3. truncatewords: n=0 behaves like n=1
func TestS2_Regression_TruncateWords_ZeroN_KeepsFirstWord(t *testing.T) {
	assert.Equal(t, "one...", s2plain(t, `{{ "one two three" | truncatewords: 0 }}`))
}

// J4. truncatewords: fewer words than n → no ellipsis added
func TestS2_Regression_TruncateWords_FewerWords_NoEllipsis(t *testing.T) {
	assert.Equal(t, "one two", s2plain(t, `{{ "one two" | truncatewords: 5 }}`))
}

// J5. divided_by: float / int = float (not integer floor division)
func TestS2_Regression_DividedBy_FloatDividend_FloatResult(t *testing.T) {
	s2eq(t, "0.5", `{{ n | divided_by: 4 }}`, map[string]any{"n": float64(2.0)})
}

// J6. divided_by: int / int = floor (remains integer division)
func TestS2_Regression_DividedBy_IntDividend_IntResult(t *testing.T) {
	assert.Equal(t, "3", s2plain(t, `{{ 10 | divided_by: 3 }}`))
}

// J7. strip_newlines removes \r\n (Windows line endings)
func TestS2_Regression_StripNewlines_CRLF(t *testing.T) {
	s2eq(t, "ab", `{{ s | strip_newlines }}`, map[string]any{"s": "a\r\nb"})
}

// J8. newline_to_br normalizes \r\n → single <br />
func TestS2_Regression_NewlineToBr_CRLF_NoDuplicate(t *testing.T) {
	s2eq(t, "a<br />\nb", `{{ s | newline_to_br }}`, map[string]any{"s": "a\r\nb"})
}

// J9. first/last on strings return first/last rune
func TestS2_Regression_First_OnString(t *testing.T) {
	assert.Equal(t, "h", s2plain(t, `{{ "hello" | first }}`))
}

func TestS2_Regression_Last_OnString(t *testing.T) {
	assert.Equal(t, "o", s2plain(t, `{{ "hello" | last }}`))
}

// J10. sort: nil values go last (not first)
func TestS2_Regression_Sort_NilLast(t *testing.T) {
	arr := []any{3, nil, 1, nil, 2}
	// After sort, nils should be at the end
	out := s2render(t, `{{ arr | sort | last }}`, map[string]any{"arr": arr})
	assert.Equal(t, "", out) // nil renders as ""
}

// J11. sort_natural: nil elements in array must not cause panic
func TestS2_Regression_SortNatural_NilElements_NoPanic(t *testing.T) {
	arr := []any{nil, "banana", nil, "apple", "cherry"}
	// Must not panic; nils go last
	out := s2render(t, `{{ arr | sort_natural | first }}`, map[string]any{"arr": arr})
	assert.Equal(t, "apple", out)
}

// J12. slice: negative length clamps to zero (no panic)
func TestS2_Regression_Slice_NegativeLength_Empty(t *testing.T) {
	assert.Equal(t, "", s2plain(t, `{{ "foobar" | slice: 0, -1 }}`))
}

// J13. date: nil input returns nil (renders as "")
func TestS2_Regression_Date_NilInput(t *testing.T) {
	s2eq(t, "", `{{ v | date: "%Y" }}`, map[string]any{"v": nil})
}

// J14. truncatewords: internal whitespace is normalized to single spaces
func TestS2_Regression_TruncateWords_InternalWhitespaceNormalized(t *testing.T) {
	s2eq(t, "one two three...", `{{ s | truncatewords: 3 }}`,
		map[string]any{"s": "one  two\tthree\nfour"})
}
