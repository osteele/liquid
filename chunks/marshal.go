package chunks

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// MustYAML is like yaml.Marshal, but panics if the value cannot be marshalled.
func MustYAML(value interface{}) string {
	b, err := yaml.Marshal(value)
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
		return map[string]interface{}{"tag": c.Name, "args": c.Args}, nil
	case ObjChunkType:
		return map[string]interface{}{"obj": c.Args}, nil
	default:
		return nil, fmt.Errorf("unknown chunk tag type: %v", c.Type)
	}
}

// MarshalYAML marshalls a chunk for debugging.
func (n ASTBlockNode) MarshalYAML() (interface{}, error) {
	return map[string]map[string]interface{}{
		n.cd.name: {
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
