//go:generate stringer -type=ChunkType

package chunks

import (
	"regexp"
	"strings"
)

var chunkMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}|{%\s*(\w+)(?:\s+((?:[^%]|%[^}])+?))?\s*%}`)

// Scan breaks a string into a sequence of Chunks.
func Scan(data string, pathname string) []Chunk {
	// TODO error on unterminated {{ and {%
	// TODO probably an error when a tag contains a {{ or {%, at least outside of a string
	var (
		p, pe = 0, len(data)
		si    = SourceInfo{pathname, 1}
		out   = make([]Chunk, 0)
	)
	for _, m := range chunkMatcher.FindAllStringSubmatchIndex(data, -1) {
		ts, te := m[0], m[1]
		if p < ts {
			out = append(out, Chunk{Type: TextChunkType, SourceInfo: si, Source: data[p:ts]})
			si.lineNo += strings.Count(data[p:ts], "\n")
		}
		source := data[ts:te]
		switch data[ts+1] {
		case '{':
			out = append(out, Chunk{
				Type:       ObjChunkType,
				SourceInfo: si,
				Source:     source,
				Args:       data[m[2]:m[3]],
			})
		case '%':
			c := Chunk{
				Type:       TagChunkType,
				SourceInfo: si,
				Source:     source,
				Name:       data[m[4]:m[5]],
			}
			if m[6] > 0 {
				c.Args = data[m[6]:m[7]]
			}
			out = append(out, c)
		}
		si.lineNo += strings.Count(source, "\n")
		p = te
	}
	if p < pe {
		out = append(out, Chunk{Type: TextChunkType, SourceInfo: si, Source: data[p:]})
	}
	return out
}
