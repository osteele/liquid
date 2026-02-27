package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/values"
)

func TestContext_AddFilter(t *testing.T) {
	cfg := NewConfig()

	require.NotPanics(t, func() { cfg.AddFilter("f", func(int) int { return 0 }) })
	require.NotPanics(t, func() { cfg.AddFilter("f", func(int) (a int, e error) { return }) })
	require.Panics(t, func() { cfg.AddFilter("f", func() int { return 0 }) })
	require.Panics(t, func() { cfg.AddFilter("f", func(int) {}) })
	// require.Panics(t, func() { cfg.AddFilter("f", func(int) (a int, b int) { return }) })
	//nolint:staticcheck
	require.Panics(t, func() { cfg.AddFilter("f", func(int) (a int, e error, b int) { return }) })
	require.Panics(t, func() { cfg.AddFilter("f", 10) })
}

func TestContext_runFilter(t *testing.T) {
	cfg := NewConfig()
	constant := func(value any) valueFn {
		return func(Context) values.Value { return values.ValueOf(value) }
	}
	receiver := constant("self")

	// basic
	cfg.AddFilter("f1", func(s string) string {
		return "<" + s + ">"
	})
	ctx := NewContext(map[string]any{"x": 10}, cfg)
	out, err := ctx.ApplyFilter("f1", receiver, nil)
	require.NoError(t, err)
	require.Equal(t, "<self>", out)

	// filter argument
	cfg.AddFilter("with_arg", func(a, b string) string {
		return fmt.Sprintf("(%s, %s)", a, b)
	})
	ctx = NewContext(map[string]any{"x": 10}, cfg)
	out, err = ctx.ApplyFilter("with_arg", receiver, &filterArgs{positional: []valueFn{constant("arg")}})
	require.NoError(t, err)
	require.Equal(t, "(self, arg)", out)

	// TODO optional argument
	// TODO error return

	// extra argument
	_, err = ctx.ApplyFilter("with_arg", receiver, &filterArgs{positional: []valueFn{constant(1), constant(2)}})
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
	ctx = NewContext(map[string]any{"x": 10}, cfg)
	out, err = ctx.ApplyFilter("closure", receiver, &filterArgs{positional: []valueFn{constant("x |add: y")}})
	require.NoError(t, err)
	require.Equal(t, "(self, 11)", out)
}

func TestContext_runFilter_keywordArgs(t *testing.T) {
	cfg := NewConfig()
	constant := func(value any) valueFn {
		return func(Context) values.Value { return values.ValueOf(value) }
	}
	receiver := constant("self")

	cfg.AddFilter("with_kwargs", func(s string, kwargs ...map[string]any) string {
		if len(kwargs) > 0 {
			if v, ok := kwargs[0]["option"]; ok {
				return fmt.Sprintf("(%s, option=%v)", s, v)
			}
		}
		return fmt.Sprintf("(%s)", s)
	})

	ctx := NewContext(map[string]any{}, cfg)

	// without keyword args
	out, err := ctx.ApplyFilter("with_kwargs", receiver, nil)
	require.NoError(t, err)
	require.Equal(t, "(self)", out)

	// with keyword args
	out, err = ctx.ApplyFilter("with_kwargs", receiver, &filterArgs{
		keyword: []keywordArg{{name: "option", val: constant(true)}},
	})
	require.NoError(t, err)
	require.Equal(t, "(self, option=true)", out)
}

// TestAddSafeFilterNilMap verifies that AddSafeFilter doesn't panic
// when called on a Config with nil filters map
func TestAddSafeFilterNilMap(t *testing.T) {
	// Create a config without initializing filters map
	cfg := &Config{}

	// This should not panic even though filters map is nil
	require.NotPanics(t, func() {
		cfg.AddSafeFilter()
	}, "AddSafeFilter should not panic with nil filters map")

	// Verify the safe filter was added
	require.NotNil(t, cfg.filters)
	require.NotNil(t, cfg.filters["safe"])

	// Test that calling AddSafeFilter again doesn't duplicate
	cfg.AddSafeFilter()
	require.NotNil(t, cfg.filters["safe"])

	// Test the safe filter works correctly
	safeFilter := cfg.filters["safe"].(func(interface{}) interface{})

	// Test with regular value
	result := safeFilter("test")
	safeVal, ok := result.(values.SafeValue)
	require.True(t, ok, "Should return SafeValue")
	require.Equal(t, "test", safeVal.Value)

	// Test with already safe value
	alreadySafe := values.SafeValue{Value: "already safe"}
	result2 := safeFilter(alreadySafe)
	require.Equal(t, alreadySafe, result2, "Should return the same SafeValue if already safe")
}
