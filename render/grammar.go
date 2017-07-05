package render

// Grammar supplies the parser with syntax information about blocks.
type Grammar interface {
	BlockSyntax(string) (BlockSyntax, bool)
}

// BlockSyntax supplies the parser with syntax information about blocks.
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

// Grammar returns a configuration's grammar.
func (c *Config) Grammar() Grammar { return c }
