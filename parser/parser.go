package parser

import (
	"fmt"
	"strings"

	"github.com/osteele/liquid/expression"
)

// A ParseError is a parse error during the template parsing.
type ParseError string

func (e ParseError) Error() string { return string(e) }

func parseErrorf(format string, a ...interface{}) ParseError {
	return ParseError(fmt.Sprintf(format, a...))
}

// Parse parses a source template. It returns an AST root, that can be compiled and evaluated.
func (c Config) Parse(source string) (ASTNode, error) {
	tokens := Scan(source, c.Filename, c.LineNo)
	return c.parseTokens(tokens)
}

// Parse creates an AST from a sequence of tokens.
func (c Config) parseTokens(tokens []Token) (ASTNode, error) { // nolint: gocyclo
	// a stack of control tag state, for matching nested {%if}{%endif%} etc.
	type frame struct {
		syntax BlockSyntax
		node   *ASTBlock
		ap     *[]ASTNode
	}
	var (
		g         = c.Grammar
		root      = &ASTSeq{}      // root of AST; will be returned
		ap        = &root.Children // newly-constructed nodes are appended here
		sd        BlockSyntax      // current block syntax definition
		bn        *ASTBlock        // current block node
		stack     []frame          // stack of blocks
		rawTag    *ASTRaw          // current raw tag
		inComment = false
		inRaw     = false
	)
	for _, tok := range tokens {
		switch {
		// The parser needs to know about comment and raw, because tags inside
		// needn't match each other e.g. {%comment%}{%if%}{%endcomment%}
		// TODO is this true?
		case inComment:
			if tok.Type == TagTokenType && tok.Name == "endcomment" {
				inComment = false
			}
		case inRaw:
			if tok.Type == TagTokenType && tok.Name == "endraw" {
				inRaw = false
			} else {
				rawTag.Slices = append(rawTag.Slices, tok.Source)
			}
		case tok.Type == ObjTokenType:
			expr, err := expression.Parse(tok.Args)
			if err != nil {
				return nil, err
			}
			*ap = append(*ap, &ASTObject{tok, expr})
		case tok.Type == TextTokenType:
			*ap = append(*ap, &ASTText{Token: tok})
		case tok.Type == TagTokenType:
			if cs, ok := g.BlockSyntax(tok.Name); ok {
				switch {
				case tok.Name == "comment":
					inComment = true
				case tok.Name == "raw":
					inRaw = true
					rawTag = &ASTRaw{}
					*ap = append(*ap, rawTag)
				case cs.RequiresParent() && (sd == nil || !cs.CanHaveParent(sd)):
					suffix := ""
					if sd != nil {
						suffix = "; immediate parent is " + sd.TagName()
					}
					return nil, parseErrorf("%s not inside %s%s", tok.Name, strings.Join(cs.ParentTags(), " or "), suffix)
				case cs.IsBlockStart():
					push := func() {
						stack = append(stack, frame{syntax: sd, node: bn, ap: ap})
						sd, bn = cs, &ASTBlock{Token: tok, syntax: cs}
						*ap = append(*ap, bn)
					}
					push()
					ap = &bn.Body
				case cs.IsBranch():
					n := &ASTBlock{Token: tok, syntax: cs}
					bn.Branches = append(bn.Branches, n)
					ap = &n.Body
				case cs.IsBlockEnd():
					pop := func() {
						f := stack[len(stack)-1]
						stack = stack[:len(stack)-1]
						sd, bn, ap = f.syntax, f.node, f.ap
					}
					pop()
				default:
					panic(fmt.Errorf("block type %q", tok.Name))
				}
			} else {
				*ap = append(*ap, &ASTTag{tok})
			}
		}
	}
	if bn != nil {
		return nil, parseErrorf("unterminated %s block at %s", bn.Name, bn.SourceInfo)
	}
	return root, nil
}
