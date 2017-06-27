package chunks

import (
	"fmt"
	"io"
)

// ControlTagDefinitions is a map of tag names to control tag definitions.
var ControlTagDefinitions = map[string]*ControlTagDefinition{}

// ControlTagAction runs the interpreter.
type ControlTagAction func(ASTControlTag) func(io.Writer, Context) error

// ControlTagDefinition tells the parser how to parse control tags.
type ControlTagDefinition struct {
	Name        string
	IsBranchTag bool
	IsEndTag    bool
	SyntaxModel *ControlTagDefinition
	Parent      *ControlTagDefinition
	action      ControlTagAction
}

func (c *ControlTagDefinition) CompatibleParent(p *ControlTagDefinition) bool {
	if p == nil {
		return false
	}
	if p.SyntaxModel != nil {
		p = p.SyntaxModel
	}
	return c.Parent == p
}

func (c *ControlTagDefinition) RequiresParent() bool {
	return c.IsBranchTag || c.IsEndTag
}

func (c *ControlTagDefinition) IsStartTag() bool {
	return !c.IsBranchTag && !c.IsEndTag
}

// DefineControlTag defines a control tag and its matching end tag.
func DefineControlTag(name string) *ControlTagDefinition {
	ct := &ControlTagDefinition{Name: name}
	addControlTagDefinition(ct)
	addControlTagDefinition(&ControlTagDefinition{Name: "end" + name, IsEndTag: true, Parent: ct})
	return ct
}

func FindControlDefinition(name string) (*ControlTagDefinition, bool) {
	ct, found := ControlTagDefinitions[name]
	return ct, found
}

func addControlTagDefinition(ct *ControlTagDefinition) {
	ControlTagDefinitions[ct.Name] = ct
}

// Branch tells the parser that the named tag can appear immediately between this tag and its end tag,
// so long as it is not nested within any other control tags.
func (ct *ControlTagDefinition) Branch(name string) *ControlTagDefinition {
	addControlTagDefinition(&ControlTagDefinition{Name: name, IsBranchTag: true, Parent: ct})
	return ct
}

// Governs tells the parser that the tags can appear anywhere between this tag and its end tag.
func (ct *ControlTagDefinition) Governs(_ []string) *ControlTagDefinition {
	return ct
}

// SameSyntaxAs tells the parser that this tag has the same syntax as the named tag.
func (ct *ControlTagDefinition) SameSyntaxAs(name string) *ControlTagDefinition {
	ot := ControlTagDefinitions[name]
	if ot == nil {
		panic(fmt.Errorf("undefined: %s", name))
	}
	ct.SyntaxModel = ot
	return ct
}

// Action sets the action for a control tag definition.
func (ct *ControlTagDefinition) Action(fn ControlTagAction) {
	ct.action = fn
}
