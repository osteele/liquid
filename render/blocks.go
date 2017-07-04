package render

import (
	"fmt"
	"io"
)

// BlockParser builds a renderer for the tag instance.
type BlockParser func(ASTBlock) (func(io.Writer, RenderContext) error, error)

// blockDef tells the parser how to parse control tags.
type blockDef struct {
	name                  string
	isBranchTag, isEndTag bool
	syntaxModel           *blockDef
	parent                *blockDef
	parser                BlockParser
}

func (c *blockDef) compatibleParent(p *blockDef) bool {
	if p == nil {
		return false
	}
	if !c.isEndTag && p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return c.parent == p
}

func (c *blockDef) requiresParent() bool {
	return c.isBranchTag || c.isEndTag
}

func (c *blockDef) isStartTag() bool {
	return !c.isBranchTag && !c.isEndTag
}

func (s Config) addBlockDef(ct *blockDef) {
	s.controlTags[ct.name] = ct
}

func (s Config) findBlockDef(name string) (*blockDef, bool) {
	ct, found := s.controlTags[name]
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
	rt := b.s.controlTags[name]
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
func (b blockDefBuilder) Renderer(fn func(io.Writer, RenderContext) error) {
	b.tag.parser = func(node ASTBlock) (func(io.Writer, RenderContext) error, error) {
		// TODO parse error if there are arguments?
		return fn, nil
	}
}
