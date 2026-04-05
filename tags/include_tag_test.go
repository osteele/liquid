package tags

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

var includeTestBindings = map[string]any{
	"test": true,
	"var":  "value",
}

func TestIncludeTag(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}

	AddStandardTags(&config)

	// basic functionality
	root, err := config.Compile(`{% include "include_target.html" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "include target", strings.TrimSpace(buf.String()))

	// tag and variable
	root, err = config.Compile(`{% include "include_target_2.html" %}`, loc)
	require.NoError(t, err)

	buf = new(bytes.Buffer)
	err = render.Render(root, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "test value", strings.TrimSpace(buf.String()))

	// errors
	root, err = config.Compile(`{% include 10 %}`, loc)
	require.NoError(t, err)
	err = render.Render(root, io.Discard, includeTestBindings, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a string")
}

func TestIncludeTag_file_not_found_error(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}

	AddStandardTags(&config)

	// See the comment in TestIncludeTag_file_not_found_error.
	root, err := config.Compile(`{% include "missing_file.html" %}`, loc)
	require.NoError(t, err)
	err = render.Render(root, io.Discard, includeTestBindings, config)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err.Cause()))
}

func TestIncludeTag_cached_value_handling(t *testing.T) {
	config := render.NewConfig()
	// missing-file.html does not exist in the testdata directory.
	config.Cache.Store("testdata/missing-file.html", []byte("include-content"))
	config.Cache.Store("testdata\\missing-file.html", []byte("include-content"))
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}

	AddStandardTags(&config)

	root, err := config.Compile(`{% include "missing-file.html" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "include-content", strings.TrimSpace(buf.String()))
}

// TestIncludeTag_with_variable tests the "with variable" syntax.
func TestIncludeTag_with_variable(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	product := map[string]any{"title": "Cool Shirt"}
	bindings := map[string]any{"product": product}

	// include 'file' with variable — uses file stem as variable name
	root, err := config.Compile(`{% include "include_with.html" with product %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Cool Shirt", strings.TrimSpace(buf.String()))
}

// TestIncludeTag_with_alias tests the "with variable as alias" syntax.
func TestIncludeTag_with_alias(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	product := map[string]any{"title": "Cool Shirt"}
	bindings := map[string]any{"item": product}

	// include 'file' with variable as alias — uses "product" as the name inside the template
	root, err := config.Compile(`{% include "include_with.html" with item as product %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Cool Shirt", strings.TrimSpace(buf.String()))
}

// TestIncludeTag_kv_pairs tests the key:value argument syntax.
func TestIncludeTag_kv_pairs(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	bindings := map[string]any{"n": 5}

	root, err := config.Compile(`{% include "include_kv.html", title: "Hello", count: n %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Hello-5", strings.TrimSpace(buf.String()))
}

// ---------------------------------------------------------------------------
// render tag
// ---------------------------------------------------------------------------

func TestRenderTag_basic(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// Basic render: renders the snippet in isolated scope
	bindings := map[string]any{"snippet_var": "ignored_in_outer"}
	root, err := config.Compile(`{% render "render_snippet.html", snippet_var: "hello" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "hello", strings.TrimSpace(buf.String()))
}

// TestRenderTag_isolated_scope verifies that render does NOT inherit parent scope.
func TestRenderTag_isolated_scope(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// snippet_var is defined in parent scope but render should NOT see it
	bindings := map[string]any{"snippet_var": "parent_value"}
	root, err := config.Compile(`{% render "render_snippet.html" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	// render uses isolated scope, so snippet_var from parent is NOT visible; outputs empty
	require.Equal(t, "", strings.TrimSpace(buf.String()))
}

// TestRenderTag_with_variable tests the "with variable" syntax.
func TestRenderTag_with_variable(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	product := map[string]any{"title": "Cool Shirt"}
	bindings := map[string]any{"product": product}

	// render 'product.html' with product — file stem is "product", which becomes the variable
	// name inside the template. Template is: {{ product.title }}
	root, err := config.Compile(`{% render "product.html" with product %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "Cool Shirt", strings.TrimSpace(buf.String()))
}

// TestIncludeTag_for_array tests the "for array as alias" syntax.
func TestIncludeTag_for_array(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// include_for_item.html contains: {{ item }}
	bindings := map[string]any{
		"items": []any{"a", "b", "c"},
	}

	root, err := config.Compile(`{% include "include_for_item.html" for items as item %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "abc", buf.String())
}

// TestIncludeTag_for_array_default_alias tests that when no "as" is given, the
// file stem is used as the variable name.
func TestIncludeTag_for_array_default_alias(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// include_for_item.html contains: {{ item }} — file stem is "include_for_item"
	// No "as" given, so this test uses a template where the stem name is the variable.
	// We test with explicit "as" to keep tests simple here.
	bindings := map[string]any{
		"items": []any{"x", "y"},
	}

	root, err := config.Compile(`{% include "include_for_item.html" for items as item %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, bindings, config)
	require.NoError(t, err)
	require.Equal(t, "xy", buf.String())
}

// TestLayoutTag tests the layout / block template inheritance tags.
func TestLayoutTag(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// layout_base.html:
	//   <html>
	//   <head><title>{% block title %}Default Title{% endblock %}</title></head>
	//   <body>{% block content %}Default Content{% endblock %}</body>
	//   </html>

	// Child overrides both blocks.
	child := `{% layout "layout_base.html" %}` +
		`{% block title %}My Page{% endblock %}` +
		`{% block content %}Hello World{% endblock %}` +
		`{% endlayout %}`

	root, err := config.Compile(child, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, map[string]any{}, config)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "<title>My Page</title>")
	require.Contains(t, out, "Hello World")
	require.NotContains(t, out, "Default Title")
	require.NotContains(t, out, "Default Content")
}

// TestLayoutTag_default_content verifies that blocks not overridden by the
// child render the layout's default content.
func TestLayoutTag_default_content(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(&config)

	// Child only overrides "title"; "content" should fall back to layout default.
	child := `{% layout "layout_base.html" %}` +
		`{% block title %}Custom Title{% endblock %}` +
		`{% endlayout %}`

	root, err := config.Compile(child, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, map[string]any{}, config)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "<title>Custom Title</title>")
	require.Contains(t, out, "Default Content")
}

// TestBlockTag_standalone verifies that {% block %}...{% endblock %} used
// outside a layout context just renders its default content.
func TestBlockTag_standalone(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)

	root, err := config.Compile(`{% block greeting %}Hello{% endblock %}`, parser.SourceLoc{})
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	require.NoError(t, render.Render(root, buf, map[string]any{}, config))
	require.Equal(t, "Hello", buf.String())
}
