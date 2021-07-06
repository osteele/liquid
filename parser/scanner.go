package parser

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, loc SourceLoc, delims []string) (tokens []Token) {

	// Apply defaults
	if len(delims) != 4 {
		delims = []string{"{{", "}}", "{%", "%}"}
	}
	tokenMatcher := formTokenMatcher(delims)

	// TODO error on unterminated {{ and {%
	// TODO probably an error when a tag contains a {{ or {%, at least outside of a string
	p, pe := 0, len(data)
	for _, m := range tokenMatcher.FindAllStringSubmatchIndex(data, -1) {
		ts, te := m[0], m[1]
		source := data[ts:te]
		if p < ts {
			tokens = append(tokens, Token{Type: TextTokenType, SourceLoc: loc, Source: data[p:ts]})
		}
		switch {
		case rune(data[ts]) == '\n':
			tok := Token{Type: WhitespaceTokenType, Name: "New Line", SourceLoc: loc, Source: source}
			loc.LineNo++
			tokens = append(tokens, tok)
		case unicode.IsSpace(rune(data[ts])):
			tok := Token{Type: WhitespaceTokenType, Name: "Whitespace", SourceLoc: loc, Source: source}
			tokens = append(tokens, tok)
		case data[ts:ts+len(delims[0])] == delims[0]:
			tok := Token{
				Type:      ObjTokenType,
				SourceLoc: loc,
				Source:    source,
				Args:      data[m[2]:m[3]],
				TrimLeft:  source[2] == '-',
				TrimRight: source[len(source)-3] == '-',
			}
			tokens = append(tokens, tok)
		case data[ts:ts+len(delims[2])] == delims[2]:
			tok := Token{
				Type:      TagTokenType,
				SourceLoc: loc,
				Source:    source,
				Name:      data[m[8]:m[9]],
				TrimLeft:  source[2] == '-',
				TrimRight: source[len(source)-3] == '-',
			}
			if m[10] > 0 {
				tok.Args = data[m[10]:m[11]]
			}
			tokens = append(tokens, tok)
		}
		p = te
	}
	if p < pe {
		tokens = append(tokens, Token{Type: TextTokenType, SourceLoc: loc, Source: data[p:]})
	}
	return tokens
}

func formTokenMatcher(delims []string) *regexp.Regexp {
	// On ending a tag we need to exclude anything that appears to be ending a tag that's nested
	// inside the tag. We form the exclusion expression here.
	// For example, if delims is default the exclusion expression is "[^%]|%[^}]".
	// If tagRight is "TAG!RIGHT" then expression is
	// [^T]|T[^A]|TA[^G]|TAG[^!]|TAG![^R]|TAG!R[^I]|TAG!RI[^G]|TAG!RIG[^H]|TAG!RIGH[^T]
	var exclusion []string
	for idx, val := range delims[3] {
		exclusion = append(exclusion, "[^"+string(val)+"]")
		if idx > 0 {
			exclusion[idx] = delims[3][0:idx] + exclusion[idx]
		}
	}

	p := fmt.Sprintf(`%s-?\s*(.+?)\s*-?%s|([ \t]+)|(\n)|%s-?\s*(\w+)(?:\s+((?:%v)+?))?\s*-?%s`,
		// QuoteMeta will escape any of these that are regex commands
		regexp.QuoteMeta(delims[0]), regexp.QuoteMeta(delims[1]),
		regexp.QuoteMeta(delims[2]), strings.Join(exclusion, "|"), regexp.QuoteMeta(delims[3]),
	)
	tokenMatcher := regexp.MustCompile(p)

	return tokenMatcher
}
