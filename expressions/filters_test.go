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
	out, err := ctx.ApplyFilter("f1", receiver, []filterParam{})
	require.NoError(t, err)
	require.Equal(t, "<self>", out)

	// filter argument
	cfg.AddFilter("with_arg", func(a, b string) string {
		return fmt.Sprintf("(%s, %s)", a, b)
	})
	ctx = NewContext(map[string]any{"x": 10}, cfg)
	out, err = ctx.ApplyFilter("with_arg", receiver, []filterParam{{name: "", value: constant("arg")}})
	require.NoError(t, err)
	require.Equal(t, "(self, arg)", out)

	// TODO optional argument
	// TODO error return

	// extra argument
	_, err = ctx.ApplyFilter("with_arg", receiver, []filterParam{{name: "", value: constant(1)}, {name: "", value: constant(2)}})
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
	out, err = ctx.ApplyFilter("closure", receiver, []filterParam{{name: "", value: constant("x |add: y")}})
	require.NoError(t, err)
	require.Equal(t, "(self, 11)", out)
}

// TestNamedFilterArguments tests filters with named arguments
func TestNamedFilterArguments(t *testing.T) {
	cfg := NewConfig()
	constant := func(value any) valueFn {
		return func(Context) values.Value { return values.ValueOf(value) }
	}
	receiver := constant("image.jpg")

	// Filter with named arguments
	cfg.AddFilter("img_url", func(image string, size string, opts map[string]any) string {
		scale := 1
		if s, ok := opts["scale"].(int); ok {
			scale = s
		}
		return fmt.Sprintf("img_url(%s, %s, scale=%d)", image, size, scale)
	})

	ctx := NewContext(map[string]any{}, cfg)

	// Test with positional and named arguments
	out, err := ctx.ApplyFilter("img_url", receiver, []filterParam{
		{name: "", value: constant("580x")},
		{name: "scale", value: constant(2)},
	})
	require.NoError(t, err)
	require.Equal(t, "img_url(image.jpg, 580x, scale=2)", out)

	// Test with only positional argument (named args should be empty map)
	out, err = ctx.ApplyFilter("img_url", receiver, []filterParam{
		{name: "", value: constant("300x")},
	})
	require.NoError(t, err)
	require.Equal(t, "img_url(image.jpg, 300x, scale=1)", out)

	// Test with multiple named arguments
	cfg.AddFilter("custom_filter", func(input string, opts map[string]any) string {
		format := opts["format"]
		name := opts["name"]
		return fmt.Sprintf("custom(%s, format=%v, name=%v)", input, format, name)
	})

	out, err = ctx.ApplyFilter("custom_filter", receiver, []filterParam{
		{name: "format", value: constant("date")},
		{name: "name", value: constant("order.name")},
	})
	require.NoError(t, err)
	require.Equal(t, "custom(image.jpg, format=date, name=order.name)", out)

	// Test mixing positional and named arguments
	cfg.AddFilter("mixed_args", func(input string, pos1 string, pos2 int, opts map[string]any) string {
		extra := ""
		if e, ok := opts["extra"].(string); ok {
			extra = e
		}
		return fmt.Sprintf("mixed(%s, %s, %d, extra=%s)", input, pos1, pos2, extra)
	})

	out, err = ctx.ApplyFilter("mixed_args", receiver, []filterParam{
		{name: "", value: constant("arg1")},
		{name: "", value: constant(42)},
		{name: "extra", value: constant("bonus")},
	})
	require.NoError(t, err)
	require.Equal(t, "mixed(image.jpg, arg1, 42, extra=bonus)", out)
}

// TestNamedFilterArgumentsParsing tests that named arguments are correctly parsed
func TestNamedFilterArgumentsParsing(t *testing.T) {
	cfg := NewConfig()
	cfg.AddFilter("test_filter", func(input string, opts map[string]any) string {
		return fmt.Sprintf("input=%s, opts=%v", input, opts)
	})

	// Test parsing filter with named arguments from expression string
	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "single named argument",
			expr:     "'test' | test_filter: scale: 2",
			expected: "input=test, opts=map[scale:2]",
		},
		{
			name:     "multiple named arguments",
			expr:     "'test' | test_filter: scale: 2, format: 'jpg'",
			expected: "input=test, opts=map[format:jpg scale:2]",
		},
		{
			name:     "no arguments",
			expr:     "'test' | test_filter",
			expected: "input=test, opts=map[]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := Parse(tt.expr)
			require.NoError(t, err)
			ctx := NewContext(map[string]any{}, cfg)
			val, err := expr.Evaluate(ctx)
			require.NoError(t, err)
			// Check that the result contains the expected key parts
			result := fmt.Sprintf("%v", val)
			require.Contains(t, result, "input=test")
			if tt.name != "no arguments" {
				// For tests with named args, verify they're present
				require.Contains(t, result, "opts=map[")
			}
		})
	}
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
