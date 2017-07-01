package chunks

import (
	"fmt"
	"io"
)

// ControlTagParser builds a renderer for the tag instance.
type ControlTagParser func(ASTControlTag) (func(io.Writer, RenderContext) error, error)

// controlTagDefinition tells the parser how to parse control tags.
type controlTagDefinition struct {
	name                  string
	isBranchTag, isEndTag bool
	syntaxModel           *controlTagDefinition
	parent                *controlTagDefinition
	parser                ControlTagParser
}

func (c *controlTagDefinition) compatibleParent(p *controlTagDefinition) bool {
	if p == nil {
		return false
	}
	if !c.isEndTag && p.syntaxModel != nil {
		p = p.syntaxModel
	}
	return c.parent == p
}

func (c *controlTagDefinition) requiresParent() bool {
	return c.isBranchTag || c.isEndTag
}

func (c *controlTagDefinition) isStartTag() bool {
	return !c.isBranchTag && !c.isEndTag
}

func (s Settings) addControlTagDefinition(ct *controlTagDefinition) {
	s.controlTags[ct.name] = ct
}

func (s Settings) findControlTagDefinition(name string) (*controlTagDefinition, bool) {
	ct, found := s.controlTags[name]
	return ct, found
}

type tagBuilder struct {
	s   Settings
	tag *controlTagDefinition
}

// AddStartTag defines a control tag and its matching end tag.
func (s Settings) AddStartTag(name string) tagBuilder {
	ct := &controlTagDefinition{name: name}
	s.addControlTagDefinition(ct)
	s.addControlTagDefinition(&controlTagDefinition{name: "end" + name, isEndTag: true, parent: ct})
	return tagBuilder{s, ct}
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tags.
func (b tagBuilder) Branch(name string) tagBuilder {
	b.s.addControlTagDefinition(&controlTagDefinition{name: name, isBranchTag: true, parent: b.tag})
	return b
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (b tagBuilder) Governs(_ []string) tagBuilder {
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (b tagBuilder) SameSyntaxAs(name string) tagBuilder {
	ot := b.s.controlTags[name]
	if ot == nil {
		panic(fmt.Errorf("undefined: %s", name))
	}
	b.tag.syntaxModel = ot
	return b
}

// Parser sets the parser for a control tag definition.
func (b tagBuilder) Parser(fn ControlTagParser) {
	b.tag.parser = fn
}

// Renderer sets the render action for a control tag definition.
func (b tagBuilder) Renderer(fn func(io.Writer, RenderContext) error) {
	b.tag.parser = func(node ASTControlTag) (func(io.Writer, RenderContext) error, error) {
		// TODO parse error if there are arguments?
		return fn, nil
	}
}
