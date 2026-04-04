package render

import (
	"github.com/osteele/liquid/parser"
)

// Config holds configuration information for parsing and rendering.
type Config struct {
	parser.Config
	grammar

	Cache           map[string][]byte
	StrictVariables bool
	TemplateStore   TemplateStore

	// Globals are variables that are accessible in every rendering context,
	// including isolated sub-contexts created by the {% render %} tag.
	// They have lower priority than scope bindings: if a key exists in both,
	// the scope binding wins.
	Globals map[string]any

	escapeReplacer Replacer

	// JekyllExtensions enables Jekyll-specific extensions to Liquid.
	// When true, allows dot notation in assign tags (e.g., {% assign page.canonical_url = value %})
	// This is not part of the Shopify Liquid standard but is used in Jekyll and Gojekyll.
	// Default: false (strict Shopify Liquid compatibility)
	JekyllExtensions bool
}

type grammar struct {
	tags           map[string]TagCompiler
	blockDefs      map[string]*blockSyntax
	tagAnalyzers   map[string]TagAnalyzer
	blockAnalyzers map[string]BlockAnalyzer
}

// NewConfig creates a new Settings.
// TemplateStore is initialized to a FileTemplateStore for backwards compatibility
// AddTagAnalyzer registers a static analysis function for the named tag.
func (c *Config) AddTagAnalyzer(name string, a TagAnalyzer) {
	if c.tagAnalyzers == nil {
		c.tagAnalyzers = map[string]TagAnalyzer{}
	}
	c.tagAnalyzers[name] = a
}

// AddBlockAnalyzer registers a static analysis function for the named block tag.
func (c *Config) AddBlockAnalyzer(name string, a BlockAnalyzer) {
	if c.blockAnalyzers == nil {
		c.blockAnalyzers = map[string]BlockAnalyzer{}
	}
	c.blockAnalyzers[name] = a
}

func (g grammar) findTagAnalyzer(name string) (TagAnalyzer, bool) {
	a, ok := g.tagAnalyzers[name]
	return a, ok
}

func (g grammar) findBlockAnalyzer(name string) (BlockAnalyzer, bool) {
	a, ok := g.blockAnalyzers[name]
	return a, ok
}

func NewConfig() Config {
	g := grammar{
		tags:      map[string]TagCompiler{},
		blockDefs: map[string]*blockSyntax{},
	}

	return Config{
		Config:        parser.NewConfig(g),
		grammar:       g,
		Cache:         map[string][]byte{},
		TemplateStore: &FileTemplateStore{},
	}
}

func (c *Config) SetAutoEscapeReplacer(replacer Replacer) {
	c.escapeReplacer = replacer
	c.AddSafeFilter()
}
