package chunks

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// MustYAML returns the YAML of an interface.
func MustYAML(val interface{}) string {
	b, err := yaml.Marshal(val)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// MarshalYAML is for debugging.
func (c Chunk) MarshalYAML() (interface{}, error) {
	switch c.Type {
	case TextChunkType:
		return map[string]interface{}{"text": c.Source}, nil
	case TagChunkType:
		return map[string]interface{}{"tag": c.Tag, "args": c.Args}, nil
	case ObjChunkType:
		return map[string]interface{}{"obj": c.Tag}, nil
	default:
		return nil, fmt.Errorf("unknown chunk tag type: %v", c.Type)
	}
}

// MarshalYAML marshalls a chunk for debugging.
func (n ASTChunks) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{"chunks": n.chunks}, nil
}

// MarshalYAML marshalls a chunk for debugging.
func (n ASTControlTag) MarshalYAML() (interface{}, error) {
	return map[string]map[string]interface{}{
		n.cd.Name: {
			"args":     n.Args,
			"body":     n.Body,
			"branches": n.Branches,
		}}, nil
}

// MarshalYAML marshalls a chunk for debugging.
func (n ASTText) MarshalYAML() (interface{}, error) {
	return n.Chunk.MarshalYAML()
}

// MarshalYAML marshalls a chunk for debugging.
func (n ASTObject) MarshalYAML() (interface{}, error) {
	return n.Chunk.MarshalYAML()
}
