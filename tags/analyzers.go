package tags

import (
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// makeAssignAnalyzer returns a TagAnalyzer for the assign tag.
// assign introduces a local variable (LocalScope) and references an expression (Arguments).
func makeAssignAnalyzer() render.TagAnalyzer {
	return func(args string) render.NodeAnalysis {
		stmt, err := expressions.ParseStatement(expressions.AssignStatementSelector, args)
		if err != nil {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{
			Arguments:  []expressions.Expression{stmt.Assignment.ValueFn},
			LocalScope: []string{stmt.Assignment.Variable},
		}
	}
}

// captureBlockAnalyzer handles {% capture varname %}…{% endcapture %}.
// The body is evaluated and its output stored in varname (LocalScope).
func captureBlockAnalyzer(node render.BlockNode) render.NodeAnalysis {
	varname := strings.TrimSpace(node.Args)
	if varname == "" {
		return render.NodeAnalysis{}
	}
	return render.NodeAnalysis{LocalScope: []string{varname}}
}

// ifBlockAnalyzer handles {% if expr %} and {% unless expr %}.
// All condition expressions (main + elsif) are in Arguments.
func ifBlockAnalyzer() render.BlockAnalyzer {
	return func(node render.BlockNode) render.NodeAnalysis {
		var exprs []expressions.Expression
		if expr, err := expressions.Parse(node.Args); err == nil {
			exprs = append(exprs, expr)
		}
		for _, clause := range node.Clauses {
			if clause.Name == "elsif" {
				if expr, err := expressions.Parse(clause.Args); err == nil {
					exprs = append(exprs, expr)
				}
			}
		}
		return render.NodeAnalysis{Arguments: exprs}
	}
}

// caseBlockAnalyzer handles {% case expr %}{% when val %}…{% endcase %}.
// The case expression and each when expression are all in Arguments.
func caseBlockAnalyzer(node render.BlockNode) render.NodeAnalysis {
	var exprs []expressions.Expression
	if expr, err := expressions.Parse(node.Args); err == nil {
		exprs = append(exprs, expr)
	}
	for _, clause := range node.Clauses {
		if clause.Name == "when" {
			stmt, err := expressions.ParseStatement(expressions.WhenStatementSelector, clause.Args)
			if err != nil {
				continue
			}
			exprs = append(exprs, stmt.When.Exprs...)
		}
	}
	return render.NodeAnalysis{Arguments: exprs}
}

// makeEchoAnalyzer returns a TagAnalyzer for the echo tag.
// echo evaluates and outputs an expression (like {{ expr }}), reporting the same variables.
func makeEchoAnalyzer() render.TagAnalyzer {
	return func(args string) render.NodeAnalysis {
		if strings.TrimSpace(args) == "" {
			return render.NodeAnalysis{}
		}
		expr, err := expressions.Parse(args)
		if err != nil {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{Arguments: []expressions.Expression{expr}}
	}
}

// makeIncrementAnalyzer returns a TagAnalyzer for the increment tag.
// Per LiquidJS spec, the counter name is treated as a locally-defined variable.
func makeIncrementAnalyzer() render.TagAnalyzer {
	return func(args string) render.NodeAnalysis {
		varname := strings.TrimSpace(args)
		if varname == "" {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{LocalScope: []string{varname}}
	}
}

// makeDecrementAnalyzer returns a TagAnalyzer for the decrement tag.
// Per LiquidJS spec, the counter name is treated as a locally-defined variable.
func makeDecrementAnalyzer() render.TagAnalyzer {
	return func(args string) render.NodeAnalysis {
		varname := strings.TrimSpace(args)
		if varname == "" {
			return render.NodeAnalysis{}
		}
		return render.NodeAnalysis{LocalScope: []string{varname}}
	}
}

// loopBlockAnalyzerFull handles {% for var in expr limit: lim offset: off %}
// and {% tablerow var in expr %}, reporting the collection expression and any
// limit/offset expressions as Arguments, and the loop variable as BlockScope.
func loopBlockAnalyzerFull(node render.BlockNode) render.NodeAnalysis {
	stmt, err := expressions.ParseStatement(expressions.LoopStatementSelector, node.Args)
	if err != nil {
		return render.NodeAnalysis{}
	}
	var args []expressions.Expression
	if stmt.Loop.Expr != nil {
		args = append(args, stmt.Loop.Expr)
	}
	if stmt.Loop.Limit != nil {
		args = append(args, stmt.Loop.Limit)
	}
	if stmt.Loop.Offset != nil {
		args = append(args, stmt.Loop.Offset)
	}
	return render.NodeAnalysis{
		Arguments:  args,
		BlockScope: []string{stmt.Loop.Variable},
	}
}

// makeIncludeAnalyzer returns a TagAnalyzer for the include tag.
// Reports variable references from the file expression, with/for arguments, and key-value pairs.
func makeIncludeAnalyzer() render.TagAnalyzer {
	return func(source string) render.NodeAnalysis {
		parsed, err := parseIncludeArgs(source)
		if err != nil {
			return render.NodeAnalysis{}
		}
		var exprs []expressions.Expression
		if parsed.fileExpr != nil {
			exprs = append(exprs, parsed.fileExpr)
		}
		if parsed.withExpr != nil {
			exprs = append(exprs, parsed.withExpr)
		}
		if parsed.forExpr != nil {
			exprs = append(exprs, parsed.forExpr)
		}
		for _, kv := range parsed.kvPairs {
			exprs = append(exprs, kv.valueExpr)
		}
		return render.NodeAnalysis{Arguments: exprs}
	}
}

// makeRenderAnalyzer returns a TagAnalyzer for the render tag.
// Reports variable references from with/for arguments and key-value pairs.
// The file name expression is included if it resolves to a variable.
func makeRenderAnalyzer() render.TagAnalyzer {
	return func(source string) render.NodeAnalysis {
		source = strings.TrimSpace(source)
		if source == "" {
			return render.NodeAnalysis{}
		}

		fileExprStr, rest, err := consumeFirstExpression(source)
		if err != nil {
			return render.NodeAnalysis{}
		}
		rest = strings.TrimSpace(rest)

		var exprs []expressions.Expression

		// Report file expression variables (e.g., if filename is a variable).
		if fileExpr, err2 := expressions.Parse(fileExprStr); err2 == nil {
			exprs = append(exprs, fileExpr)
		}

		// Check for 'for collection [as item]' syntax.
		if strings.HasPrefix(rest, "for ") {
			rest = strings.TrimSpace(rest[4:])
			forExprStr, afterFor, err2 := consumeFirstExpression(rest)
			if err2 == nil {
				if forExpr, err3 := expressions.Parse(forExprStr); err3 == nil {
					exprs = append(exprs, forExpr)
				}
			}
			rest = strings.TrimSpace(afterFor)
			// Skip 'as alias'
			if strings.HasPrefix(rest, "as ") {
				rest = strings.TrimSpace(rest[3:])
				aliasEnd := strings.IndexAny(rest, " ,\t")
				if aliasEnd < 0 {
					aliasEnd = len(rest)
				}
				rest = strings.TrimSpace(rest[aliasEnd:])
			}
			if strings.HasPrefix(rest, ",") {
				rest = strings.TrimSpace(rest[1:])
			}
		}

		// Use parseIncludeArgs to extract with/kvPairs from remaining source.
		synth := fileExprStr + " " + rest
		if parsed, err2 := parseIncludeArgs(synth); err2 == nil {
			if parsed.withExpr != nil {
				exprs = append(exprs, parsed.withExpr)
			}
			for _, kv := range parsed.kvPairs {
				exprs = append(exprs, kv.valueExpr)
			}
		}

		return render.NodeAnalysis{Arguments: exprs}
	}
}

// makeLiquidAnalyzer returns a TagAnalyzer for the liquid multi-line tag.
// It compiles the inner lines as a sub-template and stores the compiled nodes
// in NodeAnalysis.ChildNodes so that walkForVariables/collectLocals can recurse
// into them, giving accurate analysis of variables used inside {% liquid %} blocks.
func makeLiquidAnalyzer(cfg *render.Config) render.TagAnalyzer {
	return func(source string) render.NodeAnalysis {
		lines := strings.Split(source, "\n")

		var sb strings.Builder
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			sb.WriteString("{%")
			sb.WriteString(trimmed)
			sb.WriteString("%}")
		}

		templateStr := sb.String()
		if templateStr == "" {
			return render.NodeAnalysis{}
		}

		node, err := cfg.Compile(templateStr, parser.SourceLoc{})
		if err != nil {
			return render.NodeAnalysis{}
		}

		return render.NodeAnalysis{ChildNodes: []render.Node{node}}
	}
}
