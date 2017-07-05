package render

import (
	"fmt"
	"io"
)

// BlockParser builds a renderer for the tag instance.
type BlockParser func(ASTBlock) (func(io.Writer, Context) error, error)

// blockDef tells the parser how to parse control tags.
type blockDef struct {
	name                  string
	isBranchTag, isEndTag bool
	syntaxModel           *blockDef
	parent                *blockDef
	parser                BlockParser
}

func (c *blockDef) CanHaveParent(parent BlockSyntax) bool {
	if parent == nil {
		return false
	}
	p := parent.(*blockDef)
	if !c.isEndTag && p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return c.parent == p
}

func (c *blockDef) requiresParent() bool {
	return c.isBranchTag || c.isEndTag
}

func (c *blockDef) IsBlock() bool        { return true }
func (c *blockDef) IsBlockEnd() bool     { return c.isEndTag }
func (c *blockDef) IsBlockStart() bool   { return !c.isBranchTag && !c.isEndTag }
func (c *blockDef) IsBranch() bool       { return c.isBranchTag }
func (c *blockDef) RequiresParent() bool { return c.isBranchTag || c.isEndTag }

func (c *blockDef) ParentTags() []string {
	if c.parent == nil {
		return []string{}
	}
	return []string{c.parent.name}
}
func (c *blockDef) TagName() string { return c.name }

func (s Config) addBlockDef(ct *blockDef) {
	s.blockDefs[ct.name] = ct
}

func (s Config) findBlockDef(name string) (*blockDef, bool) {
	ct, found := s.blockDefs[name]
	return ct, found
}

func (s Config) BlockSyntax(name string) (BlockSyntax, bool) {
	ct, found := s.blockDefs[name]
	return ct, found
}

type blockDefBuilder struct {
	s   Config
	tag *blockDef
}

// AddBlock defines a control tag and its matching end tag.
func (s Config) AddBlock(name string) blockDefBuilder { // nolint: golint
	ct := &blockDef{name: name}
	s.addBlockDef(ct)
	s.addBlockDef(&blockDef{name: "end" + name, isEndTag: true, parent: ct})
	return blockDefBuilder{s, ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tags.
func (b blockDefBuilder) Branch(name string) blockDefBuilder {
	b.s.addBlockDef(&blockDef{name: name, isBranchTag: true, parent: b.tag})
	return b
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (b blockDefBuilder) Governs(_ []string) blockDefBuilder {
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (b blockDefBuilder) SameSyntaxAs(name string) blockDefBuilder {
	rt := b.s.blockDefs[name]
	if rt == nil {
		panic(fmt.Errorf("undefined: %s", name))
	}
	b.tag.syntaxModel = rt
	return b
}

// Parser sets the parser for a control tag definition.
func (b blockDefBuilder) Parser(fn BlockParser) {
	b.tag.parser = fn
}

// Renderer sets the render action for a control tag definition.
func (b blockDefBuilder) Renderer(fn func(io.Writer, Context) error) {
	b.tag.parser = func(node ASTBlock) (func(io.Writer, Context) error, error) {
		// TODO parse error if there are arguments?
		return fn, nil
	}
}
