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

func (c *blockSyntax) CanHaveParent(parent parser.BlockSyntax) bool {
	if parent == nil {
		return false
	}
	p := parent.(*blockSyntax)
	if !c.isEndTag && p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return c.parent == p
}

func (c *blockSyntax) IsBlock() bool        { return true }
func (c *blockSyntax) IsBlockEnd() bool     { return c.isEndTag }
func (c *blockSyntax) IsBlockStart() bool   { return !c.isBranchTag && !c.isEndTag }
func (c *blockSyntax) IsBranch() bool       { return c.isBranchTag }
func (c *blockSyntax) RequiresParent() bool { return c.isBranchTag || c.isEndTag }

func (c *blockSyntax) ParentTags() []string {
	if c.parent == nil {
		return []string{}
	}
	return []string{c.parent.name}
}
func (c *blockSyntax) TagName() string { return c.name }

func (c Config) addBlockDef(ct *blockSyntax) {
	c.blockDefs[ct.name] = ct
}

func (c Config) findBlockDef(name string) (*blockSyntax, bool) {
	ct, found := c.blockDefs[name]
	return ct, found
}

// BlockSyntax is part of the Grammar interface.
func (c Config) BlockSyntax(name string) (parser.BlockSyntax, bool) {
	ct, found := c.blockDefs[name]
	return ct, found
}

type blockDefBuilder struct {
	cfg Config
	tag *blockSyntax
}

// AddBlock defines a control tag and its matching end tag.
func (c Config) AddBlock(name string) blockDefBuilder { // nolint: golint
	ct := &blockSyntax{name: name}
	c.addBlockDef(ct)
	c.addBlockDef(&blockSyntax{name: "end" + name, isEndTag: true, parent: ct})
	return blockDefBuilder{c, ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tag.
func (b blockDefBuilder) Branch(name string) blockDefBuilder {
	b.cfg.addBlockDef(&blockSyntax{name: name, isBranchTag: true, parent: b.tag})
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
