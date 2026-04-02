package render

import (
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
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

// VariableRef is a variable path paired with the source location where it is referenced.
type VariableRef struct {
	Path []string
	Loc  parser.SourceLoc
}

// AnalysisResult is the result of static analysis of a compiled template.
type AnalysisResult struct {
	// Globals contains variable paths that come from the outer scope (not defined
	// within the template itself via assign, capture, for, etc.).
	Globals [][]string
	// All contains all variable paths referenced in the template, including locals.
	All [][]string

	// GlobalRefs contains global variable references with source locations.
	GlobalRefs []VariableRef
	// AllRefs contains all variable references with source locations.
	AllRefs []VariableRef

	// Locals contains variable names defined within the template (assign, capture, for, etc.).
	Locals []string

	// Tags contains the unique tag names used in the template (e.g. "if", "for", "assign").
	Tags []string
}

// Analyze performs static analysis on a compiled template tree and returns
// the set of variable paths referenced by the template.
func Analyze(root Node) AnalysisResult {
	locals := map[string]bool{}
	var localList []string
	collectLocals(root, locals, &localList)

	collector := &analysisCollector{seen: map[string]bool{}}
	walkForVariables(root, collector)

	allRefs := collector.refs
	all := make([][]string, len(allRefs))
	for i, r := range allRefs {
		all[i] = r.Path
	}

	var globals [][]string
	var globalRefs []VariableRef
	for _, ref := range allRefs {
		if len(ref.Path) > 0 && !locals[ref.Path[0]] {
			globals = append(globals, ref.Path)
			globalRefs = append(globalRefs, ref)
		}
	}

	tagSeen := map[string]bool{}
	var tags []string
	walkForTags(root, tagSeen, &tags)

	return AnalysisResult{
		All:        all,
		Globals:    globals,
		AllRefs:    allRefs,
		GlobalRefs: globalRefs,
		Locals:     localList,
		Tags:       tags,
	}
}

// analysisCollector deduplicates variable paths across the full AST walk,
// preserving the source location of the first occurrence of each path.
type analysisCollector struct {
	refs []VariableRef
	seen map[string]bool
}

func (c *analysisCollector) addRef(path []string, loc parser.SourceLoc) {
	if len(path) == 0 {
		return
	}
	key := strings.Join(path, "\x00")
	if !c.seen[key] {
		c.seen[key] = true
		cp := make([]string, len(path))
		copy(cp, path)
		c.refs = append(c.refs, VariableRef{Path: cp, Loc: loc})
	}
}

func (c *analysisCollector) addFromExpr(expr expressions.Expression, loc parser.SourceLoc) {
	for _, path := range expr.Variables() {
		c.addRef(path, loc)
	}
}

// walkForVariables traverses the AST collecting all variable references with their locations.
func walkForVariables(node Node, collector *analysisCollector) {
	switch n := node.(type) {
	case *SeqNode:
		for _, child := range n.Children {
			walkForVariables(child, collector)
		}
	case *ObjectNode:
		collector.addFromExpr(n.GetExpr(), n.SourceLoc)
	case *TagNode:
		for _, expr := range n.Analysis.Arguments {
			collector.addFromExpr(expr, n.SourceLoc)
		}
	case *BlockNode:
		for _, expr := range n.Analysis.Arguments {
			collector.addFromExpr(expr, n.SourceLoc)
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
func collectLocals(node Node, locals map[string]bool, list *[]string) {
	addLocal := func(name string) {
		if !locals[name] {
			locals[name] = true
			*list = append(*list, name)
		}
	}
	switch n := node.(type) {
	case *SeqNode:
		for _, child := range n.Children {
			collectLocals(child, locals, list)
		}
	case *TagNode:
		for _, name := range n.Analysis.LocalScope {
			addLocal(name)
		}
	case *BlockNode:
		for _, name := range n.Analysis.LocalScope {
			addLocal(name)
		}
		for _, name := range n.Analysis.BlockScope {
			addLocal(name)
		}
		for _, child := range n.Body {
			collectLocals(child, locals, list)
		}
		for _, clause := range n.Clauses {
			collectLocals(clause, locals, list)
		}
	}
}

// walkForTags traverses the AST collecting unique tag names (e.g. "if", "for", "assign").
func walkForTags(node Node, seen map[string]bool, tags *[]string) {
	switch n := node.(type) {
	case *SeqNode:
		for _, child := range n.Children {
			walkForTags(child, seen, tags)
		}
	case *TagNode:
		if !seen[n.Name] {
			seen[n.Name] = true
			*tags = append(*tags, n.Name)
		}
	case *BlockNode:
		if !seen[n.Name] {
			seen[n.Name] = true
			*tags = append(*tags, n.Name)
		}
		for _, child := range n.Body {
			walkForTags(child, seen, tags)
		}
		for _, clause := range n.Clauses {
			walkForTags(clause, seen, tags)
		}
	}
}
