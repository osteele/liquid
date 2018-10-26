package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/urbn8/liquid/parser"
	"github.com/stretchr/testify/require"
)

var renderTests = []struct{ in, out string }{
	// literal representations
	{`{{ nil }}`, ""},
	{`{{ true }}`, "true"},
	{`{{ false }}`, "false"},
	{`{{ 12 }}`, "12"},
	{`{{ 12.3 }}`, "12.3"},
	{`{{ date }}`, "2015-07-17 15:04:05 +0000"},
	{`{{ "string" }}`, "string"},
	{`{{ array }}`, "firstsecondthird"},

	// variables and properties
	{`{{ int }}`, "123"},
	{`{{ page.title }}`, "Introduction"},
	{`{{ array[1] }}`, "second"},

	// whitespace control
	{` {{ 1 }} `, " 1 "},
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
	"array": []string{"first", "second", "third"},
	"date":  time.Date(2015, 7, 17, 15, 4, 5, 123456789, time.UTC),
	"int":   123,
	"sort_prop": []map[string]interface{}{
		{"weight": 1},
		{"weight": 5},
		{"weight": 3},
		{"weight": nil},
	},
	// for examples from liquid docs
	"animals": []string{"zebra", "octopus", "giraffe", "Sally Snake"},
	"page": map[string]interface{}{
		"title": "Introduction",
	},
	"pages": []map[string]interface{}{
		{"category": "business"},
		{"category": "celebrities"},
		{},
		{"category": "lifestyle"},
		{"category": "sports"},
		{},
		{"category": "technology"},
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
