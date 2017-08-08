// Package parser is an internal package that parses template source into an abstract syntax tree.
package parser

import (
	"fmt"
	"strings"

	"github.com/osteele/liquid/expressions"
)

// Parse parses a source template. It returns an AST root, that can be compiled and evaluated.
func (c Config) Parse(source string, loc SourceLoc) (ASTNode, Error) {
	tokens := Scan(source, loc, c.Delims)
	return c.parseTokens(tokens)
}

// Parse creates an AST from a sequence of tokens.
func (c Config) parseTokens(tokens []Token) (ASTNode, Error) { // nolint: gocyclo
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
			expr, err := expressions.Parse(tok.Args)
			if err != nil {
				return nil, WrapError(err, tok)
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
					return nil, Errorf(tok, "%s not inside %s%s", tok.Name, strings.Join(cs.ParentTags(), " or "), suffix)
				case cs.IsBlockStart():
					push := func() {
						stack = append(stack, frame{syntax: sd, node: bn, ap: ap})
						sd, bn = cs, &ASTBlock{Token: tok, syntax: cs}
						*ap = append(*ap, bn)
					}
					push()
					ap = &bn.Body
				case cs.IsClause():
					n := &ASTBlock{Token: tok, syntax: cs}
					bn.Clauses = append(bn.Clauses, n)
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
		return nil, Errorf(bn, "unterminated %q block", bn.Name)
	}
	return root, nil
}
