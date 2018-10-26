package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/urbn8/liquid/parser"
	"github.com/stretchr/testify/require"
)

func addContextTestTags(s Config) {
	s.AddTag("test_evaluate_string", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			v, err := c.EvaluateString(c.TagArgs())
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(w, v)
			return err
		}, nil
	})
	s.AddBlock("parse").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		a := c.Args
		return func(w io.Writer, c Context) error {
			_, err := io.WriteString(w, a)
			return err
		}, nil
	})
	s.AddTag("test_tag_name", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			_, err := io.WriteString(w, c.TagName())
			return err
		}, nil
	})
	s.AddTag("test_expand_tag_arg", func(string) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			s, err := c.ExpandTagArg()
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, s)
			return err
		}, nil
	})
	s.AddTag("test_render_file", func(filename string) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			s, err := c.RenderFile(filename, map[string]interface{}{"shadowed": 2})
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, s)
			return err
		}, nil
	})
}

var contextTests = []struct{ in, out string }{
	{`{% parse args %}{% endparse %}`, "args"},
	{`{% test_evaluate_string x %}`, "123"},
	{`{% test_expand_tag_arg x %}`, "x"},
	{`{% test_expand_tag_arg {{x}} %}`, "123"},
	{`{% test_tag_name %}`, "test_tag_name"},
	{`{% test_render_file testdata/render_file.txt %}; unshadowed={{ shadowed }}`,
		"rendered shadowed=2; unshadowed=1"},
}

var contextErrorTests = []struct{ in, expect string }{
	{`{% test_evaluate_string syntax error %}`, "syntax error"},
	{`{% test_expand_tag_arg {{ syntax error }} %}`, "syntax error"},
	{`{% test_expand_tag_arg {{ x | undefined_filter }} %}`, "undefined filter"},
	{`{% test_render_file testdata/render_file_syntax_error.txt %}`, "syntax error"},
	{`{% test_render_file testdata/render_file_runtime_error.txt %}`, "undefined tag"},
}

var contextTestBindings = map[string]interface{}{
	"x":        123,
	"shadowed": 1,
}

func TestContext(t *testing.T) {
	cfg := NewConfig()
	addContextTestTags(cfg)
	for i, test := range contextTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			buf := new(bytes.Buffer)
			err = Render(root, buf, contextTestBindings, cfg)
			require.NoErrorf(t, err, test.in)
			require.Equalf(t, test.out, buf.String(), test.in)
		})
	}
}

func TestContext_errors(t *testing.T) {
	cfg := NewConfig()
	addContextTestTags(cfg)
	for i, test := range contextErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			root, err := cfg.Compile(test.in, parser.SourceLoc{})
			require.NoErrorf(t, err, test.in)
			err = Render(root, ioutil.Discard, contextTestBindings, cfg)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expect, test.in)
		})
	}
}

func TestContext_file_not_found_error(t *testing.T) {
	// Test the cause instead of looking for a string, since the error message is different
	// between Darwin and Linux ("no such file") and Windows ("The system cannot find the file specified"), at least.
	//
	// Also see TestIncludeTag_file_not_found_error.
	cfg := NewConfig()
	addContextTestTags(cfg)
	root, err := cfg.Compile(`{% test_render_file testdata/missing_file %}`, parser.SourceLoc{})
	require.NoError(t, err)
	err = Render(root, ioutil.Discard, contextTestBindings, cfg)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err.Cause()))
}
