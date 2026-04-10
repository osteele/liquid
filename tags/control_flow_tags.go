package tags

import (
	"io"

	e "github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/values"
)

type caseInterpreter interface {
	body() *render.BlockNode
	test(any, render.Context) (bool, error)
}
type exprCase struct {
	e.When

	b *render.BlockNode
}

func (c exprCase) body() *render.BlockNode { return c.b }

func (c exprCase) test(caseValue any, ctx render.Context) (bool, error) {
	for _, expr := range c.Exprs {
		whenValue, err := ctx.Evaluate(expr)
		if err != nil {
			return false, err
		}

		if values.Equal(caseValue, whenValue) {
			return true, nil
		}
	}

	return false, nil
}

type elseCase struct{ b *render.BlockNode }

func (c elseCase) body() *render.BlockNode { return c.b }

func (c elseCase) test(any, render.Context) (bool, error) { return true, nil }

func caseTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// TODO syntax error on non-empty node.Body
	expr, err := e.Parse(node.Args)
	if err != nil {
		return nil, err
	}

	cases := []caseInterpreter{}

	for _, clause := range node.Clauses {
		switch clause.Name {
		case "when":
			stmt, err := e.ParseStatement(e.WhenStatementSelector, clause.Args)
			if err != nil {
				return nil, err
			}

			cases = append(cases, exprCase{stmt.When, clause})
		default: // should be a check for "else", but I like the metacircularity
			cases = append(cases, elseCase{clause})
		}
	}

	return func(w io.Writer, ctx render.Context) error {
		sel, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}

		executedIdx := -1
		var renderErr error
		hooks := ctx.AuditHooks()

		var branchItems [][]render.AuditConditionNode
		if hooks != nil && hooks.OnCondition != nil {
			branchItems = make([][]render.AuditConditionNode, len(cases))
		}

		for i, clause := range cases {
			if branchItems != nil {
				hooks.SetConditionTarget(&branchItems[i])
				hooks.SetBranchSource(clause.body().Args)
			}
			if hooks != nil {
				hooks.SetCurrentLoc(clause.body().SourceLoc, clause.body().EndLoc, clause.body().Source)
			}

			b, err := clause.test(sel, ctx)

			if branchItems != nil {
				hooks.SetConditionTarget(nil)
			}

			if err != nil {
				return err
			}

			if b {
				executedIdx = i
				renderErr = ctx.RenderBlock(w, clause.body())
				break
			}
		}

		// Emit audit event for the case block.
		if hooks != nil && hooks.OnCondition != nil {
			auditBranches := make([]render.AuditBranch, len(cases))
			for i, clause := range cases {
				bn := clause.body()
				kind := bn.Name // "when" or "else"
				ab := render.AuditBranch{
					Kind:     kind,
					LocStart: bn.SourceLoc,
					LocEnd:   bn.EndLoc,
					Source:   bn.Source,
					Executed: i == executedIdx,
				}
				if branchItems != nil {
					ab.Items = branchItems[i]
				}
				auditBranches[i] = ab
			}
			hooks.OnCondition(node.SourceLoc, node.EndLoc, node.Source, auditBranches, hooks.Depth())
		}

		return renderErr
	}, nil
}

func ifTagCompiler(polarity bool) func(render.BlockNode) (func(io.Writer, render.Context) error, error) { //nolint: gocyclo
	return func(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
		type branchRec struct {
			test e.Expression
			body *render.BlockNode
		}

		expr, err := e.Parse(node.Args)
		if err != nil {
			return nil, err
		}

		if !polarity {
			expr = e.Not(expr)
		}

		branches := []branchRec{
			{expr, &node},
		}
		for _, c := range node.Clauses {
			test := e.Constant(true)

			switch c.Name {
			case "else":
			// TODO syntax error if this isn't the last branch
			case "elsif":
				t, err := e.Parse(c.Args)
				if err != nil {
					return nil, err
				}

				test = t
			}

			branches = append(branches, branchRec{test, c})
		}

		return func(w io.Writer, ctx render.Context) error {
			executedIdx := -1
			var renderErr error
			hooks := ctx.AuditHooks()

			// Per-branch comparison collections (populated by ComparisonHook via AuditHooks).
			var branchItems [][]render.AuditConditionNode
			if hooks != nil && hooks.OnCondition != nil {
				branchItems = make([][]render.AuditConditionNode, len(branches))
			}

			for i, b := range branches {
				// Arm the conditionTarget so the ComparisonHook writes into this branch's slice.
				if branchItems != nil {
					hooks.SetConditionTarget(&branchItems[i])
					hooks.SetBranchSource(b.body.Args)
				}
				// Always set current loc when audit is active (needed for type-mismatch diagnostic).
				if hooks != nil {
					hooks.SetCurrentLoc(b.body.SourceLoc, b.body.EndLoc, b.body.Source)
				}

				value, err := ctx.Evaluate(b.test)

				if branchItems != nil {
					hooks.SetConditionTarget(nil)
				}

				if err != nil {
					return err
				}

				if values.Truthy(value) {
					executedIdx = i
					renderErr = ctx.RenderBlock(w, b.body)
					break
				}
			}

			// Emit audit event for the condition block.
			if hooks != nil && hooks.OnCondition != nil {
				auditBranches := make([]render.AuditBranch, len(branches))
				for i, b := range branches {
					kind := b.body.Name // "if", "unless", "elsif", "else"
					ab := render.AuditBranch{
						Kind:     kind,
						LocStart: b.body.SourceLoc,
						LocEnd:   b.body.EndLoc,
						Source:   b.body.Source,
						Executed: i == executedIdx,
					}
					if branchItems != nil {
						ab.Items = branchItems[i]
					}
					auditBranches[i] = ab
				}
				hooks.OnCondition(node.SourceLoc, node.EndLoc, node.Source, auditBranches, hooks.Depth())
			}

			return renderErr
		}, nil
	}
}
