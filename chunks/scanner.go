//go:generate stringer -type=ChunkType

package chunks

import (
	"fmt"
	"regexp"
)

type Chunk struct {
	Type              ChunkType
	SourceInfo        SourceInfo
	Source, Tag, Args string
}

type SourceInfo struct {
	Pathname string
	lineNo   int
}

type ChunkType int

const (
	TextChunk ChunkType = iota
	TagChunk
	ObjChunk
)

var chunkMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}|{%\s*(\w+)(?:\s+(.+?))?\s*%}`)

// MarshalYAML, for debugging
func (c Chunk) MarshalYAML() (interface{}, error) {
	switch c.Type {
	case TextChunk:
		return map[string]interface{}{"text": c.Source}, nil
	case TagChunk:
		return map[string]interface{}{"tag": c.Tag, "args": c.Args}, nil
	case ObjChunk:
		return map[string]interface{}{"obj": c.Tag}, nil
	default:
		return nil, fmt.Errorf("unknown chunk tag type: %v", c.Type)
	}
}

func ScanChunks(data string, pathname string) []Chunk {
	var (
		sourceInfo = SourceInfo{pathname, 0}
		out        = make([]Chunk, 0)
		p, pe      = 0, len(data)
		matches    = chunkMatcher.FindAllStringSubmatchIndex(data, -1)
	)
	for _, m := range matches {
		ts, te := m[0], m[1]
		if p < ts {
			out = append(out, Chunk{TextChunk, sourceInfo, data[p:ts], "", ""})
		}
		switch data[ts+1] {
		case '{':
			out = append(out, Chunk{ObjChunk, sourceInfo, data[ts:te], data[m[2]:m[3]], ""})
		case '%':
			var args string
			if m[6] > 0 {
				args = data[m[6]:m[7]]
			}
			out = append(out, Chunk{TagChunk, sourceInfo, data[ts:te], data[m[4]:m[5]], args})
		}
		p = te
	}
	if p < pe {
		out = append(out, Chunk{TextChunk, sourceInfo, data[p:], "", ""})
	}
	return out
}
