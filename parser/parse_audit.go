package parser

import (
	"fmt"
	"strings"

	"github.com/osteele/liquid/expressions"
)

// ParseDiagRelated is supplementary source info for a ParseDiag.
type ParseDiagRelated struct {
	Loc     SourceLoc
	Message string
}

// ParseDiag is an internal parse-time diagnostic.
// It is converted to the public Diagnostic type at the API boundary.
type ParseDiag struct {
	Code    string
	Message string
	Tok     Token
	Related []ParseDiagRelated
}

// ParseAudit is the error-recovering variant of Parse.
// It returns the AST, non-fatal diagnostics (syntax errors), and a fatal error
// (unclosed-tag or unexpected-tag). All three may be inspected independently:
//   - diags contains only non-fatal issues (syntax-error); it is empty when fatalErr != nil and none occurred before it
//   - fatalErr is non-nil only for the two structural errors that prevent a coherent AST
func (c *Config) ParseAudit(source string, loc SourceLoc) (ASTNode, []ParseDiag, Error) {
	tokens := Scan(source, loc, c.Delims)
	return c.parseTokensAudit(tokens)
}

// parseTokensAudit is the error-recovering variant of parseTokens.
// It treats expression parse failures in {{ }} objects as non-fatal syntax-errors.
// Only the two structural errors (unexpected-tag and unclosed-tag) remain fatal.
func (c *Config) parseTokensAudit(tokens []Token) (ASTNode, []ParseDiag, Error) { //nolint: gocyclo
	type frame struct {
		syntax BlockSyntax
		node   *ASTBlock
		ap     *[]ASTNode
	}

	var (
		g         = c.Grammar
		root      = &ASTSeq{}
		ap        = &root.Children
		sd        BlockSyntax
		bn        *ASTBlock
		stack     []frame
		rawTag    *ASTRaw
		inComment = false
		inRaw     = false
		diags     []ParseDiag
		lastTok   Token
	)

	for _, tok := range tokens {
		lastTok = tok
		switch {
		case inComment:
			if tok.Type == TagTokenType && (tok.Name == "endcomment" || tok.Name == "enddoc") {
				inComment = false
			}
		case inRaw:
			if tok.Type == TagTokenType && tok.Name == "endraw" {
				inRaw = false
			} else {
				rawTag.Slices = append(rawTag.Slices, tok.Source)
			}
		case tok.Type == ObjTokenType:
			if tok.Args == "" {
				break
			}
			expr, err := expressions.Parse(tok.Args)
			if err != nil {
				// Non-fatal: emit diagnostic and replace with a broken node.
				diags = append(diags, ParseDiag{
					Code:    "syntax-error",
					Message: err.Error(),
					Tok:     tok,
				})
				*ap = append(*ap, &ASTBroken{Token: tok, ParseErr: err})

				break
			}
			*ap = append(*ap, &ASTObject{tok, expr})
		case tok.Type == TextTokenType:
			*ap = append(*ap, &ASTText{Token: tok})
		case tok.Type == TagTokenType:
			if g == nil {
				return nil, diags, Errorf(tok, "Grammar field is nil")
			}

			if cs, ok := g.BlockSyntax(tok.Name); ok {
				switch {
				case tok.Name == "comment" || tok.Name == "doc":
					inComment = true
				case tok.Name == "raw":
					inRaw = true
					rawTag = &ASTRaw{}
					*ap = append(*ap, rawTag)
				case cs.RequiresParent() && (sd == nil || !cs.CanHaveParent(sd)):
					// unexpected-tag: fatal.
					suffix := ""
					if sd != nil {
						suffix = "; immediate parent is " + sd.TagName()
					}
					fatalTok := tok
					fatalErr := Errorf(fatalTok, "%s not inside %s%s", tok.Name, strings.Join(cs.ParentTags(), " or "), suffix)
					// Emit fatal diagnostic with code unexpected-tag then return.
					diags = append(diags, ParseDiag{
						Code:    "unexpected-tag",
						Message: fmt.Sprintf("tag %q is not inside %s%s", tok.Name, strings.Join(cs.ParentTags(), " or "), suffix),
						Tok:     fatalTok,
					})
					return nil, diags, fatalErr
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
		case tok.Type == TrimLeftTokenType:
			*ap = append(*ap, &ASTTrim{TrimDirection: Left})
		case tok.Type == TrimRightTokenType:
			*ap = append(*ap, &ASTTrim{TrimDirection: Right})
		}
	}

	if bn != nil {
		// unclosed-tag: fatal.
		// The Related entry points to the end-of-template position.
		endLoc := lastTok.EndLoc
		if endLoc.LineNo == 0 {
			endLoc = lastTok.SourceLoc
		}
		diags = append(diags, ParseDiag{
			Code:    "unclosed-tag",
			Message: fmt.Sprintf("tag %q opened here was never closed", bn.Name),
			Tok:     bn.Token,
			Related: []ParseDiagRelated{{
				Loc:     endLoc,
				Message: fmt.Sprintf("expected {%% end%s %%} before end of template", bn.Name),
			}},
		})
		return nil, diags, Errorf(bn, "unterminated %q block", bn.Name)
	}

	return root, diags, nil
}
