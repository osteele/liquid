package liquid

import (
	"bytes"

	"github.com/osteele/liquid/chunks"
)

type Engine interface {
	ParseAndRender(text []byte, scope map[string]interface{}) ([]byte, error)
	ParseAndRenderString(text string, scope map[string]interface{}) ([]byte, error)
}

type Template interface {
	Render(scope map[string]interface{}) ([]byte, error)
}

type engine struct{}

type template struct {
	ast chunks.AST
}

func NewEngine() Engine {
	return engine{}
}

func (e engine) Parse(text []byte) (Template, error) {
	tokens := chunks.Scan(string(text), "")
	ast, err := chunks.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return &template{ast}, nil
}

// ParseAndRender parses and then renders the template.
func (e engine) ParseAndRender(text []byte, scope map[string]interface{}) ([]byte, error) {
	t, err := e.Parse(text)
	if err != nil {
		return nil, err
	}
	return t.Render(scope)
}

func (e engine) ParseAndRenderString(text string, scope map[string]interface{}) ([]byte, error) {
	return e.ParseAndRender([]byte(text), scope)
}

func (t *template) Render(scope map[string]interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := t.ast.Render(buf, chunks.Context{scope})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
