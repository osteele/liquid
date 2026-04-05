package liquid

// Ported whitespace-control tests from:
//   - Ruby Liquid: test/integration/trim_mode_test.rb
//   - LiquidJS:    test/integration/liquid/whitespace-ctrl.spec.ts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// renderWS is a helper that parses and renders the template with no bindings.
func renderWS(t *testing.T, src string) string {
	t.Helper()
	eng := NewEngine()
	out, err := eng.ParseAndRenderString(src, nil)
	require.NoError(t, err)
	return out
}

// ── Ruby Liquid: trim_mode_test.rb ──────────────────────────────────────────

// test_standard_output – whitespace trimming must NOT alter standard {{ }} output.
func TestWhitespaceCtrl_StandardOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{ 'John' }}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          John\n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_variable_output_with_multiple_blank_lines – {{- -}} collapses surrounding blank lines.
func TestWhitespaceCtrl_VariableOutputMultipleBlankLines(t *testing.T) {
	text := "      <div>\n        <p>\n\n\n          {{- 'John' -}}\n\n\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>John</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_tag_output_with_multiple_blank_lines – {%- if -%} and {%- endif -%} collapse blank lines.
func TestWhitespaceCtrl_TagOutputMultipleBlankLines(t *testing.T) {
	text := "      <div>\n        <p>\n\n\n          {%- if true -%}\n          yes\n          {%- endif -%}\n\n\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>yes</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_standard_tags – standard {% %} tags must NOT trim surrounding whitespace.
func TestWhitespaceCtrl_StandardTags_TrueCondition(t *testing.T) {
	text := "      <div>\n        <p>\n          {% if true %}\n          yes\n          {% endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          \n          yes\n          \n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_StandardTags_FalseCondition(t *testing.T) {
	text := "      <div>\n        <p>\n          {% if false %}\n          no\n          {% endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          \n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_no_trim_output – {{- -}} with no surrounding whitespace leaves adjacent chars intact.
func TestWhitespaceCtrl_NoTrimOutput(t *testing.T) {
	require.Equal(t, "<p>John</p>", renderWS(t, "<p>{{- 'John' -}}</p>"))
}

// test_no_trim_tags – {%- -%} with no surrounding whitespace.
func TestWhitespaceCtrl_NoTrimTags_True(t *testing.T) {
	require.Equal(t, "<p>yes</p>", renderWS(t, "<p>{%- if true -%}yes{%- endif -%}</p>"))
}

func TestWhitespaceCtrl_NoTrimTags_False(t *testing.T) {
	require.Equal(t, "<p></p>", renderWS(t, "<p>{%- if false -%}no{%- endif -%}</p>"))
}

// test_single_line_outer_tag – left trim on open, right trim on close.
func TestWhitespaceCtrl_SingleLineOuterTag_True(t *testing.T) {
	require.Equal(t, "<p> yes </p>", renderWS(t, "<p> {%- if true %} yes {% endif -%} </p>"))
}

func TestWhitespaceCtrl_SingleLineOuterTag_False(t *testing.T) {
	require.Equal(t, "<p></p>", renderWS(t, "<p> {%- if false %} no {% endif -%} </p>"))
}

// test_single_line_inner_tag – right trim on open (consumes inner leading space), left trim on close.
func TestWhitespaceCtrl_SingleLineInnerTag_True(t *testing.T) {
	require.Equal(t, "<p> yes </p>", renderWS(t, "<p> {% if true -%} yes {%- endif %} </p>"))
}

func TestWhitespaceCtrl_SingleLineInnerTag_False(t *testing.T) {
	require.Equal(t, "<p>  </p>", renderWS(t, "<p> {% if false -%} no {%- endif %} </p>"))
}

// test_single_line_post_tag – right trim on both open and close.
func TestWhitespaceCtrl_SingleLinePostTag_True(t *testing.T) {
	require.Equal(t, "<p> yes </p>", renderWS(t, "<p> {% if true -%} yes {% endif -%} </p>"))
}

func TestWhitespaceCtrl_SingleLinePostTag_False(t *testing.T) {
	require.Equal(t, "<p> </p>", renderWS(t, "<p> {% if false -%} no {% endif -%} </p>"))
}

// test_single_line_pre_tag – left trim on both open and close.
func TestWhitespaceCtrl_SingleLinePreTag_True(t *testing.T) {
	require.Equal(t, "<p> yes </p>", renderWS(t, "<p> {%- if true %} yes {%- endif %} </p>"))
}

func TestWhitespaceCtrl_SingleLinePreTag_False(t *testing.T) {
	require.Equal(t, "<p> </p>", renderWS(t, "<p> {%- if false %} no {%- endif %} </p>"))
}

// test_pre_trim_output – left-trim only on {{ output.
func TestWhitespaceCtrl_PreTrimOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{- 'John' }}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>John\n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_pre_trim_tags – left-trim on if tag, left-trim on endif.
func TestWhitespaceCtrl_PreTrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if true %}\n          yes\n          {%- endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          yes\n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_PreTrimTags_False(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if false %}\n          no\n          {%- endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_post_trim_output – right-trim only on {{ output.
func TestWhitespaceCtrl_PostTrimOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{ 'John' -}}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          John</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_post_trim_tags – right-trim on if and endif.
func TestWhitespaceCtrl_PostTrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {% if true -%}\n          yes\n          {% endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          yes\n          </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_PostTrimTags_False(t *testing.T) {
	text := "      <div>\n        <p>\n          {% if false -%}\n          no\n          {% endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_pre_and_post_trim_tags – left trim on if, right trim on endif.
func TestWhitespaceCtrl_PreAndPostTrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if true %}\n          yes\n          {% endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          yes\n          </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_PreAndPostTrimTags_False(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if false %}\n          no\n          {% endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p></p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_post_and_pre_trim_tags – right trim on if, left trim on endif.
func TestWhitespaceCtrl_PostAndPreTrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {% if true -%}\n          yes\n          {%- endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          yes\n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_PostAndPreTrimTags_False(t *testing.T) {
	// Ruby: the space before {%- endif %} (from "          ") is preserved since it follows the
	// "no" branch which wasn't rendered (false condition). The trim on {%- doesn't trim "nothing".
	// Result: "          \n" (the whitespace that was between {% if false -%} and {%- endif %})
	text := "      <div>\n        <p>\n          {% if false -%}\n          no\n          {%- endif %}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>\n          \n        </p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_trim_output – {{- -}} trims both sides.
func TestWhitespaceCtrl_TrimOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{- 'John' -}}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>John</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_trim_tags – {%- if -%} and {%- endif -%} trim both sides.
func TestWhitespaceCtrl_TrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if true -%}\n          yes\n          {%- endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>yes</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_TrimTags_False(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if false -%}\n          no\n          {%- endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p></p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_whitespace_trim_output – adjacent {{- -}} expressions separated by a comma.
func TestWhitespaceCtrl_WhitespaceTrimOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{- 'John' -}},\n          {{- '30' -}}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>John,30</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_whitespace_trim_tags – adjacent {%- if -%} blocks.
func TestWhitespaceCtrl_WhitespaceTrimTags_True(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if true -%}\n          yes\n          {%- endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p>yes</p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

func TestWhitespaceCtrl_WhitespaceTrimTags_False(t *testing.T) {
	text := "      <div>\n        <p>\n          {%- if false -%}\n          no\n          {%- endif -%}\n        </p>\n      </div>\n    "
	want := "      <div>\n        <p></p>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_complex_trim_output – mix of directions in a single block.
func TestWhitespaceCtrl_ComplexTrimOutput(t *testing.T) {
	text := "      <div>\n        <p>\n          {{- 'John' -}}\n          {{- '30' -}}\n        </p>\n        <b>\n          {{ 'John' -}}\n          {{- '30' }}\n        </b>\n        <i>\n          {{- 'John' }}\n          {{ '30' -}}\n        </i>\n      </div>\n    "
	want := "      <div>\n        <p>John30</p>\n        <b>\n          John30\n        </b>\n        <i>John\n          30</i>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_complex_trim – deeply nested {%- if -%} blocks.
func TestWhitespaceCtrl_ComplexTrim(t *testing.T) {
	text := "      <div>\n        {%- if true -%}\n          {%- if true -%}\n            <p>\n              {{- 'John' -}}\n            </p>\n          {%- endif -%}\n        {%- endif -%}\n      </div>\n    "
	want := "      <div><p>John</p></div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_right_trim_followed_by_tag – right-trim on first output, immediately followed by second.
func TestWhitespaceCtrl_RightTrimFollowedByTag(t *testing.T) {
	require.Equal(t, "ab c", renderWS(t, `{{ "a" -}}{{ "b" }} c`))
}

// test_raw_output – trim markers inside {% raw %} must be emitted verbatim (not applied).
func TestWhitespaceCtrl_RawOutput(t *testing.T) {
	text := "      <div>\n        {% raw %}\n          {%- if true -%}\n            <p>\n              {{- 'John' -}}\n            </p>\n          {%- endif -%}\n        {% endraw %}\n      </div>\n    "
	want := "      <div>\n        \n          {%- if true -%}\n            <p>\n              {{- 'John' -}}\n            </p>\n          {%- endif -%}\n        \n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// test_pre_trim_blank_preceding_text – {%- trims the newline (and spaces) before it.
func TestWhitespaceCtrl_PreTrimBlankPrecedingText_RawBlock(t *testing.T) {
	require.Equal(t, "", renderWS(t, "\n{%- raw %}{% endraw %}"))
}

func TestWhitespaceCtrl_PreTrimBlankPrecedingText_IfBlock(t *testing.T) {
	require.Equal(t, "", renderWS(t, "\n{%- if true %}{% endif %}"))
}

func TestWhitespaceCtrl_PreTrimBlankPrecedingText_WithContent(t *testing.T) {
	require.Equal(t, "BC", renderWS(t, "{{ 'B' }} \n{%- if true %}C{% endif %}"))
}

// ── LiquidJS: whitespace-ctrl.spec.ts ───────────────────────────────────────

// From the JS spec, a case that combines {{- -}} with comma and adjacent output.
func TestWhitespaceCtrl_JS_AdjacentTrimmedOutputs(t *testing.T) {
	text := "      <div>\n        <p>\n          {{- 'John' -}}\n          {{- '30' -}}\n        </p>\n        <b>\n          {{ 'John' -}}\n          {{- '30' }}\n        </b>\n        <i>\n          {{- 'John' }}\n          {{ '30' -}}\n        </i>\n      </div>\n    "
	want := "      <div>\n        <p>John30</p>\n        <b>\n          John30\n        </b>\n        <i>John\n          30</i>\n      </div>\n    "
	require.Equal(t, want, renderWS(t, text))
}

// From JS spec: markup-based trim works the same as inline.
func TestWhitespaceCtrl_JS_MarkupTrim(t *testing.T) {
	src := "{%- assign username = \"John G. Chalmers-Smith\" -%}\n{%- if username and username.size > 10 -%}\n  Wow, {{ username }}, you have a long name!\n{%- else -%}\n  Hello there!\n{%- endif -%}"
	want := "Wow, John G. Chalmers-Smith, you have a long name!"
	require.Equal(t, want, renderWS(t, src))
}

// From JS spec: no trim when not specified.
func TestWhitespaceCtrl_JS_NoTrimWhenNotSpecified(t *testing.T) {
	src := "{% assign username = \"John G. Chalmers-Smith\" %}\n{% if username and username.size > 10 %}\n  Wow, {{ username }}, you have a long name!\n{% else %}\n  Hello there!\n{% endif %}"
	want := "\n\n  Wow, John G. Chalmers-Smith, you have a long name!\n"
	require.Equal(t, want, renderWS(t, src))
}

// ── LiquidJS: trimming.spec.ts — global trim options ────────────────────────

// tag trimming: TrimTagLeft trims whitespace before every {% %} and block open/close.
func TestWhitespaceCtrl_TrimTagLeft(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimTagLeft(true)
	out, err := eng.ParseAndRenderString(" \n \t{%if true%}foo{%endif%} ", nil)
	require.NoError(t, err)
	require.Equal(t, "foo ", out)
}

// tag trimming: TrimTagRight trims whitespace after every {% %} and block open/close.
func TestWhitespaceCtrl_TrimTagRight(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimTagRight(true)
	out, err := eng.ParseAndRenderString("\t{%if true%}foo{%endif%} \n", nil)
	require.NoError(t, err)
	require.Equal(t, "\tfoo", out)
}

// tag trimming: TrimTagLeft+TrimTagRight must NOT trim {{ output }} expressions.
func TestWhitespaceCtrl_TrimTagBoth_NoTrimOutput(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimTagLeft(true)
	eng.SetTrimTagRight(true)
	out, err := eng.ParseAndRenderString("{%if true%}a {{name}} b{%endif%}", map[string]any{"name": "harttle"})
	require.NoError(t, err)
	require.Equal(t, "a harttle b", out)
}

// value trimming: TrimOutputLeft trims whitespace before every {{ output }}.
func TestWhitespaceCtrl_TrimOutputLeft(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimOutputLeft(true)
	out, err := eng.ParseAndRenderString(" \n \t{{name}} ", map[string]any{"name": "harttle"})
	require.NoError(t, err)
	require.Equal(t, "harttle ", out)
}

// value trimming: TrimOutputRight trims whitespace after every {{ output }}.
func TestWhitespaceCtrl_TrimOutputRight(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimOutputRight(true)
	out, err := eng.ParseAndRenderString(" \n \t{{name}} ", map[string]any{"name": "harttle"})
	require.NoError(t, err)
	require.Equal(t, " \n \tharttle", out)
}

// value trimming: TrimOutputLeft+TrimOutputRight must NOT trim {% tag %} blocks.
func TestWhitespaceCtrl_TrimOutputBoth_NoTrimTag(t *testing.T) {
	eng := NewEngine()
	eng.SetTrimOutputLeft(true)
	eng.SetTrimOutputRight(true)
	out, err := eng.ParseAndRenderString("\t{% if true %} aha {%endif%}\t", nil)
	require.NoError(t, err)
	require.Equal(t, "\t aha \t", out)
}

// greedy: default (true) — all consecutive whitespace/newlines are trimmed.
func TestWhitespaceCtrl_Greedy_Default(t *testing.T) {
	eng := NewEngine()
	// greedy is true by default; explicit inline markers should trim all whitespace
	src := "\n {%-if true-%}\n a \n{{-name-}}{%-endif-%}\n "
	out, err := eng.ParseAndRenderString(src, map[string]any{"name": "harttle"})
	require.NoError(t, err)
	require.Equal(t, "aharttle", out)
}

// greedy: false — only inline blanks (space/tab) + at most one newline are trimmed.
func TestWhitespaceCtrl_Greedy_False(t *testing.T) {
	eng := NewEngine()
	eng.SetGreedy(false)
	src := "\n {%-if true-%}\n a \n{{-name-}}{%-endif-%}\n "
	out, err := eng.ParseAndRenderString(src, map[string]any{"name": "harttle"})
	require.NoError(t, err)
	require.Equal(t, "\n a \nharttle ", out)
}
