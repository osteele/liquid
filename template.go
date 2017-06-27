package liquid

import (
	"bytes"

	"github.com/osteele/liquid/chunks"
)

// Template renders a template according to scope.
//
// Scope is a map of liquid variable names to objects.
type Template interface {
	Render(scope map[string]interface{}) ([]byte, error)
	RenderString(scope map[string]interface{}) (string, error)
}

// Render applies the template to the scope.
func (t *template) Render(scope map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.ast.Render(buf, chunks.NewContext(scope))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *template) RenderString(scope map[string]interface{}) (string, error) {
	b, err := t.Render(scope)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
