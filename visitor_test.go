package liquid

// Tests for Walk and ParseTree visitor API.
//
// Ported / adapted from:
//   - Ruby Liquid: test/unit/parse_tree_visitor_test.rb
//     (test_preserve_tree_structure and tree-structural assertions)

import (
	"testing"
)

// ── Walk: basic node-kind collection ─────────────────────────────────────────

func TestWalk_CollectsNodeKinds(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name     string
		template string
		wantText bool
		wantOut  bool
		wantTag  bool
		wantBlk  bool
	}{
		{
			name:     "text only",
			template: `hello`,
			wantText: true,
		},
		{
			name:     "output only",
			template: `{{ x }}`,
			wantOut:  true,
		},
		{
			name:     "simple tag",
			template: `{% assign x = 1 %}`,
			wantTag:  true,
		},
		{
			name:     "block tag",
			template: `{% if true %}yes{% endif %}`,
			wantBlk:  true,
			wantText: true,
		},
		{
			name:     "all kinds",
			template: `hello {{ x }} {% assign y = x %}{% if y %}ok{% endif %}`,
			wantText: true,
			wantOut:  true,
			wantTag:  true,
			wantBlk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := engine.ParseString(tt.template)
			if err != nil {
				t.Fatalf("ParseString: %v", err)
			}

			var kinds [4]bool // indexed by TemplateNodeKind
			tpl.Walk(func(node *TemplateNode) bool {
				kinds[node.Kind] = true
				return true
			})

			if tt.wantText && !kinds[TemplateNodeText] {
				t.Errorf("expected TemplateNodeText")
			}
			if tt.wantOut && !kinds[TemplateNodeOutput] {
				t.Errorf("expected TemplateNodeOutput")
			}
			if tt.wantTag && !kinds[TemplateNodeTag] {
				t.Errorf("expected TemplateNodeTag")
			}
			if tt.wantBlk && !kinds[TemplateNodeBlock] {
				t.Errorf("expected TemplateNodeBlock")
			}
		})
	}
}

// ── Walk: tag names ───────────────────────────────────────────────────────────

func TestWalk_TagNames(t *testing.T) {
	engine := NewEngine()

	collectTagNames := func(src string) []string {
		tpl, err := engine.ParseString(src)
		if err != nil {
			t.Fatalf("ParseString(%q): %v", src, err)
		}
		var names []string
		seen := map[string]bool{}
		tpl.Walk(func(node *TemplateNode) bool {
			if node.TagName != "" && !seen[node.TagName] {
				seen[node.TagName] = true
				names = append(names, node.TagName)
			}
			return true
		})
		return names
	}

	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "assign tag",
			template: `{% assign x = 1 %}`,
			want:     []string{"assign"},
		},
		{
			name:     "if block",
			template: `{% if true %}yes{% endif %}`,
			want:     []string{"if"},
		},
		{
			name:     "for block",
			template: `{% for item in list %}{{ item }}{% endfor %}`,
			want:     []string{"for"},
		},
		{
			name:     "multiple tag types",
			template: `{% assign x = 1 %}{% for i in list %}{% if x %}ok{% endif %}{% endfor %}`,
			want:     []string{"assign", "for", "if"},
		},
		{
			name:     "tablerow block",
			template: `{% tablerow item in items %}{{ item }}{% endtablerow %}`,
			want:     []string{"tablerow"},
		},
		{
			name:     "case block",
			template: `{% case status %}{% when "active" %}ok{% endcase %}`,
			// "when" is a clause sub-block of "case", so Walk reports both.
			want: []string{"case", "when"},
		},
		{
			name:     "capture block",
			template: `{% capture buf %}hello{% endcapture %}`,
			want:     []string{"capture"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectTagNames(tt.template)
			if !stringSliceSetEqual(got, tt.want) {
				t.Errorf("tag names: got %v, want %v", got, tt.want)
			}
		})
	}
}

// ── Walk: returning false skips children ─────────────────────────────────────

func TestWalk_SkipChildren(t *testing.T) {
	engine := NewEngine()

	// Template has a nested for-if structure. If we return false on "for",
	// the inner "if" should NOT be visited.
	src := `{% for item in list %}{% if item %}{{ item }}{% endif %}{% endfor %}`
	tpl, err := engine.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	var visited []string
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName != "" {
			visited = append(visited, node.TagName)
		}
		// Return false when we hit the "for" block — skip its children.
		return node.TagName != "for"
	})

	// "for" should be in visited, but "if" (which is inside for) should not.
	foundFor := false
	foundIf := false
	for _, name := range visited {
		switch name {
		case "for":
			foundFor = true
		case "if":
			foundIf = true
		}
	}
	if !foundFor {
		t.Error("expected 'for' to be visited")
	}
	if foundIf {
		t.Error("expected 'if' NOT to be visited when returning false on 'for'")
	}
}

// ── Walk: visits block clauses ────────────────────────────────────────────────

func TestWalk_VisitsElseClauses(t *testing.T) {
	engine := NewEngine()

	// The else clause body contains an output node. Walk should visit it because
	// clauses are children of the block.
	src := `{% if false %}{{ a }}{% else %}{{ b }}{% endif %}`
	tpl, err := engine.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	outputCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput {
			outputCount++
		}
		return true
	})

	// Both {{ a }} and {{ b }} should be visited.
	if outputCount != 2 {
		t.Errorf("expected 2 output nodes, got %d", outputCount)
	}
}

// ── Walk: source locations ────────────────────────────────────────────────────

func TestWalk_SourceLocations(t *testing.T) {
	engine := NewEngine()

	// Line numbers should be non-zero for non-empty templates.
	tpl, err := engine.ParseString("{{ x }}\n{% if y %}ok{% endif %}")
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	allZero := true
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput || node.Kind == TemplateNodeBlock {
			if node.Location.LineNo > 0 {
				allZero = false
			}
		}
		return true
	})

	if allZero {
		t.Error("expected at least one non-zero source location")
	}
}

// ── ParseTree: structure ──────────────────────────────────────────────────────

func TestParseTree_Root(t *testing.T) {
	engine := NewEngine()

	tpl, err := engine.ParseString(`hello {{ x }}`)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	root := tpl.ParseTree()
	if root == nil {
		t.Fatal("ParseTree() returned nil")
	}
	// Root is a nameless block (the template sequence).
	if root.Kind != TemplateNodeBlock {
		t.Errorf("root Kind = %v, want TemplateNodeBlock", root.Kind)
	}
	if root.TagName != "" {
		t.Errorf("root TagName = %q, want empty", root.TagName)
	}

	// Should have two children: text "hello " and output {{ x }}.
	if len(root.Children) != 2 {
		t.Errorf("root Children count = %d, want 2", len(root.Children))
	} else {
		if root.Children[0].Kind != TemplateNodeText {
			t.Errorf("children[0].Kind = %v, want TemplateNodeText", root.Children[0].Kind)
		}
		if root.Children[1].Kind != TemplateNodeOutput {
			t.Errorf("children[1].Kind = %v, want TemplateNodeOutput", root.Children[1].Kind)
		}
	}
}

func TestParseTree_BlockChildren(t *testing.T) {
	engine := NewEngine()

	// if block should have body children and the else clause as a child.
	tpl, err := engine.ParseString(`{% if x %}{{ a }}{% else %}{{ b }}{% endif %}`)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	root := tpl.ParseTree()
	if root == nil || len(root.Children) == 0 {
		t.Fatal("expected non-empty root children")
	}

	// First child should be the "if" block.
	ifNode := root.Children[0]
	if ifNode.Kind != TemplateNodeBlock || ifNode.TagName != "if" {
		t.Fatalf("expected if block, got Kind=%v TagName=%q", ifNode.Kind, ifNode.TagName)
	}

	// The if block should have children: {{ a }} and the else clause block.
	if len(ifNode.Children) < 2 {
		t.Fatalf("if block should have ≥2 children, got %d", len(ifNode.Children))
	}

	// First body child: {{ a }}
	if ifNode.Children[0].Kind != TemplateNodeOutput {
		t.Errorf("expected output node as first child, got %v", ifNode.Children[0].Kind)
	}

	// Last child should be the else clause block.
	elseNode := ifNode.Children[len(ifNode.Children)-1]
	if elseNode.Kind != TemplateNodeBlock {
		t.Errorf("expected block node for else clause, got %v", elseNode.Kind)
	}
}

func TestParseTree_ForBlock(t *testing.T) {
	engine := NewEngine()

	tpl, err := engine.ParseString(`{% for item in list %}{{ item }}{% endfor %}`)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	root := tpl.ParseTree()
	if root == nil || len(root.Children) == 0 {
		t.Fatal("expected non-empty root children")
	}

	forNode := root.Children[0]
	if forNode.Kind != TemplateNodeBlock || forNode.TagName != "for" {
		t.Fatalf("expected for block, got Kind=%v TagName=%q", forNode.Kind, forNode.TagName)
	}
	// Body should contain {{ item }}.
	if len(forNode.Children) == 0 {
		t.Error("expected for block to have children")
	}
	if forNode.Children[0].Kind != TemplateNodeOutput {
		t.Errorf("expected output node inside for, got %v", forNode.Children[0].Kind)
	}
}

// ── Ported from Ruby: test_preserve_tree_structure ──────────────────────────
//
// Ruby original:
//   def test_preserve_tree_structure
//     assert_equal(
//       [[nil, [
//         [nil, [[nil, [["other", []]]]]],
//         ["test", []],
//         ["xs", []],
//       ]]],
//       traversal(%({% for x in xs offset: test %}{{ other }}{% endfor %})).visit,
//     )
//   end
//
// The Ruby test registers a callback only for VariableLookup nodes that returns
// node.name. In Go we use Walk to collect output-node presence (analogous) and
// verify tree shape via ParseTree.

func TestWalk_PreserveTreeStructure(t *testing.T) {
	engine := NewEngine()

	// {% for x in xs offset: test %}{{ other }}{% endfor %}
	// Expected structure:
	//   for block
	//     └─ output  ({{ other }})
	src := `{% for x in xs offset: test %}{{ other }}{% endfor %}`
	tpl, err := engine.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	root := tpl.ParseTree()
	if root == nil || len(root.Children) == 0 {
		t.Fatal("expected non-empty root")
	}

	forNode := root.Children[0]
	if forNode.Kind != TemplateNodeBlock || forNode.TagName != "for" {
		t.Fatalf("expected for block, got Kind=%v TagName=%q", forNode.Kind, forNode.TagName)
	}

	// The body of the for loop must contain exactly one output node ({{ other }}).
	outputCount := 0
	for _, child := range forNode.Children {
		if child.Kind == TemplateNodeOutput {
			outputCount++
		}
	}
	if outputCount != 1 {
		t.Errorf("expected 1 output node inside for, got %d", outputCount)
	}
}

// ── Walk: nested blocks ───────────────────────────────────────────────────────

func TestWalk_NestedBlocks(t *testing.T) {
	engine := NewEngine()

	src := `{% for a in list_a %}{% for b in list_b %}{{ a }}{{ b }}{% endfor %}{% endfor %}`
	tpl, err := engine.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	depth := 0
	maxDepth := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeBlock {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return true
	})
	// Walk is pre-order — depth counter won't naturally decrement.
	// Instead just verify that we visited two "for" blocks.
	forCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.TagName == "for" {
			forCount++
		}
		return true
	})
	if forCount != 2 {
		t.Errorf("expected 2 for blocks, got %d", forCount)
	}
}

// ── Walk: empty template ──────────────────────────────────────────────────────

func TestWalk_EmptyTemplate(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(``)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	visited := 0
	tpl.Walk(func(node *TemplateNode) bool {
		visited++
		return true
	})
	if visited != 0 {
		t.Errorf("empty template should yield 0 nodes, got %d", visited)
	}
}

// ── ParseTree: tag-only template ─────────────────────────────────────────────

func TestParseTree_TagOnly(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{% assign x = 1 %}`)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	root := tpl.ParseTree()
	if root == nil {
		t.Fatal("ParseTree() returned nil")
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
	child := root.Children[0]
	if child.Kind != TemplateNodeTag {
		t.Errorf("expected TemplateNodeTag, got %v", child.Kind)
	}
	if child.TagName != "assign" {
		t.Errorf("expected TagName 'assign', got %q", child.TagName)
	}
}

// ── Walk: unless block ────────────────────────────────────────────────────────

func TestWalk_UnlessBlock(t *testing.T) {
	engine := NewEngine()
	tpl, err := engine.ParseString(`{% unless flag %}{{ val }}{% endunless %}`)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	var tags []string
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeBlock {
			tags = append(tags, node.TagName)
		}
		return true
	})
	if len(tags) == 0 || tags[0] != "unless" {
		t.Errorf("expected 'unless' block, got %v", tags)
	}
}

// ── Walk: case/when block ─────────────────────────────────────────────────────

func TestWalk_CaseBlock(t *testing.T) {
	engine := NewEngine()
	src := `{% case status %}{% when "active" %}{{ active_msg }}{% else %}{{ default_msg }}{% endcase %}`
	tpl, err := engine.ParseString(src)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}

	outputCount := 0
	tpl.Walk(func(node *TemplateNode) bool {
		if node.Kind == TemplateNodeOutput {
			outputCount++
		}
		return true
	})
	if outputCount != 2 {
		t.Errorf("expected 2 output nodes in case, got %d", outputCount)
	}
}
