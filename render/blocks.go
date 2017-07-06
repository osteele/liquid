package render

import (
	"fmt"
	"io"
)

// BlockParser builds a renderer for the tag instance.
type BlockParser func(BlockNode) (func(io.Writer, Context) error, error)

// blockDef tells the parser how to parse control tagc.
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

func (c Config) addBlockDef(ct *blockDef) {
	c.blockDefs[ct.name] = ct
}

func (c Config) findBlockDef(name string) (*blockDef, bool) {
	ct, found := c.blockDefs[name]
	return ct, found
}

// BlockSyntax is part of the Grammar interface.
func (c Config) BlockSyntax(name string) (BlockSyntax, bool) {
	ct, found := c.blockDefs[name]
	return ct, found
}

type blockDefBuilder struct {
	cfg Config
	tag *blockDef
}

// AddBlock defines a control tag and its matching end tag.
func (c Config) AddBlock(name string) blockDefBuilder { // nolint: golint
	ct := &blockDef{name: name}
	c.addBlockDef(ct)
	c.addBlockDef(&blockDef{name: "end" + name, isEndTag: true, parent: ct})
	return blockDefBuilder{c, ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tagc.
func (b blockDefBuilder) Branch(name string) blockDefBuilder {
	b.cfg.addBlockDef(&blockDef{name: name, isBranchTag: true, parent: b.tag})
	return b
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (b blockDefBuilder) Governs(_ []string) blockDefBuilder {
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (b blockDefBuilder) SameSyntaxAs(name string) blockDefBuilder {
	rt := b.cfg.blockDefs[name]
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
	b.tag.parser = func(node BlockNode) (func(io.Writer, Context) error, error) {
		// TODO parse error if there are arguments?
		return fn, nil
	}
}
