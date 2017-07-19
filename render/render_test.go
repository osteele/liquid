package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/stretchr/testify/require"
)

func addRenderTestTags(cfg Config) {
	cfg.AddTag("y", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, _ Context) error {
			_, err := io.WriteString(w, "y")
			return err
		}, nil
	})
	cfg.AddTag("null", func(string) (func(io.Writer, Context) error, error) {
		return func(io.Writer, Context) error { return nil }, nil
	})
	cfg.AddBlock("errblock").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			return fmt.Errorf("errblock error")
		}, nil
	})
}

var renderTests = []struct{ in, out string }{
	// literals representations
	{`{{ nil }}`, ""},
	{`{{ true }}`, "true"},
	{`{{ false }}`, "false"},
	{`{{ 12 }}`, "12"},
	{`{{ 12.3 }}`, "12.3"},
	{`{{ "abc" }}`, "abc"},
	{`{{ array }}`, "firstsecondthird"},

	// variables and properties
	{`{{ x }}`, "123"},
	{`{{ page.title }}`, "Introduction"},
	{`{{ array[1] }}`, "second"},

	// whitespace control
	// {` {{ 1 }} `, " 1 "},
	{` {{- 1 }} `, "1 "},
	{` {{ 1 -}} `, " 1"},
	{` {{- 1 -}} `, "1"},
	{` {{- nil -}} `, ""},
	{`x {{ 1 }} z`, "x 1 z"},
	{`x {{- 1 }} z`, "x1 z"},
	{`x {{ 1 -}} z`, "x 1z"},
	{`x {{- 1 -}} z`, "x1z"},
	{`x {{ nil }} z`, "x  z"},
	{`x {{- nil }} z`, "x z"},
	{`x {{ nil -}} z`, "x z"},
	{`x {{- nil -}} z`, "xz"},
	{`x {% null %} z`, "x  z"},
	{`x {%- null %} z`, "x z"},
	{`x {% null -%} z`, "x z"},
	{`x {%- null -%} z`, "xz"},
	{`x {% y %} z`, "x y z"},
	{`x {%- y %} z`, "xy z"},
	{`x {% y -%} z`, "x yz"},
	{`x {%- y -%} z`, "xyz"},
}

var renderErrorTests = []struct{ in, out string }{
	{`{% errblock %}{% enderrblock %}`, "errblock error"},
}

var renderTestBindings = map[string]interface{}{
	"x":     123,
	"array": []string{"first", "second", "third"},
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

func TestRender(t *testing.T) {
	cfg := NewConfig()
	addRenderTestTags(cfg)
	for i, test := range renderTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = Render(root, buf, renderTestBindings, cfg)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.out, buf.String(), test.in)
		})
	}
}

func TestRenderErrors(t *testing.T) {
	cfg := NewConfig()
	addRenderTestTags(cfg)
	for i, test := range renderErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = Render(root, ioutil.Discard, renderTestBindings, cfg)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.out, test.in)
		})
	}
}
