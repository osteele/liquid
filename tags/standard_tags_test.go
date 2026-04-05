package tags

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{% undefined_tag %}", "undefined tag"},
	{"{% assign v x y z %}", "syntax error"},
	{"{% if syntax error %}", `unterminated "if" block`},
	{"{% echo %}", "syntax error"},
	// TODO once expression parsing is moved to template parse stage
	// {"{% if syntax error %}{% endif %}", "syntax error"},
	// {"{% for a in ar undefined %}{{ a }} {% endfor %}", "TODO"},
}

var tagTests = []struct{ in, expected string }{
	// variable tags
	{`{% assign av = 1 %}{{ av }}`, "1"},
	{`{% assign av = obj.a %}{{ av }}`, "1"},
	{`{% assign av = (1..5) %}{{ av }}`, "1..5"},
	{`{% capture x %}captured{% endcapture %}{{ x }}`, "captured"},

	// issue #76: assign with boolean expressions using 'and'/'or' operators
	{`{% assign result = x == 123 and obj.a == 1 %}{{ result }}`, "true"},
	{`{% assign result = x == 999 and obj.a == 1 %}{{ result }}`, "false"},
	{`{% assign result = x == 999 or obj.a == 1 %}{{ result }}`, "true"},
	{`{% assign result = x == 123 or obj.a == 999 %}{{ result }}`, "true"},
	{`{% assign result = x == 999 or obj.a == 999 %}{{ result }}`, "false"},
	// exact test case from issue #76
	{`{% assign con_0_Euh43 = user.name == "Ryan" and user.email == "xx@gmail.com" %}{{ con_0_Euh43 }}`, "true"},

	// TODO research whether Liquid requires matching interior tags
	{`{% comment %}{{ a }}{% undefined_tag %}{% endcomment %}`, ""},

	// TODO research whether Liquid requires matching interior tags
	{`pre{% raw %}{{ a }}{% undefined_tag %}{% endraw %}post`, "pre{{ a }}{% undefined_tag %}post"},
	{`pre{% raw %}{% if false %}anyway-{% endraw %}post`, "pre{% if false %}anyway-post"},
}

var tagErrorTests = []struct{ in, expected string }{
	{`{% assign av = x | undefined_filter %}`, "undefined filter"},
}

var echoTagTests = []struct{ in, expected string }{
	// basic expression output — same semantics as {{ expr }}
	{`{% echo x %}`, "123"},
	{`{% echo "hello" %}`, "hello"},
	{`{% echo obj.a %}`, "1"},
	// nil variable renders as empty string (same as {{ }})
	{`{% echo undefined %}`, ""},
}

// this is also used in the other test files
var tagTestBindings = map[string]any{
	"x": 123,
	"obj": map[string]any{
		"a": 1,
	},
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"pages": []map[string]any{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]any{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"page": map[string]any{
		"title": "Introduction",
		"meta": map[string]any{
			"author": "John Doe",
		},
	},
	"user": map[string]any{
		"name":  "Ryan",
		"email": "xx@gmail.com",
	},
}

func TestStandardTags_parse_errors(t *testing.T) {
	settings := render.NewConfig()
	AddStandardTags(&settings)

	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := settings.Compile(test.in, parser.SourceLoc{})
			require.Nilf(t, root, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestStandardTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range tagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestStandardTags_render_errors(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range tagErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = render.Render(root, io.Discard, tagTestBindings, config)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

// Test Jekyll extensions for assign tag with dot notation
func TestAssignTag_JekyllExtensions(t *testing.T) {
	jekyllTests := []struct{ in, expected string }{
		// dot notation assignments (Jekyll compatibility)
		{`{% assign page.canonical_url = "/about/" %}{{ page.canonical_url }}`, "/about/"},
		{`{% assign page.meta.description = "Test description" %}{{ page.meta.description }}`, "Test description"},
		{`{% assign obj.nested = 42 %}{{ obj.nested }}`, "42"},
		{`{% assign new_obj.prop = "value" %}{{ new_obj.prop }}`, "value"},
		{`{% assign page.title = "New Title" %}{{ page.title }}`, "New Title"},
	}

	t.Run("With Jekyll Extensions", func(t *testing.T) {
		config := render.NewConfig()
		config.JekyllExtensions = true // Enable Jekyll extensions
		AddStandardTags(&config)

		for i, test := range jekyllTests {
			t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
				root, err := config.Compile(test.in, parser.SourceLoc{})
				require.NoErrorf(t, err, test.in)

				buf := new(bytes.Buffer)
				err = render.Render(root, buf, tagTestBindings, config)
				require.NoErrorf(t, err, test.in)
				require.Equalf(t, test.expected, buf.String(), test.in)
			})
		}
	})

	t.Run("Without Jekyll Extensions (Standard Mode)", func(t *testing.T) {
		config := render.NewConfig()
		config.JekyllExtensions = false // Disable Jekyll extensions (default)
		AddStandardTags(&config)

		// These should all fail in standard mode
		for i, test := range jekyllTests {
			t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
				_, err := config.Compile(test.in, parser.SourceLoc{})
				require.Errorf(t, err, "Expected error for: %s", test.in)
				require.Containsf(t, err.Error(), "Jekyll extensions", "Expected Jekyll extensions error for: %s", test.in)
			})
		}
	})

	// Test that simple assignments still work in standard mode
	t.Run("Simple Assignments in Standard Mode", func(t *testing.T) {
		config := render.NewConfig()
		config.JekyllExtensions = false // Standard mode
		AddStandardTags(&config)

		simpleTests := []struct{ in, expected string }{
			{`{% assign av = 1 %}{{ av }}`, "1"},
			{`{% assign name = "John" %}{{ name }}`, "John"},
			{`{% assign val = obj.a %}{{ val }}`, "1"},
		}

		for i, test := range simpleTests {
			t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
				root, err := config.Compile(test.in, parser.SourceLoc{})
				require.NoErrorf(t, err, test.in)

				buf := new(bytes.Buffer)
				err = render.Render(root, buf, tagTestBindings, config)
				require.NoErrorf(t, err, test.in)
				require.Equalf(t, test.expected, buf.String(), test.in)
			})
		}
	})
}

func TestEchoTag(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range echoTagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

// ---------------------------------------------------------------------------
// increment / decrement
// ---------------------------------------------------------------------------

var incrementDecrementTests = []struct{ in, expected string }{
	// basic increment: counter starts at 0, outputs current then increments
	{`{% increment var %}`, "0"},
	{`{% increment var %}{% increment var %}`, "01"},
	{`{% increment var %}{% increment var %}{% increment var %}`, "012"},
	// basic decrement: counter starts at 0, decrements then outputs
	{`{% decrement var %}`, "-1"},
	{`{% decrement var %}{% decrement var %}`, "-1-2"},
	// increment and decrement are in SEPARATE namespaces (Shopify spec)
	{`{% increment shared %}{% decrement shared %}`, "0-1"},
	{`{% decrement shared %}{% increment shared %}`, "-10"},
	// counters are in a separate namespace from assign variables
	{`{% assign var = 100 %}{% increment var %}{{ var }}`, "0100"},
	{`{% assign var = 100 %}{% decrement var %}{{ var }}`, "-1100"},
	// multiple independent variables
	{`{% increment a %}{% increment b %}{% increment a %}`, "001"},
	// counter persists across if/for blocks
	{`{% for i in array %}{% increment c %}{% endfor %}`, "012"},
}

func TestIncrementDecrement(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	bindings := map[string]any{
		"array": []string{"a", "b", "c"},
	}

	for i, test := range incrementDecrementTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, bindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestIncrementDecrement_errors(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	parseErrTests := []struct{ in, expected string }{
		{`{% increment %}`, "syntax error"},
		{`{% decrement %}`, "syntax error"},
	}

	for _, test := range parseErrTests {
		_, err := config.Compile(test.in, parser.SourceLoc{})
		require.Errorf(t, err, test.in)
		require.Containsf(t, err.Error(), test.expected, test.in)
	}
}

// ---------------------------------------------------------------------------
// {%# inline comment %}
// ---------------------------------------------------------------------------

var inlineCommentTests = []struct{ in, expected string }{
	// basic comment: rendered as empty string
	{`{%# this is a comment %}`, ""},
	// comment inside content
	{`before{%# comment %}after`, "beforeafter"},
	// comment with trim left (note: no space between - and # per Liquid spec)
	{`before   {%-# comment %}after`, "beforeafter"},
	// comment with trim right
	{`before{%# comment -%}   after`, "beforeafter"},
	// comment with trim both
	{`before   {%-# comment -%}   after`, "beforeafter"},
	// comment with variables (should not be evaluated)
	{`{%# {{ x }} does not output %}result`, "result"},
	// multiple comments
	{`a{%# first %}b{%# second %}c`, "abc"},
	// empty comment
	{`{%# %}`, ""},
	// comment doesn't affect surrounding tags
	{`{% assign v = 1 %}{%# ignore me %}{{ v }}`, "1"},
	// space-separated variants: {%- # comment -%} (space between - and #)
	{`before   {%- # comment %}after`, "beforeafter"},
	{`before{%  # comment -%}   after`, "beforeafter"},
	{`before   {%- # comment -%}   after`, "beforeafter"},
	// space-only (no trim marker) with space before #
	{`before{% # comment %}after`, "beforeafter"},
	// trim with space: left only trims preceding whitespace
	{"a \n{%- # comment %}b", "ab"},
	// trim with space: both sides trims preceding and following whitespace
	{"a \n{%- # comment -%}\nb", "ab"},
}

func TestInlineComment(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range inlineCommentTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

// ---------------------------------------------------------------------------
// liquid tag (multi-line)
// ---------------------------------------------------------------------------

var liquidTagTests = []struct{ in, expected string }{
	// basic assign inside liquid tag
	{`{% liquid assign v = 1 %}{{ v }}`, "1"},
	// multi-line with newlines
	{"{% liquid\nassign v = 1\n%}{{ v }}", "1"},
	// echo inside liquid tag outputs
	{"{% liquid\nassign v = 42\necho v\n%}", "42"},
	// if/endif inside liquid tag
	{"{% liquid\nassign x = true\nif x\necho \"yes\"\nendif\n%}", "yes"},
	// if/else/endif
	{"{% liquid\nassign x = false\nif x\necho \"yes\"\nelse\necho \"no\"\nendif\n%}", "no"},
	// comments using # inside liquid tag
	{"{% liquid\n# this is a comment\nassign v = 99\n%}{{ v }}", "99"},
	// empty lines are ignored
	{"{% liquid\n\n\nassign v = 5\n\n%}{{ v }}", "5"},
	// assign propagates to outer scope
	{"{% liquid\nassign outer = \"hello\"\n%}{{ outer }}", "hello"},
	// multiple assigns
	{"{% liquid\nassign a = 1\nassign b = 2\n%}{{ a }}{{ b }}", "12"},
}

func TestLiquidTag(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	for i, test := range liquidTagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)

			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestLiquidTag_syntax_error(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	// Syntax errors inside liquid tag should propagate at compile time
	_, err := config.Compile("{% liquid\nif x\n%}", parser.SourceLoc{})
	require.Errorf(t, err, "expected error for unterminated if inside liquid tag")
}

// TestDocTag verifies that {% doc %}...{% enddoc %} renders as empty string.
func TestDocTag(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	tests := []struct{ in, expected string }{
		{`{% doc %}anything here{% enddoc %}`, ""},
		{`{% doc %}{% if x %}y{% endif %}{% enddoc %}`, ""},
		{`before{% doc %}ignored{% enddoc %}after`, "beforeafter"},
	}

	for _, tt := range tests {
		root, err := config.Compile(tt.in, parser.SourceLoc{})
		require.NoError(t, err)
		buf := new(bytes.Buffer)
		require.NoError(t, render.Render(root, buf, map[string]any{}, config))
		require.Equal(t, tt.expected, buf.String(), tt.in)
	}
}

// TestIfchangedTag verifies that {% ifchanged %} only emits output when
// it differs from the previous iteration.
func TestIfchangedTag(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	tests := []struct {
		in       string
		bindings map[string]any
		expected string
	}{
		// Unique values — all emitted.
		{
			`{% for x in arr %}{% ifchanged %}{{ x }}{% endifchanged %}{% endfor %}`,
			map[string]any{"arr": []any{1, 2, 3}},
			"123",
		},
		// Repeated values — only first occurrence emitted.
		{
			`{% for x in arr %}{% ifchanged %}{{ x }}{% endifchanged %}{% endfor %}`,
			map[string]any{"arr": []any{1, 1, 2, 2, 3}},
			"123",
		},
		// All same — only emitted once.
		{
			`{% for x in arr %}{% ifchanged %}{{ x }}{% endifchanged %}{% endfor %}`,
			map[string]any{"arr": []any{"a", "a", "a"}},
			"a",
		},
		// Static content inside ifchanged — emitted only once since it never changes.
		{
			`{% for x in arr %}{% ifchanged %}static{% endifchanged %}{{ x }}{% endfor %}`,
			map[string]any{"arr": []any{1, 2, 3}},
			"static123",
		},
		// Same static content across more iterations — still emitted only once.
		{
			`{% for x in arr %}{% ifchanged %}static{% endifchanged %}{{ x }}{% endfor %}`,
			map[string]any{"arr": []any{1, 2, 3, 4}},
			"static1234",
		},
	}

	for _, tt := range tests {
		root, err := config.Compile(tt.in, parser.SourceLoc{})
		require.NoError(t, err, tt.in)
		buf := new(bytes.Buffer)
		require.NoError(t, render.Render(root, buf, tt.bindings, config), tt.in)
		require.Equal(t, tt.expected, buf.String(), tt.in)
	}
}
