package chunks

import (
	"io"
)

// TODO instead of using the bare function definition, use a structure that defines how to parse
type TagDefinition func(expr string) (func(io.Writer, Context) error, error)

// TODO parse during definition stage, not rendering stage
func assignTagDef(source string) (func(io.Writer, Context) error, error) {
	return func(w io.Writer, ctx Context) error {
		_, err := ctx.evaluateStatement("assign", source)
		if err != nil {
			return err
		}
		return nil
	}, nil
}

func FindTagDefinition(name string) (TagDefinition, bool) {
	switch name {
	case "assign":
		return assignTagDef, true
	}
	return nil, false
}
