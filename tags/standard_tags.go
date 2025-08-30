// Package tags is an internal package that defines the standard Liquid tags.
package tags

import (
	"errors"
	"io"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

// AddStandardTags defines the standard Liquid tags.
func AddStandardTags(c *render.Config) {
	c.AddTag("assign", makeAssignTag(c))
	c.AddTag("include", includeTag)

	// blocks
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	c.AddTag("break", breakTag)
	c.AddTag("continue", continueTag)
	c.AddTag("cycle", cycleTag)
	c.AddBlock("capture").Compiler(captureTagCompiler)
	c.AddBlock("case").Clause("when").Clause("else").Compiler(caseTagCompiler)
	c.AddBlock("comment")
	c.AddBlock("for").Clause("else").Compiler(loopTagCompiler)
	c.AddBlock("if").Clause("else").Clause("elsif").Compiler(ifTagCompiler(true))
	c.AddBlock("raw")
	c.AddBlock("tablerow").Compiler(loopTagCompiler)
	c.AddBlock("unless").Clause("else").Compiler(ifTagCompiler(false))
}

func makeAssignTag(cfg *render.Config) func(string) (func(io.Writer, render.Context) error, error) {
	return func(source string) (func(io.Writer, render.Context) error, error) {
		stmt, err := expressions.ParseStatement(expressions.AssignStatementSelector, source)
		if err != nil {
			return nil, err
		}

		// Check if dot notation is used without Jekyll extensions enabled
		if len(stmt.Assignment.Path) > 1 && !cfg.JekyllExtensions {
			return nil, errors.New("syntax error: dot notation in assign tag (e.g., 'obj.property = value') requires Jekyll extensions to be enabled")
		}

		return func(w io.Writer, ctx render.Context) error {
			value, err := ctx.Evaluate(stmt.Assignment.ValueFn)
			if err != nil {
				return err
			}

			// Use Path if available (dot notation), otherwise fall back to Variable (simple assignment)
			if len(stmt.Assignment.Path) > 1 {
				return ctx.SetPath(stmt.Assignment.Path, value)
			}

			// Simple assignment (backward compatibility and standard mode)
			ctx.Set(stmt.Assignment.Variable, value)

			return nil
		}, nil
	}
}

func captureTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	// TODO verify syntax
	varname := node.Args

	return func(w io.Writer, ctx render.Context) error {
		s, err := ctx.InnerString()
		if err != nil {
			return err
		}

		ctx.Set(varname, s)

		return nil
	}, nil
}
