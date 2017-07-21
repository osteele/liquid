package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, loc SourceLoc, delims []byte) (tokens []Token) {
	// Configure the token matcher to respect the delimeters passed to it
	if len(delims) != 3 {
		delims = []byte{'{', '}', '%'}
	}
	objectLeft := string(delims[0]) + string(delims[0])
	objectRight := string(delims[1]) + string(delims[1])
	tagLeft := string(delims[0]) + string(delims[2])
	tagRight := string(delims[2]) + string(delims[1])
	var tokenMatcher = regexp.MustCompile(fmt.Sprintf(`%v-?\s*(.+?)\s*-?%v|%v-?\s*(\w+)(?:\s+((?:[^%%]|%%[^}])+?))?\s*-?%v`, objectLeft, objectRight, tagLeft, tagRight))

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
		case delims[0]:
			tok := Token{
				Type:      ObjTokenType,
				SourceLoc: loc,
				Source:    source,
				Args:      data[m[2]:m[3]],
				TrimLeft:  source[2] == '-',
				TrimRight: source[len(source)-3] == '-',
			}
			tokens = append(tokens, tok)
		case delims[2]:
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
