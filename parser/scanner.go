package parser

import (
	"regexp"
	"strings"
)

var tokenMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}|{%\s*(\w+)(?:\s+((?:[^%]|%[^}])+?))?\s*%}`)

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, pathname string) []Token {
	// TODO error on unterminated {{ and {%
	// TODO probably an error when a tag contains a {{ or {%, at least outside of a string
	var (
		p, pe = 0, len(data)
		si    = SourceInfo{pathname, 1}
		out   = make([]Token, 0)
	)
	for _, m := range tokenMatcher.FindAllStringSubmatchIndex(data, -1) {
		ts, te := m[0], m[1]
		if p < ts {
			out = append(out, Token{Type: TextTokenType, SourceInfo: si, Source: data[p:ts]})
			si.lineNo += strings.Count(data[p:ts], "\n")
		}
		source := data[ts:te]
		switch data[ts+1] {
		case '{':
			out = append(out, Token{
				Type:       ObjTokenType,
				SourceInfo: si,
				Source:     source,
				Args:       data[m[2]:m[3]],
			})
		case '%':
			c := Token{
				Type:       TagTokenType,
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
		out = append(out, Token{Type: TextTokenType, SourceInfo: si, Source: data[p:]})
	}
	return out
}
