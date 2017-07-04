package liquid

import (
	"bytes"

	"github.com/osteele/liquid/render"
)

type template struct {
	ast    render.ASTNode
	config render.Config
}

// Render executes the template within the bindings environment.
func (t *template) Render(b Bindings) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := render.Render(t.ast, buf, b, t.config)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *template) RenderString(b Bindings) (string, error) {
	bs, err := t.Render(b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func (t *template) SetSourcePath(filename string) {
	t.config.Filename = filename
}
