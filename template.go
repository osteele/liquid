package liquid

import (
	"bytes"

	"github.com/osteele/liquid/chunks"
)

type template struct {
	ast      chunks.ASTNode
	settings chunks.Settings
}

// Render executes the template within the bindings environment.
func (t *template) Render(b Bindings) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.ast.Render(buf, chunks.NewContext(b, t.settings))
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
