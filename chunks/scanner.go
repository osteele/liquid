//go:generate stringer -type=ChunkType

package chunks

import (
	"regexp"
)

// Chunk is a chunk of a template source. It is one of an object "{{…}}", a tag "{%…%}", or the text between objects and tags.
type Chunk struct {
	Type       ChunkType
	SourceInfo SourceInfo
	Source     string // Source is the entirety of the chunk, including the "{{", "{%", etc. markers.
	Tag        string // Tag is the tag name of a tag Chunk. E.g. the tag name of "{% if 1 %}" is "if".
	Args       string // Args is the tag arguments of a tag Chunk. E.g. the tag arguments of "{% if 1 %}" is "1".
}

// SourceInfo contains a Chunk's source information
type SourceInfo struct {
	Pathname string
	lineNo   int
}

// ChunkType is the type of a Chunk
type ChunkType int

const (
	TextChunkType ChunkType = iota // TextChunkType is the type of a text Chunk
	TagChunkType                   // TagChunkType is the type of a tag Chunk "{%…%}"
	ObjChunkType                   // TextChunkType is the type of an object Chunk "{{…}}"
)

var chunkMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}|{%\s*(\w+)(?:\s+(.+?))?\s*%}`)

// Scan breaks a string into a sequence of Chunks.
func Scan(data string, pathname string) []Chunk {
	// TODO error on unterminated {{ and {%
	// TODO probably an error when a tag contains a {{ or {%, at least outside of a string
	var (
		sourceInfo = SourceInfo{pathname, 0}
		out        = make([]Chunk, 0)
		p, pe      = 0, len(data)
		matches    = chunkMatcher.FindAllStringSubmatchIndex(data, -1)
	)
	for _, m := range matches {
		ts, te := m[0], m[1]
		if p < ts {
			out = append(out, Chunk{TextChunkType, sourceInfo, data[p:ts], "", ""})
		}
		switch data[ts+1] {
		case '{':
			out = append(out, Chunk{ObjChunkType, sourceInfo, data[ts:te], "", data[m[2]:m[3]]})
		case '%':
			var args string
			if m[6] > 0 {
				args = data[m[6]:m[7]]
			}
			out = append(out, Chunk{TagChunkType, sourceInfo, data[ts:te], data[m[4]:m[5]], args})
		}
		p = te
	}
	if p < pe {
		out = append(out, Chunk{TextChunkType, sourceInfo, data[p:], "", ""})
	}
	return out
}
