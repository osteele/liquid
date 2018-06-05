SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LIB = liquid
PACKAGE = github.com/osteele/liquid
LDFLAGS=

.DEFAULT_GOAL: ci
.PHONY: ci clean coverage deps generate imports install lint pre-commit setup test help

clean: ## remove binary files
	rm -f ${LIB} ${CMD}

coverage: ## test the package, with coverage
	go test -cov ./...

deps: ## list dependencies
	@go list -f '{{join .Deps "\n"}}' ./... | grep -v `go list -f '{{.ImportPath}}'` | grep '\.' | sort | uniq

generate: ## re-generate lexers and parser
	go generate ./...

imports: ## list imports
	@go list -f '{{join .Imports "\n"}}' ./... | grep -v `go list -f '{{.ImportPath}}'` | grep '\.' | sort | uniq

lint: ## lint the package
	gometalinter ./... --tests --deadline=5m --include=gofmt --exclude expressions/scanner.go --exclude y.go --exclude '.*_string.go' --disable=gotype --disable=interfacer
	@echo lint passed

pre-commit: lint test ## lint and test the package

setup: ## install dependencies and development tools
	go get golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/goyacc
	go get -t ./...
	go get github.com/alecthomas/gometalinter
	gometalinter --install

test: ## test the package
	go test ./...

# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
