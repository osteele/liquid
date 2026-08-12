package tags

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
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

func TestIncludeTagRejectsEscapingPaths(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)
	loc := parser.SourceLoc{Pathname: filepath.Join("testdata", "source.liquid"), LineNo: 1}
	root, err := config.Compile(`{% include "../secret.liquid" %}`, loc)
	require.NoError(t, err)

	err = render.Render(root, io.Discard, nil, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "escapes its source directory")
}

func TestIncludeTagResolvesNestedIncludesRelativeToIncludedFile(t *testing.T) {
	config := render.NewConfig()
	AddStandardTags(&config)
	mainPath := filepath.Join("templates", "main.liquid")
	firstPath := filepath.Join("templates", "nested", "first.liquid")
	secondPath := filepath.Join("templates", "nested", "second.liquid")
	config.Cache[firstPath] = []byte(`first:{% include "second.liquid" %}`)
	config.Cache[secondPath] = []byte("second")

	root, err := config.Compile(`{% include "nested/first.liquid" %}`, parser.SourceLoc{Pathname: mainPath, LineNo: 1})
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = render.Render(root, buf, nil, config)
	require.NoError(t, err)
	require.Equal(t, "first:second", buf.String())
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
	config.Cache["testdata/missing-file.html"] = []byte("include-content")
	config.Cache["testdata\\missing-file.html"] = []byte("include-content")
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}

	AddStandardTags(&config)

	root, err := config.Compile(`{% include "missing-file.html" %}`, loc)
	require.NoError(t, err)

	buf := new(bytes.Buffer)
	err = render.Render(root, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "include-content", strings.TrimSpace(buf.String()))
}
