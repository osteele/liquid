package render

import (
	"context"
	"sync"

	"github.com/osteele/liquid/parser"
)

// Config holds configuration information for parsing and rendering.
type Config struct {
	parser.Config
	grammar

	Cache           sync.Map // key: string, value: []byte — safe for concurrent use
	StrictVariables bool
	TemplateStore   TemplateStore

	// Globals are variables that are accessible in every rendering context,
	// including isolated sub-contexts created by the {% render %} tag.
	// They have lower priority than scope bindings: if a key exists in both,
	// the scope binding wins.
	Globals map[string]any

	escapeReplacer Replacer

	// globalFilter is a function applied to the value of every {{ }} expression
	// before it is written to the output. Analogous to Ruby's global_filter option.
	globalFilter func(any) (any, error)

	// JekyllExtensions enables Jekyll-specific extensions to Liquid.
	// When true, allows dot notation in assign tags (e.g., {% assign page.canonical_url = value %})
	// This is not part of the Shopify Liquid standard but is used in Jekyll and Gojekyll.
	// Default: false (strict Shopify Liquid compatibility)
	JekyllExtensions bool

	// TrimTagLeft, when true, automatically trims whitespace to the left of every
	// {% tag %} and block open/close tag, as if each had a {%- prefix.
	TrimTagLeft bool

	// TrimTagRight, when true, automatically trims whitespace to the right of every
	// {% tag %} and block open/close tag, as if each had a -%} suffix.
	TrimTagRight bool

	// TrimOutputLeft, when true, automatically trims whitespace to the left of every
	// {{ output }} expression, as if each had a {{- prefix.
	TrimOutputLeft bool

	// TrimOutputRight, when true, automatically trims whitespace to the right of every
	// {{ output }} expression, as if each had a -}} suffix.
	TrimOutputRight bool

	// Greedy controls whether whitespace trimming removes all consecutive blank
	// characters including newlines (true, the default), or only trims inline
	// blanks (space/tab) plus at most one newline (false).
	Greedy bool

	// SizeLimit, when positive, caps the total number of bytes written to the
	// render output. A render that would exceed this limit fails with an error.
	SizeLimit int64

	// Context is an optional Go context.Context that can be used to cancel a
	// render in-flight (e.g. for per-request timeouts). When set, each node
	// render checks for cancellation before proceeding.
	Context context.Context

	// ExceptionHandler, when non-nil, is called for each render-time error
	// encountered during node evaluation. The function receives the error and
	// returns a string to emit in place of the failed node. Returning an empty
	// string suppresses the node output. This is analogous to Ruby Liquid's
	// exception_renderer option.
	ExceptionHandler func(error) string

	// LaxTags, when true, silently ignores unknown tags instead of raising a
	// parse error. Only the render-path skips unknown tags; analysis still
	// treats them as no-ops.
	LaxTags bool
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

	cfg := Config{
		Config:        parser.NewConfig(g),
		grammar:       g,
		TemplateStore: &FileTemplateStore{},
		Greedy:        true,
	}
	// Register "raw" unconditionally — it is a LiquidJS-standard filter that marks
	// a value as safe (skips autoescape). When autoescape is off it is a no-op.
	cfg.AddSafeFilter()
	return cfg
}

func (c *Config) SetAutoEscapeReplacer(replacer Replacer) {
	c.escapeReplacer = replacer
	c.AddSafeFilter()
}

// SetGlobalFilter sets a function that is applied to the evaluated value of every
// {{ expression }} before it is written to the output. This is analogous to Ruby
// Liquid's global_filter option. The function receives the evaluated value and
// returns a transformed value or an error.
func (c *Config) SetGlobalFilter(fn func(any) (any, error)) {
	c.globalFilter = fn
}
