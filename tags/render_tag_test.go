package tags

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var renderTestBindings = map[string]any{
	"secret":  "hidden-value",
	"visible": "visible-value",
	"product": map[string]any{
		"name":  "Widget",
		"price": 9.99,
	},
	"products": []any{
		map[string]any{"name": "Item1"},
		map[string]any{"name": "Item2"},
		map[string]any{"name": "Item3"},
	},
}

// TestRenderTag_Basic tests basic render functionality
func TestRenderTag_Basic(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	root, err := config.Compile(`{% render "render_basic.html" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "Hello from render!", strings.TrimSpace(buf.String()))
}

// TestRenderTag_WithParameters tests render with explicit parameters
func TestRenderTag_WithParameters(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Test with literal and variable values
	root, err := config.Compile(`{% render "render_with_params.html", title: "Widget", price: 9.99 %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "Title: Widget, Price: 9.99", strings.TrimSpace(buf.String()))
}

// TestRenderTag_IsolatedScope tests that parent variables are not accessible
func TestRenderTag_IsolatedScope(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Render should NOT have access to 'secret' from parent, but should have 'visible' passed as parameter
	root, err := config.Compile(`{% render "render_isolated.html", visible: visible %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)

	// 'secret' should be empty (not accessible), 'visible' should be present
	result := strings.TrimSpace(buf.String())
	require.Equal(t, "Secret: , Visible: visible-value", result)
}

// TestRenderTag_WithObject tests "render with object" syntax
func TestRenderTag_WithObject(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Test: {% render "template" with object as item %}
	root, err := config.Compile(`{% render "render_with_object.html" with product as item %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "Product: Widget - $9.99", strings.TrimSpace(buf.String()))
}

// TestRenderTag_ForLoop tests "render for array" syntax
func TestRenderTag_ForLoop(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Test: {% render "template" for array as item %}
	root, err := config.Compile(`{% render "render_for_loop.html" for products as item %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)

	// Should render once for each item with forloop object
	result := strings.TrimSpace(buf.String())
	require.Contains(t, result, "Item1 (1/3)")
	require.Contains(t, result, "Item2 (2/3)")
	require.Contains(t, result, "Item3 (3/3)")
}

// TestRenderTag_ForLoopWithParams tests combining "for" with explicit parameters
func TestRenderTag_ForLoopWithParams(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Create a template that uses both the loop item and a parameter
	config.Cache["testdata/render_combined.html"] = []byte(`{{ item.name }} - {{ label }}`)

	root, err := config.Compile(`{% render "render_combined.html" for products as item, label: "Product" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, renderTestBindings, config)
	require.NoError(t, err)

	result := strings.TrimSpace(buf.String())
	require.Contains(t, result, "Item1 - Product")
	require.Contains(t, result, "Item2 - Product")
	require.Contains(t, result, "Item3 - Product")
}

// TestRenderTag_DynamicTemplateName tests variable template names
func TestRenderTag_DynamicTemplateName(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	bindings := map[string]any{
		"template_name": "render_basic.html",
	}

	root, err := config.Compile(`{% render template_name %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Hello from render!", strings.TrimSpace(buf.String()))
}

// TestRenderTag_FileNotFound tests error handling for missing files
func TestRenderTag_FileNotFound(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	root, err := config.Compile(`{% render "missing_file.html" %}`, loc)
	require.NoError(t, err)

	err = render.Render(root, io.Discard, renderTestBindings, config)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err.Cause()))
}

// TestRenderTag_InvalidTemplateName tests error handling for non-string template names
func TestRenderTag_InvalidTemplateName(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	root, err := config.Compile(`{% render 123 %}`, loc)
	require.NoError(t, err)

	err = render.Render(root, io.Discard, renderTestBindings, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "string template name")
}

// TestRenderTag_InvalidSyntax tests various syntax errors
func TestRenderTag_InvalidSyntax(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	tests := []struct {
		name     string
		template string
	}{
		{"missing template name", `{% render %}`},
		{"invalid parameter format", `{% render "test", invalid %}`},
		{"unclosed quote", `{% render "test %}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.Compile(tt.template, loc)
			// Some errors might be caught during compilation, others during rendering
			if err == nil {
				root, _ := config.Compile(tt.template, loc)
				err = render.Render(root, io.Discard, renderTestBindings, config)
			}
			require.Error(t, err, "expected error for: %s", tt.template)
		})
	}
}

// TestRenderTag_WithExpressionInParameters tests using expressions/filters in parameters
func TestRenderTag_WithExpressionInParameters(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)
	// Need to add standard filters for size and times to work
	config.AddFilter("size", func(v any) int {
		if arr, ok := v.([]any); ok {
			return len(arr)
		}
		return 0
	})
	config.AddFilter("times", func(a, b int) int {
		return a * b
	})

	bindings := map[string]any{
		"items": []any{"a", "b", "c"},
		"price": 10,
	}

	// Template that uses the evaluated parameters
	config.Cache["testdata/render_expr.html"] = []byte(`Count: {{ count }}, Total: {{ total }}`)

	// Use filters in parameter values
	root, err := config.Compile(`{% render "render_expr.html", count: items | size, total: price | times: 2 %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Count: 3, Total: 20", strings.TrimSpace(buf.String()))
}

// TestRenderTag_ForloopVariables tests all forloop object properties
func TestRenderTag_ForloopVariables(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Add a newline at the end of each iteration
	config.Cache["testdata/render_forloop.html"] = []byte(
		"{{ forloop.index }},{{ forloop.index0 }},{{ forloop.rindex }},{{ forloop.rindex0 }},{{ forloop.first }},{{ forloop.last }},{{ forloop.length }}\n",
	)

	bindings := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	root, err := config.Compile(`{% render "render_forloop.html" for items as item %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)

	// First iteration: index=1, index0=0, rindex=3, rindex0=2, first=true, last=false, length=3
	require.Contains(t, lines[0], "1,0,3,2,true,false,3")
	// Second iteration: index=2, index0=1, rindex=2, rindex0=1, first=false, last=false, length=3
	require.Contains(t, lines[1], "2,1,2,1,false,false,3")
	// Third iteration: index=3, index0=2, rindex=1, rindex0=0, first=false, last=true, length=3
	require.Contains(t, lines[2], "3,2,1,0,false,true,3")
}

// TestRenderTag_CachedTemplate tests rendering from cached templates
func TestRenderTag_CachedTemplate(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/render_source.html", LineNo: 1}

	AddStandardTags(&config)

	// Add a template to the cache that doesn't exist as a file
	config.Cache["testdata/cached_template.html"] = []byte("Cached: {{ value }}")

	root, err := config.Compile(`{% render "cached_template.html", value: "test" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, nil, config)
	require.NoError(t, err)
	require.Equal(t, "Cached: test", strings.TrimSpace(buf.String()))
}

// TestParseRenderArgs tests the argument parsing function
func TestParseRenderArgs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
		checkFunc   func(*testing.T, *renderArgs)
	}{
		{
			name:  "basic template name",
			input: `"template.html"`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.NotNil(t, args.templateName)
				require.Empty(t, args.params)
				require.Nil(t, args.withValue)
				require.Nil(t, args.forValue)
			},
		},
		{
			name:  "template with single parameter",
			input: `"template.html", key: value`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.NotNil(t, args.templateName)
				require.Len(t, args.params, 1)
				require.Contains(t, args.params, "key")
			},
		},
		{
			name:  "template with multiple parameters",
			input: `"template.html", key1: value1, key2: value2`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.Len(t, args.params, 2)
				require.Contains(t, args.params, "key1")
				require.Contains(t, args.params, "key2")
			},
		},
		{
			name:  "with syntax",
			input: `"template.html" with object as item`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.NotNil(t, args.withValue)
				require.Equal(t, "item", args.withAlias)
			},
		},
		{
			name:  "for syntax",
			input: `"template.html" for items as item`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.NotNil(t, args.forValue)
				require.Equal(t, "item", args.forAlias)
			},
		},
		{
			name:  "for syntax with parameters",
			input: `"template.html" for items as item, key: value`,
			checkFunc: func(t *testing.T, args *renderArgs) {
				require.NotNil(t, args.forValue)
				require.Equal(t, "item", args.forAlias)
				require.Len(t, args.params, 1)
			},
		},
		{
			name:        "empty input",
			input:       ``,
			shouldError: true,
		},
		{
			name:        "invalid parameter format",
			input:       `"template.html", invalid`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := parseRenderArgs(tt.input)
			if tt.shouldError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, args)
				}
			}
		})
	}
}
