package liquid_test

// s10_error_handling_e2e_test.go — Intensive E2E tests for Section 10: Tratamento de Erros
//
// Coverage matrix (regression guard: prevents silent behaviour changes):
//
//   A. ParseError / SyntaxError
//      A1  — basic "Liquid syntax error" prefix on all parse-time failures
//      A2  — SyntaxError type alias: errors.As works with both *ParseError and *SyntaxError
//      A3  — line number on single-line template
//      A4  — line number on multi-line template (error on line N ≠ 1)
//      A5  — line number correct inside nested blocks
//      A6  — line number correct when whitespace-trim markers ({%- -%}) are used
//      A7  — Path() and LineNumber() on ParseError
//      A8  — Message() strips prefix and location info
//      A9  — MarkupContext() returns exact source text of the failing token
//      A10 — unknown tag → ParseError (not a runtime/render error)
//      A11 — unclosed block → ParseError
//      A12 — invalid operator (=!) → ParseError
//
//   B. RenderError
//      B1  — "Liquid error" prefix (NOT "Liquid syntax error")
//      B2  — ZeroDivision wrapped in *render.RenderError
//      B3  — plain filter error wrapped in *render.RenderError
//      B4  — plain tag error wrapped in *render.RenderError
//      B5  — line number correct on first line and on line N
//      B6  — Message() strips "Liquid error" prefix and location
//      B7  — MarkupContext() carries the failing {{ expr }} source text
//
//   C. ZeroDivisionError
//      C1  — divided_by: 0 → *filters.ZeroDivisionError findable via errors.As
//      C2  — modulo: 0     → *filters.ZeroDivisionError findable via errors.As
//      C3  — ZeroDivisionError sits below RenderError in the chain
//      C4  — divided_by / modulo with non-zero → no error
//      C5  — ZeroDivisionError message content
//
//   D. ArgumentError / ContextError (typed leaf errors)
//      D1  — filter returning *render.ArgumentError → detectable via errors.As
//      D2  — tag returning *render.ArgumentError → detectable via errors.As
//      D3  — tag returning *render.ContextError   → detectable via errors.As
//      D4  — ArgumentError message carried through chain
//      D5  — ContextError message carried through chain
//      D6  — error from filter has "Liquid error" prefix in full string
//
//   E. UndefinedVariableError
//      E1  — default mode: undefined variable → empty string, no error
//      E2  — StrictVariables(): undefined var → *render.UndefinedVariableError
//      E3  — Name field set to root variable name
//      E4  — line number and markup context set correctly
//      E5  — per-render WithStrictVariables() same as engine-level
//      E6  — errors.As chain: UndefinedVariableError findable
//      E7  — defined variable with StrictVariables: no error
//      E8  — dotted access: root name preserved (e.g. user.name → Name="user")
//
//   F. WithErrorHandler (exception_renderer)
//      F1  — handler output replaces the failing node text
//      F2  — rendering continues after the failing node
//      F3  — multiple errors handled; output assembled in order
//      F4  — handler receives the error (errors.As works inside handler)
//      F5  — parse errors are NOT caught by the render handler
//      F6  — non-erroring nodes render correctly alongside failing nodes
//
//   G. markup_context metadata (end-to-end)
//      G1  — Error() shows markup context ({{ expr }}) when no path set
//      G2  — Error() shows path NOT markup context when path is set
//      G3  — nested render: inner markup context preserved over outer block source
//      G4  — MarkupContext() is empty when no locatable information is available
//
//   H. Error chain walking (errors.As through full chain)
//      H1  — ZeroDivisionError walkable without knowing intermediate types
//      H2  — ArgumentError walkable from top-level error
//      H3  — RenderError always present in chain for render-time failures
//      H4  — ParseError always present in chain for parse-time failures
//
//   I. Prefix invariants (regression guard)
//      I1  — every parse-time error starts with "Liquid syntax error"
//      I2  — every render-time error starts with "Liquid error" (never "Liquid syntax error")
//      I3  — render error with line N includes "(line N)" in string
//      I4  — parse error with line N includes "(line N)" in string

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/osteele/liquid"
	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func s10eng(t *testing.T) *liquid.Engine {
	t.Helper()
	return liquid.NewEngine()
}

func s10mustParse(t *testing.T, eng *liquid.Engine, src string) *liquid.Template {
	t.Helper()
	tpl, err := eng.ParseString(src)
	require.NoError(t, err, "unexpected parse error for %q", src)
	return tpl
}

func s10parseErr(t *testing.T, src string) error {
	t.Helper()
	_, err := s10eng(t).ParseString(src)
	require.Error(t, err, "expected a parse error for %q", src)
	return err
}

func s10renderErr(t *testing.T, eng *liquid.Engine, src string, binds map[string]any) error {
	t.Helper()
	_, err := eng.ParseAndRenderString(src, binds)
	require.Error(t, err, "expected a render error for %q", src)
	return err
}

func s10render(t *testing.T, eng *liquid.Engine, src string, binds map[string]any) string {
	t.Helper()
	out, err := eng.ParseAndRenderString(src, binds)
	require.NoError(t, err, "unexpected error for %q", src)
	return out
}

// ═════════════════════════════════════════════════════════════════════════════
// A. ParseError / SyntaxError
// ═════════════════════════════════════════════════════════════════════════════

// A1 — ParseError carries "Liquid syntax error" prefix for all parse failures.
func TestS10A1_ParseError_Prefix_UnclosedFor(t *testing.T) {
	err := s10parseErr(t, `{% for x in arr %}`)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"got: %q", err.Error())
}

func TestS10A1_ParseError_Prefix_UnclosedIf(t *testing.T) {
	err := s10parseErr(t, `{% if cond %}`)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"got: %q", err.Error())
}

func TestS10A1_ParseError_Prefix_UnclosedCapture(t *testing.T) {
	err := s10parseErr(t, `{% capture x %}`)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"got: %q", err.Error())
}

func TestS10A1_ParseError_Prefix_UnclosedUnless(t *testing.T) {
	err := s10parseErr(t, `{% unless cond %}`)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"got: %q", err.Error())
}

func TestS10A1_ParseError_Prefix_UnknownTag(t *testing.T) {
	err := s10parseErr(t, `{% totallynotthere %}`)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"got: %q", err.Error())
}

// A2 — SyntaxError alias: errors.As works with both *ParseError and *SyntaxError.
func TestS10A2_SyntaxError_Alias_ParseError(t *testing.T) {
	err := s10parseErr(t, `{% unclosed %}`)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "errors.As(*ParseError) failed, got %T", err)
}

func TestS10A2_SyntaxError_Alias_SyntaxError(t *testing.T) {
	err := s10parseErr(t, `{% unclosed %}`)
	var se *parser.SyntaxError
	require.True(t, errors.As(err, &se), "errors.As(*SyntaxError) failed, got %T", err)
}

func TestS10A2_SyntaxError_Alias_SamePointer(t *testing.T) {
	err := s10parseErr(t, `{% unclosed %}`)
	var pe *parser.ParseError
	var se *parser.SyntaxError
	require.True(t, errors.As(err, &pe))
	require.True(t, errors.As(err, &se))
	// SyntaxError = ParseError (type alias) — same object
	require.Equal(t, pe, se)
}

// A3 — Line number is 1 for single-line templates.
func TestS10A3_ParseError_LineNumber_SingleLine(t *testing.T) {
	err := s10parseErr(t, `{% unknowntag_s10a3 %}`)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))
	assert.Equal(t, 1, pe.LineNumber())
	assert.Contains(t, pe.Error(), "line 1")
}

// A4 — Line number correct when error is on line N > 1.
func TestS10A4_ParseError_LineNumber_Line2(t *testing.T) {
	src := "good line 1\n{% unknowntag_s10a4 %}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 2")
}

func TestS10A4_ParseError_LineNumber_Line3(t *testing.T) {
	src := "foobar\n\n{% unknowntag_s10a4_line3 %}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 3")
}

func TestS10A4_ParseError_LineNumber_Line5(t *testing.T) {
	src := "l1\nl2\nl3\nl4\n{% unknowntag_s10a4_line5 %}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 5")
}

// A5 — Line number correct for errors inside nested blocks.
func TestS10A5_ParseError_LineNumber_NestedBlock(t *testing.T) {
	// Unknown tag inside {% if %} — error must report line 4, not line 1.
	src := "foobar\n\n{% if 1 != 2 %}\n  {% unknowntag_nested %}\n{% endif %}\n\nbla"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 4",
		"nested error at line 4, full error: %q", err.Error())
}

func TestS10A5_ParseError_LineNumber_NestedFor(t *testing.T) {
	src := "before\n{% for i in arr %}\n  {% nosuchfoo %}\n{% endfor %}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 3")
}

// A6 — Whitespace-trim markers do not shift line numbers.
func TestS10A6_ParseError_WhitespaceTrim_LineNumberUnchanged(t *testing.T) {
	// Without trim markers: line 3
	src1 := "foobar\n\n{% unknowntag_s10a6 %}\n\nbla"
	err1 := s10parseErr(t, src1)
	assert.Contains(t, err1.Error(), "line 3", "without trim markers: %q", err1.Error())

	// With trim markers: still line 3
	src2 := "foobar\n\n{%- unknowntag_s10a6 -%}\n\nbla"
	err2 := s10parseErr(t, src2)
	assert.Contains(t, err2.Error(), "line 3",
		"trim markers must not shift line number: %q", err2.Error())
}

func TestS10A6_ParseError_WhitespaceTrim_MultipleLines(t *testing.T) {
	src := "{%- assign x = 1 -%}\n{%- assign y = 2 -%}\n{%- unknowntag_multiline -%}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "line 3")
}

// A7 — Path() and LineNumber() accessible on ParseError.
func TestS10A7_ParseError_Path_FromToken(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{Pathname: "theme/product.html", LineNo: 12},
		Source:    `{% badtag %}`,
	}
	err := parser.Errorf(&tok, "unknown tag 'badtag'")
	assert.Equal(t, "theme/product.html", err.Path())
	assert.Equal(t, 12, err.LineNumber())
	assert.Contains(t, err.Error(), "theme/product.html")
	assert.Contains(t, err.Error(), "line 12")
}

func TestS10A7_ParseError_NoPath_EmptyString(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{LineNo: 3},
		Source:    `{% badtag %}`,
	}
	err := parser.Errorf(&tok, "unknown tag")
	assert.Equal(t, "", err.Path())
}

// A8 — Message() strips "Liquid syntax error" prefix and "(line N)" location.
func TestS10A8_ParseError_Message_NoPrefix(t *testing.T) {
	err := s10parseErr(t, `{% for a in b %}`)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))
	msg := pe.Message()
	assert.NotEmpty(t, msg, "Message() must not be empty")
	assert.NotContains(t, msg, "Liquid syntax error")
	assert.NotContains(t, msg, "Liquid error")
	assert.NotContains(t, msg, "(line ")
}

func TestS10A8_ParseError_Message_NoLineInfo(t *testing.T) {
	src := "l1\nl2\n{% for a in b %}" // error on line 3
	err := s10parseErr(t, src)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))
	// Full Error() has "line 3", Message() must not
	assert.Contains(t, pe.Error(), "line 3")
	assert.NotContains(t, pe.Message(), "line 3")
}

// A9 — MarkupContext() returns exact source text of the failing token.
func TestS10A9_ParseError_MarkupContext_SourceText(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{LineNo: 1},
		Source:    `{% bad_tag with_args %}`,
	}
	err := parser.Errorf(&tok, "unknown tag")
	assert.Equal(t, `{% bad_tag with_args %}`, err.MarkupContext())
}

func TestS10A9_ParseError_MarkupContext_InErrorString_WhenNoPath(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{LineNo: 1},
		Source:    `{% some_special_tag %}`,
	}
	err := parser.Errorf(&tok, "not found")
	// No pathname → markup context appears in Error() string
	assert.Contains(t, err.Error(), `{% some_special_tag %}`)
}

func TestS10A9_ParseError_MarkupContext_HiddenWhenPathSet(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{Pathname: "index.html", LineNo: 1},
		Source:    `{% my_tag %}`,
	}
	err := parser.Errorf(&tok, "unknown tag")
	// With pathname, the path appears instead of the raw markup context
	assert.Contains(t, err.Error(), "index.html")
	assert.NotContains(t, err.Error(), `{% my_tag %}`)
}

// A10 — Unknown tag produces a *parser.ParseError.
func TestS10A10_UnknownTag_IsParseError(t *testing.T) {
	err := s10parseErr(t, `{% totally_unknown_tag_xyz %}`)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "got %T: %v", err, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"))
}

func TestS10A10_UnknownTag_NotARenderError(t *testing.T) {
	err := s10parseErr(t, `{% totally_unknown_tag_abc %}`)
	var re *render.RenderError
	// Must NOT be a RenderError — parse errors are not render errors
	assert.False(t, errors.As(err, &re),
		"unknown tag should be a parse error, not a render error")
}

// A11 — Unclosed block tag produces a ParseError with appropriate message.
func TestS10A11_UnclosedBlock_For(t *testing.T) {
	err := s10parseErr(t, `{% for a in b %} ... `)
	assert.Contains(t, err.Error(), "Liquid syntax error")
	assert.Contains(t, err.Error(), "for")
}

func TestS10A11_UnclosedBlock_If(t *testing.T) {
	err := s10parseErr(t, `{% if x %}`)
	assert.Contains(t, err.Error(), "Liquid syntax error")
}

func TestS10A11_UnclosedBlock_TableRow(t *testing.T) {
	err := s10parseErr(t, `{% tablerow i in arr %}cell`)
	assert.Contains(t, err.Error(), "Liquid syntax error")
}

// A12 — Invalid operator (=!) in expression causes a ParseError.
func TestS10A12_InvalidOperator_IsParseError(t *testing.T) {
	err := s10parseErr(t, `{% if 1 =! 2 %}yes{% endif %}`)
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "=! must cause a ParseError, got %T: %v", err, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"))
}

// ═════════════════════════════════════════════════════════════════════════════
// B. RenderError
// ═════════════════════════════════════════════════════════════════════════════

// B1 — RenderError carries "Liquid error" prefix, never "Liquid syntax error".
func TestS10B1_RenderError_Prefix_ZeroDivision(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 10 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid error"),
		"render error must start with 'Liquid error', got: %q", err.Error())
	assert.False(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
		"render error must NOT start with 'Liquid syntax error', got: %q", err.Error())
}

func TestS10B1_RenderError_Prefix_CustomFilter(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("err_filter_b1", func(v any) (any, error) {
		return nil, errors.New("deliberate failure")
	})
	err := s10renderErr(t, eng, `{{ "x" | err_filter_b1 }}`, nil)
	assert.True(t, strings.HasPrefix(err.Error(), "Liquid error"),
		"got: %q", err.Error())
}

// B2 — ZeroDivision is wrapped in *render.RenderError.
func TestS10B2_RenderError_ZeroDivision_WrappedType(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 1 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var re *render.RenderError
	require.True(t, errors.As(err, &re), "ZeroDivision must be wrapped in *render.RenderError, got %T", err)
}

// B3 — Plain filter error wrapped in *render.RenderError.
func TestS10B3_RenderError_FilterPlainError(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("plain_err_b3", func(v any) (any, error) {
		return nil, errors.New("plain error from filter")
	})
	err := s10renderErr(t, eng, `{{ "x" | plain_err_b3 }}`, nil)
	var re *render.RenderError
	require.True(t, errors.As(err, &re), "plain filter error must be *render.RenderError, got %T", err)
	assert.Contains(t, err.Error(), "plain error from filter")
}

// B4 — Plain tag error wrapped in *render.RenderError.
func TestS10B4_RenderError_TagPlainError(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterTag("plain_err_b4", func(c render.Context) (string, error) {
		return "", errors.New("plain error from tag")
	})
	err := s10renderErr(t, eng, `{% plain_err_b4 %}`, nil)
	var re *render.RenderError
	require.True(t, errors.As(err, &re), "plain tag error must be *render.RenderError, got %T", err)
	assert.Contains(t, err.Error(), "plain error from tag")
}

// B5 — Line number correct in RenderError.
func TestS10B5_RenderError_LineNumber_Line1(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 5 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 1")
}

func TestS10B5_RenderError_LineNumber_LineN(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, "line1\nline2\n{{ 5 | divided_by: 0 }}")
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3")
}

// B6 — Message() strips "Liquid error" prefix and location info.
func TestS10B6_RenderError_Message_NoPrefix(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 10 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var re *render.RenderError
	require.True(t, errors.As(err, &re))
	msg := re.Message()
	assert.NotEmpty(t, msg)
	assert.NotContains(t, msg, "Liquid error")
	assert.NotContains(t, msg, "(line ")
}

// B7 — MarkupContext() carries source text of the failing {{ expr }}.
func TestS10B7_RenderError_MarkupContext_ExprSource(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ product.price | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var re *render.RenderError
	require.True(t, errors.As(err, &re))
	// MarkupContext must contain the expression source
	assert.Contains(t, re.MarkupContext(), "product.price")
	// And the full error string shows the markup context (no path set)
	assert.Contains(t, err.Error(), "product.price")
}

func TestS10B7_RenderError_MarkupContext_TagSource(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterTag("err_tag_b7", func(c render.Context) (string, error) {
		return "", errors.New("b7 tag error")
	})
	err := s10renderErr(t, eng, `{% err_tag_b7 %}`, nil)
	var re *render.RenderError
	require.True(t, errors.As(err, &re))
	assert.Contains(t, re.MarkupContext(), "err_tag_b7")
}

// ═════════════════════════════════════════════════════════════════════════════
// C. ZeroDivisionError
// ═════════════════════════════════════════════════════════════════════════════

// C1 — divided_by: 0 produces *filters.ZeroDivisionError findable via errors.As.
func TestS10C1_ZeroDivisionError_DividedBy_ErrorsAs(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 10 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(err, &zde), "divided_by: 0 must yield *filters.ZeroDivisionError, got %T", err)
}

// C2 — modulo: 0 produces *filters.ZeroDivisionError findable via errors.As.
func TestS10C2_ZeroDivisionError_Modulo_ErrorsAs(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 10 | modulo: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(err, &zde), "modulo: 0 must yield *filters.ZeroDivisionError, got %T", err)
}

// C3 — ZeroDivisionError sits below *render.RenderError in the chain.
func TestS10C3_ZeroDivisionError_BelowRenderError(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 7 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var re *render.RenderError
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(err, &re), "outer wrapper must be *render.RenderError")
	require.True(t, errors.As(err, &zde), "inner cause must be *filters.ZeroDivisionError")
}

// C4 — Non-zero divisor: no error, correct result.
func TestS10C4_ZeroDivisionError_NonZero_NoError(t *testing.T) {
	out := s10render(t, s10eng(t), `{{ 10 | divided_by: 2 }}`, nil)
	assert.Equal(t, "5", out)
}

func TestS10C4_ZeroDivisionError_Modulo_NonZero_NoError(t *testing.T) {
	out := s10render(t, s10eng(t), `{{ 10 | modulo: 3 }}`, nil)
	assert.Equal(t, "1", out)
}

// C5 — ZeroDivisionError has a meaningful Error() message.
func TestS10C5_ZeroDivisionError_Message(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 1 | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(err, &zde))
	assert.NotEmpty(t, zde.Error())
	// Typically "divided by 0" or similar phrasing
	assert.True(t,
		strings.Contains(zde.Error(), "0") || strings.Contains(strings.ToLower(zde.Error()), "divis"),
		"ZeroDivisionError message should mention zero or division: %q", zde.Error())
}

// ═════════════════════════════════════════════════════════════════════════════
// D. ArgumentError / ContextError
// ═════════════════════════════════════════════════════════════════════════════

// D1 — Filter returning *render.ArgumentError → detectable via errors.As.
func TestS10D1_ArgumentError_FromFilter(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("bad_args_d1", func(v any) (any, error) {
		return nil, render.NewArgumentError("argument error from filter")
	})
	err := s10renderErr(t, eng, `{{ "x" | bad_args_d1 }}`, nil)
	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae), "expected *render.ArgumentError, got %T: %v", err, err)
}

// D2 — Tag returning *render.ArgumentError → detectable via errors.As.
func TestS10D2_ArgumentError_FromTag(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterTag("bad_tag_d2", func(c render.Context) (string, error) {
		return "", render.NewArgumentError("argument error from tag")
	})
	err := s10renderErr(t, eng, `{% bad_tag_d2 %}`, nil)
	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae), "expected *render.ArgumentError, got %T: %v", err, err)
}

// D3 — Tag returning *render.ContextError → detectable via errors.As.
func TestS10D3_ContextError_FromTag(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterTag("ctx_err_d3", func(c render.Context) (string, error) {
		return "", render.NewContextError("context error from tag")
	})
	err := s10renderErr(t, eng, `{% ctx_err_d3 %}`, nil)
	var ce *render.ContextError
	require.True(t, errors.As(err, &ce), "expected *render.ContextError, got %T: %v", err, err)
}

// D4 — ArgumentError message propagated through chain.
func TestS10D4_ArgumentError_MessageInChain(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("bad_args_d4", func(v any) (any, error) {
		return nil, render.NewArgumentError("this is my specific argument error message")
	})
	err := s10renderErr(t, eng, `{{ x | bad_args_d4 }}`, map[string]any{"x": 1})
	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae))
	assert.Equal(t, "this is my specific argument error message", ae.Error())
}

// D5 — ContextError message propagated through chain.
func TestS10D5_ContextError_MessageInChain(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterTag("ctx_err_d5", func(c render.Context) (string, error) {
		return "", render.NewContextError("ctx-specific error message")
	})
	err := s10renderErr(t, eng, `{% ctx_err_d5 %}`, nil)
	var ce *render.ContextError
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "ctx-specific error message", ce.Error())
}

// D6 — Error string from filter has "Liquid error" prefix (not "Liquid syntax error").
func TestS10D6_ArgumentError_FullErrorHasLiquidErrorPrefix(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("bad_args_d6", func(v any) (any, error) {
		return nil, render.NewArgumentError("bad arg")
	})
	err := s10renderErr(t, eng, `{{ 1 | bad_args_d6 }}`, nil)
	assert.Contains(t, err.Error(), "Liquid error")
	assert.NotContains(t, err.Error(), "Liquid syntax error")
}

// ═════════════════════════════════════════════════════════════════════════════
// E. UndefinedVariableError
// ═════════════════════════════════════════════════════════════════════════════

// E1 — Default (non-strict) mode: undefined variable → empty string, no error.
func TestS10E1_UndefinedVar_DefaultMode_NoError(t *testing.T) {
	out := s10render(t, s10eng(t), `X{{ missing_var_e1 }}Y`, nil)
	assert.Equal(t, "XY", out)
}

func TestS10E1_UndefinedVar_DefaultMode_NestedProperty(t *testing.T) {
	out := s10render(t, s10eng(t), `{{ user.name }}`, nil)
	assert.Equal(t, "", out)
}

// E2 — StrictVariables(): undefined var → *render.UndefinedVariableError.
func TestS10E2_UndefinedVar_StrictMode_ReturnsError(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	err := s10renderErr(t, eng, `{{ missing_var_e2 }}`, map[string]any{})
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "strict mode must produce *render.UndefinedVariableError, got %T", err)
}

// E3 — Name field set to the root variable name (not a property path).
func TestS10E3_UndefinedVar_NameField_Simple(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	err := s10renderErr(t, eng, `{{ my_missing_var }}`, map[string]any{})
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue))
	assert.Equal(t, "my_missing_var", ue.RootName)
}

func TestS10E3_UndefinedVar_NameField_DottedAccessPreservesRoot(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	// user.name → root Name should be "user"
	err := s10renderErr(t, eng, `{{ user.name }}`, map[string]any{})
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue))
	assert.Equal(t, "user", ue.RootName)
}

// E4 — Line number and markup context correct.
func TestS10E4_UndefinedVar_LineNumber(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	src := "before\n{{ missing_e4 }}\nafter"
	err := s10renderErr(t, eng, src, map[string]any{})
	assert.Contains(t, err.Error(), "line 2")
}

func TestS10E4_UndefinedVar_MarkupContext(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	err := s10renderErr(t, eng, `{{ my_undefined_e4 }}`, map[string]any{})
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue))
	// MarkupContext should be the expression source
	assert.Contains(t, ue.MarkupContext(), "my_undefined_e4")
}

// E5 — Per-render WithStrictVariables() works the same as engine-level.
func TestS10E5_UndefinedVar_PerRender_WithStrictVariables(t *testing.T) {
	eng := s10eng(t) // engine is non-strict
	tpl := s10mustParse(t, eng, `{{ missing_e5 }}`)
	// Per-render option enforces strict
	_, err := tpl.RenderString(map[string]any{}, liquid.WithStrictVariables())
	require.Error(t, err)
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "WithStrictVariables() must produce UndefinedVariableError")
}

func TestS10E5_UndefinedVar_PerRender_WithStrictVariables_DefinedIsOk(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ defined_e5 }}`)
	out, err := tpl.RenderString(map[string]any{"defined_e5": "hello"}, liquid.WithStrictVariables())
	require.NoError(t, err)
	assert.Equal(t, "hello", string(out))
}

// E6 — errors.As chain walking: UndefinedVariableError findable from outer error.
func TestS10E6_UndefinedVar_ErrorsAs_Chain(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	err := s10renderErr(t, eng, `{{ e6_var }}`, map[string]any{})
	// Must find via errors.As regardless of intermediate wrapping
	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "UndefinedVariableError must be findable via errors.As, got %T", err)
}

// E7 — Defined variable with StrictVariables: no error, correct output.
func TestS10E7_UndefinedVar_DefinedVar_NoError(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	out := s10render(t, eng, `{{ greeting_e7 }}`, map[string]any{"greeting_e7": "hi"})
	assert.Equal(t, "hi", out)
}

// E8 — Error prefix for UndefinedVariableError is "Liquid error", not "Liquid syntax error".
func TestS10E8_UndefinedVar_ErrorPrefix(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	err := s10renderErr(t, eng, `{{ missing_e8 }}`, map[string]any{})
	assert.Contains(t, err.Error(), "Liquid error")
	assert.NotContains(t, err.Error(), "Liquid syntax error")
}

// ═════════════════════════════════════════════════════════════════════════════
// F. WithErrorHandler (exception_renderer)
// ═════════════════════════════════════════════════════════════════════════════

// F1 — Handler output replaces the failing node text.
func TestS10F1_ErrorHandler_ReplacesFailingNode(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_f1", func(v any) (any, error) {
		return nil, errors.New("boom")
	})
	tpl := s10mustParse(t, eng, `before {{ "x" | fail_f1 }} after`)
	out, err := tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		return "[ERROR]"
	}))
	require.NoError(t, err, "handler must absorb the error")
	assert.Equal(t, "before [ERROR] after", out)
}

// F2 — Rendering continues after the failing node.
func TestS10F2_ErrorHandler_ContinuesAfterFailure(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_f2", func(v any) (any, error) {
		return nil, errors.New("f2 failure")
	})
	tpl := s10mustParse(t, eng, `A{{ "x" | fail_f2 }}B{{ "y" | upcase }}C`)
	out, err := tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		return "X"
	}))
	require.NoError(t, err)
	assert.Equal(t, "AXBYC", out)
}

// F3 — Multiple errors handled; output assembled in order.
func TestS10F3_ErrorHandler_MultipleErrors(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_f3", func(v any) (any, error) {
		return nil, errors.New("f3 error")
	})
	tpl := s10mustParse(t, eng, `{{ 1 | fail_f3 }}+{{ 2 | fail_f3 }}+{{ 3 | fail_f3 }}`)

	var collected []error
	out, err := tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		collected = append(collected, e)
		return "E"
	}))
	require.NoError(t, err)
	assert.Equal(t, "E+E+E", out)
	assert.Len(t, collected, 3, "handler must be called once per failing node")
}

// F4 — Handler receives the error; errors.As works inside handler.
func TestS10F4_ErrorHandler_ReceivesTypedError(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_f4", func(v any) (any, error) {
		return nil, render.NewArgumentError("typed arg error")
	})
	tpl := s10mustParse(t, eng, `{{ "x" | fail_f4 }}`)

	var sawArgErr bool
	_, err := tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		var ae *render.ArgumentError
		if errors.As(e, &ae) {
			sawArgErr = true
		}
		return ""
	}))
	require.NoError(t, err)
	assert.True(t, sawArgErr, "handler must receive the ArgumentError through the chain")
}

// F5 — Parse errors are NOT caught by the render error handler.
func TestS10F5_ErrorHandler_ParseErrorsNotCaught(t *testing.T) {
	eng := s10eng(t)
	// Parse error (unclosed block) happens before render; handler cannot intercept it
	_, parseErr := eng.ParseString(`{% for x in arr %}`)
	require.Error(t, parseErr, "a parse error must occur")
	// The error must be a parse error, not absorbed by any handler
	var pe *parser.ParseError
	require.True(t, errors.As(parseErr, &pe))
}

// F6 — Non-erroring nodes render correctly alongside failing nodes.
func TestS10F6_ErrorHandler_HealthyNodesUnaffected(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_f6", func(v any) (any, error) {
		return nil, errors.New("failure")
	})
	tpl := s10mustParse(t, eng, `{{ greeting }} world {{ "x" | fail_f6 }} !!`)
	out, err := tpl.RenderString(map[string]any{"greeting": "hello"},
		liquid.WithErrorHandler(func(e error) string { return "" }))
	require.NoError(t, err)
	assert.Equal(t, "hello world  !!", out)
}

// ═════════════════════════════════════════════════════════════════════════════
// G. markup_context metadata
// ═════════════════════════════════════════════════════════════════════════════

// G1 — Error() shows markup context of failing expression when no path set.
func TestS10G1_MarkupContext_InErrorString_WhenNoPath(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ product.cost | divided_by: 0 }}`)
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	// No path → markup context should appear in Error() string
	assert.Contains(t, err.Error(), "product.cost")
}

// G2 — When multiple nodes fail, each carries its own markup context.
func TestS10G2_MarkupContext_EachNodeHasOwnContext(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("ctx_fail_g2", func(v any) (any, error) {
		return nil, errors.New("ctx_fail_g2 err")
	})

	var contexts []string
	tpl := s10mustParse(t, eng, `{{ alpha | ctx_fail_g2 }} {{ beta | ctx_fail_g2 }}`)
	_, _ = tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		var re *render.RenderError
		if errors.As(e, &re) {
			contexts = append(contexts, re.MarkupContext())
		}
		return ""
	}))
	// Each of the two failing nodes must have a different markup context
	require.Len(t, contexts, 2)
	assert.NotEqual(t, contexts[0], contexts[1], "each node must have its own markup context")
	assert.Contains(t, contexts[0], "alpha")
	assert.Contains(t, contexts[1], "beta")
}

// G3 — Inner markup context preserved over outer block source in nested structure.
func TestS10G3_MarkupContext_InnerPreservedThroughBlock(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, "{% if true %}\n  {{ 1 | divided_by: 0 }}\n{% endif %}")
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	var re *render.RenderError
	require.True(t, errors.As(err, &re))
	// MarkupContext must refer to the inner {{ expr }}, not the outer {% if %}
	mc := re.MarkupContext()
	assert.Contains(t, mc, "divided_by",
		"inner markup context must be preserved over outer block context, got: %q", mc)
	assert.NotContains(t, mc, "if true",
		"outer block source must NOT overwrite inner context, got: %q", mc)
}

// G4 — MarkupContext() returns empty string when error has no locatable info.
func TestS10G4_MarkupContext_EmptyWhenNoSource(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{}, // no pathname, no line
		Source:    "",
	}
	err := parser.Errorf(&tok, "some error")
	assert.Equal(t, "", err.MarkupContext())
}

// ═════════════════════════════════════════════════════════════════════════════
// H. Error chain walking
// ═════════════════════════════════════════════════════════════════════════════

// H1 — ZeroDivisionError walkable from top-level error without knowing intermediate types.
func TestS10H1_Chain_ZeroDivision_Walkable(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, `{{ 8 | divided_by: 0 }}`)
	_, top := tpl.RenderString(nil)
	require.Error(t, top)
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(top, &zde),
		"ZeroDivisionError must be findable via errors.As from top-level error, chain: %T → %v", top, top)
}

// H2 — ArgumentError walkable from top-level error.
func TestS10H2_Chain_ArgumentError_Walkable(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("chain_test_h2", func(v any) (any, error) {
		return nil, render.NewArgumentError("chain arg error")
	})
	top := s10renderErr(t, eng, `{{ 1 | chain_test_h2 }}`, nil)
	var ae *render.ArgumentError
	require.True(t, errors.As(top, &ae),
		"ArgumentError must be findable via errors.As from top-level error, got %T", top)
}

// H3 — *render.RenderError always present in chain for render-time failures.
func TestS10H3_Chain_RenderError_AlwaysPresent(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{"zero_division", `{{ 1 | divided_by: 0 }}`},
		{"modulo_zero", `{{ 5 | modulo: 0 }}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			eng := s10eng(t)
			tpl := s10mustParse(t, eng, tc.src)
			_, err := tpl.RenderString(nil)
			require.Error(t, err)
			var re *render.RenderError
			require.True(t, errors.As(err, &re),
				"*render.RenderError must be in chain for %s, got %T", tc.name, err)
		})
	}
}

// H4 — *parser.ParseError always present in chain for parse-time failures.
func TestS10H4_Chain_ParseError_AlwaysPresent(t *testing.T) {
	testCases := []struct {
		name string
		src  string
	}{
		{"unclosed_for", `{% for x in y %}`},
		{"unclosed_if", `{% if cond %}`},
		{"unknown_tag", `{% no_such_tag_h4 %}`},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s10eng(t).ParseString(tc.src)
			require.Error(t, err)
			var pe *parser.ParseError
			require.True(t, errors.As(err, &pe),
				"*parser.ParseError must be in chain for %s, got %T", tc.name, err)
		})
	}
}

// ═════════════════════════════════════════════════════════════════════════════
// I. Prefix invariants (regression guard)
// ═════════════════════════════════════════════════════════════════════════════

// I1 — Every parse-time error starts with "Liquid syntax error".
func TestS10I1_Prefix_AllParseErrors_HaveSyntaxErrorPrefix(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"unclosed_for", `{% for a in b %}`},
		{"unclosed_if", `{% if x %}`},
		{"unclosed_case", `{% case x %}`},
		{"unclosed_unless", `{% unless x %}`},
		{"unclosed_capture", `{% capture v %}`},
		{"unknown_tag_solo", `{% xyz_notregistered %}`},
		{"unknown_tag_in_if", "{% if true %}\n{% xyz_in_if %}\n{% endif %}"},
		{"invalid_operator", `{% if 1 =! 2 %}y{% endif %}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := s10eng(t).ParseString(c.src)
			require.Error(t, err, "expected parse error for %q", c.src)
			assert.True(t,
				strings.HasPrefix(err.Error(), "Liquid syntax error"),
				"parse error must start with 'Liquid syntax error', got: %q", err.Error())
		})
	}
}

// I2 — Every render-time error starts with "Liquid error", never "Liquid syntax error".
func TestS10I2_Prefix_AllRenderErrors_HaveLiquidErrorPrefix(t *testing.T) {
	buildEng := func(t *testing.T) *liquid.Engine {
		t.Helper()
		eng := s10eng(t)
		eng.RegisterFilter("fail_i2", func(v any) (any, error) {
			return nil, errors.New("i2 render-time failure")
		})
		eng.RegisterTag("tag_fail_i2", func(c render.Context) (string, error) {
			return "", errors.New("i2 tag failure")
		})
		return eng
	}

	cases := []struct {
		name string
		src  string
	}{
		{"zero_division", `{{ 1 | divided_by: 0 }}`},
		{"modulo_zero", `{{ 1 | modulo: 0 }}`},
		{"filter_fails", `{{ "x" | fail_i2 }}`},
		{"tag_fails", `{% tag_fail_i2 %}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			eng := buildEng(t)
			tpl, parseErr := eng.ParseString(c.src)
			require.NoError(t, parseErr)
			_, err := tpl.RenderString(nil)
			require.Error(t, err)
			assert.True(t,
				strings.HasPrefix(err.Error(), "Liquid error"),
				"render error must start with 'Liquid error', got: %q", err.Error())
			assert.False(t,
				strings.HasPrefix(err.Error(), "Liquid syntax error"),
				"render error must NOT start with 'Liquid syntax error', got: %q", err.Error())
		})
	}
}

// I3 — Render error with line N includes "(line N)" in Error() string.
func TestS10I3_Prefix_RenderError_LineN_InString(t *testing.T) {
	eng := s10eng(t)
	tpl := s10mustParse(t, eng, "ok\n{{ 1 | divided_by: 0 }}")
	_, err := tpl.RenderString(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "(line 2)",
		"render error on line 2 must contain '(line 2)', got: %q", err.Error())
}

// I4 — Parse error with line N includes "(line N)" in Error() string.
func TestS10I4_Prefix_ParseError_LineN_InString(t *testing.T) {
	src := "ok\nok\n{% for_never_closed %}"
	err := s10parseErr(t, src)
	assert.Contains(t, err.Error(), "(line 3)",
		"parse error on line 3 must contain '(line 3)', got: %q", err.Error())
}

// ═════════════════════════════════════════════════════════════════════════════
// Integration: realistic templates combining multiple section-10 features
// ═════════════════════════════════════════════════════════════════════════════

// TestS10_Integration_ErrorHandlerCollectsAllErrors demonstrates the canonical
// pattern for collecting all render errors without stopping the output.
func TestS10_Integration_ErrorHandlerCollectsAllErrors(t *testing.T) {
	eng := s10eng(t)
	eng.RegisterFilter("fail_collect", func(v any) (any, error) {
		return nil, fmt.Errorf("item %v failed", v)
	})

	src := "start\n{{ 1 | fail_collect }}\nmiddle\n{{ 2 | fail_collect }}\nend"
	tpl := s10mustParse(t, eng, src)

	var errs []error
	out, err := tpl.RenderString(nil, liquid.WithErrorHandler(func(e error) string {
		errs = append(errs, e)
		return ""
	}))
	require.NoError(t, err)
	assert.Equal(t, "start\n\nmiddle\n\nend", out)
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0].Error(), "1")
	assert.Contains(t, errs[1].Error(), "2")
}

// TestS10_Integration_StrictVariables_MultipleUndefined collects all
// UndefinedVariableErrors from a template in a single render via handler.
func TestS10_Integration_StrictVariables_MultipleUndefined(t *testing.T) {
	eng := s10eng(t)
	eng.StrictVariables()
	tpl := s10mustParse(t, eng, `{{ a }} and {{ b }} and {{ c }}`)

	var names []string
	out, err := tpl.RenderString(map[string]any{}, liquid.WithErrorHandler(func(e error) string {
		var ue *render.UndefinedVariableError
		if errors.As(e, &ue) {
			names = append(names, ue.RootName)
		}
		return "?"
	}))
	require.NoError(t, err)
	assert.Equal(t, "? and ? and ?", out)
	require.Len(t, names, 3)
	assert.Contains(t, names, "a")
	assert.Contains(t, names, "b")
	assert.Contains(t, names, "c")
}

// TestS10_Integration_ZeroDivision_LineAndContext verifies that a ZeroDivision
// error in a multi-line template has correct line number AND markup context.
func TestS10_Integration_ZeroDivision_LineAndContext(t *testing.T) {
	eng := s10eng(t)
	src := "{% assign price = 100 %}\n{% assign discount = 0 %}\n{{ price | divided_by: discount }}"
	tpl := s10mustParse(t, eng, src)
	_, err := tpl.RenderString(map[string]any{"price": 100, "discount": 0})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "line 3", "error must be on line 3")
	assert.Contains(t, err.Error(), "divided_by", "error must mention the filter")

	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(err, &zde))
}

// TestS10_Integration_NestedBlock_ErrorBubbles validates that an error deep in
// a nested block structure carries accurate line and context metadata.
func TestS10_Integration_NestedBlock_ErrorBubbles(t *testing.T) {
	eng := s10eng(t)
	src := "{% if true %}\n  {% for i in arr %}\n    {{ i | divided_by: 0 }}\n  {% endfor %}\n{% endif %}"
	tpl := s10mustParse(t, eng, src)
	_, err := tpl.RenderString(map[string]any{"arr": []int{1}})
	require.Error(t, err)

	// Error must be attributed to line 3 (the divided_by: 0 expression)
	assert.Contains(t, err.Error(), "line 3")
	// The inner markup context must be preserved (not replaced by {% for %} or {% if %} source)
	var re *render.RenderError
	require.True(t, errors.As(err, &re))
	mc := re.MarkupContext()
	assert.Contains(t, mc, "divided_by",
		"inner context must survive bubbling through nested blocks: %q", mc)
}
