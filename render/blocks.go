package render

import (
	"io"
	"sort"

	"github.com/osteele/liquid/parser"
)

// BlockCompiler builds a renderer for the tag instance.
type BlockCompiler func(BlockNode) (func(io.Writer, Context) error, error)

// blockSyntax tells the parser how to parse a control tag.
type blockSyntax struct {
	name                  string
	isClauseTag, isEndTag bool
	startName             string          // for an end tag, the name of the correspondign start tag
	parents               map[string]bool // if non-nil, must be an immediate clause of one of these
	parser                BlockCompiler
}

func (s *blockSyntax) CanHaveParent(parent parser.BlockSyntax) bool {
	switch {
	case s.isClauseTag:
		return parent != nil && s.parents[parent.TagName()]
	case s.isEndTag:
		return parent != nil && parent.TagName() == s.startName
	default:
		return true
	}
}

func (s *blockSyntax) IsBlock() bool        { return true }
func (s *blockSyntax) IsBlockEnd() bool     { return s.isEndTag }
func (s *blockSyntax) IsBlockStart() bool   { return !s.isClauseTag && !s.isEndTag }
func (s *blockSyntax) IsClause() bool       { return s.isClauseTag }
func (s *blockSyntax) RequiresParent() bool { return s.isClauseTag || s.isEndTag }

func (s *blockSyntax) ParentTags() (parents []string) {
	for k := range s.parents {
		parents = append(parents, k)
	}
	sort.Strings(parents)
	return
}
func (s *blockSyntax) TagName() string { return s.name }

func (g grammar) addBlockDef(ct *blockSyntax) {
	if g.blockDefs[ct.name] != nil {
		panic("duplicate definition of " + ct.name)
	}
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
	g.addBlockDef(&blockSyntax{name: "end" + name, isEndTag: true, startName: name})
	return blockDefBuilder{g, ct}
}

// Clause tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tag.
func (b blockDefBuilder) Clause(name string) blockDefBuilder {
	if b.blockDefs[name] == nil {
		b.addBlockDef(&blockSyntax{name: name, isClauseTag: true})
	}
	c := b.blockDefs[name]
	if !c.isClauseTag {
		panic(name + " has already been defined as a non-clause")
	}
	if len(c.parents) == 0 {
		c.parents = make(map[string]bool)
	}
	c.parents[b.tag.name] = true
	return b
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
// func (b blockDefBuilder) SameSyntaxAs(name string) blockDefBuilder {
// 	rt := b.blockDefs[name]
// 	if rt == nil {
// 		panic(fmt.Errorf("undefined: %s", name))
// 	}
// 	b.tag.syntaxModel = rt
// 	return b
// }

// Compiler sets the parser for a control tag definition.
func (b blockDefBuilder) Compiler(fn BlockCompiler) {
	b.tag.parser = fn
}

// Renderer sets the render action for a control tag definition.
func (b blockDefBuilder) Renderer(fn func(io.Writer, Context) error) {
	b.tag.parser = func(node BlockNode) (func(io.Writer, Context) error, error) {
		// TODO syntax error if there are arguments?
		return fn, nil
	}
}
