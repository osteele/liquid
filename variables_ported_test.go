package liquid_test

// Ported variable-access tests from:
//   - Ruby Liquid: test/integration/variable_test.rb
//   - LiquidJS:    test/e2e/issues.spec.ts  (issue #259, #486, #643, #655)
//   - LiquidJS:    test/integration/liquid/liquid.spec.ts
//
// Covers checklist section 5 — Acesso a Variáveis:
//   5a. obj.prop, obj[key], array[0]
//   5b. array[-1] — negative indexing
//   5c. array.first, array.last, obj.size
//   5d. {{ [key] }}  — dynamic variable lookup (Ruby)
//   5e. {{ test . test }} — dot with spaces (Ruby)
//   5f. {{ ["Key"].sub }} — top-level bracket + dot (LiquidJS #643)

import (
	"testing"

	"github.com/osteele/liquid"
	"github.com/stretchr/testify/require"
)

// ── 5a. obj.prop · obj[key] · array[0] ──────────────────────────────────────

// Ruby: test_simple_variable
func TestVariables_SimpleVariable(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{test}}`, map[string]any{"test": "worked"})
	require.NoError(t, err)
	require.Equal(t, "worked", out)

	out, err = eng.ParseAndRenderString(`{{test}}`, map[string]any{"test": "worked wonderfully"})
	require.NoError(t, err)
	require.Equal(t, "worked wonderfully", out)
}

// Ruby: test_simple_with_whitespaces
func TestVariables_WhitespacePadding(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`  {{ test }}  `, map[string]any{"test": "worked"})
	require.NoError(t, err)
	require.Equal(t, "  worked  ", out)

	out, err = eng.ParseAndRenderString(`  {{ test }}  `, map[string]any{"test": "worked wonderfully"})
	require.NoError(t, err)
	require.Equal(t, "  worked wonderfully  ", out)
}

// Ruby: test_ignore_unknown
func TestVariables_UnknownRendersEmpty(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ test }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// Ruby: test_hash_scoping – dot notation
func TestVariables_DotNotation(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ test.test }}`, map[string]any{
		"test": map[string]any{"test": "worked"},
	})
	require.NoError(t, err)
	require.Equal(t, "worked", out)
}

// Ruby: test_false_renders_as_false
func TestVariables_FalseRendersAsFalse(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ foo }}`, map[string]any{"foo": false})
	require.NoError(t, err)
	require.Equal(t, "false", out)

	// literal false in template
	out, err = eng.ParseAndRenderString(`{{ false }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "false", out)
}

// Ruby: test_nil_renders_as_empty_string
func TestVariables_NilRendersEmpty(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ nil }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)

	// nil | append still works (treats nil as empty string)
	out, err = eng.ParseAndRenderString(`{{ nil | append: 'cat' }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "cat", out)
}

// Ruby: test_multiline_variable
func TestVariables_MultilineTag(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString("{{\ntest\n}}", map[string]any{"test": "worked"})
	require.NoError(t, err)
	require.Equal(t, "worked", out)
}

// Ruby: test_expression_with_whitespace_in_square_brackets
// Whitespace inside bracket notation should be ignored.
func TestVariables_BracketNotationWhitespace(t *testing.T) {
	eng := liquid.NewEngine()

	// {{ a[ 'b' ] }} – spaces inside brackets
	out, err := eng.ParseAndRenderString(`{{ a[ 'b' ] }}`, map[string]any{
		"a": map[string]any{"b": "result"},
	})
	require.NoError(t, err)
	require.Equal(t, "result", out)

	// {{ a[ [ 'b' ] ] }} – inner bracket-lookup used as outer key
	out, err = eng.ParseAndRenderString(`{{ a[ b ] }}`, map[string]any{
		"b": "c",
		"a": map[string]any{"c": "result"},
	})
	require.NoError(t, err)
	require.Equal(t, "result", out)
}

// Array index 0 – explicit
func TestVariables_ArrayIndex(t *testing.T) {
	eng := liquid.NewEngine()

	arr := []string{"first", "second", "third"}

	out, err := eng.ParseAndRenderString(`{{ array[0] }}`, map[string]any{"array": arr})
	require.NoError(t, err)
	require.Equal(t, "first", out)

	out, err = eng.ParseAndRenderString(`{{ array[1] }}`, map[string]any{"array": arr})
	require.NoError(t, err)
	require.Equal(t, "second", out)

	out, err = eng.ParseAndRenderString(`{{ array[2] }}`, map[string]any{"array": arr})
	require.NoError(t, err)
	require.Equal(t, "third", out)

	// out-of-bounds → empty
	out, err = eng.ParseAndRenderString(`{{ array[100] }}`, map[string]any{"array": arr})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// String key via bracket notation
func TestVariables_MapBracketAccess(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ page["title"] }}`, map[string]any{
		"page": map[string]any{"title": "Introduction"},
	})
	require.NoError(t, err)
	require.Equal(t, "Introduction", out)
}

// Deep nested: obj.a.b.c
func TestVariables_DeepNesting(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ a.b.c }}`, map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": "deep"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "deep", out)
}

// Mixed: array inside map, accessed by computed key
func TestVariables_MixedIndexing(t *testing.T) {
	eng := liquid.NewEngine()

	// {{ hash["b"].c }} – bracket then dot
	out, err := eng.ParseAndRenderString(`{{ hash["b"].c }}`, map[string]any{
		"hash": map[string]any{
			"b": map[string]any{"c": "d"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "d", out)
}

// LiquidJS issue #259 — complex key access via string literal in brackets
func TestVariables_ComplexKeyAccess(t *testing.T) {
	eng := liquid.NewEngine()

	// Variable with spaces in key, accessed via string literal in brackets
	out, err := eng.ParseAndRenderString(`{{ hash["complex key"] }}`, map[string]any{
		"hash": map[string]any{"complex key": "foo"},
	})
	require.NoError(t, err)
	require.Equal(t, "foo", out)
}

// ── 5b. array[-1] — negative indexing ───────────────────────────────────────

// LiquidJS issue #486 — negative index
func TestVariables_NegativeIndex(t *testing.T) {
	eng := liquid.NewEngine()

	bindings := map[string]any{
		"a": []string{"x", "y", "z"},
	}

	// -1 → last element
	out, err := eng.ParseAndRenderString(`{{ a[-1] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "z", out)

	// -2 → second to last
	out, err = eng.ParseAndRenderString(`{{ a[-2] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "y", out)

	// -3 → first element
	out, err = eng.ParseAndRenderString(`{{ a[-3] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "x", out)

	// out-of-bounds → empty string
	out, err = eng.ParseAndRenderString(`{{ a[-8] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// negative index on empty array → empty string
func TestVariables_NegativeIndexEmptyArray(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ arr[-1] }}`, map[string]any{
		"arr": []string{},
	})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// ── 5c. array.first · array.last · obj.size ─────────────────────────────────

// Source: expressions_test.go (unit) + Ruby/JS integration
func TestVariables_ArrayFirst(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ fruits.first }}`, map[string]any{
		"fruits": []string{"apples", "oranges", "peaches", "plums"},
	})
	require.NoError(t, err)
	require.Equal(t, "apples", out)
}

func TestVariables_ArrayLast(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ fruits.last }}`, map[string]any{
		"fruits": []string{"apples", "oranges", "peaches", "plums"},
	})
	require.NoError(t, err)
	require.Equal(t, "plums", out)
}

// Empty array → .first and .last both render as empty string
func TestVariables_EmptyArrayFirstLast(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ arr.first }}`, map[string]any{"arr": []string{}})
	require.NoError(t, err)
	require.Equal(t, "", out)

	out, err = eng.ParseAndRenderString(`{{ arr.last }}`, map[string]any{"arr": []string{}})
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// .size on arrays
func TestVariables_ArraySize(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ fruits.size }}`, map[string]any{
		"fruits": []string{"apples", "oranges", "peaches", "plums"},
	})
	require.NoError(t, err)
	require.Equal(t, "4", out)
}

// .size on strings — rune count
func TestVariables_StringSize(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ word.size }}`, map[string]any{"word": "abc"})
	require.NoError(t, err)
	require.Equal(t, "3", out)

	// emoji is a single rune
	out, err = eng.ParseAndRenderString(`{{ word.size }}`, map[string]any{"word": "héllo"})
	require.NoError(t, err)
	require.Equal(t, "5", out)
}

// .size on maps
func TestVariables_MapSize(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ h.size }}`, map[string]any{
		"h": map[string]any{"a": 1, "b": 2, "c": 3},
	})
	require.NoError(t, err)
	require.Equal(t, "3", out)
}

// .size key overridden by a real value in the map
func TestVariables_SizeKeyOverride(t *testing.T) {
	eng := liquid.NewEngine()

	// When a map has an explicit "size" key, that should win over the built-in count
	out, err := eng.ParseAndRenderString(`{{ h.size }}`, map[string]any{
		"h": map[string]any{"size": "key_value"},
	})
	require.NoError(t, err)
	require.Equal(t, "key_value", out)
}

// array.first == array[0]  and  array.last == array[array.size-1]
// (README invariant)
func TestVariables_FirstLastEquivalence(t *testing.T) {
	eng := liquid.NewEngine()

	bindings := map[string]any{"arr": []string{"a", "b", "c"}}

	first0, err := eng.ParseAndRenderString(`{{ arr.first }}`, bindings)
	require.NoError(t, err)
	first1, err := eng.ParseAndRenderString(`{{ arr[0] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, first0, first1)

	last0, err := eng.ParseAndRenderString(`{{ arr.last }}`, bindings)
	require.NoError(t, err)
	last1, err := eng.ParseAndRenderString(`{{ arr[2] }}`, bindings)
	require.NoError(t, err)
	require.Equal(t, last0, last1)
}

// ── 5d. {{ [key] }} — dynamic variable lookup (Ruby) ─────────────────────────

// Ruby: test_dynamic_find_var / test_raw_value_variable
func TestVariables_DynamicFindVar(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ [key] }}`, map[string]any{"key": "foo", "foo": "bar"})
	require.NoError(t, err)
	require.Equal(t, "bar", out)
}

// Ruby: test_dynamic_find_var_with_drop — nested indirection
func TestVariables_DynamicFindVarNested(t *testing.T) {
	eng := liquid.NewEngine()

	// {{ [list[settings.zero]] }} — uses the result of list[0] as the var name
	out, err := eng.ParseAndRenderString(`{{ [list[0]] }}`, map[string]any{
		"list": []string{"foo"},
		"foo":  "bar",
	})
	require.NoError(t, err)
	require.Equal(t, "bar", out)
}

// Ruby: test_double_nested_variable_lookup — bracket chain
func TestVariables_DoubleNestedLookup(t *testing.T) {
	eng := liquid.NewEngine()

	// {{ list[list[0]]["foo"] }} — uses result of list[0] (=1) as index into list
	out, err := eng.ParseAndRenderString(`{{ list[list[0]]["foo"] }}`, map[string]any{
		"list": []any{1, map[string]any{"foo": "bar"}},
	})
	require.NoError(t, err)
	require.Equal(t, "bar", out)
}

// ── 5e. dot with spaces (Ruby) ────────────────────────────────────────────────

// Ruby: test_hash_scoping — dot with surrounding whitespace
func TestVariables_DotWithSpaces(t *testing.T) {
	eng := liquid.NewEngine()

	// {{ test . test }} — spaces around the dot
	out, err := eng.ParseAndRenderString(`{{ test . test }}`, map[string]any{
		"test": map[string]any{"test": "worked"},
	})
	require.NoError(t, err)
	require.Equal(t, "worked", out)
}

// ── 5f. Top-level bracket + dot (LiquidJS #643) ───────────────────────────────

// LiquidJS issue #643 — {{ ["Key String with Spaces"].subpropertyKey }}
func TestVariables_BracketRootPlusDot(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ ["Key String with Spaces"].subpropertyKey }}`, map[string]any{
		"Key String with Spaces": map[string]any{"subpropertyKey": "FOO"},
	})
	require.NoError(t, err)
	require.Equal(t, "FOO", out)
}

// LiquidJS issue #655 — Unicode variable names
func TestVariables_UnicodeVariableName(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ÜLKE}}`, map[string]any{"ÜLKE": "Türkiye"})
	require.NoError(t, err)
	require.Equal(t, "Türkiye", out)
}

// Ruby: using blank/empty as variable names
func TestVariables_BlankAsVariableName(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{% assign foo = blank %}{{ foo }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

func TestVariables_EmptyAsVariableName(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{% assign foo = empty %}{{ foo }}`, nil)
	require.NoError(t, err)
	require.Equal(t, "", out)
}

// Ruby: nested array like [[nil]] should render as empty string
func TestVariables_NestedArrayRenders(t *testing.T) {
	eng := liquid.NewEngine()

	out, err := eng.ParseAndRenderString(`{{ foo }}`, map[string]any{"foo": [][]any{{nil}}})
	require.NoError(t, err)
	require.Equal(t, "", out)
}
