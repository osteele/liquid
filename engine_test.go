package liquid

import (
	"bytes"
	"encoding/json"
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
	{`{% if page.nil-number < x %}true{% endif %}`, "true"},
}

var testBindings = map[string]interface{}{
	"x":  123,
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"ar":    []interface{}{"first", "second", "third"},
		"title": "Introduction",
	},
	"set": map[string]interface{}{
		"chars": []string{"a", "b", "c"},
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

func TestDateFilter(t *testing.T) {
	engine := NewEngine()
	template := `{% assign vardays = 30 | times: 24 | times: 60 | times: 60 %}{{ 'now' | date: "%s" | plus: vardays | date: "%d/%m/%Y" }}`
	str, err := engine.ParseAndRenderString(template, nil)
	require.NoError(t, err)
	t.Log(str)
	if len(str) == 0 {
		t.Error("date filter error")
	}
}

func TestDecimalFilter(t *testing.T) {
	engine := NewEngine()
	template := `{{ 12345 | decimal: 'one', '$' }}`
	str, err := engine.ParseAndRenderString(template, nil)
	require.NoError(t, err)
	t.Log(str)
	if str != "$12.3" {
		t.Error("decimal filter error")
	}
}

func TestDecimalWithDelimiterFilter(t *testing.T) {
	engine := NewEngine()
	tests := []struct {
		name          string
		liquid        string
		expectedValue string
	}{
		{
			name:          "currency symbol",
			liquid:        `{{ 12345 | decimalWithDelimiter: 'one', '€', 'de' }}`,
			expectedValue: "€12,3",
		},
		{
			name:          "norwegian Krone",
			liquid:        `{{ 12345 | decimalWithDelimiter: 'one', 'NOK', 'no-NO' }}`,
			expectedValue: "12,3 kr",
		},
		{
			name:          "us dollar in english",
			liquid:        `{{ 12345 | decimalWithDelimiter: 'one', 'USD', 'en' }}`,
			expectedValue: "$12.3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str, err := engine.ParseAndRenderString(test.liquid, nil)
			require.NoError(t, err)
			if str != test.expectedValue {
				t.Errorf("For %s, expected %s, got %s", test.name, test.expectedValue, str)
			}
		})
	}
}

func TestFindVariables(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name         string
		liquid       string
		expectedVars string
	}{
		{
			name: "2 levels loop",
			liquid: `{% for company in people.companies %}
		{% for instance in people.instances %}
			{{ company.name }}
			{{ instance.name }}
		{% endfor %}
	{% endfor %}`,
			expectedVars: `{"people.companies":{"Loop":true,"Attributes":{"name":{"Loop":false,"Attributes":null}}},"people.instances":{"Loop":true,"Attributes":{"name":{"Loop":false,"Attributes":null}}}}`,
		},
		{
			name: "2 levels loop which uses var of top loop",
			liquid: `{% for company in people.companies %}
		{% for instance in company.instances %}
			{{ company.name }}
			{{ instance.name }}
		{% endfor %}
	{% endfor %}`,
			expectedVars: `{"people.companies":{"Loop":true,"Attributes":{"instances":{"Loop":true,"Attributes":{"name":{"Loop":false,"Attributes":null}}},"name":{"Loop":false,"Attributes":null}}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := engine.ParseString(test.liquid)
			if err != nil {
				t.Fatalf("Expected no err: but got %s", err)
			}
			vars, err := tmpl.FindVariables()
			if err != nil {
				t.Fatalf("Expected no FindVariables err: but got %s", err)
			}
			varsJSON, jerr := json.Marshal(vars)
			if jerr != nil {
				t.Fatalf("Expected no Marshal err: but got %s", jerr)
			}

			if string(varsJSON) != test.expectedVars {
				t.Errorf("Expected:\n%s\nbut got:\n%s", test.expectedVars, string(varsJSON))
			}
		})
	}
}

func TestStartsWith(t *testing.T) {
	engine := NewEngine()
	template := `{{ 'hello' | startsWith: 'he' }}`
	str, err := engine.ParseAndRenderString(template, nil)
	require.NoError(t, err)
	t.Log(str)
	if str != "true" {
		t.Error("startsWith filter error")
	}
}

func TestEndsWith(t *testing.T) {
	engine := NewEngine()
	template := `{{ 'hello' | endsWith: 'lo' }}`
	str, err := engine.ParseAndRenderString(template, nil)
	require.NoError(t, err)
	t.Log(str)
	if str != "true" {
		t.Error("endsWith filter error")
	}
}

func TestSetContains(t *testing.T) {
	engine := NewEngine()
	template := `{{ set.chars | setContains: 'a' }}`
	str, err := engine.ParseAndRenderString(template, testBindings)
	require.NoError(t, err)
	t.Log(str)
	if str != "true" {
		t.Error("set contains filter error")
	}
	template = `{{ set.chars | setContains: 'd' }}`
	str, err = engine.ParseAndRenderString(template, testBindings)
	require.NoError(t, err)
	t.Log(str)
	if str != "false" {
		t.Error("set contains filter error")
	}
	template = `{{ set.chars | setContains: 'a', 'b' }}`
	str, err = engine.ParseAndRenderString(template, testBindings)
	require.NoError(t, err)
	t.Log(str)
	if str != "true" {
		t.Error("set contains filter error")
	}
}
