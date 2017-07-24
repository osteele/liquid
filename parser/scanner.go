package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, loc SourceLoc, delims []byte) (tokens []Token) {
	// delims = {, }, % => delimiters = {{, }}, {%, %}
	if len(delims) != 3 {
		delims = []byte{'{', '}', '%'}
	}
	delimiters := formFullDelimiters(delims)
	tokenMatcher := regexp.MustCompile(
		fmt.Sprintf(`%s-?\s*(.+?)\s*-?%s|%s-?\s*(\w+)(?:\s+((?:[^%%]|%%[^}])+?))?\s*-?%s`,
			// QuoteMeta will escape any of these that are regex commands
			regexp.QuoteMeta(delimiters[0]), regexp.QuoteMeta(delimiters[1]),
			regexp.QuoteMeta(delimiters[2]), regexp.QuoteMeta(delimiters[3]),
		),
	)

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

// formFullDelimiters converts the single character byte delimiters into the full string actual
// delimiters.
func formFullDelimiters(delims []byte) []string {
	// Configure the token matcher to respect the delimiters passed to it. The default delims are '{',
	// '}', '%' which turn into "{{" and "}}" for objects and "{%" and "%}" for tags
	fullDelimiters := make([]string, 4, 4)
	fullDelimiters[0] = string([]byte{delims[0], delims[0]})
	fullDelimiters[1] = string([]byte{delims[1], delims[1]})
	fullDelimiters[2] = string([]byte{delims[0], delims[2]})
	fullDelimiters[3] = string([]byte{delims[2], delims[1]})
	return fullDelimiters
}
