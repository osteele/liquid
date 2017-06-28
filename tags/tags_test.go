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
	// TODO once expression parsing is moved to template parse stage
	// {"{%if syntax error%}{%endif%}", "parse error"},
	// {"{%for a in ar unknown%}{{a}} {%endfor%}", "TODO"},
}

var tagTests = []struct{ in, expected string }{
	{`{%assign av = 1%}{{av}}`, "1"},
	{`{%assign av = obj.a%}{{av}}`, "1"},

	// TODO test whether this requires matching interior tags
	{`{%comment%}{{a}}{%unknown%}{%endcomment%}`, ""},

	{`{%capture x%}captured{%endcapture%}{{x}}`, "captured"},

	{`{%case 1%}{%when 1%}a{%when 2%}b{%endcase%}`, "a"},
	{`{%case 2%}{%when 1%}a{%when 2%}b{%endcase%}`, "b"},
	{`{%case 3%}{%when 1%}a{%when 2%}b{%endcase%}`, ""},
	// {`{%case 2%}{%when 1%}a{%else 2%}b{%endcase%}`, "captured"},

	{`{%for a in ar%}{{a}} {%endfor%}`, "first second third "},
	{`{%for a in ar reversed%}{{a}} {%endfor%}`, "third second first "},
	{`{%for a in ar limit:2%}{{a}} {%endfor%}`, "first second "},
	{`{%for a in ar offset:1%}{{a}} {%endfor%}`, "second third "},
	{`{%for a in ar reversed offset:1%}{{a}} {%endfor%}`, "second first "},
	{`{%for a in ar reversed limit:1%}{{a}} {%endfor%}`, "third "},
	{`{%for a in ar limit:1 offset:1%}{{a}} {%endfor%}`, "second "},
	{`{%for a in ar reversed limit:1 offset:1%}{{a}} {%endfor%}`, "second "},

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

	// TODO test whether this requires matching interior tags
	{`pre{%raw%}{{a}}{%unknown%}{%endraw%}post`, "pre{{a}}{%unknown%}post"},
	{`pre{%raw%}{%if false%}anyway-{%endraw%}post`, "pre{%if false%}anyway-post"},

	{`{%unless true%}false{%endunless%}`, ""},
	{`{%unless false%}true{%endunless%}`, "true"},
	{`{%unless true%}false{%else%}true{%endunless%}`, "true"},
	{`{%unless false%}true{%else%}false{%endunless%}`, "true"},
	{`{%unless false%}0{%elsif true%}1{%else%}2{%endunless%}`, "0"},
	{`{%unless true%}0{%elsif true%}1{%else%}2{%endunless%}`, "1"},
	{`{%unless true%}0{%elsif false%}1{%else%}2{%endunless%}`, "2"},
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
