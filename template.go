package liquid

import (
	"bytes"

	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// A Template is a compiled Liquid template. It knows how to evaluate itself within a variable binding environment, to create a rendered byte slice.
//
// Use Engine.ParseTemplate to create a template.
type Template struct {
	root render.Node
	cfg  *render.Config
}

func newTemplate(cfg *render.Config, source []byte, path string, line int) (*Template, SourceError) {
	loc := parser.SourceLoc{Pathname: path, LineNo: line}
	root, err := cfg.Compile(string(source), loc)
	if err != nil {
		return nil, err
	}
	return &Template{root, cfg}, nil
}

// Render executes the template with the specified variable bindings.
func (t *Template) Render(vars Bindings) ([]byte, SourceError) {
	buf := new(bytes.Buffer)
	err := render.Render(t.root, buf, vars, *t.cfg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
func (t *Template) RenderString(b Bindings) (string, SourceError) {
	bs, err := t.Render(b)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
