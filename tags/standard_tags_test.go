package tags

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var parseErrorTests = []struct{ in, expected string }{
	{"{% undefined_tag %}", "undefined tag"},
	{"{% assign v x y z %}", "syntax error"},
	{"{% if syntax error %}", `unterminated "if" block`},
	// TODO once expression parsing is moved to template parse stage
	// {"{% if syntax error %}{% endif %}", "syntax error"},
	// {"{% for a in ar undefined %}{{ a }} {% endfor %}", "TODO"},
}

var tagTests = []struct{ in, expected string }{
	// variable tags
	{`{% assign av = 1 %}{{ av }}`, "1"},
	{`{% assign av = obj.a %}{{ av }}`, "1"},
	{`{% assign av = (1..5) %}{{ av }}`, "{1 5}"},
	{`{% capture x %}captured{% endcapture %}{{ x }}`, "captured"},
	
	// dot notation assignments
	{`{% assign page.canonical_url = "/about/" %}{{ page.canonical_url }}`, "/about/"},
	{`{% assign page.meta.description = "Test description" %}{{ page.meta.description }}`, "Test description"},
	{`{% assign obj.nested = 42 %}{{ obj.nested }}`, "42"},
	{`{% assign new_obj.prop = "value" %}{{ new_obj.prop }}`, "value"},
	{`{% assign page.title = "New Title" %}{{ page.title }}`, "New Title"},

	// TODO research whether Liquid requires matching interior tags
	{`{% comment %}{{ a }}{% undefined_tag %}{% endcomment %}`, ""},

	// TODO research whether Liquid requires matching interior tags
	{`pre{% raw %}{{ a }}{% undefined_tag %}{% endraw %}post`, "pre{{ a }}{% undefined_tag %}post"},
	{`pre{% raw %}{% if false %}anyway-{% endraw %}post`, "pre{% if false %}anyway-post"},
}

var tagErrorTests = []struct{ in, expected string }{
	{`{% assign av = x | undefined_filter %}`, "undefined filter"},
}

// this is also used in the other test files
var tagTestBindings = map[string]any{
	"x": 123,
	"obj": map[string]any{
		"a": 1,
	},
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"pages": []map[string]any{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
	},
	"sort_prop": []map[string]any{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	"page": map[string]any{
		"title": "Introduction",
		"meta": map[string]any{
			"author": "John Doe",
		},
	},
}

func TestStandardTags_parse_errors(t *testing.T) {
	settings := render.NewConfig()
	AddStandardTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := settings.Compile(test.in, parser.SourceLoc{})
			require.Nilf(t, root, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestStandardTags(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range tagTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = render.Render(root, buf, tagTestBindings, config)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.expected, buf.String(), test.in)
		})
	}
}

func TestStandardTags_render_errors(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(config)
	for i, test := range tagErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := config.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = render.Render(root, io.Discard, tagTestBindings, config)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
