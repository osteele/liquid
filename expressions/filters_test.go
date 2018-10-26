package expressions

import (
	"fmt"
	"testing"

	"github.com/urbn8/liquid/values"
	"github.com/stretchr/testify/require"
)

func TestContext_AddFilter(t *testing.T) {
	cfg := NewConfig()
	require.NotPanics(t, func() { cfg.AddFilter("f", func(int) int { return 0 }) })
	require.NotPanics(t, func() { cfg.AddFilter("f", func(int) (a int, e error) { return }) })
	require.Panics(t, func() { cfg.AddFilter("f", func() int { return 0 }) })
	require.Panics(t, func() { cfg.AddFilter("f", func(int) {}) })
	// require.Panics(t, func() { cfg.AddFilter("f", func(int) (a int, b int) { return }) })
	require.Panics(t, func() { cfg.AddFilter("f", func(int) (a int, e error, b int) { return }) })
	require.Panics(t, func() { cfg.AddFilter("f", 10) })
}

func TestContext_runFilter(t *testing.T) {
	cfg := NewConfig()
	constant := func(value interface{}) valueFn {
		return func(Context) values.Value { return values.ValueOf(value) }
	}
	receiver := constant("self")

	// basic
	cfg.AddFilter("f1", func(s string) string {
		return "<" + s + ">"
	})
	ctx := NewContext(map[string]interface{}{"x": 10}, cfg)
	out, err := ctx.ApplyFilter("f1", receiver, []valueFn{})
	require.NoError(t, err)
	require.Equal(t, "<self>", out)

	// filter argument
	cfg.AddFilter("with_arg", func(a, b string) string {
		return fmt.Sprintf("(%s, %s)", a, b)
	})
	ctx = NewContext(map[string]interface{}{"x": 10}, cfg)
	out, err = ctx.ApplyFilter("with_arg", receiver, []valueFn{constant("arg")})
	require.NoError(t, err)
	require.Equal(t, "(self, arg)", out)

	// TODO optional argument
	// TODO error return

	// extra argument
	_, err = ctx.ApplyFilter("with_arg", receiver, []valueFn{constant(1), constant(2)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of arguments")
	require.Contains(t, err.Error(), "given 2")
	require.Contains(t, err.Error(), "expected 1")

	// closure
	cfg.AddFilter("add", func(a, b int) int {
		return a + b
	})
	cfg.AddFilter("closure", func(a string, expr Closure) (string, error) {
		value, e := expr.Bind("y", 1).Evaluate()
		if e != nil {
			return "", e
		}
		return fmt.Sprintf("(%v, %v)", a, value), nil
	})
	ctx = NewContext(map[string]interface{}{"x": 10}, cfg)
	out, err = ctx.ApplyFilter("closure", receiver, []valueFn{constant("x |add: y")})
	require.NoError(t, err)
	require.Equal(t, "(self, 11)", out)
}
