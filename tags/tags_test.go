package tags

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if syntax error%}", "unterminated if block"},
	// TODO once expression parsing is moved to template parse stage
	// {"{%if syntax error%}{%endif%}", "parse error"},
	// {"{%for a in ar unknown%}{{a}} {%endfor%}", "TODO"},
}

var tagTests = []struct{ in, expected string }{
	// variables
	{`{%assign av = 1%}{{av}}`, "1"},
	{`{%assign av = obj.a%}{{av}}`, "1"},
	{`{%capture x%}captured{%endcapture%}{{x}}`, "captured"},

	// TODO test whether this requires matching interior tags
	{`{%comment%}{{a}}{%unknown%}{%endcomment%}`, ""},

	// conditionals
	{`{%case 1%}{%when 1%}a{%when 2%}b{%endcase%}`, "a"},
	{`{%case 2%}{%when 1%}a{%when 2%}b{%endcase%}`, "b"},
	{`{%case 3%}{%when 1%}a{%when 2%}b{%endcase%}`, ""},
	// {`{%case 2%}{%when 1%}a{%else 2%}b{%endcase%}`, "captured"},

	{`{%if true%}true{%endif%}`, "true"},
	{`{%if false%}false{%endif%}`, ""},
	{`{%if 0%}true{%endif%}`, "true"},
	{`{%if 1%}true{%endif%}`, "true"},
	{`{%if x%}true{%endif%}`, "true"},
	{`{%if y%}true{%endif%}`, ""},
	{`{%if true%}true{%endif%}`, "true"},
	{`{%if false%}false{%endif%}`, ""},
	{`{%if true%}true{%else%}false{%endif%}`, "true"},
	{`{%if false%}false{%else%}true{%endif%}`, "true"},
	{`{%if true%}0{%elsif true%}1{%else%}2{%endif%}`, "0"},
	{`{%if false%}0{%elsif true%}1{%else%}2{%endif%}`, "1"},
	{`{%if false%}0{%elsif false%}1{%else%}2{%endif%}`, "2"},

	{`{%unless true%}false{%endunless%}`, ""},
	{`{%unless false%}true{%endunless%}`, "true"},
	{`{%unless true%}false{%else%}true{%endunless%}`, "true"},
	{`{%unless false%}true{%else%}false{%endunless%}`, "true"},
	{`{%unless false%}0{%elsif true%}1{%else%}2{%endunless%}`, "0"},
	{`{%unless true%}0{%elsif true%}1{%else%}2{%endunless%}`, "1"},
	{`{%unless true%}0{%elsif false%}1{%else%}2{%endunless%}`, "2"},

	// TODO test whether this requires matching interior tags
	{`pre{%raw%}{{a}}{%unknown%}{%endraw%}post`, "pre{{a}}{%unknown%}post"},
	{`pre{%raw%}{%if false%}anyway-{%endraw%}post`, "pre{%if false%}anyway-post"},
}

var tagTestBindings = map[string]interface{}{
	"x": 123,
	"obj": map[string]interface{}{
		"a": 1,
	},
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"pages": []map[string]interface{}{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
}

func TestParseErrors(t *testing.T) {
	settings := render.NewConfig()
	AddStandardTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := settings.Compile(test.in)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
func TestTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range tagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := config.Compile(test.in)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(ast, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
