package tags

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/osteele/liquid/chunks"
	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if syntax error%}", "unterminated if tag"},
	// {"{%if syntax error%}{%endif%}", "parse error"},
}

var tagTests = []struct{ in, expected string }{
	// TODO test whether this requires matching interior tags
	{"{%comment%}{{a}}{%unknown%}{%endcomment%}", ""},

	{"{%if true%}true{%endif%}", "true"},
	{"{%if false%}false{%endif%}", ""},
	{"{%if 0%}true{%endif%}", "true"},
	{"{%if 1%}true{%endif%}", "true"},
	{"{%if x%}true{%endif%}", "true"},
	{"{%if y%}true{%endif%}", ""},
	{"{%if true%}true{%endif%}", "true"},
	{"{%if false%}false{%endif%}", ""},
	{"{%if true%}true{%else%}false{%endif%}", "true"},
	{"{%if false%}false{%else%}true{%endif%}", "true"},
	{"{%if true%}0{%elsif true%}1{%else%}2{%endif%}", "0"},
	{"{%if false%}0{%elsif true%}1{%else%}2{%endif%}", "1"},
	{"{%if false%}0{%elsif false%}1{%else%}2{%endif%}", "2"},

	{"{%unless true%}false{%endif%}", ""},
	{"{%unless false%}true{%endif%}", "true"},
	{"{%unless true%}false{%else%}true{%endif%}", "true"},
	{"{%unless false%}true{%else%}false{%endif%}", "true"},
	{"{%unless false%}0{%elsif true%}1{%else%}2{%endif%}", "0"},
	{"{%unless true%}0{%elsif true%}1{%else%}2{%endif%}", "1"},
	{"{%unless true%}0{%elsif false%}1{%else%}2{%endif%}", "2"},

	{"{%assign av = 1%}{{av}}", "1"},
	{"{%assign av = obj.a%}{{av}}", "1"},

	{"{%for a in ar%}{{a}} {%endfor%}", "first second third "},
}

var tagTestContext = chunks.NewContext(map[string]interface{}{
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
	"ar": []string{"first", "second", "third"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
})

func init() {
	DefineStandardTags()
}
func TestParseErrors(t *testing.T) {
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := chunks.Scan(test.in, "")
			ast, err := chunks.Parse(tokens)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
func TestRender(t *testing.T) {
	for i, test := range tagTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := chunks.Scan(test.in, "")
			// fmt.Println(tokens)
			ast, err := chunks.Parse(tokens)
			require.NoErrorf(t, err, test.in)
			// fmt.Println(MustYAML(ast))
			buf := new(bytes.Buffer)
			err = ast.Render(buf, tagTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
