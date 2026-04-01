package render

import (
	"strings"

	"github.com/osteele/liquid/expressions"
)

// NodeAnalysis holds static analysis metadata for a compiled node.
// Populated at compile time by tag/block analyzers.
type NodeAnalysis struct {
	// Arguments are expressions whose variable references are "used" by this node.
	// Analogous to LiquidJS tag.arguments().
	Arguments []expressions.Expression

	// LocalScope lists variable names DEFINED by this node in the current scope.
	// Analogous to LiquidJS tag.localScope(). E.g. assign, capture.
	LocalScope []string

	// BlockScope lists variable names added to the scope for this node's BODY only.
	// Analogous to LiquidJS tag.blockScope(). E.g. the loop variable in for.
	BlockScope []string
}

// TagAnalyzer provides static analysis metadata for a simple tag.
type TagAnalyzer func(args string) NodeAnalysis

// BlockAnalyzer provides static analysis metadata for a block tag.
// It receives the already-compiled BlockNode (with Body and Clauses populated).
type BlockAnalyzer func(node BlockNode) NodeAnalysis

// AnalysisResult is the result of static analysis of a compiled template.
type AnalysisResult struct {
	// Globals contains variable paths that come from the outer scope (not defined
	// within the template itself via assign, capture, for, etc.).
	Globals [][]string
	// All contains all variable paths referenced in the template, including locals.
	All [][]string
}

// Analyze performs static analysis on a compiled template tree and returns
// the set of variable paths referenced by the template.
func Analyze(root Node) AnalysisResult {
	locals := map[string]bool{}
	collectLocals(root, locals)

	collector := &analysisCollector{seen: map[string]bool{}}
	walkForVariables(root, collector)

	all := collector.paths

	var globals [][]string
	for _, path := range all {
		if len(path) > 0 && !locals[path[0]] {
			globals = append(globals, path)
		}
	}

	return AnalysisResult{All: all, Globals: globals}
}

// analysisCollector deduplicates variable paths across the full AST walk.
type analysisCollector struct {
	paths [][]string
	seen  map[string]bool
}

func (c *analysisCollector) addPath(path []string) {
	if len(path) == 0 {
		return
	}
	key := strings.Join(path, "\x00")
	if !c.seen[key] {
		c.seen[key] = true
		cp := make([]string, len(path))
		copy(cp, path)
		c.paths = append(c.paths, cp)
	}
}

func (c *analysisCollector) addFromExpr(expr expressions.Expression) {
	for _, path := range expr.Variables() {
		c.addPath(path)
	}
}

// walkForVariables traverses the AST collecting all variable references.
func walkForVariables(node Node, collector *analysisCollector) {
	switch n := node.(type) {
	case *SeqNode:
		for _, child := range n.Children {
			walkForVariables(child, collector)
		}
	case *ObjectNode:
		collector.addFromExpr(n.GetExpr())
	case *TagNode:
		for _, expr := range n.Analysis.Arguments {
			collector.addFromExpr(expr)
		}
	case *BlockNode:
		for _, expr := range n.Analysis.Arguments {
			collector.addFromExpr(expr)
		}
		for _, child := range n.Body {
			walkForVariables(child, collector)
		}
		for _, clause := range n.Clauses {
			walkForVariables(clause, collector)
		}
	}
}

// collectLocals traverses the AST collecting all locally-defined variable names.
// These are names introduced by assign, capture, for (BlockScope), etc.
func collectLocals(node Node, locals map[string]bool) {
	switch n := node.(type) {
	case *SeqNode:
		for _, child := range n.Children {
			collectLocals(child, locals)
		}
	case *TagNode:
		for _, name := range n.Analysis.LocalScope {
			locals[name] = true
		}
	case *BlockNode:
		for _, name := range n.Analysis.LocalScope {
			locals[name] = true
		}
		for _, name := range n.Analysis.BlockScope {
			locals[name] = true
		}
		for _, child := range n.Body {
			collectLocals(child, locals)
		}
		for _, clause := range n.Clauses {
			collectLocals(clause, locals)
		}
	}
}
