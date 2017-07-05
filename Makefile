SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LIB = liquid
PACKAGE = github.com/osteele/liquid
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_TIME = `date +%FT%T%z`

VERSION=0.0.0

LDFLAGS=

.DEFAULT_GOAL: ci
.PHONY: build clean ci command dependencies setup install lint test help

ci: setup test #lint

$(LIB): $(SOURCES)
	go build ${LDFLAGS} -o ${LIB} ${PACKAGE}

build: $(LIB) ## compile the package

clean: ## remove binary files
	rm -f ${LIB} ${CMD}

deps: ## list dependencies
	go list -f '{{join .Imports "\n"}}' ./... | grep -v ${PACKAGE} | grep '\.' | sort | uniq

install-dev-tools: ## install dependencies and development tools
	go get github.com/alecthomas/gometalinter
	go get golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/goyacc
	gometalinter --install

install: $(SOURCES)  ## install the liquid command-line tool
	go install ${LDFLAGS} ${PACKAGE}/cmd/liquid

lint: ## lint the package
	gometalinter ./... --deadline=5m --exclude expression/scanner.go --exclude y.go --exclude '.*_string.go' --disable=gotype
	@echo lint passed

test: ## test the package
	go test ./...

# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
