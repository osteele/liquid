package parser

import (
	"regexp"
	"strings"
)

var tokenMatcher = regexp.MustCompile(`{{-?\s*(.+?)\s*-?}}|{%-?\s*(\w+)(?:\s+((?:[^%]|%[^}])+?))?\s*-?%}`)

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, loc SourceLoc) (tokens []Token) {
	// TODO error on unterminated {{ and {%
	// TODO probably an error when a tag contains a {{ or {%, at least outside of a string
	p, pe := 0, len(data)
	for _, m := range tokenMatcher.FindAllStringSubmatchIndex(data, -1) {
		ts, te := m[0], m[1]
		if p < ts {
			tokens = append(tokens, Token{Type: TextTokenType, SourceLoc: loc, Source: data[p:ts]})
			loc.LineNo += strings.Count(data[p:ts], "\n")
		}
		source := data[ts:te]
		switch data[ts+1] {
		case '{':
			tok := Token{
				Type:      ObjTokenType,
				SourceLoc: loc,
				Source:    source,
				Args:      data[m[2]:m[3]],
				TrimLeft:  source[2] == '-',
				TrimRight: source[len(source)-3] == '-',
			}
			tokens = append(tokens, tok)
		case '%':
			tok := Token{
				Type:      TagTokenType,
				SourceLoc: loc,
				Source:    source,
				Name:      data[m[4]:m[5]],
				TrimLeft:  source[2] == '-',
				TrimRight: source[len(source)-3] == '-',
			}
			if m[6] > 0 {
				tok.Args = data[m[6]:m[7]]
			}
			tokens = append(tokens, tok)
		}
		loc.LineNo += strings.Count(source, "\n")
		p = te
	}
	if p < pe {
		tokens = append(tokens, Token{Type: TextTokenType, SourceLoc: loc, Source: data[p:]})
	}
	return tokens
}
