package chunks

import (
	"io"

	yaml "gopkg.in/yaml.v2"
)

type AST interface {
	Render(io.Writer, Context) error
}

type ASTSeq struct {
	Children []AST
}

type ASTChunks struct {
	chunks []Chunk
}

type ASTText struct {
	chunk Chunk
}

type ASTObject struct {
	chunk Chunk
}

type ASTControlTag struct {
	chunk    Chunk
	cd       *ControlTagDefinition
	body     []AST
	branches []*ASTControlTag
}

func MustYAML(val interface{}) string {
	b, err := yaml.Marshal(val)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (n ASTChunks) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{"leaf": n.chunks}, nil
}

func (n ASTControlTag) MarshalYAML() (interface{}, error) {
	return map[string]map[string]interface{}{
		n.cd.Name: {
			"args":     n.chunk.Args,
			"body":     n.body,
			"branches": n.branches,
		}}, nil
}

func (n ASTText) MarshalYAML() (interface{}, error) {
	return n.chunk.MarshalYAML()
}

func (n ASTObject) MarshalYAML() (interface{}, error) {
	return n.chunk.MarshalYAML()
}
