package liquid

import (
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// TemplateNodeKind identifies the kind of a node in the template parse tree.
type TemplateNodeKind int

const (
	// TemplateNodeText represents a literal text chunk rendered verbatim.
	TemplateNodeText TemplateNodeKind = iota
	// TemplateNodeOutput represents an {{ expr }} output expression.
	TemplateNodeOutput
	// TemplateNodeTag represents a simple {% tag %} with no body.
	TemplateNodeTag
	// TemplateNodeBlock represents a block tag with a body, e.g. {% if %}...{% endif %}.
	TemplateNodeBlock
)

// TemplateNode is a public representation of a single node in the template parse tree.
// It provides a lightweight, stable view over the internal AST for tree inspection
// and custom traversal.
//
// Clause nodes (e.g. elsif, else, when) appear as children of their containing block.
type TemplateNode struct {
	// Kind identifies the type of this node.
	Kind TemplateNodeKind
	// TagName is non-empty for Tag and Block nodes; it holds the tag name (e.g. "if", "for").
	// It is empty for Text and Output nodes.
	TagName string
	// Location is the source location of this node in the template source.
	Location parser.SourceLoc
	// Children contains child nodes for Block nodes (body nodes followed by clause nodes).
	// For Text and Output nodes, Children is nil.
	Children []*TemplateNode
}

// WalkFunc is a callback invoked for each TemplateNode during a tree walk.
// Returning false prevents descending into that node's children; returning true
// continues the traversal into children.
type WalkFunc func(node *TemplateNode) bool

// Walk traverses the template parse tree in depth-first, pre-order, calling fn
// for each node. If fn returns false for a given node, its children are skipped.
//
// Tags are visited in document order. Block clauses (e.g. elsif, else, for-else)
// appear as direct children of their enclosing block node.
func (t *Template) Walk(fn WalkFunc) {
	visitRenderNode(t.root, fn)
}

// ParseTree returns the root of the template's parse tree as a *TemplateNode with
// all Children populated. The returned tree is a snapshot that can be inspected
// independently of the live template.
//
// The root node is always of kind TemplateNodeBlock with an empty TagName; it
// represents the top-level sequence of the template.
func (t *Template) ParseTree() *TemplateNode {
	return buildParseTree(t.root)
}

// ── internal helpers ──────────────────────────────────────────────────────────

// visitRenderNode walks the internal render.Node tree, translating each node to a
// TemplateNode and calling fn. render.SeqNode is transparent (its children are
// visited directly).
func visitRenderNode(node render.Node, fn WalkFunc) {
	switch n := node.(type) {
	case *render.SeqNode:
		for _, child := range n.Children {
			visitRenderNode(child, fn)
		}
	case *render.TextNode:
		fn(&TemplateNode{Kind: TemplateNodeText, Location: n.SourceLocation()})
	case *render.ObjectNode:
		fn(&TemplateNode{Kind: TemplateNodeOutput, Location: n.SourceLocation()})
	case *render.TagNode:
		fn(&TemplateNode{Kind: TemplateNodeTag, TagName: n.Name, Location: n.SourceLocation()})
	case *render.BlockNode:
		tn := &TemplateNode{Kind: TemplateNodeBlock, TagName: n.Name, Location: n.SourceLocation()}
		if !fn(tn) {
			return
		}
		for _, child := range n.Body {
			visitRenderNode(child, fn)
		}
		for _, clause := range n.Clauses {
			visitRenderNode(clause, fn)
		}
	}
}

// buildParseTree constructs a TemplateNode tree from the internal render node tree.
func buildParseTree(node render.Node) *TemplateNode {
	switch n := node.(type) {
	case *render.SeqNode:
		// SeqNode is the template root — represent it as a nameless block.
		return &TemplateNode{
			Kind:     TemplateNodeBlock,
			TagName:  "",
			Children: collectParseTreeChildren(n.Children),
		}
	case *render.TextNode:
		return &TemplateNode{Kind: TemplateNodeText, Location: n.SourceLocation()}
	case *render.ObjectNode:
		return &TemplateNode{Kind: TemplateNodeOutput, Location: n.SourceLocation()}
	case *render.TagNode:
		return &TemplateNode{Kind: TemplateNodeTag, TagName: n.Name, Location: n.SourceLocation()}
	case *render.BlockNode:
		children := collectParseTreeChildren(n.Body)
		for _, clause := range n.Clauses {
			if clauseNode := buildParseTree(clause); clauseNode != nil {
				children = append(children, clauseNode)
			}
		}
		return &TemplateNode{
			Kind:     TemplateNodeBlock,
			TagName:  n.Name,
			Location: n.SourceLocation(),
			Children: children,
		}
	}
	return nil
}

// collectParseTreeChildren converts a slice of render.Node into a slice of *TemplateNode,
// flattening transparent SeqNode containers.
func collectParseTreeChildren(nodes []render.Node) []*TemplateNode {
	var result []*TemplateNode
	for _, child := range nodes {
		tn := buildParseTree(child)
		if tn == nil {
			continue
		}
		// Flatten nested SeqNode (nameless blocks from buildParseTree).
		if tn.Kind == TemplateNodeBlock && tn.TagName == "" {
			result = append(result, tn.Children...)
		} else {
			result = append(result, tn)
		}
	}
	return result
}
