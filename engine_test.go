package liquid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"

	"github.com/stretchr/testify/require"
)

var emptyBindings = map[string]any{}

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
}

var echoIntegrationTests = []struct{ in, expected string }{
	// echo behaves identically to {{ expr }}
	{`{% echo x %}`, "123"},
	{`{% echo "hello" | upcase %}`, "HELLO"},
	{`{% echo x | plus: 1 %}`, "124"},
	// nil renders as empty string
	{`{% echo undefined %}`, ""},
}

var testBindings = map[string]any{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]any{
		"title": "Introduction",
	},
}

func TestEngine_ParseAndRenderString(t *testing.T) {
	engine := NewEngine()

	for i, test := range liquidTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func TestEngine_EchoTag(t *testing.T) {
	engine := NewEngine()

	for i, test := range echoIntegrationTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
		})
	}
}

func TestBasicEngine_ParseAndRenderString(t *testing.T) {
	engine := NewBasicEngine()

	t.Run("1", func(t *testing.T) {
		test := liquidTests[0]
		out, err := engine.ParseAndRenderString(test.in, testBindings)
		require.NoErrorf(t, err, test.in)
		require.Equalf(t, test.expected, out, test.in)
	})

	for i, test := range liquidTests[1:] {
		t.Run(strconv.Itoa(i+2), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.Errorf(t, err, test.in)
			require.Emptyf(t, out, test.in)
		})
	}
}

type capWriter struct {
	bytes.Buffer
}

func (c *capWriter) Write(bs []byte) (int, error) {
	return c.Buffer.Write([]byte(strings.ToUpper(string(bs))))
}

func TestEngine_ParseAndFRender(t *testing.T) {
	engine := NewEngine()

	for i, test := range liquidTests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			wr := capWriter{}
			err := engine.ParseAndFRender(&wr, []byte(test.in), testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, strings.ToUpper(test.expected), wr.String(), test.in)
		})
	}
}

func TestEngine_ParseAndRenderString_ptr_to_hash(t *testing.T) {
	params := map[string]any{
		"message": &map[string]any{
			"Text":       "hello",
			"jsonNumber": json.Number("123"),
		},
	}
	engine := NewEngine()
	template := "{{ message.Text }} {{message.jsonNumber}}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "hello 123", str)
}

type testStruct struct{ Text string }

func TestEngine_ParseAndRenderString_struct(t *testing.T) {
	params := map[string]any{
		"message": testStruct{
			Text: "hello",
		},
	}
	engine := NewEngine()
	template := "{{ message.Text }}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "hello", str)
}

func TestEngine_ParseAndRender_errors(t *testing.T) {
	_, err := NewEngine().ParseAndRenderString("{{ syntax error }}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% if %}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% undefined_tag %}", emptyBindings)
	require.Error(t, err)
	_, err = NewEngine().ParseAndRenderString("{% a | undefined_filter %}", emptyBindings)
	require.Error(t, err)
}

func BenchmarkEngine_Parse(b *testing.B) {
	engine := NewEngine()

	buf := new(bytes.Buffer)
	for range 1000 {
		_, err := io.WriteString(buf, `if{% if true %}true{% elsif %}elsif{% else %}else{% endif %}`)
		require.NoError(b, err)
		_, err = io.WriteString(buf, `loop{% for item in array %}loop{% break %}{% endfor %}`)
		require.NoError(b, err)
		_, err = io.WriteString(buf, `case{% case value %}{% when a %}{% when b %{% endcase %}`)
		require.NoError(b, err)
		_, err = io.WriteString(buf, `expr{{ a and b }}{{ a add: b }}`)
		require.NoError(b, err)
	}

	s := buf.Bytes()

	b.ResetTimer()

	for range b.N {
		_, err := engine.ParseTemplate(s)
		require.NoError(b, err)
	}
}

func TestEngine_ParseTemplateAndCache(t *testing.T) {
	// Given two templates...
	templateA := []byte("Foo")
	templateB := []byte(`{% include "template_a.html" %}, Bar`)

	// Cache the first
	eng := NewEngine()
	_, err := eng.ParseTemplateAndCache(templateA, "template_a.html", 1)
	require.NoError(t, err)

	// ...and execute the second.
	result, err := eng.ParseAndRender(templateB, Bindings{})
	require.NoError(t, err)
	require.Equal(t, "Foo, Bar", string(result))
}

type MockTemplateStore struct{}

func (tl *MockTemplateStore) ReadTemplate(filename string) ([]byte, error) {
	template := fmt.Appendf(nil, "Message Text: {{ message.Text }} from: %v.", filename)
	return template, nil
}

func Test_template_store(t *testing.T) {
	template := []byte(`{% include "template.liquid" %}`)
	mockstore := &MockTemplateStore{}
	params := map[string]any{
		"message": testStruct{
			Text: "filename",
		},
	}
	engine := NewEngine()
	engine.RegisterTemplateStore(mockstore)
	out, _ := engine.ParseAndRenderString(string(template), params)
	require.Equal(t, "Message Text: filename from: template.liquid.", out)
}

func TestEngine_LaxFilters(t *testing.T) {
	// Default: undefined filters cause an error
	engine := NewEngine()
	_, err := engine.ParseAndRenderString(`{{ "hello" | nofilter }}`, emptyBindings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined filter")

	// LaxFilters: undefined filters pass through the value
	engine = NewEngine()
	engine.LaxFilters()
	out, err := engine.ParseAndRenderString(`{{ "hello" | nofilter }}`, emptyBindings)
	require.NoError(t, err)
	require.Equal(t, "hello", out)

	// LaxFilters: defined filters still work
	out, err = engine.ParseAndRenderString(`{{ "hello" | upcase }}`, emptyBindings)
	require.NoError(t, err)
	require.Equal(t, "HELLO", out)
}

func TestEngine_Delims(t *testing.T) {
	engine := NewEngine()
	engine.Delims("<%=", "%>", "<%", "%>")

	out, err := engine.ParseAndRenderString(`<%= x %>`, testBindings)
	require.NoError(t, err)
	require.Equal(t, "123", out)

	// standard delimiters should not work
	out, err = engine.ParseAndRenderString(`{{ x }}`, testBindings)
	require.NoError(t, err)
	require.Equal(t, "{{ x }}", out)
}

func TestEngine_StrictVariables(t *testing.T) {
	engine := NewEngine()
	engine.StrictVariables()

	// defined variable works
	out, err := engine.ParseAndRenderString(`{{ x }}`, testBindings)
	require.NoError(t, err)
	require.Equal(t, "123", out)

	// undefined variable causes error
	_, err = engine.ParseAndRenderString(`{{ undefined_var }}`, testBindings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "undefined variable")
}

func TestParseError_Type(t *testing.T) {
	engine := NewEngine()

	_, err := engine.ParseString(`{% if unclosed %}`)
	require.Error(t, err)

	var pe *parser.ParseError
	require.True(t, errors.As(err, &pe), "expected *parser.ParseError, got %T", err)
}

func TestRenderError_Type(t *testing.T) {
	engine := NewEngine()
	template, parseErr := engine.ParseString(`{{ x | modulo: 0 }}`)
	require.NoError(t, parseErr)

	_, renderErr := template.RenderString(map[string]any{"x": 10})
	require.Error(t, renderErr)

	var re *render.RenderError
	require.True(t, errors.As(renderErr, &re), "expected *render.RenderError, got %T", renderErr)
}

func TestUndefinedVariableError_Type(t *testing.T) {
	engine := NewEngine()
	engine.StrictVariables()

	_, err := engine.ParseAndRenderString(`{{ my_missing_var }}`, map[string]any{})
	require.Error(t, err)

	var ue *render.UndefinedVariableError
	require.True(t, errors.As(err, &ue), "expected *render.UndefinedVariableError, got %T", err)
	require.Equal(t, "my_missing_var", ue.RootName)
	require.Contains(t, err.Error(), "my_missing_var")
}

func TestZeroDivisionError_Type(t *testing.T) {
	engine := NewEngine()

	t.Run("divided_by", func(t *testing.T) {
		template, parseErr := engine.ParseString(`{{ 10 | divided_by: 0 }}`)
		require.NoError(t, parseErr)
		_, renderErr := template.RenderString(map[string]any{})
		require.Error(t, renderErr)
		var zde *filters.ZeroDivisionError
		require.True(t, errors.As(renderErr, &zde), "expected *filters.ZeroDivisionError, got %T", renderErr)
	})

	t.Run("modulo", func(t *testing.T) {
		template, parseErr := engine.ParseString(`{{ 10 | modulo: 0 }}`)
		require.NoError(t, parseErr)
		_, renderErr := template.RenderString(map[string]any{})
		require.Error(t, renderErr)
		var zde *filters.ZeroDivisionError
		require.True(t, errors.As(renderErr, &zde), "expected *filters.ZeroDivisionError, got %T", renderErr)
	})
}

func TestEngine_EnableJekyllExtensions(t *testing.T) {
	engine := NewEngine()
	engine.EnableJekyllExtensions()

	out, err := engine.ParseAndRenderString(
		`{% assign page.url = "/about/" %}{{ page.url }}`,
		map[string]any{"page": map[string]any{}},
	)
	require.NoError(t, err)
	require.Equal(t, "/about/", out)
}

func TestEngine_SetAutoEscapeReplacer(t *testing.T) {
	engine := NewEngine()
	engine.SetAutoEscapeReplacer(render.HtmlEscaper)

	// HTML should be escaped
	out, err := engine.ParseAndRenderString(`{{ html }}`, map[string]any{
		"html": "<b>bold</b>",
	})
	require.NoError(t, err)
	require.Equal(t, "&lt;b&gt;bold&lt;/b&gt;", out)

	// safe filter bypasses escaping
	out, err = engine.ParseAndRenderString(`{{ html | safe }}`, map[string]any{
		"html": "<b>bold</b>",
	})
	require.NoError(t, err)
	require.Equal(t, "<b>bold</b>", out)
}

func TestEngine_SetGlobalFilter(t *testing.T) {
	// global_filter applies a function to every {{ expression }} output [ruby: global_filter option]
	engine := NewEngine()
	engine.SetGlobalFilter(func(v any) (any, error) {
		if s, ok := v.(string); ok {
			return strings.ToUpper(s), nil
		}
		return v, nil
	})

	// string values are transformed
	out, err := engine.ParseAndRenderString(`{{ name }}`, map[string]any{"name": "world"})
	require.NoError(t, err)
	require.Equal(t, "WORLD", out)

	// non-string values pass through untouched
	out, err = engine.ParseAndRenderString(`{{ count }}`, map[string]any{"count": 42})
	require.NoError(t, err)
	require.Equal(t, "42", out)

	// nil values pass through (rendered as empty)
	out, err = engine.ParseAndRenderString(`{{ missing }}`, emptyBindings)
	require.NoError(t, err)
	require.Equal(t, "", out)

	// filter is applied after Liquid filter chain
	out, err = engine.ParseAndRenderString(`{{ name | prepend: "hello " }}`, map[string]any{"name": "world"})
	require.NoError(t, err)
	require.Equal(t, "HELLO WORLD", out)

	// filter error propagates as a render error
	errorEngine := NewEngine()
	errorEngine.SetGlobalFilter(func(v any) (any, error) {
		return nil, fmt.Errorf("global filter error")
	})
	_, err = errorEngine.ParseAndRenderString(`{{ name }}`, map[string]any{"name": "world"})
	require.Error(t, err)
}

func TestEngine_UnregisterTag(t *testing.T) {
	engine := NewEngine()
	engine.RegisterTag("custom_test_tag", func(c render.Context) (string, error) {
		return c.TagArgs(), nil
	})
	source := `{% custom_test_tag hello world %}`

	_, err := engine.ParseAndRenderString(source, emptyBindings)
	require.NoError(t, err)

	engine.UnregisterTag("custom_test_tag")

	_, err = engine.ParseAndRenderString(source, emptyBindings)
	require.Error(t, err)
}
