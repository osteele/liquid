package tags

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var includeTestBindings = map[string]interface{}{
	"test": true,
	"var":  "value",
}

func TestIncludeTag(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(config)

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
	err = render.Render(root, ioutil.Discard, includeTestBindings, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a string")
}

func TestIncludeTag_file_not_found_error(t *testing.T) {
	config := render.NewConfig()
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(config)

	// See the comment in TestIncludeTag_file_not_found_error.
	root, err := config.Compile(`{% include "missing_file.html" %}`, loc)
	require.NoError(t, err)
	err = render.Render(root, ioutil.Discard, includeTestBindings, config)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err.Cause()))
}

func TestIncludeTag_cached_value_handling(t *testing.T) {
	config := render.NewConfig()
	// foo.html does not exist on testdata.
	config.Cache["testdata/foo.html"] = []byte("bar")
	loc := parser.SourceLoc{Pathname: "testdata/include_source.html", LineNo: 1}
	AddStandardTags(config)

	root, err := config.Compile(`{% include "foo.html" %}`, loc)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = render.Render(root, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "bar", strings.TrimSpace(buf.String()))
}
