//go:build tools

// Package tools documents the code-generation tools tracked by go.mod tool directives.
package tools

import (
	// Code generation tools
	_ "golang.org/x/tools/cmd/goyacc"
	_ "golang.org/x/tools/cmd/stringer"
)
