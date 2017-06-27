package chunks

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	// {"{%if syntax error%}", "unterminated if tag"},
	// {"{%if syntax error%}{%endif%}", "parse error"},
}

var renderTests = []struct{ in, expected string }{
	// {"{%if syntax error%}{%endif%}", "parse error"},
	{"{{12}}", "12"},
	{"{{x}}", "123"},
	{"{{page.title}}", "Introduction"},
	{"{{ar[1]}}", "second"},
}

var filterTests = []struct{ in, expected string }{
	// Jekyll extensions
	{`{{ obj | inspect }}`, `{"a":1}`},

	// filters
	// {{ product_price | default: 2.99 }}

	// list filters
	// {{ site.pages | map: 'category' | compact | join "," %}
	// {% assign my_array = "apples, oranges, peaches, plums" | split: ", " %}{{ my_array.first }}
	{`{{"John, Paul, George, Ringo" | split: ", " | join: " and "}}`, "John and Paul and George and Ringo"},
	{`{{ animals | sort | join: ", " }}`, "Sally Snake, giraffe, octopus, zebra"},
	{`{{ sort_prop | sort: "weight" | inspect }}`, `[{"weight":null},{"weight":1},{"weight":3},{"weight":5}]`},

	// last, map, slice, sort_natural, reverse, size, uniq

	// string filters
	// {{ "/my/fancy/url" | append: ".html" }}
	// {% assign filename = "/index.html" %}{{ "website.com" | append: filename }}

	// {{ "title" | capitalize }}
	// {{ "my great title" | capitalize }}

	// {{ "Parker Moore" | downcase }}

	// {{ "Have you read 'James & the Giant Peach'?" | escape }}
	// {{ "1 < 2 & 3" | escape_once }}
	// {{ "1 &lt; 2 &amp; 3" | escape_once }}

	// lstrip, newline_to_br, prepend, remove, remove_first, replace, replace_first
	// rstrip, split, strip, strip_html, strip_newlines, truncate, truncatewords, upcase
	// url_decode, url_encode

	// number filters
	// {{ -17 | abs }}
	// {{ 4 | abs }}
	// {{ "-19.86" | abs }}

	// {{ 1.2 | ceil }}
	// {{ 2.0 | ceil }}
	// {{ 183.357 | ceil }}
	// {{ "3.5" | ceil }}

	// {{ 16 | divided_by: 4 }}
	// {{ 5 | divided_by: 3 }}
	// {{ 20 | divided_by: 7.0 }}

	// {{ 1.2 | floor }}
	// {{ 2.0 | floor }}
	// {{ 183.357 | floor }}
	// minus, modulo, plus, round,times

	// date filters
	// {{ article.published_at | date: "%a, %b %d, %y" }}
	// {{ article.published_at | date: "%Y" }}
	// {{ "March 14, 2016" | date: "%b %d, %y" }}
	// {{ "now" | date: "%Y-%m-%d %H:%M" }
}

var renderTestContext = Context{map[string]interface{}{
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
},
}

func TestParseErrors(t *testing.T) {
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			ast, err := Parse(tokens)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
func TestRender(t *testing.T) {
	for i, test := range renderTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			// fmt.Println(tokens)
			ast, err := Parse(tokens)
			require.NoErrorf(t, err, test.in)
			// fmt.Println(MustYAML(ast))
			buf := new(bytes.Buffer)
			err = ast.Render(buf, renderTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestFilters(t *testing.T) {
	for i, test := range filterTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			ast, err := Parse(tokens)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = ast.Render(buf, renderTestContext)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}
