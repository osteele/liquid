package liquid

import (
	"bytes"

	"github.com/osteele/liquid/chunks"
)

// Template renders a template according to scope.
//
// Bindings is a map of liquid variable names to objects.
type Template interface {
	// Render executes the template with the specified bindings.
	Render(bindings map[string]interface{}) ([]byte, error)
	// RenderString is a convenience wrapper for Render, that has string input and output.
	RenderString(bindings map[string]interface{}) (string, error)
}

// Render executes the template within the bindings environment.
func (t *template) Render(bindings map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.ast.Render(buf, chunks.NewContext(bindings))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *template) RenderString(bindings map[string]interface{}) (string, error) {
	b, err := t.Render(bindings)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
