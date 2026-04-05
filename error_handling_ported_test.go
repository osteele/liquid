package liquid

// Ported error-handling tests from:
//   - Ruby Liquid: test/integration/error_handling_test.rb
//   - LiquidJS:    test/integration/misc/error.spec.ts
//   - LiquidJS:    src/util/error.spec.ts
//
// Covers checklist section 10: Tratamento de Erros
//   10.1  SourceError with Path(), LineNumber(), Cause()    — ✅ (extended here)
//   10.2  ZeroDivisionError typed error                     — ✅ (tested in filters/ and engine_test.go, not re-ported here)
//   10.3  SyntaxError, ArgumentError, ContextError types    — ✅ (ported here)
//   10.4  markup_context metadata                           — ✅ (ported here)

import (
	"errors"
	"strings"
	"testing"

	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"

	"github.com/stretchr/testify/require"
)

// ── 10.1 SourceError / ParseError extended ───────────────────────────────────
// Ruby: test_missing_endtag_parse_time_error, test_with_line_numbers_adds_numbers_to_parser_errors

// TestPortedErrors_ParseError_SyntaxErrorPrefix verifies that parse-time errors
// carry the "Liquid syntax error" prefix, matching Ruby Liquid behaviour.
// Ruby: test_missing_endtag_parse_time_error
//
//	assert_match_syntax_error(/: 'for' tag was never closed\z/, ' {% for a in b %} ... ')
func TestPortedErrors_ParseError_SyntaxErrorPrefix(t *testing.T) {
	engine := NewEngine()

	t.Run("unclosed_for_is_syntax_error", func(t *testing.T) {
		_, err := engine.ParseString(`{% for a in b %} ... `)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Liquid syntax error")
		require.Contains(t, err.Error(), "for")
	})

	t.Run("unclosed_if_is_syntax_error", func(t *testing.T) {
		_, err := engine.ParseString(`{% if test %}`)
		require.Error(t, err)
		require.Contains(t, err.Error(), "Liquid syntax error")
	})

	t.Run("parse_error_does_not_say_liquid_error", func(t *testing.T) {
		_, err := engine.ParseString(`{% for a in b %}`)
		require.Error(t, err)
		// Must NOT start with plain "Liquid error:" — it is a syntax error
		require.NotContains(t, err.Error(), "Liquid error:")
	})
}

// TestPortedErrors_ParseError_LineNumber verifies line numbers appear in parse
// errors when the source is multi-line.
// Ruby: test_with_line_numbers_adds_numbers_to_parser_errors
//
//	assert_match_syntax_error(/Liquid syntax error \(line 3\)/, source)
func TestPortedErrors_ParseError_LineNumber(t *testing.T) {
	engine := NewEngine()

	src := "foobar\n\n{% unclosed_block_goes_here %}"
	_, err := engine.ParseString(src)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
	require.Contains(t, err.Error(), "line 3")
}

// TestPortedErrors_ParseError_LineNumber_Nested verifies line numbers are
// correct for errors inside nested blocks.
// Ruby: test_syntax_errors_in_nested_blocks_have_correct_line_number
//
//	assert_match_syntax_error("Liquid syntax error (line 4): Unknown tag 'foo'", source)
func TestPortedErrors_ParseError_LineNumber_Nested(t *testing.T) {
	engine := NewEngine()

	src := "foobar\n\n{% if 1 != 2 %}\n  {% foo %}\n{% endif %}\n\nbla"
	_, err := engine.ParseString(src)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
	require.Contains(t, err.Error(), "line 4")
}

// ── 10.3 SyntaxError type alias ───────────────────────────────────────────────

// TestPortedErrors_SyntaxError_Alias confirms that *parser.SyntaxError is
// identical to *parser.ParseError for errors.As matching.
// SyntaxError = ParseError (type alias), so both patterns must work.
func TestPortedErrors_SyntaxError_Alias(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseString(`{% if unclosed %}`)
	require.Error(t, err)

	// *parser.ParseError must match
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "expected *parser.ParseError, got %T", err)

	// *parser.SyntaxError must also match (it IS ParseError — type alias)
	var se *parser.SyntaxError
	require.True(t, errors.As(err, &se), "expected *parser.SyntaxError, got %T", err)

	// Both pointers should be the same underlying object
	require.Equal(t, pe, se)
}

// ── 10.3 ArgumentError type ───────────────────────────────────────────────────

// TestPortedErrors_ArgumentError_FilterReturnsIt shows that when a filter
// returns a *render.ArgumentError, errors.As can detect it in the chain and
// the top-level error message still uses the "Liquid error:" prefix.
// Ruby: test_argument — "Liquid error: argument error"
func TestPortedErrors_ArgumentError_FilterReturnsIt(t *testing.T) {
	engine := NewEngine()
	// Register a filter that raises an ArgumentError
	engine.RegisterFilter("bad_args", func(n any) (any, error) {
		return nil, render.NewArgumentError("invalid argument supplied")
	})

	_, err := engine.ParseAndRenderString(`{{ val | bad_args }}`, map[string]any{"val": 10})
	require.Error(t, err)

	// Must be detectable via errors.As
	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae), "expected *render.ArgumentError in chain, got %T", err)
	require.Contains(t, ae.Error(), "invalid argument supplied")

	// The top-level error should be a render error ("Liquid error") not a parse error ("Liquid syntax error")
	require.Contains(t, err.Error(), "Liquid error")
	require.NotContains(t, err.Error(), "Liquid syntax error:")
}

// TestPortedErrors_ArgumentError_TagReturnsIt shows that a tag renderer can
// also return ArgumentError and it surfaces correctly.
func TestPortedErrors_ArgumentError_TagReturnsIt(t *testing.T) {
	engine := NewEngine()
	engine.RegisterTag("bad_arg_tag", func(c render.Context) (string, error) {
		return "", render.NewArgumentError("tag got bad arg")
	})

	_, err := engine.ParseAndRenderString(`{% bad_arg_tag %}`, map[string]any{})
	require.Error(t, err)

	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae), "expected *render.ArgumentError in chain, got %T", err)
}

// ── 10.3 ContextError type ────────────────────────────────────────────────────

// TestPortedErrors_ContextError_TagReturnsIt shows that a tag returning a
// *render.ContextError is detectable via errors.As in the chain.
func TestPortedErrors_ContextError_TagReturnsIt(t *testing.T) {
	engine := NewEngine()
	engine.RegisterTag("ctx_error_tag", func(c render.Context) (string, error) {
		return "", render.NewContextError("context lookup failed")
	})

	_, err := engine.ParseAndRenderString(`{% ctx_error_tag %}`, map[string]any{})
	require.Error(t, err)

	var ce *render.ContextError
	require.True(t, errors.As(err, &ce), "expected *render.ContextError in chain, got %T", err)
	require.Contains(t, ce.Error(), "context lookup failed")
}

// ── 10.4 markup_context metadata ─────────────────────────────────────────────

// TestPortedErrors_ParseError_Message verifies that Message() returns the
// error text without the "Liquid syntax error" prefix or location info.
// Ruby: error.to_s(false) — returns message without prefix
func TestPortedErrors_ParseError_Message(t *testing.T) {
	engine := NewEngine()
	_, err := engine.ParseString(`{% for a in b %}`)
	require.Error(t, err)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))

	msg := pe.Message()
	// Must contain the error text
	require.NotEmpty(t, msg)
	// Must NOT contain the "Liquid" prefix
	require.NotContains(t, msg, "Liquid syntax error")
	require.NotContains(t, msg, "Liquid error")
	// Must NOT contain "(line N)" location
	require.NotContains(t, msg, "(line ")
}

// TestPortedErrors_ParseError_MarkupContext verifies that MarkupContext()
// surfaces the source text of the offending tag/expression.
// Ruby: error.markup_context — "in tag '{% for a in b %}'"
func TestPortedErrors_ParseError_MarkupContext(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{Pathname: "t.html", LineNo: 3},
		Source:    `{% bad_tag with_args %}`,
	}
	err := parser.Errorf(&tok, "unknown tag")

	// MarkupContext returns the token's Source field
	require.Equal(t, `{% bad_tag with_args %}`, err.MarkupContext())

	// When no pathname, the source context appears in Error() output
	tokNoPath := parser.Token{
		SourceLoc: parser.SourceLoc{LineNo: 1},
		Source:    `{% no_path_tag %}`,
	}
	errNoPath := parser.Errorf(&tokNoPath, "unrecognised")
	require.Contains(t, errNoPath.Error(), `{% no_path_tag %}`)
}

// TestPortedErrors_RenderError_Message verifies Message() on a render error.
func TestPortedErrors_RenderError_Message(t *testing.T) {
	engine := NewEngine()
	tpl, parseErr := engine.ParseString(`{{ 10 | divided_by: 0 }}`)
	require.NoError(t, parseErr)

	_, renderErr := tpl.RenderString(map[string]any{})
	require.Error(t, renderErr)

	var re *render.RenderError
	require.True(t, errors.As(renderErr, &re), "expected *render.RenderError")

	// Message() should not contain the "Liquid error" prefix
	require.NotContains(t, re.Message(), "Liquid error")
	require.NotContains(t, re.Message(), "(line ")

	// The full Error() string should still have "Liquid error"
	require.Contains(t, renderErr.Error(), "Liquid error")
}

// TestPortedErrors_RenderError_LiquidErrorPrefix verifies that render errors
// use "Liquid error" prefix even when they arise from (internally) wrapping a
// parser.ParseError. We do NOT want "Liquid syntax error" in a render path.
// Ruby: test_standard_error — "Liquid error: standard error"
func TestPortedErrors_RenderError_LiquidErrorPrefix(t *testing.T) {
	engine := NewEngine()

	// ZeroDivision is a render-time error
	tpl, parseErr := engine.ParseString(`{{ x | divided_by: 0 }}`)
	require.NoError(t, parseErr)
	_, err := tpl.RenderString(map[string]any{"x": 10})
	require.Error(t, err)

	require.Contains(t, err.Error(), "Liquid error")
	require.NotContains(t, err.Error(), "Liquid syntax error")
}

// ── 10.1 Error chain walking ──────────────────────────────────────────────────

// TestPortedErrors_ErrorChain_ZeroDivision verifies the full error chain
// can be walked via errors.As from the top-level engine error.
// Ruby: test demonstrates ZeroDivisionError is a specific error type.
func TestPortedErrors_ErrorChain_ZeroDivision(t *testing.T) {
	engine := NewEngine()

	tpl, parseErr := engine.ParseString(`{{ 10 | divided_by: 0 }}`)
	require.NoError(t, parseErr)
	_, renderErr := tpl.RenderString(map[string]any{})
	require.Error(t, renderErr)

	// Outer wrapper is RenderError
	var re *render.RenderError
	require.True(t, errors.As(renderErr, &re), "expected *render.RenderError in chain")

	// Inner cause is ZeroDivisionError
	var zde *filters.ZeroDivisionError
	require.True(t, errors.As(renderErr, &zde), "expected *filters.ZeroDivisionError in chain")
}

// TestPortedErrors_ErrorChain_ArgumentError confirms the chain walk for
// ArgumentError through RenderError.
func TestPortedErrors_ErrorChain_ArgumentError(t *testing.T) {
	engine := NewEngine()
	engine.RegisterFilter("chain_test_arg", func(n any) (any, error) {
		return nil, render.NewArgumentError("bad arg")
	})

	_, err := engine.ParseAndRenderString(`{{ 1 | chain_test_arg }}`, nil)
	require.Error(t, err)

	// Must find RenderError in chain
	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	// Must find ArgumentError deeper in chain
	var ae *render.ArgumentError
	require.True(t, errors.As(err, &ae))
}

// TestPortedErrors_ErrorChain_UndefinedVariable confirms that UndefinedVariableError
// carries Name field and is walkable.
// Ruby: UndefinedVariable error carries the variable name.
func TestPortedErrors_ErrorChain_UndefinedVariable(t *testing.T) {
	engine := NewEngine()
	engine.StrictVariables()

	_, err := engine.ParseAndRenderString(`{{ my_missing_var }}`, map[string]any{})
	require.Error(t, err)

	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "expected *render.UndefinedVariableError in chain")
	require.Equal(t, "my_missing_var", ue.Name)

	// Message() includes the variable name but not a "Liquid" prefix
	require.Contains(t, ue.Message(), "my_missing_var")
	require.NotContains(t, ue.Message(), "Liquid error")
}

// ── 10.1 ParseError with path vs without path ─────────────────────────────────

// TestPortedErrors_ParseError_Path verifies that when a template file path is
// set, the error includes the path (not the raw source text).
func TestPortedErrors_ParseError_Path(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{Pathname: "products/detail.html", LineNo: 7},
		Source:    `{% bad %}`,
	}
	err := parser.Errorf(&tok, "unknown tag 'bad'")

	require.Equal(t, "products/detail.html", err.Path())
	require.Equal(t, 7, err.LineNumber())
	require.Contains(t, err.Error(), "products/detail.html")
	require.Contains(t, err.Error(), "line 7")
	require.Contains(t, err.Error(), "unknown tag 'bad'")
}

// TestPortedErrors_ParseError_NoPath verifies that without a path, the
// source text (markup context) appears in the error string.
func TestPortedErrors_ParseError_NoPath(t *testing.T) {
	tok := parser.Token{
		SourceLoc: parser.SourceLoc{LineNo: 2},
		Source:    `{{ product.price | divided_by: 0 }}`,
	}
	err := parser.Errorf(&tok, "divided by 0")

	require.Equal(t, "", err.Path())
	require.Equal(t, 2, err.LineNumber())
	// Source text should appear as the location context
	require.Contains(t, err.Error(), `{{ product.price | divided_by: 0 }}`)
}

// ── 10.1 Line numbers with whitespace-trim markers ────────────────────────────

// TestPortedErrors_ParseError_LineNumber_WhitespaceTrim verifies that using
// the whitespace-trim markers ({%- -%}) does NOT shift reported line numbers.
// Ruby: test_with_line_numbers_adds_numbers_to_parser_errors_with_whitespace_trim
//
//	assert_match_syntax_error(/Liquid syntax error \(line 3\)/, source)
//
// The source has `{%- "cat" | foobar -%}` on line 3. It must still report
// "line 3" regardless of the trim dashes.
func TestPortedErrors_ParseError_LineNumber_WhitespaceTrim(t *testing.T) {
	engine := NewEngine()

	// "cat" is a literal, not a value that can be piped into an unknown tag context,
	// and "foobar" is an unknown filter — this should produce a parse/render error on line 3.
	src := "foobar\n\n{%- assign x = 1 -%}\n{%- nosuchtagxyz -%}\n\nbla"
	_, err := engine.ParseString(src)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
	require.Contains(t, err.Error(), "line 4",
		"trim markers must not shift line numbers: got %q", err.Error())
}

// TestPortedErrors_ParseError_LineNumber_WhitespaceTrimSimple validates the
// exact Ruby test case: tag on line 3 with trim markers still reports line 3.
// Ruby: foobar\n\n{%- "cat" | foobar -%}\n\nbla  → "Liquid syntax error (line 3)"
func TestPortedErrors_ParseError_LineNumber_WhitespaceTrimSimple(t *testing.T) {
	engine := NewEngine()

	// Use the unknown-tag variant since Go's expression errors are parse-time
	src := "foobar\n\n{%- nosuchtagxyz -%}\n\nbla"
	_, err := engine.ParseString(src)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
	require.Contains(t, err.Error(), "line 3",
		"trim markers must not shift line numbers: got %q", err.Error())
}

// ── 10.3 Unrecognized operator ────────────────────────────────────────────────

// TestPortedErrors_UnrecognizedOperator_IsSyntaxError documents that the `=!`
// operator (invalid in both Ruby strict mode and Go) produces a ParseError.
// Ruby: test_unrecognized_operator (strict) — raises SyntaxError
// Go always produces a ParseError for malformed expressions regardless of mode.
func TestPortedErrors_UnrecognizedOperator_IsSyntaxError(t *testing.T) {
	engine := NewEngine()

	_, err := engine.ParseString(`{% if 1 =! 2 %}ok{% endif %}`)
	require.Error(t, err, "=! is not a valid operator and should cause an error")
	// In Go, this is always a parse-time error
	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "expected *parser.ParseError, got %T: %v", err, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
}

// ── 10.3 SyntaxError raised from a Drop ──────────────────────────────────────

// TestPortedErrors_Drop_SyntaxErrorSurfacesCorrectly confirms that when a Drop
// raises a SyntaxError (parser.ParseError), the error is detectable via errors.As
// and carries the correct prefix.
// Ruby: test_syntax — template renders ' Liquid syntax error: syntax error '
// In Go, this is tested by a custom tag that returns a *parser.ParseError.
func TestPortedErrors_Drop_SyntaxErrorSurfacesCorrectly(t *testing.T) {
	engine := NewEngine()
	engine.RegisterFilter("raise_syntax_err", func(v any) (any, error) {
		// Simulate a drop that raises SyntaxError — return a *parser.ParseError
		// (which is also *parser.SyntaxError via type alias)
		tok := parser.Token{
			SourceLoc: parser.SourceLoc{LineNo: 1},
			Source:    `{{ errors.syntax_error }}`,
		}
		return nil, parser.Errorf(&tok, "syntax error")
	})

	_, err := engine.ParseAndRenderString(`{{ "x" | raise_syntax_err }}`, map[string]any{})
	require.Error(t, err)

	// Must be detectable as SyntaxError (= ParseError alias)
	var se *parser.SyntaxError
	require.True(t, errors.As(err, &se), "expected *parser.SyntaxError in error chain, got %T: %v", err, err)
}

// ── JS: RenderError from a plain-error filter ────────────────────────────────

// TestPortedErrors_FilterPlainError_WrappedInRenderError verifies that a filter
// returning a plain (non-ArgumentError) error is still wrapped in *render.RenderError.
// JS: "should throw RenderError when filter throws"
func TestPortedErrors_FilterPlainError_WrappedInRenderError(t *testing.T) {
	engine := NewEngine()
	engine.RegisterFilter("throwing_filter", func(v any) (any, error) {
		return nil, errors.New("intended filter error")
	})

	_, err := engine.ParseAndRenderString(`{{ "hello" | throwing_filter }}`, map[string]any{})
	require.Error(t, err)

	var re *render.RenderError
	require.True(t, errors.As(err, &re), "plain error from filter must be wrapped in *render.RenderError, got %T: %v", err, err)
	require.Contains(t, err.Error(), "intended filter error")
	require.Contains(t, err.Error(), "Liquid error")
}

// ── JS: ParseError for unknown tag ───────────────────────────────────────────

// TestPortedErrors_UnknownTag_IsParseError verifies that referencing an
// unregistered tag name produces a *parser.ParseError.
// JS: "should throw ParseError when tag not exist"
//
//	await expect(engine.parseAndRender('{% a %}')).rejects.toMatchObject({
//	    name: 'ParseError',
//	    message: expect.stringContaining('tag "a" not found')
//	})
func TestPortedErrors_UnknownTag_IsParseError(t *testing.T) {
	engine := NewEngine()

	_, err := engine.ParseString(`{% nosuchtagatall %}`)
	require.Error(t, err)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "unknown tag must produce *parser.ParseError, got %T: %v", err, err)
	require.Contains(t, err.Error(), "Liquid syntax error")
}

// TestPortedErrors_UnknownTag_MultiLine_CorrectLine verifies the line number is
// correct when an unknown tag appears mid-template.
// JS: "should throw ParseError when tag not found" (multi-line test)
//
//	src = '{%if true%}\naaa{%endif%}\n{% -a %}\n3'  → ParseError on line 3
func TestPortedErrors_UnknownTag_MultiLine_CorrectLine(t *testing.T) {
	engine := NewEngine()

	src := "{%if true%}\naaa{%endif%}\n{% nosuchtagmultiline %}\n3"
	_, err := engine.ParseString(src)
	require.Error(t, err)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "expected *parser.ParseError, got %T", err)
	require.Contains(t, err.Error(), "line 3")
}

// ── JS: RenderError from a tag renderer ──────────────────────────────────────

// TestPortedErrors_TagPlainError_WrappedInRenderError verifies that a tag
// returning a plain error from its renderer is wrapped in *render.RenderError.
// JS: "should throw RenderError when tag throws"
func TestPortedErrors_TagPlainError_WrappedInRenderError(t *testing.T) {
	engine := NewEngine()
	engine.RegisterTag("throwing_tag", func(c render.Context) (string, error) {
		return "", errors.New("intended tag error")
	})

	_, err := engine.ParseAndRenderString(`{% throwing_tag %}`, map[string]any{})
	require.Error(t, err)

	var re *render.RenderError
	require.True(t, errors.As(err, &re), "plain error from tag must be wrapped in *render.RenderError, got %T: %v", err, err)
	require.Contains(t, err.Error(), "intended tag error")
	require.Contains(t, err.Error(), "Liquid error")
}

// ── 10.1 UndefinedVariableError — strict mode prefix ────────────────────────

// TestPortedErrors_UndefinedVariable_LiquidErrorPrefix verifies that strict-mode
// UndefinedVariableError uses the "Liquid error" prefix, not "Liquid syntax error".
// JS: "should throw RenderError when variable not defined" →
//
//	name: 'UndefinedVariableError', message: 'undefined variable: a, line:1, col:3'
//
// Ruby: undefined variable errors are render-time (not parse-time) errors.
func TestPortedErrors_UndefinedVariable_LiquidErrorPrefix(t *testing.T) {
	engine := NewEngine()
	engine.StrictVariables()

	_, err := engine.ParseAndRenderString(`{{ undefined_var }}`, map[string]any{})
	require.Error(t, err)

	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "expected *render.UndefinedVariableError")

	// UndefinedVariableError is a render-time error
	require.Contains(t, err.Error(), "Liquid error")
	require.NotContains(t, err.Error(), "Liquid syntax error")
	require.Contains(t, err.Error(), "undefined_var")
}

// TestPortedErrors_UndefinedVariable_NoError_DefaultMode documents that
// by default (non-strict), undefined variables produce empty output, not errors.
// JS: "should not throw when variable undefined by default"
//
//	const html = await engine.parseAndRender('X{{a}}Y')
//	return expect(html).toBe('XY')
func TestPortedErrors_UndefinedVariable_NoError_DefaultMode(t *testing.T) {
	engine := NewEngine() // default: non-strict
	out, err := engine.ParseAndRenderString(`X{{ a }}Y`, map[string]any{})
	require.NoError(t, err, "undefined variable must not error in default mode")
	require.Equal(t, "XY", out)
}

// ── Integration: parse errors should NOT contain "Liquid error" ───────────────

// TestPortedErrors_PrefixDistinction ensures parse errors and render errors
// use different prefixes, matching Ruby Liquid's distinction between
// "Liquid syntax error" (parse) and "Liquid error" (render).
// Ruby: test_standard_error vs test_syntax
func TestPortedErrors_PrefixDistinction(t *testing.T) {
	engine := NewEngine()

	t.Run("parse_error_says_syntax_error", func(t *testing.T) {
		// Unclosed block is a parse error
		_, err := engine.ParseString(`{% if foo %}`)
		require.Error(t, err)
		require.True(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
			"expected 'Liquid syntax error' prefix, got: %q", err.Error())
	})

	t.Run("render_error_says_liquid_error", func(t *testing.T) {
		// ZeroDivision is a render error
		tpl, _ := engine.ParseString(`{{ 4 | divided_by: 0 }}`)
		_, err := tpl.RenderString(nil)
		require.Error(t, err)
		require.True(t, strings.HasPrefix(err.Error(), "Liquid error"),
			"expected 'Liquid error' prefix, got: %q", err.Error())
		require.False(t, strings.HasPrefix(err.Error(), "Liquid syntax error"),
			"render error must not start with 'Liquid syntax error', got: %q", err.Error())
	})
}
