package liquid

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIterationKeyedMap(t *testing.T) {
	vars := map[string]any{
		"keyed_map": IterationKeyedMap(map[string]any{"a": 1, "b": 2}),
	}
	engine := NewEngine()
	tpl, err := engine.ParseTemplate([]byte(`{% for k in keyed_map %}{{ k }}={{ keyed_map[k] }}.{% endfor %}`))
	require.NoError(t, err)
	out, err := tpl.RenderString(vars)
	require.NoError(t, err)
	require.Equal(t, "a=1.b=2.", out)
}

// TestNamedFilterArgumentsIntegration tests issue #42 - named filter arguments
func TestNamedFilterArgumentsIntegration(t *testing.T) {
	engine := NewEngine()

	// Register test filters that support named arguments
	engine.RegisterFilter("img_url", func(image string, size string, opts map[string]any) string {
		scale := 1
		if s, ok := opts["scale"].(int); ok {
			scale = s
		}
		return fmt.Sprintf("https://cdn.example.com/%s?size=%s&scale=%d", image, size, scale)
	})

	engine.RegisterFilter("date", func(dateStr string, opts map[string]any) string {
		format := "default"
		if f, ok := opts["format"].(string); ok {
			format = f
		}
		return fmt.Sprintf("date(%s, format=%s)", dateStr, format)
	})

	engine.RegisterFilter("t", func(key string, opts map[string]any) string {
		name := ""
		if n, ok := opts["name"].(string); ok {
			name = n
		}
		return fmt.Sprintf("translate(%s, name=%s)", key, name)
	})

	vars := map[string]any{
		"image": "product.jpg",
		"order": map[string]any{
			"created_at": "2023-01-15",
			"name":       "Order #123",
		},
	}

	// Test case from issue #42: {{image | img_url: '580x', scale: 2}}
	out, err := engine.ParseAndRenderString(`{{image | img_url: '580x', scale: 2}}`, vars)
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example.com/product.jpg?size=580x&scale=2", out)

	// Test case from issue #42: {{ order.created_at | date: format: 'date' }}
	out, err = engine.ParseAndRenderString(`{{ order.created_at | date: format: 'date' }}`, vars)
	require.NoError(t, err)
	require.Equal(t, "date(2023-01-15, format=date)", out)

	// Test case from issue #42: {{ 'customer.order.title' | t: name: order.name }}
	out, err = engine.ParseAndRenderString(`{{ 'customer.order.title' | t: name: order.name }}`, vars)
	require.NoError(t, err)
	require.Equal(t, "translate(customer.order.title, name=Order #123)", out)

	// Test with mixed positional and named arguments
	out, err = engine.ParseAndRenderString(`{{image | img_url: '300x'}}`, vars)
	require.NoError(t, err)
	require.Equal(t, "https://cdn.example.com/product.jpg?size=300x&scale=1", out)
}

func ExampleIterationKeyedMap() {
	vars := map[string]any{
		"map":       map[string]any{"a": 1},
		"keyed_map": IterationKeyedMap(map[string]any{"a": 1}),
	}
	engine := NewEngine()

	out, err := engine.ParseAndRenderString(
		`{% for k in map %}{{ k[0] }}={{ k[1] }}.{% endfor %}`, vars)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)

	out, err = engine.ParseAndRenderString(
		`{% for k in keyed_map %}{{ k }}={{ keyed_map[k] }}.{% endfor %}`, vars)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(out)
	// Output: a=1.
	// a=1.
}
