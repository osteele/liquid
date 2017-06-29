//go:generate stringer -type=ChunkType

package chunks

import (
	"regexp"
)

var chunkMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}|{%\s*(\w+)(?:\s+((?:[^%]|%[^}])+?))?\s*%}`)

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
			out = append(out, Chunk{Type: TextChunkType, SourceInfo: sourceInfo, Source: data[p:ts]})
		}
		switch data[ts+1] {
		case '{':
			out = append(out, Chunk{
				Type:       ObjChunkType,
				SourceInfo: sourceInfo,
				Source:     data[ts:te],
				Parameters: data[m[2]:m[3]],
			})
		case '%':
			c := Chunk{
				Type:       TagChunkType,
				SourceInfo: sourceInfo,
				Source:     data[ts:te],
				Name:       data[m[4]:m[5]],
			}
			if m[6] > 0 {
				c.Parameters = data[m[6]:m[7]]
			}
			out = append(out, c)
		}
		p = te
	}
	if p < pe {
		out = append(out, Chunk{Type: TextChunkType, SourceInfo: sourceInfo, Source: data[p:]})
	}
	return out
}
