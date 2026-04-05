package tags

import (
	"bytes"
	"io"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

// Register keys used to pass block override data between layout and block tags.
// The null-byte prefix prevents collision with user-defined variables.
const (
	// blockCapturesKey holds the in-progress map[string]string while the layout
	// tag is collecting block content from its children.
	blockCapturesKey = "\x00block_captures"

	// blockOverridesKey holds the completed map[string]string when the layout
	// file is being rendered so that {% block %} tags can inject overridden content.
	blockOverridesKey = "\x00block_overrides"
)

// makeLayoutTag creates the layout block tag compiler.
//
// Usage (child template):
//
//	{% layout 'base.html' %}
//	  {% block title %}My Page{% endblock %}
//	  {% block content %}Hello World{% endblock %}
//	{% endlayout %}
//
// The layout file (base.html) uses the same {% block %} … {% endblock %}
// syntax to define slot positions with optional default content:
//
//	<html><body>
//	  <title>{% block title %}Default{% endblock %}</title>
//	  {% block content %}Default{% endblock %}
//	</body></html>
func makeLayoutTag(cfg *render.Config) func(render.BlockNode) (func(io.Writer, render.Context) error, error) {
	return func(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
		fileExprStr := strings.TrimSpace(node.Args)

		fileExpr, err := expressions.Parse(fileExprStr)
		if err != nil {
			return nil, err
		}

		return func(w io.Writer, ctx render.Context) error {
			// Evaluate the layout filename.
			fileVal, err := ctx.Evaluate(fileExpr)
			if err != nil {
				return err
			}

			rel, ok := fileVal.(string)
			if !ok {
				return ctx.Errorf("layout requires a string argument; got %v", fileVal)
			}

			filename := filepath.Join(filepath.Dir(ctx.SourceFile()), rel)

			// Phase 1: render the layout tag's body (the block definitions) into a
			// scratch buffer.  Any {% block %} tags that see blockCapturesKey in the
			// context will capture their content into the shared map instead of
			// writing to the writer.
			captures := map[string]string{}
			ctx.Set(blockCapturesKey, captures)

			var captureBuf bytes.Buffer
			if err := ctx.RenderChildren(&captureBuf); err != nil {
				return err
			}

			// Signal that capture mode is finished.
			ctx.Set(blockCapturesKey, nil)

			// Phase 2: render the layout file with the captured block overrides
			// injected as a special binding.
			s, err := ctx.RenderFile(filename, map[string]any{blockOverridesKey: captures})
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, s)
			return err
		}, nil
	}
}

// blockTagCompiler implements {% block name %}...{% endblock %}.
//
// Behaviour depends on context:
//   - Inside {% layout %} body (capture mode): renders inner content and stores it
//     in the captures map keyed by block name.
//   - Inside the layout file (render mode): checks for an override from the child
//     template; uses it when present, otherwise renders the default inner content.
//   - Outside both (standalone use): renders the inner content as-is.
func blockTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	name := strings.TrimSpace(node.Args)

	return func(w io.Writer, ctx render.Context) error {
		// Capture mode: content goes into the layout's captures map.
		captures, _ := ctx.Get(blockCapturesKey).(map[string]string)
		if captures != nil {
			var buf bytes.Buffer
			if err := ctx.RenderChildren(&buf); err != nil {
				return err
			}

			captures[name] = buf.String()
			return nil
		}

		// Render mode (inside layout file): inject the child's override when available.
		overrides, _ := ctx.Get(blockOverridesKey).(map[string]string)
		if overrides != nil {
			if content, ok := overrides[name]; ok {
				_, err := io.WriteString(w, content)
				return err
			}
		}

		// Default: render inner content directly.
		return ctx.RenderChildren(w)
	}, nil
}
