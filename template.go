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
func (t *template) Render(c Context) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.ast.Render(buf, chunks.NewContext(c.Bindings(), t.settings))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *template) RenderString(c Context) (string, error) {
	b, err := t.Render(c)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
