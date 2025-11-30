//go:build tools

// Package tools tracks tool dependencies for the project.
// These imports ensure that tool dependencies are tracked in go.mod.
// Install with: go install -tags tools ./...
package tools

import (
	// Code generation tools
	_ "golang.org/x/tools/cmd/goyacc"
	_ "golang.org/x/tools/cmd/stringer"
	// Linting and formatting - commented out due to compatibility issues
	// _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	// Note: golangci-lint v1.x has compatibility issues with Go 1.23
	// Install globally with: brew install golangci-lint
	// or: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
)
