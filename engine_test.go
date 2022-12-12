package liquid

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var emptyBindings = map[string]interface{}{}

// There's a lot more tests in the filters and tags sub-packages.
// This collects a minimal set for testing end-to-end.
var liquidTests = []struct{ in, expected string }{
	{`{{ page.title }}`, "Introduction"},
	{`{% if x %}true{% endif %}`, "true"},
	{`{{ "upper" | upcase }}`, "UPPER"},
	{`{{ page.ar | first }}`, "first"},
}

var testBindings = map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"ar":    []interface{}{"first", "second", "third"},
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

func TestEngine_ParseAndRenderString_ptr_to_hash(t *testing.T) {
	params := map[string]interface{}{
		"message": &map[string]interface{}{
			"Text": "hello",
		},
	}
	engine := NewEngine()
	template := "{{ message.Text }}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "hello", str)
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
		io.WriteString(buf, `if{% if true %}true{% elsif %}elsif{% else %}else{% endif %}`)
		io.WriteString(buf, `loop{% for item in array %}loop{% break %}{% endfor %}`)
		io.WriteString(buf, `case{% case value %}{% when a %}{% when b %{% endcase %}`)
		io.WriteString(buf, `expr{{ a and b }}{{ a add: b }}`)
	}
	s := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ParseTemplate(s)
	}
}

func TestCustomFilter(t *testing.T) {
	// GMT	Sun Dec 11 2022 14:02:03 GMT+0000
	// Your Time Zone Mon Dec 12 2022 01:02:03 GMT+1100 (Australian Eastern Daylight Time)
	params := map[string]interface{}{
		"message": &map[string]interface{}{
			"created_at": time.Unix(1670767323, 0),
		},
	}
	engine := NewEngine()
	template := "{{ message.created_at | timeInTimezone: 'Australia/Sydney', 'mdy12' }}"
	str, err := engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "Dec 12 2022 1:02 AM", str)
	template = "{{ message.created_at | timeInTimezone: 'Asia/Shanghai', 'mdy12' }}"
	str, err = engine.ParseAndRenderString(template, params)
	require.NoError(t, err)
	require.Equal(t, "Dec 11 2022 10:02 PM", str)
}
