package liquid_test

// s11_whitespace_e2e_test.go — Intensive E2E tests for Section 11: Whitespace Control
//
// Coverage matrix:
//   A. Inline trim markers: {%- -%} and {{- -}} in every meaningful direction and context
//   B. {{-}} trim-blank (empty expression with trim marker) — regression guard for the fix
//   C. Global trim options: TrimTagLeft, TrimTagRight, TrimOutputLeft, TrimOutputRight
//   D. Greedy vs. non-greedy trim semantics
//   E. Interaction: inline markers + global options (must not double-apply)
//   F. All tag types with inline trim: for, if, unless, case, assign, capture, liquid, raw, comment
//   G. Edge cases: empty output, multi-line, adjacent markers, strings with whitespace

import (
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func wsEngine(t *testing.T, opts ...func(*liquid.Engine)) *liquid.Engine {
	t.Helper()
	eng := liquid.NewEngine()
	for _, o := range opts {
		o(eng)
	}
	return eng
}

func wsRender(t *testing.T, eng *liquid.Engine, tpl string, binds map[string]any) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(tpl, binds)
	require.NoError(t, err, "template: %q", tpl)
	return out
}

func wsRenderPlain(t *testing.T, tpl string) string {
	t.Helper()
	return wsRender(t, wsEngine(t), tpl, nil)
}

// ─────────────────────────────────────────────────────────────────────────────
// A. Inline trim markers — all four combinations on tags
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Inline_Tag_NoTrim_PreservesAll(t *testing.T) {
	// {% if %}...{% endif %} preserves surrounding whitespace completely
	got := wsRenderPlain(t, " \n {% if true %} yes {% endif %} \n ")
	require.Equal(t, " \n  yes  \n ", got)
}

func TestS11_Inline_Tag_TrimLeft_OnOpen(t *testing.T) {
	// {%- if %}: trims whitespace to the LEFT of the opening tag
	got := wsRenderPlain(t, " \n {%- if true %} yes {% endif %} ")
	require.Equal(t, " yes  ", got)
}

func TestS11_Inline_Tag_TrimRight_OnOpen(t *testing.T) {
	// {% if -%}: trims whitespace to the RIGHT of the opening tag.
	// -%} consumes the " " between the tag and "yes"; outer " " (before the {%if%}) is kept.
	require.Equal(t, " yes  ", wsRenderPlain(t, " {% if true -%} yes {% endif %} "))
}

func TestS11_Inline_Tag_TrimRight_OnOpen_Correct(t *testing.T) {
	// {% if true -%} eats " yes " up to the next non-whitespace — NO, it trims the
	// whitespace text node that follows the tag, not the content. " yes " is only
	// whitespace before the literal "yes" text — so only the space after -%} is eaten.
	got := wsRenderPlain(t, "<p>{% if true -%} yes {%- endif %}</p>")
	require.Equal(t, "<p>yes</p>", got)
}

func TestS11_Inline_Tag_TrimLeft_OnClose(t *testing.T) {
	// {%- endif %}: trims whitespace to the LEFT of the closing tag
	got := wsRenderPlain(t, "<p>{% if true %} yes {%- endif %}</p>")
	require.Equal(t, "<p> yes</p>", got)
}

func TestS11_Inline_Tag_TrimRight_OnClose(t *testing.T) {
	// {% endif -%}: trims whitespace to the RIGHT of the closing tag
	got := wsRenderPlain(t, "<p>{% if true %} yes {% endif -%} </p>")
	require.Equal(t, "<p> yes </p>", got)
}

func TestS11_Inline_Tag_TrimBoth_CollapseAll(t *testing.T) {
	// {%- if -%}...{%- endif -%}: no surrounding whitespace survives
	got := wsRenderPlain(t, "  {%- if true -%}  yes  {%- endif -%}  ")
	require.Equal(t, "yes", got)
}

func TestS11_Inline_Tag_TrimBoth_FalseBranch_EmitsNothing(t *testing.T) {
	// false branch: nothing rendered — surrounding ws is still consumed
	got := wsRenderPlain(t, "  {%- if false -%}  no  {%- endif -%}  ")
	require.Equal(t, "", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// A. Inline trim markers — output expressions {{ }}
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Inline_Output_NoTrim_PreservesAll(t *testing.T) {
	got := wsRenderPlain(t, " {{ 'x' }} ")
	require.Equal(t, " x ", got)
}

func TestS11_Inline_Output_TrimLeft(t *testing.T) {
	// {{- 'x' }} eats whitespace before the output
	got := wsRenderPlain(t, " \n  {{- 'x' }} ")
	require.Equal(t, "x ", got)
}

func TestS11_Inline_Output_TrimRight(t *testing.T) {
	// {{ 'x' -}} eats whitespace after the output
	got := wsRenderPlain(t, " {{ 'x' -}} \n  ")
	require.Equal(t, " x", got)
}

func TestS11_Inline_Output_TrimBoth(t *testing.T) {
	got := wsRenderPlain(t, "  \n  {{- 'x' -}}  \n  ")
	require.Equal(t, "x", got)
}

func TestS11_Inline_Output_TrimBoth_MultipleBlankLines(t *testing.T) {
	// {{- -}} with several blank lines on both sides: all consumed
	got := wsRenderPlain(t, "a\n\n\n{{- 'mid' -}}\n\n\nb")
	require.Equal(t, "amidb", got)
}

func TestS11_Inline_Output_TrimRight_AdjacentOutput(t *testing.T) {
	// right-trim on first output, no trim on second: whitespace between them consumed
	got := wsRenderPlain(t, `{{ "a" -}}{{ "b" }} c`)
	require.Equal(t, "ab c", got)
}

func TestS11_Inline_Output_TrimBoth_CommaJoined(t *testing.T) {
	// Two trimmed outputs separated by a comma: both collapse to adjacent values
	got := wsRenderPlain(t, "  {{- 'John' -}},\n  {{- '30' -}}  ")
	require.Equal(t, "John,30", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// A. Mixed tag + output trim directions
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Mixed_TagLeft_OutputRight(t *testing.T) {
	// {%- if %} (trim left on tag open), {{ v -}} (trim right on output)
	// {%- eats "\n " before the if; -}} eats " \n" after v; endif has no trim.
	eng := wsEngine(t)
	got := wsRender(t, eng, "\n {%- if true %}a{{ v -}} \n{% endif %}", map[string]any{"v": 1})
	require.Equal(t, "a1", got)
}

func TestS11_Mixed_TrimRightTag_TrimLeftOutput(t *testing.T) {
	// {% if -%} (trim right on open) followed by {{- v }} (trim left on output)
	// -%} eats "  " before {{-, so {{- has nothing left to trim.
	eng := wsEngine(t)
	got := wsRender(t, eng, "{% if true -%}  {{- v }}{% endif %}", map[string]any{"v": "hi"})
	require.Equal(t, "hi", got)
}

func TestS11_Mixed_ComplexInterleaved(t *testing.T) {
	// Full interleaved scenario from Ruby test_complex_trim_output
	src := "      <div>\n" +
		"        <p>\n" +
		"          {{- 'John' -}}\n" +
		"          {{- '30' -}}\n" +
		"        </p>\n" +
		"        <b>\n" +
		"          {{ 'John' -}}\n" +
		"          {{- '30' }}\n" +
		"        </b>\n" +
		"        <i>\n" +
		"          {{- 'John' }}\n" +
		"          {{ '30' -}}\n" +
		"        </i>\n" +
		"      </div>\n    "
	want := "      <div>\n        <p>John30</p>\n        <b>\n          John30\n        </b>\n        <i>John\n          30</i>\n      </div>\n    "
	require.Equal(t, want, wsRenderPlain(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// B. {{-}} trim blank — regression guard
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_TrimBlank_Basic(t *testing.T) {
	// Ruby test_trim_blank: {{-}} trims surrounding whitespace, outputs nothing.
	got := wsRenderPlain(t, "foo {{-}} bar")
	require.Equal(t, "foobar", got)
}

func TestS11_TrimBlank_MultipleSpaces(t *testing.T) {
	// Multiple surrounding spaces all consumed
	got := wsRenderPlain(t, "a   {{-}}   b")
	require.Equal(t, "ab", got)
}

func TestS11_TrimBlank_WithNewlines(t *testing.T) {
	// Newlines on both sides consumed
	got := wsRenderPlain(t, "a\n\n{{-}}\n\nb")
	require.Equal(t, "ab", got)
}

func TestS11_TrimBlank_InMiddleOfText(t *testing.T) {
	// {{-}} in the middle of a sentence collapses the space
	got := wsRenderPlain(t, "hello {{-}} world")
	require.Equal(t, "helloworld", got)
}

func TestS11_TrimBlank_AdjacentToContent(t *testing.T) {
	// {{-}} immediately adjacent to content — no space to trim, no output
	got := wsRenderPlain(t, "AB{{-}}CD")
	require.Equal(t, "ABCD", got)
}

func TestS11_TrimBlank_Multiple(t *testing.T) {
	// Multiple {{-}} in sequence — each is a no-op output with trim
	got := wsRenderPlain(t, "a {{-}} {{-}} b")
	require.Equal(t, "ab", got)
}

func TestS11_TrimBlank_InsideForLoop(t *testing.T) {
	// {{-}} inside a for loop body: TrimLeft=nothing (no preceding ws), TrimRight eats " " before {{ i }}
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{% for i in arr %}{{-}} {{ i }}{% endfor %}",
		map[string]any{"arr": []int{1, 2, 3}})
	// Per iteration: TrimLeft(nothing), TrimRight eats " " before {{ i }} → "1", "2", "3"
	require.Equal(t, "123", got)
}

func TestS11_TrimBlank_NoParseError(t *testing.T) {
	// Regression: {{-}} must NOT produce a parse/syntax error
	eng := wsEngine(t)
	_, err := eng.ParseString("{{-}}")
	require.NoError(t, err, "{{-}} should parse without error")
}

func TestS11_TrimBlank_EmptyExpression_NoOutput(t *testing.T) {
	// Explicitly verify that {{-}} produces no output bytes
	got := wsRenderPlain(t, "{{-}}")
	require.Equal(t, "", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimTagLeft
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimTagLeft_Basic(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	got := wsRender(t, eng, " \n \t{%if true%}foo{%endif%} ", nil)
	require.Equal(t, "foo ", got)
}

func TestS11_Global_TrimTagLeft_MultipleSpaces(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	got := wsRender(t, eng, "   {%if true%}ok{%endif%}", nil)
	require.Equal(t, "ok", got)
}

func TestS11_Global_TrimTagLeft_DoesNotTrimOutput(t *testing.T) {
	// TrimTagLeft trims whitespace text nodes before {% tags %}, but does NOT
	// alter the VALUE rendered by {{ output }} expressions.
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	// Whitespace inside the body between tag-bound content is NOT affected by TrimTagLeft
	got := wsRender(t, eng, "{%if true%}a {{name}} b{%endif%}", map[string]any{"name": "harttle"})
	require.Equal(t, "a harttle b", got)
}

func TestS11_Global_TrimTagLeft_OnlyTrimsTagSide(t *testing.T) {
	// Text AFTER the tag is not trimmed; only before is
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	got := wsRender(t, eng, "  {%assign x = 1%}  after", nil)
	require.Equal(t, "  after", got)
}

func TestS11_Global_TrimTagLeft_FalseBranch(t *testing.T) {
	// Even when if renders nothing, the left trim still applied
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	got := wsRender(t, eng, "   {%if false%}no{%endif%}done", nil)
	require.Equal(t, "done", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimTagRight
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimTagRight_Basic(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	got := wsRender(t, eng, "\t{%if true%}foo{%endif%} \n", nil)
	require.Equal(t, "\tfoo", got)
}

func TestS11_Global_TrimTagRight_MultiLine(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	// TrimRight is in the OUTER sequence after the block. It trims the text FOLLOWING
	// the block tag. Text inside the body is not affected.
	got := wsRender(t, eng, "{%if true%}foo{%endif%}  after", nil)
	// "  " between endif and "after" consumed by TrimRight
	require.Equal(t, "fooafter", got)
}

func TestS11_Global_TrimTagRight_DoesNotTrimOutput(t *testing.T) {
	// TrimTagRight must NOT trim whitespace adjacent to {{ }} expressions
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	got := wsRender(t, eng, "{%if true%}a {{name}} b{%endif%}", map[string]any{"name": "harttle"})
	require.Equal(t, "a harttle b", got)
}

func TestS11_Global_TrimTagRight_DoesNotTrimOutputRight(t *testing.T) {
	// After an output expression, TrimTagRight doesn't trigger (no tag right)
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	got := wsRender(t, eng, "{{ 'x' }}  suffix", nil)
	require.Equal(t, "x  suffix", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimTagLeft + TrimTagRight combined
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimTagBoth_CollapsesAroundTags(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimTagLeft(true)
		e.SetTrimTagRight(true)
	})
	// Empty body: TrimLeft eats leading ws before {%if%}; TrimRight eats trailing ws after {%endif%}
	got := wsRender(t, eng, "  {%if true%}{%endif%}  ", nil)
	require.Equal(t, "", got)
}

func TestS11_Global_TrimTagBoth_ContentBetweenTags(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimTagLeft(true)
		e.SetTrimTagRight(true)
	})
	got := wsRender(t, eng, "  {%if true%}content{%endif%}  ", nil)
	require.Equal(t, "content", got)
}

func TestS11_Global_TrimTagBoth_PreservesOutputExpression(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimTagLeft(true)
		e.SetTrimTagRight(true)
	})
	got := wsRender(t, eng, "{%if true%}a {{name}} b{%endif%}", map[string]any{"name": "harttle"})
	require.Equal(t, "a harttle b", got)
}

func TestS11_Global_TrimTagBoth_MultipleStatements(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimTagLeft(true)
		e.SetTrimTagRight(true)
	})
	// Each tag's left+right whitespace trimmed; content text preserved
	got := wsRender(t, eng,
		"  {%assign a = 1%}  {%assign b = 2%}  {{ a }}+{{ b }}",
		nil)
	require.Equal(t, "1+2", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimOutputLeft
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimOutputLeft_Basic(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputLeft(true) })
	got := wsRender(t, eng, " \n \t{{name}} ", map[string]any{"name": "harttle"})
	require.Equal(t, "harttle ", got)
}

func TestS11_Global_TrimOutputLeft_DoesNotTrimTag(t *testing.T) {
	// TrimOutputLeft must NOT trim whitespace adjacent to {% %} tags
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputLeft(true) })
	got := wsRender(t, eng, "\t{% if true %} aha {%endif%}\t", nil)
	require.Equal(t, "\t aha \t", got)
}

func TestS11_Global_TrimOutputLeft_MultipleOutputs(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputLeft(true) })
	got := wsRender(t, eng, " {{a}}  {{b}} ", map[string]any{"a": 1, "b": 2})
	// Left trim before each output: " {{a}}" → "1", "  {{b}}" → "2"; trailing " " kept
	require.Equal(t, "12 ", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimOutputRight
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimOutputRight_Basic(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputRight(true) })
	got := wsRender(t, eng, " \n \t{{name}} ", map[string]any{"name": "harttle"})
	require.Equal(t, " \n \tharttle", got)
}

func TestS11_Global_TrimOutputRight_DoesNotTrimTag(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputRight(true) })
	got := wsRender(t, eng, "\t{% if true %} aha {%endif%}\t", nil)
	require.Equal(t, "\t aha \t", got)
}

func TestS11_Global_TrimOutputRight_TrailingContentPreserved(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputRight(true) })
	got := wsRender(t, eng, " {{v}} text", map[string]any{"v": "hi"})
	// TrimOutputRight eats " " between output and "text"
	require.Equal(t, " hitext", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// C. Global trim options — TrimOutputLeft + TrimOutputRight combined
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Global_TrimOutputBoth_CollapsesBothSides(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimOutputLeft(true)
		e.SetTrimOutputRight(true)
	})
	got := wsRender(t, eng, "  {{v}}  ", map[string]any{"v": "mid"})
	require.Equal(t, "mid", got)
}

func TestS11_Global_TrimOutputBoth_DoesNotTrimTags(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimOutputLeft(true)
		e.SetTrimOutputRight(true)
	})
	got := wsRender(t, eng, "\t{% if true %} aha {%endif%}\t", nil)
	require.Equal(t, "\t aha \t", got)
}

func TestS11_Global_TrimOutputBoth_MultipleOutputsTouching(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) {
		e.SetTrimOutputLeft(true)
		e.SetTrimOutputRight(true)
	})
	// "  {{a}}  {{b}}  " → both outputs trimmed; "a" and "b" touch each other
	got := wsRender(t, eng, "  {{a}}  {{b}}  ", map[string]any{"a": "A", "b": "B"})
	require.Equal(t, "AB", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// D. Greedy vs. non-greedy semantics
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Greedy_Default_IsTrue(t *testing.T) {
	// Default greedy=true: all consecutive whitespace (incl. multiple newlines) trimmed
	eng := wsEngine(t)
	got := wsRender(t, eng, "\n\n\n{%- if true -%}\nhello\n{%- endif -%}\n\n\n", nil)
	require.Equal(t, "hello", got)
}

func TestS11_Greedy_True_ConsumesAllNewlines(t *testing.T) {
	eng := wsEngine(t, func(e *liquid.Engine) { /* default greedy=true */ })
	got := wsRender(t, eng, "a\n\n\n{%- assign x = 1 -%}\n\n\nb", nil)
	require.Equal(t, "ab", got)
}

func TestS11_Greedy_False_ConsumesOnlyOneNewline(t *testing.T) {
	// non-greedy {%- and -%} behavior:
	// - TrimLeftNonGreedy removes only trailing INLINE-BLANK (space/tab) from buffer
	// - TrimRightNonGreedy removes leading inline-blank + AT MOST 1 newline from next text
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetGreedy(false) })
	// Template: trailing spaces before {%-, spaces+newline after -%}, then second newline
	// -%} eats "  " (inline blanks) + 1 newline → leaves second "\n"+"b"
	src := "a  {%- assign x = 1 -%}  \n\nb"
	got := wsRender(t, eng, src, nil)
	// non-greedy: TrimLeft eats trailing "  " from "a  " → "a"
	// non-greedy TrimRight: eats "  " (inline blanks) + 1 "\n" → leaves "\nb"
	require.Equal(t, "a\nb", got)
}

func TestS11_Greedy_False_InlineBlankBeforeNewline(t *testing.T) {
	// Non-greedy TrimRight eats inline-blank + 1 newline; extra newlines are preserved.
	// non-greedy TrimLeft eats only trailing inline-blank chars NOT newlines.
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetGreedy(false) })
	// Source: spaces before {%-, two newlines after -%}
	// TrimLeft(NG): "a  " → trailing spaces removed → "a" written.
	// TrimRight(NG) on "  \n\nb": spaces eaten, then 1 newline eaten → "\nb" remains.
	got := wsRender(t, eng, "a  {%- assign x = 1 -%}  \n\nb", nil)
	require.Equal(t, "a\nb", got)
}

func TestS11_Greedy_False_PreservesExtraNewlines(t *testing.T) {
	// Non-greedy: two trailing newlines — only one consumed, second preserved
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetGreedy(false) })
	src := "\n {%-if true-%}\n a \n{{-name-}}{%-endif-%}\n "
	got := wsRender(t, eng, src, map[string]any{"name": "harttle"})
	// Exactly matches ported test TestWhitespaceCtrl_Greedy_False
	require.Equal(t, "\n a \nharttle ", got)
}

func TestS11_Greedy_True_SameSource(t *testing.T) {
	// Same source as above with greedy=true
	eng := wsEngine(t)
	src := "\n {%-if true-%}\n a \n{{-name-}}{%-endif-%}\n "
	got := wsRender(t, eng, src, map[string]any{"name": "harttle"})
	require.Equal(t, "aharttle", got)
}

func TestS11_Greedy_ToggleProducesDistinctOutputs(t *testing.T) {
	// The same template must NEVER produce the same output in greedy vs non-greedy
	// whenever there are multiple consecutive whitespace chars.
	src := "a\n\n{%- assign x = 1 -%}\n\nb"
	engG := wsEngine(t)
	engNG := wsEngine(t, func(e *liquid.Engine) { e.SetGreedy(false) })
	outG := wsRender(t, engG, src, nil)
	outNG := wsRender(t, engNG, src, nil)
	assert.NotEqual(t, outG, outNG, "greedy and non-greedy must differ on multi-newline input")
}

// ─────────────────────────────────────────────────────────────────────────────
// E. Interaction: inline markers + global options
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Interaction_InlineMarkerWithGlobalTagTrim_NoDoubleApply(t *testing.T) {
	// Global TrimTagLeft + inline {%- should still work correctly (not double-trim)
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	got := wsRender(t, eng, "  {%- if true %}content{%endif%}", nil)
	// Both the global trim and inline {%- trim the left — result is same: "content"
	require.Equal(t, "content", got)
}

func TestS11_Interaction_InlineOutputMarkerNotAffectedByGlobalTagTrim(t *testing.T) {
	// Global TrimTagLeft trims whitespace TEXT NODES before {% tags %}.
	// It does NOT trim the value rendered by {{ output }} expressions.
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagLeft(true) })
	// The " " between {{ 'x' }} and {%if%} is a text node: TrimLeft eats it.
	// But the value of 'x' itself and the outer " " before {{ }} are not touched.
	got := wsRender(t, eng, " {{ 'x' }} {%if true%}y{%endif%}", nil)
	require.Equal(t, " xy", got)
}

func TestS11_Interaction_GlobalOutputTrim_WithInlineTagTrim(t *testing.T) {
	// Global TrimOutputLeft + inline {%- tag trim: both active independently
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimOutputLeft(true) })
	got := wsRender(t, eng, "  {{ v }}  {%- if true %}ok{% endif %}", map[string]any{"v": 1})
	// Output left trim: "  {{ v }}" → leading "  " consumed → "1"
	// {%- tag: trim left before the if → "  " before if consumed
	// Result: "1ok"
	require.Equal(t, "1ok", got)
}

func TestS11_Interaction_GlobalTagTrim_WithInlineOutputTrim(t *testing.T) {
	// Global TrimTagRight trims the text FOLLOWING a block (in outer context).
	// It does NOT affect text inside the body that follows an output expression.
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	got := wsRender(t, eng, "{%if true%}  {{- v }}  {%endif%}  end", map[string]any{"v": "hi"})
	// Body: TrimLeft (from {{-) eats leading "  "; outputs "hi"; trailing "  " stays in body.
	// TrimRight (global, after endif in outer) eats "  " before "end".
	require.Equal(t, "hi  end", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// F. All tag types — inline trim works for every supported tag
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Tags_For_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"\n{%- for i in arr -%}\n{{ i }}\n{%- endfor -%}\n",
		map[string]any{"arr": []int{1, 2, 3}})
	// Body: TrimRight (from -%} on for open) eats "\n" before {{ i }};
	// TrimLeft (from {%- on endfor) eats "\n" after {{ i }}.
	// Per-iteration: "1", "2", "3" — all adjacent.
	require.Equal(t, "123", got)
}

func TestS11_Tags_For_TrimRight_Open_TrimLeft_Close(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{% for i in arr -%}\n{{ i }}\n{%- endfor %}",
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "123", got)
}

func TestS11_Tags_For_Reversed_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- for i in arr reversed -%}{{ i }}{%- endfor -%}",
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "321", got)
}

func TestS11_Tags_For_WithRange_TrimBoth(t *testing.T) {
	got := wsRenderPlain(t, "{%- for i in (1..3) -%}{{ i }}{%- endfor -%}")
	require.Equal(t, "123", got)
}

func TestS11_Tags_For_Limit_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- for i in arr limit: 2 -%}{{ i }}{%- endfor -%}",
		map[string]any{"arr": []int{1, 2, 3, 4}})
	require.Equal(t, "12", got)
}

func TestS11_Tags_For_Offset_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- for i in arr offset: 1 -%}{{ i }}{%- endfor -%}",
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "23", got)
}

func TestS11_Tags_For_Else_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"\n{%- for i in arr -%}{{ i }}{%- else -%} none {%- endfor -%}\n",
		map[string]any{"arr": []int{}})
	require.Equal(t, "none", got)
}

func TestS11_Tags_If_AllForms_TrimBoth(t *testing.T) {
	// if / elsif / else / endif all with trim
	eng := wsEngine(t)
	for _, tc := range []struct {
		v    int
		want string
	}{
		{1, "one"},
		{2, "two"},
		{3, "other"},
	} {
		t.Run(fmt.Sprintf("v=%d", tc.v), func(t *testing.T) {
			got := wsRender(t, eng,
				"{%- if v == 1 -%}one{%- elsif v == 2 -%}two{%- else -%}other{%- endif -%}",
				map[string]any{"v": tc.v})
			require.Equal(t, tc.want, got)
		})
	}
}

func TestS11_Tags_Unless_TrimBoth(t *testing.T) {
	got := wsRenderPlain(t, "  {%- unless false -%}  yes  {%- endunless -%}  ")
	require.Equal(t, "yes", got)
}

func TestS11_Tags_Case_TrimBoth(t *testing.T) {
	eng := wsEngine(t)
	for _, tc := range []struct {
		v    int
		want string
	}{
		{1, "one"},
		{2, "two"},
		{99, "other"},
	} {
		t.Run(fmt.Sprintf("v=%d", tc.v), func(t *testing.T) {
			got := wsRender(t, eng,
				" {%- case v -%} {%- when 1 -%}one{%- when 2 -%}two{%- else -%}other{%- endcase -%} ",
				map[string]any{"v": tc.v})
			require.Equal(t, tc.want, got)
		})
	}
}

func TestS11_Tags_Assign_TrimBoth_Invisible(t *testing.T) {
	// assign renders nothing; TrimLeft eats all trailing whitespace from before;
	// TrimRight eats all leading whitespace from after.
	got := wsRenderPlain(t, "before \n {%- assign x = 42 -%} \n after{{ x }}")
	// TrimLeft: "before \n " → trailing ws trimmed → "before"
	// TrimRight: " \n after" → leading ws trimmed → "after"
	require.Equal(t, "beforeafter42", got)
}

func TestS11_Tags_Capture_TrimBoth(t *testing.T) {
	// capture with trim on both block delimiters
	got := wsRenderPlain(t, "{%- capture msg -%}  hello world  {%- endcapture -%}[{{ msg }}]")
	require.Equal(t, "[hello world]", got)
}

func TestS11_Tags_Increment_TrimBoth(t *testing.T) {
	// increment outputs a value; trim markers collapse surrounding whitespace
	got := wsRenderPlain(t, "  {%- increment v -%}  ")
	require.Equal(t, "0", got)
}

func TestS11_Tags_Decrement_TrimBoth(t *testing.T) {
	got := wsRenderPlain(t, "  {%- decrement v -%}  ")
	require.Equal(t, "-1", got)
}

func TestS11_Tags_Echo_TrimBoth(t *testing.T) {
	got := wsRenderPlain(t, "  {%- echo 'hi' -%}  ")
	require.Equal(t, "hi", got)
}

func TestS11_Tags_InlineComment_TrimBoth(t *testing.T) {
	// {%- # comment -%} trims both sides, outputs nothing
	got := wsRenderPlain(t, "a \n{%- # inline comment -%}\n b")
	require.Equal(t, "ab", got)
}

func TestS11_Tags_InlineComment_WithSpace_TrimBoth(t *testing.T) {
	// {%- # comment -%} (space after dash) — the B3 bug fix variant
	got := wsRenderPlain(t, "a \n{%- # comment with space -%}\n b")
	require.Equal(t, "ab", got)
}

func TestS11_Tags_LiquidTag_TrimBoth(t *testing.T) {
	// {%- liquid ... -%} multi-statement tag with trim
	got := wsRenderPlain(t, " \n{%- liquid\n  assign a = 1\n  assign b = 2\n-%}\n {{ a }}+{{ b }}")
	require.Equal(t, "1+2", got)
}

func TestS11_Tags_Raw_ExternalTrim_DoesNotAffectContent(t *testing.T) {
	// {%- raw -%}: trim markers on the raw tag trim OUTSIDE whitespace only
	// Content inside raw is emitted verbatim (trim markers inside are literal)
	got := wsRenderPlain(t, "before \n{%- raw -%}\n{%- {{- verbatim -}} -%}\n{%- endraw -%}\n after")
	require.Equal(t, "before\n{%- {{- verbatim -}} -%}\nafter", got)
}

// ─────────────────────────────────────────────────────────────────────────────
// F. Deeply nested tag combinations
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Nested_ForInIf_TrimAll(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- if show -%}\n{%- for i in arr -%}{{ i }}{%- endfor -%}\n{%- endif -%}",
		map[string]any{"show": true, "arr": []int{1, 2, 3}})
	require.Equal(t, "123", got)
}

func TestS11_Nested_ForInIf_TrimAll_FalseBranch(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"x{%- if show -%}{%- for i in arr -%}{{ i }}{%- endfor -%}{%- endif -%}y",
		map[string]any{"show": false, "arr": []int{1, 2, 3}})
	require.Equal(t, "xy", got)
}

func TestS11_Nested_IfInFor_ConditionFilter(t *testing.T) {
	// Filter applied in a nested if condition inside a for loop with trim
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- for item in arr -%}{%- if item.size > 3 -%}{{ item }}{%- endif -%}{%- endfor -%}",
		map[string]any{"arr": []string{"hi", "hello", "hey", "world"}})
	require.Equal(t, "helloworld", got)
}

func TestS11_Nested_ThreeLevels_TrimAll(t *testing.T) {
	eng := wsEngine(t)
	got := wsRender(t, eng,
		"{%- for i in arr -%}{%- for j in arr -%}{%- if i == j -%}{{ i }}{%- endif -%}{%- endfor -%}{%- endfor -%}",
		map[string]any{"arr": []int{1, 2, 3}})
	require.Equal(t, "123", got)
}

func TestS11_Nested_DeepStructure_AllTrimmed(t *testing.T) {
	// Ruby test_complex_trim: nested if + output markers fully collapse whitespace
	src := "      <div>\n" +
		"        {%- if true -%}\n" +
		"          {%- if true -%}\n" +
		"            <p>\n" +
		"              {{- 'John' -}}\n" +
		"            </p>\n" +
		"          {%- endif -%}\n" +
		"        {%- endif -%}\n" +
		"      </div>\n    "
	want := "      <div><p>John</p></div>\n    "
	require.Equal(t, want, wsRenderPlain(t, src))
}

// ─────────────────────────────────────────────────────────────────────────────
// G. Edge cases
// ─────────────────────────────────────────────────────────────────────────────

func TestS11_Edge_EmptyTemplate(t *testing.T) {
	got := wsRenderPlain(t, "")
	require.Equal(t, "", got)
}

func TestS11_Edge_OnlyTrimMarkers(t *testing.T) {
	// A template that is only trim markers with no content
	got := wsRenderPlain(t, "{%- assign x = 1 -%}")
	require.Equal(t, "", got)
}

func TestS11_Edge_TrimDoesNotAffectStringContent(t *testing.T) {
	// Trim should not remove whitespace WITHIN string literal values
	got := wsRenderPlain(t, "{%- assign v = '  hello  ' -%}[{{ v }}]")
	require.Equal(t, "[  hello  ]", got)
}

func TestS11_Edge_TrimAcrossMultipleNodes(t *testing.T) {
	// Right trim on one output, left trim on next output — space between consumed
	got := wsRenderPlain(t, "{{ 'a' -}}   {{ 'b' -}}   {{ 'c' }}")
	require.Equal(t, "abc", got)
}

func TestS11_Edge_TrimLeftPreservesNonWhitespaceOnLeft(t *testing.T) {
	// {%- tag does NOT trim non-whitespace characters to the left
	got := wsRenderPlain(t, "abc{%- if true %}yes{% endif %}")
	require.Equal(t, "abcyes", got)
}

func TestS11_Edge_TrimRightPreservesNonWhitespaceOnRight(t *testing.T) {
	// tag -%} does NOT trim non-whitespace characters to the right
	got := wsRenderPlain(t, "{% if true -%}abc{% endif %}")
	require.Equal(t, "abc", got)
}

func TestS11_Edge_TrimWithTabCharacters(t *testing.T) {
	// Trim should also consume tab characters
	got := wsRenderPlain(t, "\t\t{%- if true -%}\t\tcontent\t\t{%- endif -%}\t\t")
	require.Equal(t, "content", got)
}

func TestS11_Edge_TrimWithCarriageReturn(t *testing.T) {
	// \r\n line endings: trim should handle them
	got := wsRenderPlain(t, "a\r\n{%- assign x = 1 -%}\r\nb")
	require.Equal(t, "ab", got)
}

func TestS11_Edge_TrimBlank_InsideCapture(t *testing.T) {
	// {{-}} inside a capture block: trims whitespace, outputs nothing
	got := wsRenderPlain(t, "{%- capture c -%}  {{-}}  hello{{-}}  {%- endcapture -%}[{{ c }}]")
	require.Equal(t, "[hello]", got)
}

func TestS11_Edge_TrimBlank_Next_ToOtherOutput(t *testing.T) {
	// {{-}} immediately before a real output: trims the space, output follows
	got := wsRender(t, wsEngine(t), "{{-}} {{ v }}", map[string]any{"v": "x"})
	require.Equal(t, "x", got)
}

func TestS11_Edge_GlobalTrimTag_DoesNotAffectRawContent(t *testing.T) {
	// Global TrimTagRight is NOT applied to RawNode (special case: raw blocks bypass
	// the TrimNode injection in compileNodes). Raw content is always emitted verbatim.
	eng := wsEngine(t, func(e *liquid.Engine) { e.SetTrimTagRight(true) })
	got := wsRender(t, eng, "{% raw %}  {{- verbatim -}}  {% endraw %}", nil)
	// Raw block: content emitted verbatim, no TrimRight applied.
	require.Equal(t, "  {{- verbatim -}}  ", got)
}

func TestS11_Edge_LargeWhitespaceBlob_GreedyConsumesAll(t *testing.T) {
	// Greedy mode should consume an arbitrarily large whitespace blob
	bigWS := strings.Repeat("\n   \t", 20) // 20 repetitions of \n + spaces + tab
	src := "a" + bigWS + "{%- assign x = 1 -%}" + bigWS + "b"
	got := wsRenderPlain(t, src)
	require.Equal(t, "ab", got)
}

func TestS11_Edge_TrimTag_EmptyForBody(t *testing.T) {
	// for loop over empty array with trim — should produce nothing cleanly
	got := wsRender(t, wsEngine(t),
		"  {%- for i in arr -%}{{ i }}{%- endfor -%}  ",
		map[string]any{"arr": []int{}})
	require.Equal(t, "", got)
}

func TestS11_Edge_TrimTag_PreservesOutputInsideTag(t *testing.T) {
	// Trim on the tag itself does not alter the VALUE of output inside the tag body
	got := wsRender(t, wsEngine(t),
		"{%- for i in arr -%} {{ i }} {%- endfor -%}",
		map[string]any{"arr": []int{1, 2, 3}})
	// -%} after for eats " " before "{{ i }}", {%- before endfor eats " " after "{{ i }}"
	require.Equal(t, "123", got)
}

func TestS11_Edge_TrimBothSides_MultipleConsecutiveTags(t *testing.T) {
	// Multiple consecutive tags all with trim: the whitespace between them collapses
	got := wsRenderPlain(t,
		"  {%- assign a = 1 -%}  {%- assign b = 2 -%}  {%- assign c = 3 -%}  {{ a }}{{ b }}{{ c }}")
	require.Equal(t, "123", got)
}
