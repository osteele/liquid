package liquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

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
	template := []byte(fmt.Sprintf("Message Text: {{ message.Text }} from: %v.", filename))
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
