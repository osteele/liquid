package expression

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContext_runFilter(t *testing.T) {
	cfg := NewConfig()
	constant := func(value interface{}) valueFn {
		return func(Context) interface{} { return value }
	}
	receiver := constant("self")

	// basic
	cfg.AddFilter("f1", func(s string) string {
		return "<" + s + ">"
	})
	ctx := NewContext(map[string]interface{}{"x": 10}, cfg)
	out := ctx.ApplyFilter("f1", receiver, []valueFn{})
	require.Equal(t, "<self>", out)

	// filter argument
	cfg.AddFilter("with_arg", func(a, b string) string {
		return fmt.Sprintf("(%s, %s)", a, b)
	})
	ctx = NewContext(map[string]interface{}{"x": 10}, cfg)
	out = ctx.ApplyFilter("with_arg", receiver, []valueFn{constant("arg")})
	require.Equal(t, "(self, arg)", out)

	// TODO optional argument
	// TODO error return

	// closure
	cfg.AddFilter("add", func(a, b int) int {
		return a + b
	})
	cfg.AddFilter("closure", func(a string, expr Closure) (string, error) {
		value, err := expr.Bind("y", 1).Evaluate()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%v, %v)", a, value), nil
	})
	ctx = NewContext(map[string]interface{}{"x": 10}, cfg)
	out = ctx.ApplyFilter("closure", receiver, []valueFn{constant("x |add: y")})
	require.Equal(t, "(self, 11)", out)
}
