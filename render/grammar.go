package render

type Grammar interface {
	BlockSyntax(string) (BlockSyntax, bool)
}

type BlockSyntax interface {
	IsBlock() bool
	CanHaveParent(BlockSyntax) bool
	IsBlockEnd() bool
	IsBlockStart() bool
	IsBranch() bool
	ParentTags() []string
	RequiresParent() bool
	TagName() string
}

func (c *Config) Grammar() Grammar { return c }

// func (c *Config) BlockSyntax(tagName string) BlockSyntax {
// 	s, _ := c.findBlockDef(tagName)
// 	return s
// }

// 	IsBlockTag(string) bool
// func (c *Config) IsBranch(tag string) bool {
// 	// RequiresBalancedChildren(string) bool
// }

// // func (c *Config) RequiresBalancedChildren(tag string) bool {
// // 	return tag != "comment" && tag != "raw"
// // }

// func (c *Config) IsBlockTag(tag string) bool {
// 	return c.findBlockDef(c.Name) != nil
// }

// func (c *Config) CanHaveParent(tag, parent string) bool {
// 	cd := c.findBlockDef(c.Name)
// 	p := c.findBlockDef(parent)
// 	return cd.compatibleParent(ccd)
// }

// func (c *Config) IsBlockStart(tag string) bool {
// 	return c.findBlockDef(tag).cd.isStartTag()
// }
// func (c *Config) IsBlockEnd(tag string) bool {
// 	return c.findBlockDef(tag).cd.isEndTag
// }

// func (c *Config) IsBranch(tag string) bool {
// 	return c.findBlockDef(tag).isBranchTag
// }
