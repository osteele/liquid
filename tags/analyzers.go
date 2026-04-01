package tags

import (
	"strings"

	"github.com/osteele/liquid/expressions"
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

// loopBlockAnalyzer handles {% for var in expr %} and {% tablerow var in expr %}.
// The collection expression is in Arguments; the loop variable is in BlockScope.
func loopBlockAnalyzer(node render.BlockNode) render.NodeAnalysis {
	stmt, err := expressions.ParseStatement(expressions.LoopStatementSelector, node.Args)
	if err != nil {
		return render.NodeAnalysis{}
	}
	return render.NodeAnalysis{
		Arguments:  []expressions.Expression{stmt.Loop.Expr},
		BlockScope: []string{stmt.Loop.Variable},
	}
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
