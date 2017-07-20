SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LIB = liquid
PACKAGE = github.com/osteele/liquid
LDFLAGS=

.DEFAULT_GOAL: ci
.PHONY: clean ci deps setup install lint test help

ci: setup test #lint

clean: ## remove binary files
	rm -f ${LIB} ${CMD}

deps: ## list dependencies
	@go list -f '{{join .Deps "\n"}}' ./... | grep -v `go list -f '{{.ImportPath}}'` | grep '\.' | sort | uniq

imports: ## list imports
	@go list -f '{{join .Imports "\n"}}' ./... | grep -v `go list -f '{{.ImportPath}}'` | grep '\.' | sort | uniq

generate:
	go generate ./...

setup: ## install dependencies and development tools
	go get golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/goyacc
	go get -t ./...
	go get github.com/alecthomas/gometalinter
	gometalinter --install

lint: ## lint the package
	gometalinter ./... --deadline=5m --include=gofmt --exclude expressions/scanner.go --exclude y.go --exclude '.*_string.go' --disable=gotype --disable=interfacer
	@echo lint passed

test: ## test the package
	go test ./...

# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
