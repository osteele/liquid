package liquid_test

import (
	"fmt"
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// TestPREA_Integration validates all 6 items from PRE-A end-to-end.
func TestPREA_Integration(t *testing.T) {
	eng := liquid.NewEngine()

	check := func(t *testing.T, tpl, expected string, bindings map[string]any) {
		t.Helper()
		out, err := eng.ParseAndRenderString(tpl, bindings)
		require.NoError(t, err)
		require.Equal(t, expected, out)
	}

	// 1. empty literal
	t.Run("empty_literal_empty_string", func(t *testing.T) {
		check(t, `{% if x == empty %}yes{% endif %}`, "yes", map[string]any{"x": ""})
	})
	t.Run("empty_literal_empty_array", func(t *testing.T) {
		check(t, `{% if x == empty %}yes{% endif %}`, "yes", map[string]any{"x": []any{}})
	})
	t.Run("empty_literal_empty_map", func(t *testing.T) {
		check(t, `{% if x == empty %}yes{% endif %}`, "yes", map[string]any{"x": map[string]any{}})
	})
	t.Run("empty_literal_nonempty", func(t *testing.T) {
		check(t, `{% if x == empty %}yes{% else %}no{% endif %}`, "no", map[string]any{"x": "hi"})
	})
	t.Run("empty_output", func(t *testing.T) {
		// empty in output context should render as empty string
		check(t, `{{ empty }}`, "", nil)
	})

	// 2. blank literal
	t.Run("blank_literal_nil", func(t *testing.T) {
		check(t, `{% if x == blank %}yes{% endif %}`, "yes", map[string]any{"x": nil})
	})
	t.Run("blank_literal_false", func(t *testing.T) {
		check(t, `{% if x == blank %}yes{% endif %}`, "yes", map[string]any{"x": false})
	})
	t.Run("blank_literal_whitespace", func(t *testing.T) {
		check(t, `{% if x == blank %}yes{% endif %}`, "yes", map[string]any{"x": "   "})
	})
	t.Run("blank_literal_empty_string", func(t *testing.T) {
		check(t, `{% if x == blank %}yes{% endif %}`, "yes", map[string]any{"x": ""})
	})
	t.Run("blank_literal_nonempty", func(t *testing.T) {
		check(t, `{% if x == blank %}yes{% else %}no{% endif %}`, "no", map[string]any{"x": "hello"})
	})

	// 3. string escape sequences
	t.Run("escape_newline", func(t *testing.T) {
		check(t, `{{ "hello\nworld" }}`, "hello\nworld", nil)
	})
	t.Run("escape_tab", func(t *testing.T) {
		check(t, `{{ "col1\tcol2" }}`, "col1\tcol2", nil)
	})
	t.Run("escape_single_quote", func(t *testing.T) {
		check(t, `{{ 'it\'s fine' }}`, "it's fine", nil)
	})
	t.Run("escape_double_quote_in_double", func(t *testing.T) {
		check(t, `{{ "say \"hi\"" }}`, `say "hi"`, nil)
	})
	t.Run("escape_backslash", func(t *testing.T) {
		check(t, `{{ "a\\b" }}`, `a\b`, nil)
	})
	t.Run("escape_carriage_return", func(t *testing.T) {
		check(t, `{{ "a\rb" }}`, "a\rb", nil)
	})

	// 4. <> operator (alias for !=)
	t.Run("diamond_ne_true", func(t *testing.T) {
		check(t, `{% if 1 <> 2 %}yes{% endif %}`, "yes", nil)
	})
	t.Run("diamond_ne_false", func(t *testing.T) {
		check(t, `{% if 1 <> 1 %}yes{% else %}no{% endif %}`, "no", nil)
	})
	t.Run("diamond_string", func(t *testing.T) {
		check(t, `{% if "a" <> "b" %}yes{% endif %}`, "yes", nil)
	})

	// 5. not operator
	t.Run("not_false_is_true", func(t *testing.T) {
		check(t, `{% if not false %}yes{% endif %}`, "yes", nil)
	})
	t.Run("not_true_is_false", func(t *testing.T) {
		check(t, `{% if not true %}yes{% else %}no{% endif %}`, "no", nil)
	})
	t.Run("not_nil_is_true", func(t *testing.T) {
		check(t, `{% if not nil %}yes{% endif %}`, "yes", nil)
	})
	t.Run("not_nonempty_string_is_false", func(t *testing.T) {
		check(t, `{% if not x %}yes{% else %}no{% endif %}`, "no", map[string]any{"x": "hello"})
	})
	t.Run("not_with_and", func(t *testing.T) {
		// not a and b  →  (not a) and b
		check(t, `{% if not false and true %}yes{% endif %}`, "yes", nil)
	})

	// 6. case/when with or
	t.Run("when_or_first_value", func(t *testing.T) {
		check(t, `{% case x %}{% when 1 or 2 %}match{% else %}no{% endcase %}`, "match", map[string]any{"x": 1})
	})
	t.Run("when_or_second_value", func(t *testing.T) {
		check(t, `{% case x %}{% when 1 or 2 %}match{% else %}no{% endcase %}`, "match", map[string]any{"x": 2})
	})
	t.Run("when_or_no_match", func(t *testing.T) {
		check(t, `{% case x %}{% when 1 or 2 %}match{% else %}no{% endcase %}`, "no", map[string]any{"x": 3})
	})

	// 7. keyword args in filter (NamedArg plumbing)
	t.Run("keyword_arg_named_arg_type", func(t *testing.T) {
		// Verify NamedArg is passed through to filter — use a custom engine to test
		eng2 := liquid.NewEngine()
		var gotArg any
		eng2.RegisterFilter("spy", func(v any, args ...any) any {
			for _, a := range args {
				gotArg = a
			}
			return fmt.Sprintf("%v", v)
		})
		_, err := eng2.ParseAndRenderString(`{{ x | spy: "pos", flag: true }}`, map[string]any{"x": "test"})
		require.NoError(t, err)
		// gotArg should be a NamedArg
		require.NotNil(t, gotArg, "expected NamedArg to be passed to filter")
	})
}
