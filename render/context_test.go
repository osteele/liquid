package render

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
)

func addContextTestTags(s Config) {
	s.AddTag("test_bindings", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			b := c.Bindings()
			_, err := fmt.Fprintf(w, "%v", b["x"])
			return err
		}, nil
	})
	s.AddTag("test_get", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			v := c.Get("x")
			_, err := fmt.Fprintf(w, "%v", v)
			return err
		}, nil
	})
	s.AddTag("test_set", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			c.Set("x", 999)
			_, err := fmt.Fprintf(w, "%v", c.Get("x"))
			return err
		}, nil
	})
	s.AddBlock("test_inner_string").Compiler(func(bn BlockNode) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			s, err := c.InnerString()
			if err != nil {
				return err
			}
			_, err = io.WriteString(w, "inner:"+s)
			return err
		}, nil
	})
	s.AddBlock("test_render_children").Compiler(func(bn BlockNode) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			_, _ = io.WriteString(w, "before:")
			rerr := c.RenderChildren(w)
			if rerr != nil {
				return rerr
			}
			_, err := io.WriteString(w, ":after")
			return err
		}, nil
	})
	s.AddTag("test_set_path", func(string) (func(io.Writer, Context) error, error) {
		return func(w io.Writer, c Context) error {
			err := c.SetPath([]string{"page", "url"}, "/about/")
			if err != nil {
				return err
			}
			v := c.Get("page")
			m, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("page is not a map")
			}
			_, err = fmt.Fprintf(w, "%v", m["url"])
			return err
		}, nil
	})
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
			s, err := c.RenderFile(filename, map[string]any{"shadowed": 2})
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, s)

			return err
		}, nil
	})
	s.AddBlock("test_block_sourcefile").Compiler(func(c BlockNode) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			_, err := io.WriteString(w, c.SourceFile())
			return err
		}, nil
	})
	s.AddBlock("test_block_wraperror").Compiler(func(c BlockNode) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			return c.WrapError(errors.New("giftwrapped"))
		}, nil
	})
	s.AddBlock("test_block_errorf").Compiler(func(c BlockNode) (func(w io.Writer, c Context) error, error) {
		return func(w io.Writer, c Context) error {
			return c.Errorf("giftwrapped")
		}, nil
	})
}

var contextTests = []struct{ in, out string }{
	{`{% parse args %}{% endparse %}`, "args"},
	{`{% test_evaluate_string x %}`, "123"},
	{`{% test_expand_tag_arg x %}`, "x"},
	{`{% test_expand_tag_arg {{x}} %}`, "123"},
	{`{% test_tag_name %}`, "test_tag_name"},
	{
		`{% test_render_file testdata/render_file.txt %}; unshadowed={{ shadowed }}`,
		"rendered shadowed=2; unshadowed=1",
	},
	{`{% test_block_sourcefile %}x{% endtest_block_sourcefile %}`, ``},
	{`{% test_bindings %}`, "123"},
	{`{% test_get %}`, "123"},
	{`{% test_set %}`, "999"},
	{`{% test_inner_string %}hello world{% endtest_inner_string %}`, "inner:hello world"},
	{`{% test_render_children %}content{% endtest_render_children %}`, "before:content:after"},
	{`{% test_set_path %}`, "/about/"},
}

var contextErrorTests = []struct{ in, expect string }{
	{`{% test_evaluate_string syntax error %}`, "syntax error"},
	{`{% test_expand_tag_arg {{ syntax error }} %}`, "syntax error"},
	{`{% test_expand_tag_arg {{ x | undefined_filter }} %}`, "undefined filter"},
	{`{% test_render_file testdata/render_file_syntax_error.txt %}`, "syntax error"},
	{`{% test_render_file testdata/render_file_runtime_error.txt %}`, "undefined tag"},
	{`{% test_block_wraperror %}{% endtest_block_wraperror %}`, "giftwrapped"},
	{`{% test_block_errorf %}{% endtest_block_errorf %}`, "giftwrapped"},
}

var contextTestBindings = map[string]any{
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
			err = Render(root, io.Discard, contextTestBindings, cfg)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expect, test.in)
		})
	}
}

func TestSetPath(t *testing.T) {
	cfg := NewConfig()
	addContextTestTags(cfg)

	t.Run("single path", func(t *testing.T) {
		cfg.AddTag("sp_single", func(string) (func(io.Writer, Context) error, error) {
			return func(w io.Writer, c Context) error {
				err := c.SetPath([]string{"newvar"}, 42)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintf(w, "%v", c.Get("newvar"))
				return err
			}, nil
		})
		root, err := cfg.Compile(`{% sp_single %}`, parser.SourceLoc{})
		require.NoError(t, err)
		buf := new(bytes.Buffer)
		err = Render(root, buf, map[string]any{}, cfg)
		require.NoError(t, err)
		require.Equal(t, "42", buf.String())
	})

	t.Run("intermediate creation", func(t *testing.T) {
		cfg.AddTag("sp_create", func(string) (func(io.Writer, Context) error, error) {
			return func(w io.Writer, c Context) error {
				err := c.SetPath([]string{"a", "b", "c"}, "deep")
				if err != nil {
					return err
				}
				a := c.Get("a")
				m1 := a.(map[string]any)
				m2 := m1["b"].(map[string]any)
				_, err = fmt.Fprintf(w, "%v", m2["c"])
				return err
			}, nil
		})
		root, err := cfg.Compile(`{% sp_create %}`, parser.SourceLoc{})
		require.NoError(t, err)
		buf := new(bytes.Buffer)
		err = Render(root, buf, map[string]any{}, cfg)
		require.NoError(t, err)
		require.Equal(t, "deep", buf.String())
	})

	t.Run("error on non-map", func(t *testing.T) {
		cfg.AddTag("sp_nonmap", func(string) (func(io.Writer, Context) error, error) {
			return func(w io.Writer, c Context) error {
				return c.SetPath([]string{"x", "sub"}, "val")
			}, nil
		})
		root, err := cfg.Compile(`{% sp_nonmap %}`, parser.SourceLoc{})
		require.NoError(t, err)
		// x=123 (int), so SetPath should fail
		err = Render(root, io.Discard, map[string]any{"x": 123}, cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot set property on non-object")
	})

	t.Run("empty path", func(t *testing.T) {
		cfg.AddTag("sp_empty", func(string) (func(io.Writer, Context) error, error) {
			return func(w io.Writer, c Context) error {
				return c.SetPath([]string{}, "val")
			}, nil
		})
		root, err := cfg.Compile(`{% sp_empty %}`, parser.SourceLoc{})
		require.NoError(t, err)
		err = Render(root, io.Discard, map[string]any{}, cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty path")
	})
}

func TestContext_file_not_found_error(t *testing.T) {
	// Test the cause instead of looking for a string, since the error message is
	// different between Darwin and Linux ("no such file") and Windows ("The
	// system cannot find the file specified"), at least.
	//
	// Also see TestIncludeTag_file_not_found_error.
	cfg := NewConfig()
	addContextTestTags(cfg)
	root, err := cfg.Compile(`{% test_render_file testdata/missing_file %}`, parser.SourceLoc{})
	require.NoError(t, err)
	err = Render(root, io.Discard, contextTestBindings, cfg)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err.Cause()))
}
