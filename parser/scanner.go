package parser

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// defaultTokenMatcher is the compiled regex for the default delimiters, cached at package level.
var defaultTokenMatcher = formTokenMatcher([]string{"{{", "}}", "{%", "%}"})

// customTokenMatchers caches compiled regexps for custom delimiter sets.
var customTokenMatchers sync.Map // key: [4]string, value: *regexp.Regexp

// Scan breaks a string into a sequence of Tokens.
func Scan(data string, loc SourceLoc, delims []string) (tokens []Token) {
	// Apply defaults
	if len(delims) != 4 {
		delims = []string{"{{", "}}", "{%", "%}"}
	}

	var tokenMatcher *regexp.Regexp
	if delims[0] == "{{" && delims[1] == "}}" && delims[2] == "{%" && delims[3] == "%}" {
		tokenMatcher = defaultTokenMatcher
	} else {
		key := [4]string{delims[0], delims[1], delims[2], delims[3]}
		if cached, ok := customTokenMatchers.Load(key); ok {
			tokenMatcher = cached.(*regexp.Regexp)
		} else {
			tokenMatcher = formTokenMatcher(delims)
			customTokenMatchers.Store(key, tokenMatcher)
		}
	}

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
		switch {
		case data[ts:ts+len(delims[0])] == delims[0]:
			if source[2] == '-' {
				tokens = append(tokens, Token{
					Type: TrimLeftTokenType,
				})
			}

			tokens = append(tokens, Token{
				Type:      ObjTokenType,
				SourceLoc: loc,
				Source:    source,
				Args:      data[m[2]:m[3]],
			})
			if source[len(source)-3] == '-' {
				tokens = append(tokens, Token{
					Type: TrimRightTokenType,
				})
			}
		case data[ts:ts+len(delims[2])] == delims[2]:
			if source[2] == '-' {
				tokens = append(tokens, Token{
					Type: TrimLeftTokenType,
				})
			}

			tok := Token{
				Type:      TagTokenType,
				SourceLoc: loc,
				Source:    source,
				Name:      data[m[4]:m[5]],
			}
			if m[6] > 0 {
				tok.Args = data[m[6]:m[7]]
			}

			tokens = append(tokens, tok)
			if source[len(source)-3] == '-' {
				tokens = append(tokens, Token{
					Type: TrimRightTokenType,
				})
			}
		}

		loc.LineNo += strings.Count(source, "\n")
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
	exclusion := make([]string, 0, len(delims[3]))
	for idx, val := range delims[3] {
		exclusion = append(exclusion, "[^"+string(val)+"]")
		if idx > 0 {
			exclusion[idx] = delims[3][0:idx] + exclusion[idx]
		}
	}

	tokenMatcher := regexp.MustCompile(
		fmt.Sprintf(`%s-?\s*(.+?)\s*-?%s|%s-?\s*(\w+)(?:\s+((?:%v)+?))?\s*-?%s`,
			// QuoteMeta will escape any of these that are regex commands
			regexp.QuoteMeta(delims[0]), regexp.QuoteMeta(delims[1]),
			regexp.QuoteMeta(delims[2]), strings.Join(exclusion, "|"), regexp.QuoteMeta(delims[3]),
		),
	)

	return tokenMatcher
}
