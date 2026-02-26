package liquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/autopilot3/ap3-types-go/types/date"
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

func TestTimeInTimezoneFilter(t *testing.T) {
	engine := NewEngine()
	// Tuesday, June 4, 2024, 3:15:00 PM UTC
	testTime := time.Date(2024, time.June, 4, 15, 15, 0, 0, time.UTC)
	params := map[string]interface{}{
		"time": testTime,
	}

	tests := []struct {
		name     string
		format   string
		tz       string
		expected string
	}{
		// US formats with weekday
		{name: "mdy24aw", format: "mdy24aw", tz: "UTC", expected: "15:15 Tuesday, June 4, 2024"},
		{name: "mdy12aw", format: "mdy12aw", tz: "UTC", expected: "3:15pm Tuesday, June 4, 2024"},
		// US formats without weekday
		{name: "mdyaw", format: "mdyaw", tz: "UTC", expected: "Tuesday, June 4, 2024"},
		{name: "mdy24a", format: "mdy24a", tz: "UTC", expected: "15:15 June 4, 2024"},
		{name: "mdy12a", format: "mdy12a", tz: "UTC", expected: "3:15pm June 4, 2024"},
		{name: "mdya", format: "mdya", tz: "UTC", expected: "June 4, 2024"},
		{name: "mdy24n", format: "mdy24n", tz: "UTC", expected: "15:15 06/04/2024"},
		{name: "mdy12n", format: "mdy12n", tz: "UTC", expected: "3:15pm 06/04/2024"},
		{name: "mdy24nd", format: "mdy24nd", tz: "UTC", expected: "06/04/2024 15:15"},
		{name: "mdy12nd", format: "mdy12nd", tz: "UTC", expected: "06/04/2024 3:15pm"},
		{name: "mdy", format: "mdy", tz: "UTC", expected: "06/04/2024"},
		{name: "mdys24", format: "mdys24", tz: "UTC", expected: "15:15 6/4/24"},
		{name: "mdys12", format: "mdys12", tz: "UTC", expected: "3:15pm 6/4/24"},
		{name: "mdys24d", format: "mdys24d", tz: "UTC", expected: "6/4/24 15:15"},
		{name: "mdys12d", format: "mdys12d", tz: "UTC", expected: "6/4/24 3:15pm"},
		{name: "mdys", format: "mdys", tz: "UTC", expected: "6/4/24"},
		// Everyone else formats with weekday
		{name: "dmy24aw", format: "dmy24aw", tz: "UTC", expected: "15:15 Tuesday, 4 June, 2024"},
		{name: "dmy12aw", format: "dmy12aw", tz: "UTC", expected: "3:15pm Tuesday, 4 June, 2024"},
		// Everyone else formats without weekday
		{name: "dmyaw", format: "dmyaw", tz: "UTC", expected: "Tuesday, 4 June, 2024"},
		{name: "dmy24a", format: "dmy24a", tz: "UTC", expected: "15:15 4 June, 2024"},
		{name: "dmy12a", format: "dmy12a", tz: "UTC", expected: "3:15pm 4 June, 2024"},
		{name: "dmya", format: "dmya", tz: "UTC", expected: "4 June, 2024"},
		{name: "dmy24n", format: "dmy24n", tz: "UTC", expected: "15:15 04/06/2024"},
		{name: "dmy12n", format: "dmy12n", tz: "UTC", expected: "3:15pm 04/06/2024"},
		{name: "dmy24nd", format: "dmy24nd", tz: "UTC", expected: "04/06/2024 15:15"},
		{name: "dmy12nd", format: "dmy12nd", tz: "UTC", expected: "04/06/2024 3:15pm"},
		{name: "dmy", format: "dmy", tz: "UTC", expected: "04/06/2024"},
		{name: "dmys24", format: "dmys24", tz: "UTC", expected: "15:15 4/6/24"},
		{name: "dmys12", format: "dmys12", tz: "UTC", expected: "3:15pm 4/6/24"},
		{name: "dmys24d", format: "dmys24d", tz: "UTC", expected: "4/6/24 15:15"},
		{name: "dmys12d", format: "dmys12d", tz: "UTC", expected: "4/6/24 3:15pm"},
		{name: "dmys", format: "dmys", tz: "UTC", expected: "4/6/24"},
		// Individual pieces
		{name: "h24", format: "h24", tz: "UTC", expected: "15"},
		{name: "h12", format: "h12", tz: "UTC", expected: "3"},
		{name: "min", format: "min", tz: "UTC", expected: "15"},
		{name: "p", format: "p", tz: "UTC", expected: "pm"},
		{name: "d", format: "d", tz: "UTC", expected: "4"},
		{name: "dd", format: "dd", tz: "UTC", expected: "04"},
		{name: "dow", format: "dow", tz: "UTC", expected: "Tuesday"},
		{name: "m", format: "m", tz: "UTC", expected: "6"},
		{name: "mm", format: "mm", tz: "UTC", expected: "06"},
		{name: "mon", format: "mon", tz: "UTC", expected: "June"},
		{name: "yy", format: "yy", tz: "UTC", expected: "24"},
		{name: "yyyy", format: "yyyy", tz: "UTC", expected: "2024"},
		// Legacy formats
		{name: "mdy12", format: "mdy12", tz: "UTC", expected: "Jun 04 2024 3:15 PM"},
		{name: "mdy24", format: "mdy24", tz: "UTC", expected: "Jun 04 2024 15:15"},
		{name: "dmy12", format: "dmy12", tz: "UTC", expected: "04 Jun 2024 3:15 PM"},
		{name: "dmy24", format: "dmy24", tz: "UTC", expected: "04 Jun 2024 15:15"},
		{name: "ymd12", format: "ymd12", tz: "UTC", expected: "2024 Jun 04 3:15 PM"},
		{name: "ymd24", format: "ymd24", tz: "UTC", expected: "2024 Jun 04 15:15"},
		{name: "ydm12", format: "ydm12", tz: "UTC", expected: "2024 04 Jun 3:15 PM"},
		{name: "ydm24", format: "ydm24", tz: "UTC", expected: "2024 04 Jun 15:15"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			template := "{{ time | timeInTimezone: '" + test.tz + "', '" + test.format + "' }}"
			str, err := engine.ParseAndRenderString(template, params)
			require.NoError(t, err)
			require.Equal(t, test.expected, str, "format=%s", test.format)
		})
	}
}

func TestDateTimeFormatOrDefaultFilter(t *testing.T) {
	engine := NewEngine()
	// Tuesday, June 4, 2024, 3:15:00 PM UTC
	testTime := time.Date(2024, time.June, 4, 15, 15, 0, 0, time.UTC)
	zeroTime := time.Time{}
	params := map[string]interface{}{
		"time": testTime,
		"zero": zeroTime,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		// US formats with weekday
		{name: "mdy24aw", template: "{{ time | dateTimeFormatOrDefault: 'mdy24aw', 'DEFAULT' }}", expected: "15:15 Tuesday, June 4, 2024"},
		{name: "mdy12aw", template: "{{ time | dateTimeFormatOrDefault: 'mdy12aw', 'DEFAULT' }}", expected: "3:15pm Tuesday, June 4, 2024"},
		// US formats without weekday
		{name: "mdyaw", template: "{{ time | dateTimeFormatOrDefault: 'mdyaw', 'DEFAULT' }}", expected: "Tuesday, June 4, 2024"},
		{name: "mdy24a", template: "{{ time | dateTimeFormatOrDefault: 'mdy24a', 'DEFAULT' }}", expected: "15:15 June 4, 2024"},
		{name: "mdy12a", template: "{{ time | dateTimeFormatOrDefault: 'mdy12a', 'DEFAULT' }}", expected: "3:15pm June 4, 2024"},
		{name: "mdya", template: "{{ time | dateTimeFormatOrDefault: 'mdya', 'DEFAULT' }}", expected: "June 4, 2024"},
		{name: "mdy24n", template: "{{ time | dateTimeFormatOrDefault: 'mdy24n', 'DEFAULT' }}", expected: "15:15 06/04/2024"},
		{name: "mdy12n", template: "{{ time | dateTimeFormatOrDefault: 'mdy12n', 'DEFAULT' }}", expected: "3:15pm 06/04/2024"},
		{name: "mdy24nd", template: "{{ time | dateTimeFormatOrDefault: 'mdy24nd', 'DEFAULT' }}", expected: "06/04/2024 15:15"},
		{name: "mdy12nd", template: "{{ time | dateTimeFormatOrDefault: 'mdy12nd', 'DEFAULT' }}", expected: "06/04/2024 3:15pm"},
		{name: "mdy", template: "{{ time | dateTimeFormatOrDefault: 'mdy', 'DEFAULT' }}", expected: "06/04/2024"},
		{name: "mdys24", template: "{{ time | dateTimeFormatOrDefault: 'mdys24', 'DEFAULT' }}", expected: "15:15 6/4/24"},
		{name: "mdys12", template: "{{ time | dateTimeFormatOrDefault: 'mdys12', 'DEFAULT' }}", expected: "3:15pm 6/4/24"},
		{name: "mdys24d", template: "{{ time | dateTimeFormatOrDefault: 'mdys24d', 'DEFAULT' }}", expected: "6/4/24 15:15"},
		{name: "mdys12d", template: "{{ time | dateTimeFormatOrDefault: 'mdys12d', 'DEFAULT' }}", expected: "6/4/24 3:15pm"},
		{name: "mdys", template: "{{ time | dateTimeFormatOrDefault: 'mdys', 'DEFAULT' }}", expected: "6/4/24"},
		// Everyone else formats with weekday
		{name: "dmy24aw", template: "{{ time | dateTimeFormatOrDefault: 'dmy24aw', 'DEFAULT' }}", expected: "15:15 Tuesday, 4 June, 2024"},
		{name: "dmy12aw", template: "{{ time | dateTimeFormatOrDefault: 'dmy12aw', 'DEFAULT' }}", expected: "3:15pm Tuesday, 4 June, 2024"},
		// Everyone else formats without weekday
		{name: "dmyaw", template: "{{ time | dateTimeFormatOrDefault: 'dmyaw', 'DEFAULT' }}", expected: "Tuesday, 4 June, 2024"},
		{name: "dmy24a", template: "{{ time | dateTimeFormatOrDefault: 'dmy24a', 'DEFAULT' }}", expected: "15:15 4 June, 2024"},
		{name: "dmy12a", template: "{{ time | dateTimeFormatOrDefault: 'dmy12a', 'DEFAULT' }}", expected: "3:15pm 4 June, 2024"},
		{name: "dmya", template: "{{ time | dateTimeFormatOrDefault: 'dmya', 'DEFAULT' }}", expected: "4 June, 2024"},
		{name: "dmy24n", template: "{{ time | dateTimeFormatOrDefault: 'dmy24n', 'DEFAULT' }}", expected: "15:15 04/06/2024"},
		{name: "dmy12n", template: "{{ time | dateTimeFormatOrDefault: 'dmy12n', 'DEFAULT' }}", expected: "3:15pm 04/06/2024"},
		{name: "dmy24nd", template: "{{ time | dateTimeFormatOrDefault: 'dmy24nd', 'DEFAULT' }}", expected: "04/06/2024 15:15"},
		{name: "dmy12nd", template: "{{ time | dateTimeFormatOrDefault: 'dmy12nd', 'DEFAULT' }}", expected: "04/06/2024 3:15pm"},
		{name: "dmy", template: "{{ time | dateTimeFormatOrDefault: 'dmy', 'DEFAULT' }}", expected: "04/06/2024"},
		{name: "dmys24", template: "{{ time | dateTimeFormatOrDefault: 'dmys24', 'DEFAULT' }}", expected: "15:15 4/6/24"},
		{name: "dmys12", template: "{{ time | dateTimeFormatOrDefault: 'dmys12', 'DEFAULT' }}", expected: "3:15pm 4/6/24"},
		{name: "dmys24d", template: "{{ time | dateTimeFormatOrDefault: 'dmys24d', 'DEFAULT' }}", expected: "4/6/24 15:15"},
		{name: "dmys12d", template: "{{ time | dateTimeFormatOrDefault: 'dmys12d', 'DEFAULT' }}", expected: "4/6/24 3:15pm"},
		{name: "dmys", template: "{{ time | dateTimeFormatOrDefault: 'dmys', 'DEFAULT' }}", expected: "4/6/24"},
		// Individual pieces
		{name: "h24", template: "{{ time | dateTimeFormatOrDefault: 'h24', 'DEFAULT' }}", expected: "15"},
		{name: "h12", template: "{{ time | dateTimeFormatOrDefault: 'h12', 'DEFAULT' }}", expected: "3"},
		{name: "min", template: "{{ time | dateTimeFormatOrDefault: 'min', 'DEFAULT' }}", expected: "15"},
		{name: "p", template: "{{ time | dateTimeFormatOrDefault: 'p', 'DEFAULT' }}", expected: "pm"},
		{name: "d", template: "{{ time | dateTimeFormatOrDefault: 'd', 'DEFAULT' }}", expected: "4"},
		{name: "dd", template: "{{ time | dateTimeFormatOrDefault: 'dd', 'DEFAULT' }}", expected: "04"},
		{name: "dow", template: "{{ time | dateTimeFormatOrDefault: 'dow', 'DEFAULT' }}", expected: "Tuesday"},
		{name: "m", template: "{{ time | dateTimeFormatOrDefault: 'm', 'DEFAULT' }}", expected: "6"},
		{name: "mm", template: "{{ time | dateTimeFormatOrDefault: 'mm', 'DEFAULT' }}", expected: "06"},
		{name: "mon", template: "{{ time | dateTimeFormatOrDefault: 'mon', 'DEFAULT' }}", expected: "June"},
		{name: "yy", template: "{{ time | dateTimeFormatOrDefault: 'yy', 'DEFAULT' }}", expected: "24"},
		{name: "yyyy", template: "{{ time | dateTimeFormatOrDefault: 'yyyy', 'DEFAULT' }}", expected: "2024"},
		// Legacy formats
		{name: "mdy12", template: "{{ time | dateTimeFormatOrDefault: 'mdy12', 'DEFAULT' }}", expected: "Jun 04 2024 3:15 PM"},
		{name: "mdy24", template: "{{ time | dateTimeFormatOrDefault: 'mdy24', 'DEFAULT' }}", expected: "Jun 04 2024 15:15"},
		{name: "dmy12", template: "{{ time | dateTimeFormatOrDefault: 'dmy12', 'DEFAULT' }}", expected: "04 Jun 2024 3:15 PM"},
		{name: "dmy24", template: "{{ time | dateTimeFormatOrDefault: 'dmy24', 'DEFAULT' }}", expected: "04 Jun 2024 15:15"},
		{name: "ymd12", template: "{{ time | dateTimeFormatOrDefault: 'ymd12', 'DEFAULT' }}", expected: "2024 Jun 04 3:15 PM"},
		{name: "ymd24", template: "{{ time | dateTimeFormatOrDefault: 'ymd24', 'DEFAULT' }}", expected: "2024 Jun 04 15:15"},
		{name: "ydm12", template: "{{ time | dateTimeFormatOrDefault: 'ydm12', 'DEFAULT' }}", expected: "2024 04 Jun 3:15 PM"},
		{name: "ydm24", template: "{{ time | dateTimeFormatOrDefault: 'ydm24', 'DEFAULT' }}", expected: "2024 04 Jun 15:15"},
		// Zero time should return default
		{name: "zero_time_default", template: "{{ zero | dateTimeFormatOrDefault: 'mdya', 'MY_DEFAULT' }}", expected: "MY_DEFAULT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str, err := engine.ParseAndRenderString(test.template, params)
			require.NoError(t, err)
			require.Equal(t, test.expected, str)
		})
	}
}

func TestDateFormatOrDefaultFilter(t *testing.T) {
	engine := NewEngine()
	// Tuesday, June 4, 2024
	d, err := date.New(2024, 6, 4, "UTC")
	require.NoError(t, err)
	var zeroDate date.Date // zero value
	params := map[string]interface{}{
		"date":     d,
		"zeroDate": zeroDate,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		// US formats
		{name: "mdyaw", template: "{{ date | dateFormatOrDefault: 'mdyaw', 'DEFAULT' }}", expected: "Tuesday, June 4, 2024"},
		{name: "mdya", template: "{{ date | dateFormatOrDefault: 'mdya', 'DEFAULT' }}", expected: "June 4, 2024"},
		{name: "mdy", template: "{{ date | dateFormatOrDefault: 'mdy', 'DEFAULT' }}", expected: "06/04/2024"},
		{name: "mdys", template: "{{ date | dateFormatOrDefault: 'mdys', 'DEFAULT' }}", expected: "6/4/24"},
		// Everyone else formats
		{name: "dmyaw", template: "{{ date | dateFormatOrDefault: 'dmyaw', 'DEFAULT' }}", expected: "Tuesday, 4 June, 2024"},
		{name: "dmya", template: "{{ date | dateFormatOrDefault: 'dmya', 'DEFAULT' }}", expected: "4 June, 2024"},
		{name: "dmy", template: "{{ date | dateFormatOrDefault: 'dmy', 'DEFAULT' }}", expected: "04/06/2024"},
		{name: "dmys", template: "{{ date | dateFormatOrDefault: 'dmys', 'DEFAULT' }}", expected: "4/6/24"},
		// Legacy order variants
		{name: "ymd", template: "{{ date | dateFormatOrDefault: 'ymd', 'DEFAULT' }}", expected: "2024/06/04"},
		{name: "ydm", template: "{{ date | dateFormatOrDefault: 'ydm', 'DEFAULT' }}", expected: "2024/04/06"},
		// Individual pieces
		{name: "d", template: "{{ date | dateFormatOrDefault: 'd', 'DEFAULT' }}", expected: "4"},
		{name: "dd", template: "{{ date | dateFormatOrDefault: 'dd', 'DEFAULT' }}", expected: "04"},
		{name: "m", template: "{{ date | dateFormatOrDefault: 'm', 'DEFAULT' }}", expected: "6"},
		{name: "mm", template: "{{ date | dateFormatOrDefault: 'mm', 'DEFAULT' }}", expected: "06"},
		{name: "yy", template: "{{ date | dateFormatOrDefault: 'yy', 'DEFAULT' }}", expected: "24"},
		{name: "yyyy", template: "{{ date | dateFormatOrDefault: 'yyyy', 'DEFAULT' }}", expected: "2024"},
		// Zero date should return default
		{name: "zero_date_default", template: "{{ zeroDate | dateFormatOrDefault: 'mdya', 'MY_DEFAULT' }}", expected: "MY_DEFAULT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str, err := engine.ParseAndRenderString(test.template, params)
			require.NoError(t, err)
			require.Equal(t, test.expected, str)
		})
	}
}
