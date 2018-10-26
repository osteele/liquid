package liquid

import (
	"fmt"
	"sync"
	"testing"

	"github.com/urbn8/liquid/render"
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
	tpl, err := engine.ParseTemplateLocation([]byte(`{% sourcepath %}`), "source.md", 1)
	require.NoError(t, err)
	out, err := tpl.RenderString(testBindings)
	require.NoError(t, err)
	require.Equal(t, "source.md", out)

	src := []byte(`{{ n | undefined_filter }}`)
	t1, err := engine.ParseTemplateLocation(src, "path1", 1)
	require.NoError(t, err)
	t2, err := engine.ParseTemplateLocation(src, "path2", 1)
	require.NoError(t, err)
	_, err = t1.Render(Bindings{})
	require.Error(t, err)
	require.Equal(t, "path1", err.Path())
	_, err = t2.Render(Bindings{})
	require.Error(t, err)
	require.Equal(t, "path2", err.Path())
}

func TestTemplate_Parse_race(t *testing.T) {
	var (
		engine = NewEngine()
		count  = 10
		wg     sync.WaitGroup
	)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			path := fmt.Sprintf("path %d", i)
			_, err := engine.ParseTemplateLocation([]byte("{{ syntax error }}"), path, i)
			require.Error(t, err)
			require.Equal(t, path, err.Path())
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestTemplate_Render_race(t *testing.T) {
	src := []byte(`{{ n | undefined_filter }}`)
	engine := NewEngine()

	var (
		count = 10
		paths = make([]string, count)
		ts    = make([]*Template, count)
		wg    sync.WaitGroup
	)
	for i := 0; i < count; i++ {
		paths[i] = fmt.Sprintf("path %d", i)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var err error
			ts[i], err = engine.ParseTemplateLocation(src, paths[i], i)
			require.NoError(t, err)
		}(i)
	}
	wg.Wait()

	var wg2 sync.WaitGroup
	for i := 0; i < count; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()
			_, err := ts[i].Render(Bindings{})
			require.Error(t, err)
			require.Equal(t, paths[i], err.Path())
		}(i)
	}
	wg2.Wait()
}

func BenchmarkTemplate_Render(b *testing.B) {
	engine := NewEngine()
	bindings := Bindings{"a": "string value"}
	tpl, err := engine.ParseString(`{% for i in (1..1000) %}{% if i > 500 %}{{a}}{% else %}0{% endif %}{% endfor %}`)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tpl.Render(bindings)
	}
}
