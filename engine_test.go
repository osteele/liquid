package liquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var emptyBindings = map[string]interface{}{}

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
}

var testBindings = map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestEngine_ParseAndRenderString(t *testing.T) {
	engine := NewEngine()
	for i, test := range liquidTests {
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			out, err := engine.ParseAndRenderString(test.in, testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, out, test.in)
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
		t.Run(fmt.Sprint(i+1), func(t *testing.T) {
			wr := capWriter{}
			err := engine.ParseAndFRender(&wr, []byte(test.in), testBindings)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, wr.String(), test.in)
		})
	}
}

func TestEngine_ParseAndRenderString_ptr_to_hash(t *testing.T) {
	params := map[string]interface{}{
		"message": &map[string]interface{}{
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
	params := map[string]interface{}{
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
	for i := 0; i < 1000; i++ {
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
	for i := 0; i < b.N; i++ {
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
	require.Equal(t, string(result), "Foo, Bar")
}
