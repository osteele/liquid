package render

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/stretchr/testify/require"
)

func addContextTestTags(s Config) {
	s.AddBlock("eval").Renderer(func(w io.Writer, c Context) error {
		v, err := c.EvaluateString(c.TagArgs())
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(fmt.Sprint(v)))
		return err
	})
	s.AddBlock("parse").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		a := c.Args
		return func(w io.Writer, c Context) error {
			_, err := w.Write([]byte(a))
			return err
		}, nil
	})
	s.AddTag("tag_name", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			_, err := w.Write([]byte(c.TagName()))
			return err
		}, nil
	})
	s.AddTag("expand_arg", func(string) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			s, err := c.ExpandTagArg()
			if err != nil {
				return err
			}
			_, err = w.Write([]byte(s))
			return err
		}, nil
	})
}

var contextTests = []struct{ in, out string }{
	{`{% parse args %}{% endparse %}`, "args"},
	{`{% eval x %}{% endeval %}`, "123"},
	{`{% expand_arg x %}`, "x"},
	{`{% expand_arg {{x}} %}`, "123"},
	{`{% tag_name %}`, "tag_name"},
}

var contextTestBindings = map[string]interface{}{
	"x": 123,
}

func TestContext(t *testing.T) {
	cfg := NewConfig()
	addContextTestTags(cfg)
	context := newNodeContext(contextTestBindings, cfg)
	for i, test := range contextTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = renderNode(ast, buf, context)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.out, buf.String(), test.in)
		})
	}
}
