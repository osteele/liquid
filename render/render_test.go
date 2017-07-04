package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func addRenderTestTags(s Config) {
	s.AddBlock("parse").Parser(func(c ASTBlock) (func(io.Writer, Context) error, error) {
		a := c.Args
		return func(w io.Writer, c Context) error {
			_, err := w.Write([]byte(a))
			return err
		}, nil
	})
	s.AddBlock("eval").Renderer(func(w io.Writer, c Context) error {
		v, err := c.EvaluateString(c.TagArgs())
		if err != nil {
			return err
		}
		s := fmt.Sprint(v)
		_, err = w.Write([]byte(s))
		return err
	})
	s.AddBlock("err2").Parser(func(c ASTBlock) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			return fmt.Errorf("stage 2 error")
		}, nil
	})
}

var renderTests = []struct{ in, out string }{
	{`{{ 12 }}`, "12"},
	{`{{ x }}`, "123"},
	{`{{ page.title }}`, "Introduction"},
	{`{{ ar[1] }}`, "second"},
	{`{% parse args %}{% endparse %}`, "args"},
	{`{% eval x %}{% endeval %}`, "123"},
}

var renderErrorTests = []struct{ in, out string }{
	// {"{%if syntax error%}{%endif%}", "parse error"},
	{`{% err2 %}{% enderr2 %}`, "stage 2 error"},
}

var renderTestBindings = map[string]interface{}{
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
}

func TestRender(t *testing.T) {
	settings := NewConfig()
	addRenderTestTags(settings)
	context := newNodeContext(renderTestBindings, settings)
	for i, test := range renderTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := settings.Parse(test.in)
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = renderNode(ast, buf, context)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.out, buf.String(), test.in)
		})
	}
}

func TestRenderErrors(t *testing.T) {
	settings := NewConfig()
	addRenderTestTags(settings)
	context := newNodeContext(renderTestBindings, settings)
	for i, test := range renderErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := settings.Parse(test.in)
			require.NoErrorf(t, err, test.in)
			err = renderNode(ast, ioutil.Discard, context)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.out, test.in)
		})
	}
}
