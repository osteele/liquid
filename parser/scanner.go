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

	// lastNL is the byte offset of the most recent '\n' in data, or -1 before the first one.
	// Column of byte at position pos is: pos - lastNL (1-based).
	lastNL := -1

	// If the initial loc already has a ColNo set, back-compute the effective lastNL so that
	// position 0 maps to that column. Otherwise column 1 starts at position 0.
	if loc.ColNo > 1 {
		lastNL = -(loc.ColNo - 1)
	}

	colOf := func(pos int) int { return pos - lastNL }

	// advanceNL updates lastNL and loc.LineNo for the newlines in data[from:to].
	advanceNL := func(from, to int) {
		chunk := data[from:to]
		n := strings.Count(chunk, "\n")
		if n > 0 {
			loc.LineNo += n
			lastNL = from + strings.LastIndex(chunk, "\n")
		}
	}

	p, pe := 0, len(data)
	for _, m := range tokenMatcher.FindAllStringSubmatchIndex(data, -1) {
		ts, te := m[0], m[1]
		if p < ts {
			textLoc := loc
			textLoc.ColNo = colOf(p)
			text := data[p:ts]
			tokens = append(tokens, Token{
				Type:      TextTokenType,
				SourceLoc: textLoc,
				EndLoc:    tokenEndLoc(textLoc, text),
				Source:    text,
			})
			advanceNL(p, ts)
		}

		source := data[ts:te]
		tokLoc := loc
		tokLoc.ColNo = colOf(ts)
		tokEndLoc := tokenEndLoc(tokLoc, source)

		switch {
		case data[ts:ts+len(delims[0])] == delims[0]:
			leftTrim := source[2] == '-'
			rightTrim := source[len(source)-3] == '-'
			if leftTrim {
				tokens = append(tokens, Token{
					Type: TrimLeftTokenType,
				})
			}

			// When the only captured content is the trim marker itself (e.g. {{-}} or {{- -}}),
			// treat the expression as empty so it renders nothing rather than producing a parse error.
			args := data[m[2]:m[3]]
			if args == "-" && (leftTrim || rightTrim) {
				args = ""
			}

			tokens = append(tokens, Token{
				Type:      ObjTokenType,
				SourceLoc: tokLoc,
				EndLoc:    tokEndLoc,
				Source:    source,
				Args:      args,
			})
			if rightTrim {
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

			// m[4] < 0 means the (\w+) tag-name group didn't match.
			// This happens for inline comments: {%# ... %} where '#' is not \w.
			// In that case we emit only trim markers (if any) but no tag token.
			if m[4] >= 0 {
				tok := Token{
					Type:      TagTokenType,
					SourceLoc: tokLoc,
					EndLoc:    tokEndLoc,
					Source:    source,
					Name:      data[m[4]:m[5]],
				}
				if m[6] > 0 {
					tok.Args = data[m[6]:m[7]]
				}

				tokens = append(tokens, tok)
			}

			if source[len(source)-3] == '-' {
				tokens = append(tokens, Token{
					Type: TrimRightTokenType,
				})
			}
		}

		advanceNL(ts, te)
		p = te
	}

	if p < pe {
		textLoc := loc
		textLoc.ColNo = colOf(p)
		text := data[p:]
		tokens = append(tokens, Token{
			Type:      TextTokenType,
			SourceLoc: textLoc,
			EndLoc:    tokenEndLoc(textLoc, text),
			Source:    text,
		})
	}

	return tokens
}

// tokenEndLoc computes the exclusive end location of a token given its start
// location and source text.
func tokenEndLoc(start SourceLoc, source string) SourceLoc {
	nls := strings.Count(source, "\n")
	if nls == 0 {
		return SourceLoc{
			Pathname: start.Pathname,
			LineNo:   start.LineNo,
			ColNo:    start.ColNo + len(source),
		}
	}
	lastNL := strings.LastIndex(source, "\n")
	return SourceLoc{
		Pathname: start.Pathname,
		LineNo:   start.LineNo + nls,
		ColNo:    len(source) - lastNL, // 1-based col of character after last \n
	}
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

	// Build the same exclusion pattern for the OUTPUT right delimiter (delims[1], e.g. "}}").
	// This prevents the lazy content group from matching across an intermediate closing delimiter,
	// which would otherwise cause adjacent {{-}} tokens to merge into a single (broken) match.
	outputExclusion := make([]string, 0, len(delims[1]))
	for idx, val := range delims[1] {
		oe := "[^" + string(val) + "]"
		if idx > 0 {
			oe = delims[1][0:idx] + oe
		}
		outputExclusion = append(outputExclusion, oe)
	}

	tokenMatcher := regexp.MustCompile(
		fmt.Sprintf(`%s-?\s*((?:%v)+?)\s*-?%s|%s-?\s*#(?:(?:%v)*)-?%s|%s-?\s*(\w+)(?:\s+((?:%v)+?))?\s*-?%s`,
			// Output token: content must not contain the closing delimiter (outputExclusion).
			regexp.QuoteMeta(delims[0]), strings.Join(outputExclusion, "|"), regexp.QuoteMeta(delims[1]),
			// Inline comment alternative: {%#...%} or {%- # ...%} — optional whitespace between trim marker and #.
			// No capturing groups so existing group indices are unchanged.
			regexp.QuoteMeta(delims[2]), strings.Join(exclusion, "|"), regexp.QuoteMeta(delims[3]),
			regexp.QuoteMeta(delims[2]), strings.Join(exclusion, "|"), regexp.QuoteMeta(delims[3]),
		),
	)

	return tokenMatcher
}
