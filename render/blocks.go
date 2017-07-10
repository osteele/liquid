package render

import (
	"fmt"
	"io"

	"github.com/osteele/liquid/parser"
)

// BlockCompiler builds a renderer for the tag instance.
type BlockCompiler func(BlockNode) (func(io.Writer, Context) error, error)

// blockSyntax tells the parser how to parse a control tag.
type blockSyntax struct {
	name                  string
	isBranchTag, isEndTag bool
	syntaxModel           *blockSyntax
	parent                *blockSyntax
	parser                BlockCompiler
}

func (s *blockSyntax) CanHaveParent(parent parser.BlockSyntax) bool {
	if parent == nil {
		return false
	}
	p := parent.(*blockSyntax)
	if !s.isEndTag && p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return s.parent == p
}

func (s *blockSyntax) IsBlock() bool        { return true }
func (s *blockSyntax) IsBlockEnd() bool     { return s.isEndTag }
func (s *blockSyntax) IsBlockStart() bool   { return !s.isBranchTag && !s.isEndTag }
func (s *blockSyntax) IsBranch() bool       { return s.isBranchTag }
func (s *blockSyntax) RequiresParent() bool { return s.isBranchTag || s.isEndTag }

func (s *blockSyntax) ParentTags() []string {
	if s.parent == nil {
		return []string{}
	}
	return []string{s.parent.name}
}
func (s *blockSyntax) TagName() string { return s.name }

func (g grammar) addBlockDef(ct *blockSyntax) {
	g.blockDefs[ct.name] = ct
}

func (g grammar) findBlockDef(name string) (*blockSyntax, bool) {
	ct, found := g.blockDefs[name]
	return ct, found
}

// BlockSyntax is part of the Grammar interface.
func (g grammar) BlockSyntax(name string) (parser.BlockSyntax, bool) {
	ct, found := g.blockDefs[name]
	return ct, found
}

type blockDefBuilder struct {
	grammar
	tag *blockSyntax
}

// AddBlock defines a control tag and its matching end tag.
func (g grammar) AddBlock(name string) blockDefBuilder { // nolint: golint
	ct := &blockSyntax{name: name}
	g.addBlockDef(ct)
	g.addBlockDef(&blockSyntax{name: "end" + name, isEndTag: true, parent: ct})
	return blockDefBuilder{g, ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tag.
func (b blockDefBuilder) Branch(name string) blockDefBuilder {
	b.addBlockDef(&blockSyntax{name: name, isBranchTag: true, parent: b.tag})
	return b
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (b blockDefBuilder) Governs(_ []string) blockDefBuilder {
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (b blockDefBuilder) SameSyntaxAs(name string) blockDefBuilder {
	rt := b.blockDefs[name]
	if rt == nil {
		panic(fmt.Errorf("undefined: %s", name))
	}
	b.tag.syntaxModel = rt
	return b
}

// Compiler sets the parser for a control tag definition.
func (b blockDefBuilder) Compiler(fn BlockCompiler) {
	b.tag.parser = fn
}

// Renderer sets the render action for a control tag definition.
func (b blockDefBuilder) Renderer(fn func(io.Writer, Context) error) {
	b.tag.parser = func(node BlockNode) (func(io.Writer, Context) error, error) {
		// TODO parse error if there are arguments?
		return fn, nil
	}
}
