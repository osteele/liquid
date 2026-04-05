package liquid

import (
	"bytes"
	"io"

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
	startLine := line
	if startLine == 0 {
		startLine = 1
	}
	loc := parser.SourceLoc{Pathname: path, LineNo: startLine}

	root, err := cfg.Compile(string(source), loc)
	if err != nil {
		return nil, err
	}

	return &Template{root, cfg}, nil
}

// GetRoot returns the root node of the abstract syntax tree (AST) representing
// the parsed template.
func (t *Template) GetRoot() render.Node {
	return t.root
}

// Render executes the template with the specified variable bindings.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. For example, adding WithStrictVariables() enables strict variable
// checking even if StrictVariables was not called on the engine.
func (t *Template) Render(vars Bindings, opts ...RenderOption) ([]byte, SourceError) {
	buf := new(bytes.Buffer)

	cfg := t.applyRenderOptions(opts)
	err := render.Render(t.root, buf, vars, cfg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// FRender executes the template with the specified variable bindings and renders it into w.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. See Render for details.
func (t *Template) FRender(w io.Writer, vars Bindings, opts ...RenderOption) SourceError {
	cfg := t.applyRenderOptions(opts)
	err := render.Render(t.root, w, vars, cfg)
	if err != nil {
		return err
	}

	return nil
}

// RenderString is a convenience wrapper for Render, that has string input and output.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. See Render for details.
func (t *Template) RenderString(b Bindings, opts ...RenderOption) (string, SourceError) {
	bs, err := t.Render(b, opts...)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

// applyRenderOptions returns a copy of the template config with render options applied.
func (t *Template) applyRenderOptions(opts []RenderOption) render.Config {
	if len(opts) == 0 {
		return *t.cfg
	}
	cfg := *t.cfg
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
