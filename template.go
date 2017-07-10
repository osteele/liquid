package liquid

import (
	"bytes"

	"github.com/osteele/liquid/render"
)

// A Template is a compiled Liquid template. It knows how to evaluate itself within a variable binding environment, to create a rendered byte slice.
type Template interface {
	// Render executes the template with the specified variable bindings.
	Render(Bindings) ([]byte, error)
	// RenderString is a convenience wrapper for Render, that has string input and output.
	RenderString(Bindings) (string, error)
	// SetSourcePath sets the filename. This is used for error reporting,
	// and as the reference directory for relative pathnames in the {% include %} tag.
	SetSourcePath(string)
	// SetSourceLocation sets the source path as SetSourcePath, and also
	// the line number of the first line of the template text, for use in
	// error reporting.
	SetSourceLocation(string, int)
}

type template struct {
	root   render.Node
	config *render.Config
}

// Render executes the template within the bindings environment.
func (t *template) Render(vars Bindings) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := render.Render(t.root, buf, vars, *t.config)
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

func (t *template) SetSourceLocation(filename string, lineNo int) {
	t.config.Filename = filename
	t.config.LineNo = lineNo
}
