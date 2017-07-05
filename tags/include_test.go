package tags

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

var includeTestBindings = map[string]interface{}{}

func TestIncludeTag(t *testing.T) {
	config := render.NewConfig()
	config.Filename = "testdata/include_source.html"
	AddStandardTags(config)

	ast, err := config.Parse(`{% include "include_target.html" %}`)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = render.Render(ast, buf, includeTestBindings, config)
	require.NoError(t, err)
	require.Equal(t, "include target", strings.TrimSpace(buf.String()))

	ast, err = config.Parse(`{% include 10 %}`)
	require.NoError(t, err)
	err = render.Render(ast, ioutil.Discard, includeTestBindings, config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a string")
}
