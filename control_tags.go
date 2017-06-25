package main

import (
	"fmt"
	"io"
)

func init() {
	loopTags := []string{"break", "continue", "cycle"}
	DefineControlTag("comment").Action(unimplementedControlTag)
	DefineControlTag("if").Branch("else").Branch("elsif").Action(unimplementedControlTag)
	DefineControlTag("unless").Action(unimplementedControlTag)
	DefineControlTag("case").Branch("when").Action(unimplementedControlTag)
	DefineControlTag("for").Governs(loopTags).Action(unimplementedControlTag)
	DefineControlTag("tablerow").Governs(loopTags).Action(unimplementedControlTag)
	DefineControlTag("capture").Action(unimplementedControlTag)
}

// ControlTagDefinitions is a map of tag names to control tag definitions.
var ControlTagDefinitions = map[string]*ControlTagDefinition{}

// ControlTagAction runs the interpreter.
type ControlTagAction func(io.Writer, Context) error

// ControlTagDefinition tells the parser how to parse control tags.
type ControlTagDefinition struct {
	Name        string
	IsBranchTag bool
	IsEndTag    bool
	Parent      *ControlTagDefinition
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

func (ct *ControlTagDefinition) Branch(name string) *ControlTagDefinition {
	addControlTagDefinition(&ControlTagDefinition{Name: name, IsBranchTag: true, Parent: ct})
	return ct
}

func (ct *ControlTagDefinition) Governs(_ []string) *ControlTagDefinition {
	return ct
}

func (ct *ControlTagDefinition) Action(_ ControlTagAction) {
}

func unimplementedControlTag(io.Writer, Context) error {
	return fmt.Errorf("unimplementedControlTag")
}
