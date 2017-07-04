package liquid

import (
	"testing"

	"github.com/osteele/liquid/render"
	"github.com/stretchr/testify/require"
)

func TestTemplate_RenderString(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseTemplate([]byte(`{{ "hello world" | capitalize }}`))
	require.NoError(t, err)
	out, err := tpl.RenderString(testBindings)
	require.NoError(t, err)
	require.Equal(t, "Hello world", out)
}

func TestTemplate_SetSourcePath(t *testing.T) {
	engine := NewEngine()
	engine.RegisterTag("sourcepath", func(c render.Context) (string, error) {
		return c.SourceFile(), nil
	})
	tpl, err := engine.ParseTemplate([]byte(`{% sourcepath %}`))
	require.NoError(t, err)
	tpl.SetSourcePath("source.md")
	out, err := tpl.RenderString(testBindings)
	require.NoError(t, err)
	require.Equal(t, "source.md", out)
}
