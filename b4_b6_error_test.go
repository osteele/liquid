package liquid

// Intensive test suite for:
//
//   B4 — Distinct error types: ParseError, RenderError, UndefinedVariableError,
//         ZeroDivisionError, ArgumentError, ContextError.
//
//   B6 — Markup-context preserved when errors bubble up through block tags.
//        Before the fix, wrapping via BlockNode/for/if/etc. replaced the inner
//        `{{ expr }}` markup context with the outer `{% if … %}` source.
//        The fix: wrapRenderError now preserves any error that already has a
//        LineNumber > 0, so the most-specific inner context is kept.
//
// These tests have NO reference ports (Ruby/JS do not expose the same typed
// error API or markup-context field). They are original E2E specs.

import (
	"errors"
	"strings"
	"testing"

	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// renderErr is a helper that asserts rendering produces an error and returns it.
func renderErr(t *testing.T, eng *Engine, tpl string, binds map[string]any) error {
	t.Helper()
	_, err := eng.ParseAndRenderString(tpl, binds)
	require.Error(t, err, "expected error from template %q", tpl)
	return err
}

// parseErr is a helper that asserts parsing produces an error and returns it.
func parseErr(t *testing.T, eng *Engine, tpl string) error {
	t.Helper()
	_, err := eng.ParseString(tpl)
	require.Error(t, err, "expected parse error from template %q", tpl)
	return err
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  B4 – TYPED ERROR API                                                       ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// ── B4.1: ParseError (SyntaxError) ────────────────────────────────────────────

func TestB4_ParseError_Prefix(t *testing.T) {
	eng := NewEngine()

	cases := []struct {
		name, tpl string
	}{
		{"unclosed_for", `{% for x in arr %}`},
		{"unclosed_if", `{% if true %}`},
		{"unclosed_capture", `{% capture x %}`},
		{"unclosed_unless", `{% unless cond %}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := parseErr(t, eng, c.tpl)
			assert.Contains(t, err.Error(), "Liquid syntax error",
				"parse errors must carry 'Liquid syntax error' prefix")
			assert.NotContains(t, err.Error(), "Liquid error:",
				"parse errors must NOT say plain 'Liquid error:'")
		})
	}
}

func TestB4_ParseError_errorsAs(t *testing.T) {
	eng := NewEngine()
	err := parseErr(t, eng, `{% for x in arr %}`)

	var pe *parser.ParseError
	assert.True(t, errors.As(err, &pe), "errors.As must find *parser.ParseError, got %T", err)

	var se *parser.SyntaxError // type alias — same underlying type
	assert.True(t, errors.As(err, &se), "errors.As must find *parser.SyntaxError, got %T", err)

	// Both should refer to the same object (alias identity)
	require.Equal(t, pe, se)
}

func TestB4_ParseError_LineNumber(t *testing.T) {
	eng := NewEngine()

	t.Run("single_line", func(t *testing.T) {
		err := parseErr(t, eng, `{% unclosed_tag %}`)
		var pe *parser.ParseError
		require.True(t, errors.As(err, &pe))
		assert.Equal(t, 1, pe.LineNumber())
	})

	t.Run("multi_line_error_on_line3", func(t *testing.T) {
		src := "line1\nline2\n{% unclosed_tag %}"
		err := parseErr(t, eng, src)
		assert.Contains(t, err.Error(), "line 3")
	})

	t.Run("nested_block_reports_inner_line", func(t *testing.T) {
		// Unknown tag inside an if block — error should be line 4, not line 1.
		src := "l1\n\n{% if true %}\n  {% unknown_tag %}\n{% endif %}"
		err := parseErr(t, eng, src)
		assert.Contains(t, err.Error(), "line 4")
	})
}

func TestB4_ParseError_MarkupContext(t *testing.T) {
	eng := NewEngine()

	// MarkupContext should contain the exact source text of the failing token.
	src := `{% for x in arr %}`
	err := parseErr(t, eng, src)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))
	// The error string mentions the tag context.
	assert.NotEmpty(t, pe.MarkupContext(), "MarkupContext must not be empty")
}

func TestB4_ParseError_Message(t *testing.T) {
	eng := NewEngine()
	err := parseErr(t, eng, `{% for x in arr %}`)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe))

	msg := pe.Message()
	// Message() must not be empty and must not include the "Liquid syntax error" prefix
	assert.NotEmpty(t, msg)
	assert.NotContains(t, msg, "Liquid syntax error",
		"Message() must be the raw message without prefix/location")
	assert.NotContains(t, msg, "line ", "Message() must not include line info")
}

// ── B4.2: RenderError ─────────────────────────────────────────────────────────

func TestB4_RenderError_Prefix(t *testing.T) {
	eng := NewEngine()

	cases := []struct {
		name, tpl string
		binds     map[string]any
	}{
		{"divided_by_zero", `{{ x | divided_by: 0 }}`, map[string]any{"x": 4}},
		{"modulo_zero", `{{ x | modulo: 0 }}`, map[string]any{"x": 4}},
		{"custom_filter_error", `{{ x | bad }}`, map[string]any{"x": 1}},
	}

	// Register a filter that always fails.
	eng.RegisterFilter("bad", func(v any) (any, error) {
		return nil, render.NewArgumentError("always fails")
	})

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := renderErr(t, eng, c.tpl, c.binds)
			assert.Contains(t, err.Error(), "Liquid error",
				"render errors must carry 'Liquid error' prefix")
			assert.NotContains(t, err.Error(), "Liquid syntax error",
				"render errors must NOT say 'Liquid syntax error'")
		})
	}
}

func TestB4_RenderError_errorsAs(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ x | divided_by: 0 }}`, map[string]any{"x": 4})

	var re *render.RenderError
	assert.True(t, errors.As(err, &re), "errors.As must find *render.RenderError, got %T", err)
}

func TestB4_RenderError_LineNumber(t *testing.T) {
	eng := NewEngine()

	t.Run("single_line_is_1", func(t *testing.T) {
		err := renderErr(t, eng, `{{ x | divided_by: 0 }}`, map[string]any{"x": 4})
		var re *render.RenderError
		require.True(t, errors.As(err, &re))
		assert.Equal(t, 1, re.LineNumber())
		assert.Contains(t, err.Error(), "line 1")
	})

	t.Run("multi_line_correct_line", func(t *testing.T) {
		src := "ok\n{{ x | divided_by: 0 }}"
		err := renderErr(t, eng, src, map[string]any{"x": 4})
		assert.Contains(t, err.Error(), "line 2")
	})
}

func TestB4_RenderError_MarkupContext(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ price | divided_by: 0 }}`, map[string]any{"price": 100})

	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	mc := re.MarkupContext()
	assert.Contains(t, mc, "divided_by",
		"MarkupContext must contain the expression that failed")
	assert.Contains(t, mc, "{{",
		"MarkupContext must look like an object expression")
}

func TestB4_RenderError_Message(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ x | divided_by: 0 }}`, map[string]any{"x": 4})

	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	msg := re.Message()
	assert.NotEmpty(t, msg)
	assert.NotContains(t, msg, "Liquid error", "Message() must not include prefix")
	assert.NotContains(t, msg, "line ", "Message() must not include location")
}

func TestB4_RenderError_Cause_IsZeroDivision(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ x | divided_by: 0 }}`, map[string]any{"x": 4})

	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	cause := re.Cause()
	require.NotNil(t, cause, "Cause() must not be nil for a filter error")

	var zde *filters.ZeroDivisionError
	assert.True(t, errors.As(err, &zde),
		"errors.As must find *filters.ZeroDivisionError through the chain")
}

// ── B4.3: UndefinedVariableError ──────────────────────────────────────────────

func TestB4_UndefinedVariableError_StrictVariables(t *testing.T) {
	eng := NewEngine()

	_, err := eng.ParseAndRenderString(
		`{{ missing_var }}`,
		map[string]any{},
		WithStrictVariables(),
	)
	require.Error(t, err)

	var uve *render.UndefinedVariableError
	assert.True(t, errors.As(err, &uve),
		"errors.As must find *render.UndefinedVariableError, got %T", err)
	assert.Equal(t, "missing_var", uve.RootName,
		"UndefinedVariableError.Name must be the exact variable name")
}

func TestB4_UndefinedVariableError_NamePreservation(t *testing.T) {
	eng := NewEngine()

	cases := []struct {
		name, tpl, varName string
	}{
		{"simple", `{{ product }}`, "product"},
		// With dotted path: Name is the root variable ("user"), not the full path.
		{"with_dot", `{{ user.age }}`, "user"},
		{"with_filter", `{{ items | first }}`, "items"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := eng.ParseAndRenderString(c.tpl, map[string]any{}, WithStrictVariables())
			require.Error(t, err)
			var uve *render.UndefinedVariableError
			require.True(t, errors.As(err, &uve))
			assert.Contains(t, uve.RootName, strings.SplitN(c.varName, ".", 2)[0],
				"UndefinedVariableError.Name must contain the variable root")
		})
	}
}

func TestB4_UndefinedVariableError_LineNumber(t *testing.T) {
	eng := NewEngine()

	t.Run("line1", func(t *testing.T) {
		_, err := eng.ParseAndRenderString(`{{ x }}`, map[string]any{}, WithStrictVariables())
		require.Error(t, err)
		var uve *render.UndefinedVariableError
		require.True(t, errors.As(err, &uve))
		assert.Equal(t, 1, uve.LineNumber())
	})

	t.Run("line2", func(t *testing.T) {
		_, err := eng.ParseAndRenderString("ok\n{{ x }}", map[string]any{}, WithStrictVariables())
		require.Error(t, err)
		var uve *render.UndefinedVariableError
		require.True(t, errors.As(err, &uve))
		assert.Equal(t, 2, uve.LineNumber())
	})
}

func TestB4_UndefinedVariableError_Message(t *testing.T) {
	eng := NewEngine()
	_, err := eng.ParseAndRenderString(`{{ ghost }}`, map[string]any{}, WithStrictVariables())
	require.Error(t, err)

	var uve *render.UndefinedVariableError
	require.True(t, errors.As(err, &uve))

	msg := uve.Message()
	assert.Contains(t, msg, "ghost", "Message() must name the missing variable")
	assert.NotContains(t, msg, "Liquid error", "Message() must not have prefix")
	assert.NotContains(t, msg, "line ", "Message() must not include location")
}

func TestB4_UndefinedVariableError_MarkupContext(t *testing.T) {
	eng := NewEngine()
	_, err := eng.ParseAndRenderString(`{{ ghost }}`, map[string]any{}, WithStrictVariables())
	require.Error(t, err)

	var uve *render.UndefinedVariableError
	require.True(t, errors.As(err, &uve))

	mc := uve.MarkupContext()
	assert.Contains(t, mc, "ghost",
		"MarkupContext must reference the object expression")
}

// ── B4.4: ZeroDivisionError ───────────────────────────────────────────────────

func TestB4_ZeroDivisionError_DividedBy(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ 10 | divided_by: 0 }}`, nil)

	var zde *filters.ZeroDivisionError
	assert.True(t, errors.As(err, &zde),
		"divided_by: 0 must produce a ZeroDivisionError in the chain, got %T", err)
}

func TestB4_ZeroDivisionError_Modulo(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ 10 | modulo: 0 }}`, nil)

	var zde *filters.ZeroDivisionError
	assert.True(t, errors.As(err, &zde),
		"modulo: 0 must produce a ZeroDivisionError in the chain, got %T", err)
}

func TestB4_ZeroDivisionError_Message(t *testing.T) {
	eng := NewEngine()
	err := renderErr(t, eng, `{{ 10 | divided_by: 0 }}`, nil)

	assert.Contains(t, err.Error(), "0",
		"'zero' or '0' should appear in the error message")
}

// ── B4.5: ArgumentError ───────────────────────────────────────────────────────

func TestB4_ArgumentError_FromFilter(t *testing.T) {
	eng := NewEngine()
	eng.RegisterFilter("picky", func(v any) (any, error) {
		return nil, render.NewArgumentError("argument must be positive")
	})

	err := renderErr(t, eng, `{{ x | picky }}`, map[string]any{"x": -1})

	var ae *render.ArgumentError
	assert.True(t, errors.As(err, &ae),
		"errors.As must detect *render.ArgumentError, got %T", err)
	assert.Contains(t, ae.Error(), "positive",
		"ArgumentError.Error() must preserve the original message")

	// The top-level error must say "Liquid error", not "Liquid syntax error"
	assert.Contains(t, err.Error(), "Liquid error")
	assert.NotContains(t, err.Error(), "Liquid syntax error")
}

func TestB4_ArgumentError_FromTag(t *testing.T) {
	eng := NewEngine()
	eng.RegisterTag("strict_tag", func(c render.Context) (string, error) {
		return "", render.NewArgumentError("bad tag arguments")
	})

	err := renderErr(t, eng, `{% strict_tag %}`, nil)

	var ae *render.ArgumentError
	assert.True(t, errors.As(err, &ae),
		"errors.As must detect *render.ArgumentError from a tag, got %T", err)
}

func TestB4_ArgumentError_LineNumberAndContext(t *testing.T) {
	eng := NewEngine()
	eng.RegisterFilter("fail_filter", func(v any) (any, error) {
		return nil, render.NewArgumentError("filter failed")
	})

	src := "line1\nline2\n{{ x | fail_filter }}"
	err := renderErr(t, eng, src, map[string]any{"x": 1})

	assert.Contains(t, err.Error(), "line 3",
		"error line number must pinpoint the failing expression (line 3)")
	assert.Contains(t, err.Error(), "fail_filter",
		"error markup context must reference the expression")
}

// ── B4.6: ContextError ────────────────────────────────────────────────────────

func TestB4_ContextError_FromTag(t *testing.T) {
	eng := NewEngine()
	eng.RegisterTag("ctx_error_tag", func(c render.Context) (string, error) {
		return "", render.NewContextError("context look-up failed")
	})

	err := renderErr(t, eng, `{% ctx_error_tag %}`, nil)

	var ce *render.ContextError
	assert.True(t, errors.As(err, &ce),
		"errors.As must detect *render.ContextError, got %T", err)
	assert.Contains(t, ce.Error(), "context look-up failed")
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  B6 – MARKUP-CONTEXT PRESERVED THROUGH BLOCK TAGS                          ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// TestB6_ErrorContextNotLostInBlock is the core regression test for B6.
// Before the fix: BlockNode.render() re-wrapped inner errors, replacing the
// specific "{{ expr }}" markup context with the parent "{% if … %}" context.
func TestB6_ErrorContextNotLostInBlock(t *testing.T) {
	eng := NewEngine()
	binds := map[string]any{"x": 4}

	cases := []struct {
		name       string
		tpl        string
		wantInErr  string // must appear somewhere in the error string
		wantAbsent string // must NOT appear (the block tag that would be wrong)
	}{
		{
			name:       "filter_error_inside_if_oneliner",
			tpl:        `{% if true %}{{ x | divided_by: 0 }}{% endif %}`,
			wantInErr:  "divided_by",
			wantAbsent: "{% if",
		},
		{
			name:       "filter_error_inside_if_multiline",
			tpl:        "{% if true %}\n{{ x | divided_by: 0 }}\n{% endif %}",
			wantInErr:  "divided_by",
			wantAbsent: "{% if",
		},
		{
			name:       "filter_error_inside_unless",
			tpl:        `{% unless false %}{{ x | divided_by: 0 }}{% endunless %}`,
			wantInErr:  "divided_by",
			wantAbsent: "{% unless",
		},
		{
			name:       "filter_error_inside_for",
			tpl:        `{% for i in arr %}{{ i | divided_by: 0 }}{% endfor %}`,
			wantInErr:  "divided_by",
			wantAbsent: "{% for",
		},
		{
			name:       "filter_error_inside_nested_blocks",
			tpl:        `{% if true %}{% unless false %}{{ x | divided_by: 0 }}{% endunless %}{% endif %}`,
			wantInErr:  "divided_by",
			wantAbsent: "{% if",
		},
	}

	// for-loop test needs an array binding.
	bindsWithArr := map[string]any{"x": 4, "arr": []int{1, 2, 3}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := binds
			if strings.Contains(c.tpl, "arr") {
				b = bindsWithArr
			}
			_, err := eng.ParseAndRenderString(c.tpl, b)
			require.Error(t, err)
			assert.Contains(t, err.Error(), c.wantInErr,
				"error must mention the failing expression, not the outer block tag")
			if c.wantAbsent != "" {
				assert.NotContains(t, err.Error(), c.wantAbsent,
					"error must NOT show the outer block tag as the error source")
			}
		})
	}
}

// TestB6_ObjectExpressionContextPreserved checks that {{ expr }} markup context
// is preserved verbatim (not replaced by an outer tag's source).
func TestB6_ObjectExpressionContextPreserved(t *testing.T) {
	eng := NewEngine()

	t.Run("divide_by_zero_in_if", func(t *testing.T) {
		_, err := eng.ParseAndRenderString(
			`{% if true %}{{ price | divided_by: 0 }}{% endif %}`,
			map[string]any{"price": 100},
		)
		require.Error(t, err)

		var re *render.RenderError
		require.True(t, errors.As(err, &re))

		mc := re.MarkupContext()
		assert.Contains(t, mc, "price",
			"MarkupContext must contain the variable name from the expression")
		assert.Contains(t, mc, "divided_by",
			"MarkupContext must contain the filter name from the expression")
		assert.NotContains(t, mc, "if true",
			"MarkupContext must NOT be the outer if-block source")
	})

	t.Run("divide_by_zero_in_for", func(t *testing.T) {
		_, err := eng.ParseAndRenderString(
			`{% for item in items %}{{ item | divided_by: 0 }}{% endfor %}`,
			map[string]any{"items": []int{5}},
		)
		require.Error(t, err)

		var re *render.RenderError
		require.True(t, errors.As(err, &re))

		mc := re.MarkupContext()
		assert.Contains(t, mc, "item",
			"MarkupContext must reference the expression inside the for body")
		assert.NotContains(t, mc, "for item",
			"MarkupContext must NOT be the for-tag source")
	})
}

// TestB6_LineNumberPreservedInBlocks verifies that multi-line templates report
// the exact line number of the failing expression, not the line of the block tag.
func TestB6_LineNumberPreservedInBlocks(t *testing.T) {
	eng := NewEngine()

	t.Run("expression_on_line2_inside_if_on_line1", func(t *testing.T) {
		src := "{% if true %}\n{{ x | divided_by: 0 }}\n{% endif %}"
		_, err := eng.ParseAndRenderString(src, map[string]any{"x": 1})
		require.Error(t, err)
		// The expression is on line 2; the if-tag is on line 1.
		assert.Contains(t, err.Error(), "line 2",
			"error line number must be the line of the failing expression (2), not the if-tag (1)")
	})

	t.Run("expression_on_line3_inside_for_on_line1", func(t *testing.T) {
		src := "{% for i in arr %}\nok\n{{ i | divided_by: 0 }}\n{% endfor %}"
		_, err := eng.ParseAndRenderString(src, map[string]any{"arr": []int{1}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line 3",
			"error line number must be the line of the failing expression (3)")
	})

	t.Run("deeply_nested_correct_line", func(t *testing.T) {
		src := "{% if true %}\n{% unless false %}\n{{ x | divided_by: 0 }}\n{% endunless %}\n{% endif %}"
		_, err := eng.ParseAndRenderString(src, map[string]any{"x": 1})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "line 3",
			"error line number must point to the innermost failing expression (line 3)")
	})
}

// TestB6_ConditionErrorAttributedToBlock checks that when the ERROR is in the
// block's own CONDITION (not its body), it is correctly attributed to the
// block tag itself — this is correct behavior and must not regress.
func TestB6_ConditionErrorAttributedToBlock(t *testing.T) {
	eng := NewEngine()

	// A filter error inside the {% if %} condition should point to that if-tag,
	// not to some inner expression that doesn't exist.
	tpl := `{% if x | divided_by: 0 %}yes{% endif %}`
	_, err := eng.ParseAndRenderString(tpl, map[string]any{"x": 4})
	require.Error(t, err)
	// The error originates from evaluating the if-condition, so it's wrapped
	// by BlockNode(if). The markup context should be the {% if … %} tag.
	assert.Contains(t, err.Error(), "if",
		"condition error should mention the if-tag in the error message")
}

// TestB6_AssignTagErrorContext verifies that a filter error inside an {% assign %}
// tag is attributed to the assign-tag itself (correct: assign IS the failing node).
func TestB6_AssignTagErrorContext(t *testing.T) {
	eng := NewEngine()

	tpl := `{% assign y = x | divided_by: 0 %}{{ y }}`
	_, err := eng.ParseAndRenderString(tpl, map[string]any{"x": 4})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assign",
		"error from an assign-tag expression must mention 'assign' in its context")
}

// TestB6_UndefinedVariableInBlock verifies that UndefinedVariableError raised
// inside a block tag still carries the variable name and the {{ expr }} context.
func TestB6_UndefinedVariableInBlock(t *testing.T) {
	eng := NewEngine()

	t.Run("inside_if", func(t *testing.T) {
		_, err := eng.ParseAndRenderString(
			`{% if true %}{{ ghost_var }}{% endif %}`,
			map[string]any{},
			WithStrictVariables(),
		)
		require.Error(t, err)

		var uve *render.UndefinedVariableError
		require.True(t, errors.As(err, &uve),
			"errors.As must still find UndefinedVariableError inside an if-block")
		assert.Equal(t, "ghost_var", uve.RootName)
		assert.Equal(t, 1, uve.LineNumber())
	})

	t.Run("inside_for", func(t *testing.T) {
		_, err := eng.ParseAndRenderString(
			`{% for i in items %}{{ missing }}{% endfor %}`,
			map[string]any{"items": []int{1}},
			WithStrictVariables(),
		)
		require.Error(t, err)

		var uve *render.UndefinedVariableError
		require.True(t, errors.As(err, &uve),
			"errors.As must find UndefinedVariableError inside a for-block")
		assert.Equal(t, "missing", uve.RootName)
	})

	t.Run("multiline_correct_line", func(t *testing.T) {
		src := "{% if true %}\nok\n{{ ghost_var }}\n{% endif %}"
		_, err := eng.ParseAndRenderString(src, map[string]any{}, WithStrictVariables())
		require.Error(t, err)

		var uve *render.UndefinedVariableError
		require.True(t, errors.As(err, &uve))
		assert.Equal(t, 3, uve.LineNumber(),
			"UndefinedVariableError must report line 3, the location of the expression")
	})
}

// TestB6_CaseTagErrorContext verifies that a filter error inside a {% case %}
// block body is attributed to the specific {{ expr }} node, not the {% case %} tag.
func TestB6_CaseTagErrorContext(t *testing.T) {
	eng := NewEngine()

	tpl := `{% case v %}{% when 1 %}{{ x | divided_by: 0 }}{% endcase %}`
	_, err := eng.ParseAndRenderString(tpl, map[string]any{"v": 1, "x": 4})
	require.Error(t, err)

	var re *render.RenderError
	require.True(t, errors.As(err, &re))

	mc := re.MarkupContext()
	assert.Contains(t, mc, "divided_by",
		"MarkupContext must name the failing expression, not the case-tag")
	assert.NotContains(t, mc, "case v",
		"MarkupContext must NOT be the case-tag source")
}

// TestB6_ErrorSingleVsMultiLineConsistency verifies that the same logical error
// produces a consistent MarkupContext regardless of whether the template is
// written on one line or spread across multiple lines.
func TestB6_ErrorSingleVsMultiLineConsistency(t *testing.T) {
	eng := NewEngine()
	binds := map[string]any{"x": 4}

	single := `{% if true %}{{ x | divided_by: 0 }}{% endif %}`
	multi := "{% if true %}\n{{ x | divided_by: 0 }}\n{% endif %}"

	_, errSingle := eng.ParseAndRenderString(single, binds)
	_, errMulti := eng.ParseAndRenderString(multi, binds)

	require.Error(t, errSingle)
	require.Error(t, errMulti)

	// Extract MarkupContext from both
	var reSingle, reMulti *render.RenderError
	require.True(t, errors.As(errSingle, &reSingle))
	require.True(t, errors.As(errMulti, &reMulti))

	// Both should reference the {{ expr }} not {% if %}
	for _, re := range []*render.RenderError{reSingle, reMulti} {
		mc := re.MarkupContext()
		assert.Contains(t, mc, "divided_by",
			"MarkupContext must be the {{ expr }}, consistent across layouts")
		assert.NotContains(t, mc, "if true",
			"MarkupContext must NOT be the {% if %} tag regardless of layout")
	}
}

// TestB6_ArgumentErrorLineNumberInBlock verifies that a custom filter's
// ArgumentError inside a block reports the correct line and markup context.
func TestB6_ArgumentErrorLineNumberInBlock(t *testing.T) {
	eng := NewEngine()
	eng.RegisterFilter("strict_filter", func(v any) (any, error) {
		return nil, render.NewArgumentError("strict rejection")
	})

	src := "{% if true %}\nfirst line ok\n{{ product | strict_filter }}\n{% endif %}"
	_, err := eng.ParseAndRenderString(src, map[string]any{"product": "item"})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "line 3",
		"error must point to line 3 where the failing expression lives")
	assert.Contains(t, err.Error(), "strict_filter",
		"error must mention the failing filter by name")

	var ae *render.ArgumentError
	assert.True(t, errors.As(err, &ae),
		"errors.As must detect *render.ArgumentError through block wrapping")
}

// TestB6_Cause_ZeroDivisionThroughBlockChain verifies that errors.As can still
// find the root cause (ZeroDivisionError) after it bubbles through block tags.
func TestB6_Cause_ZeroDivisionThroughBlockChain(t *testing.T) {
	eng := NewEngine()

	tpl := `{% if true %}{% unless false %}{{ x | divided_by: 0 }}{% endunless %}{% endif %}`
	err := renderErr(t, eng, tpl, map[string]any{"x": 1})

	var zde *filters.ZeroDivisionError
	assert.True(t, errors.As(err, &zde),
		"errors.As must find ZeroDivisionError through nested block chain, got %T", err)
}

// ╔══════════════════════════════════════════════════════════════════════════════╗
// ║  B6.2 – UNDEFINEDVARIABLEERROR CONSISTENCY ACROSS ALL TEMPLATE LAYOUTS     ║
// ╚══════════════════════════════════════════════════════════════════════════════╝

// TestB6_UndefinedVarConsistentAcrossFormats is the core regression test for the
// original B6 complaint: UndefinedVariableError had different behaviour depending
// on indentation and template layout.
//
// Three invariants must hold regardless of formatting:
//  1. An error IS raised (not silently swallowed).
//  2. The error Name is always the root variable name, not the full expression.
//  3. The line number is always the line containing the {{ expr }}.
func TestB6_UndefinedVarConsistentAcrossFormats(t *testing.T) {
	eng := NewEngine()
	strict := WithStrictVariables()

	type tc struct {
		name     string
		tpl      string
		binds    map[string]any
		wantVar  string // expected UndefinedVariableError.Name
		wantLine int
	}

	cases := []tc{
		// ── Whitespace variants in {{ expr }} ─────────────────────────────────
		{"no_spaces", `{{ghost}}`, nil, "ghost", 1},
		{"normal_spaces", `{{ ghost }}`, nil, "ghost", 1},
		{"extra_spaces", `{{   ghost   }}`, nil, "ghost", 1},
		{"tab_inside", "{{ \tghost\t }}", nil, "ghost", 1},

		// ── With filter — root var must be identified, not the whole pipeline ──
		{"with_one_filter", `{{ ghost | upcase }}`, nil, "ghost", 1},
		{"with_two_filters", `{{ ghost | upcase | strip }}`, nil, "ghost", 1},
		{"with_arg_filter", `{{ ghost | truncate: 10 }}`, nil, "ghost", 1},

		// ── Inside if — same line vs multi-line ───────────────────────────────
		{"if_same_line",
			`{% if true %}{{ ghost }}{% endif %}`, nil, "ghost", 1},
		{"if_next_line",
			"{% if true %}\n{{ ghost }}\n{% endif %}", nil, "ghost", 2},
		{"if_indented_spaces",
			"{% if true %}\n  {{ ghost }}\n{% endif %}", nil, "ghost", 2},
		{"if_indented_tab",
			"{% if true %}\n\t{{ ghost }}\n{% endif %}", nil, "ghost", 2},
		{"if_deep_nested",
			"{% if true %}\n  {% unless false %}\n    {{ ghost }}\n  {% endunless %}\n{% endif %}", nil, "ghost", 3},
		{"if_many_lines_before",
			"a\nb\nc\n{% if true %}\n  {{ ghost }}\n{% endif %}", nil, "ghost", 5},

		// ── Inside for ─────────────────────────────────────────────────────────
		{"for_same_line",
			`{% for i in arr %}{{ ghost }}{% endfor %}`,
			map[string]any{"arr": []int{1}}, "ghost", 1},
		{"for_next_line",
			"{% for i in arr %}\n{{ ghost }}\n{% endfor %}",
			map[string]any{"arr": []int{1}}, "ghost", 2},
		{"for_indented",
			"{% for i in arr %}\n  {{ ghost }}\n{% endfor %}",
			map[string]any{"arr": []int{1}}, "ghost", 2},

		// ── Nested properties ──────────────────────────────────────────────────
		// For dotted paths (user.name), Name is the ROOT variable only ("user"),
		// matching Ruby Liquid behaviour: the undefined thing is "user", not the path.
		{"dotted_property", `{{ user.name }}`, nil, "user", 1},
		{"dotted_inside_if", `{% if true %}{{ user.name }}{% endif %}`, nil, "user", 1},
		{"dotted_multiline_if",
			"{% if true %}\n{{ user.name }}\n{% endif %}", nil, "user", 2},

		// ── "Full" template in one line (original user case) ───────────────────
		{"long_single_line",
			`text {% if true %}more {{ ghost }} end{% endif %} after`, nil, "ghost", 1},
		{"entire_template_indented",
			"  {% if true %}\n    {{ ghost }}\n  {% endif %}", nil, "ghost", 2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := eng.ParseAndRenderString(c.tpl, c.binds, strict)
			require.Error(t, err,
				"UndefinedVariableError must be raised regardless of template formatting")

			var uve *render.UndefinedVariableError
			require.True(t, errors.As(err, &uve),
				"errors.As must find *render.UndefinedVariableError, got %T", err)

			assert.Equal(t, c.wantVar, uve.RootName,
				"Name must be the root variable, not the full expression text")
			assert.Equal(t, c.wantLine, uve.LineNumber(),
				"LineNumber must point to the {{ expr }} line, not the block tag")
		})
	}
}

// TestB6_UndefinedVar_NilBindingIsUndefined verifies that a variable explicitly
// bound to nil in the bindings map DOES trigger UndefinedVariableError in strict
// mode, treating nil the same as a missing key.
func TestB6_UndefinedVar_NilBindingIsUndefined(t *testing.T) {
	eng := NewEngine()
	strict := WithStrictVariables()

	// "product" is in the map with an explicit nil value — in strict mode, nil
	// is treated as undefined (same as a missing key).
	_, err := eng.ParseAndRenderString(
		`{{ product }}`,
		map[string]any{"product": nil},
		strict,
	)
	assert.Error(t, err,
		"explicit nil binding in strict mode must raise UndefinedVariableError")
}

// TestB6_UndefinedVar_FilterDoesNotHideError verifies the key fix: when a
// variable is undefined but a filter chain would transform nil → "", StrictVariables
// must still raise UndefinedVariableError before the filter runs.
func TestB6_UndefinedVar_FilterDoesNotHideError(t *testing.T) {
	eng := NewEngine()
	strict := WithStrictVariables()

	cases := []string{
		`{{ ghost | upcase }}`,
		`{{ ghost | default: "fallback" }}`,
		`{{ ghost | size }}`,
		`{% if true %}{{ ghost | upcase }}{% endif %}`,
		"{% if true %}\n{{ ghost | upcase }}\n{% endif %}",
	}

	for _, tpl := range cases {
		_, err := eng.ParseAndRenderString(tpl, map[string]any{}, strict)
		require.Error(t, err,
			"filter chain must not hide UndefinedVariableError for template: %q", tpl)

		var uve *render.UndefinedVariableError
		assert.True(t, errors.As(err, &uve),
			"error must be *render.UndefinedVariableError even with filters, tpl: %q", tpl)
		if uve != nil {
			assert.Equal(t, "ghost", uve.RootName,
				"Name must be the root variable name, not the filter pipeline, tpl: %q", tpl)
		}
	}
}
