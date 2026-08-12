package tags

import (
	"io"

	"github.com/osteele/liquid/render"
)

func includeTag(source string) (func(io.Writer, render.Context) error, error) {
	return func(w io.Writer, ctx render.Context) error {
		// It might be more efficient to add a context interface to render bytes
		// to a writer. The status quo keeps the interface light at the expense of some overhead
		// here.
		value, err := ctx.EvaluateString(ctx.TagArgs())
		if err != nil {
			return err
		}

		rel, ok := value.(string)
		if !ok {
			return ctx.Errorf("include requires a string argument; got %v", value)
		}

		filename, err := resolveTemplatePath(ctx.SourceFile(), rel)
		if err != nil {
			return ctx.WrapError(err)
		}

		s, err := ctx.RenderFile(filename, map[string]any{})
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, s)

		return err
	}, nil
}
