SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

LIBRARY = liquid
PACKAGE = github.com/osteele/liquid
COMMIT_HASH = `git rev-parse --short HEAD 2>/dev/null`
BUILD_TIME = `date +%FT%T%z`

VERSION=0.0.0

LDFLAGS=-ldflags "-X ${PACKAGE}.Version=${VERSION} -X ${PACKAGE}.BuildTime=${BUILD_TIME}"

.DEFAULT_GOAL: $(LIBRARY)
.PHONY: build clean dependencies setup install lint test help

$(LIBRARY): $(SOURCES)
	go build ${LDFLAGS} -o ${LIBRARY} ${PACKAGE}/cmd/liquid

build: $(LIBRARY) ## compile the package

clean: ## remove binary files
	rm -fI ${LIBRARY}

setup: ## install dependencies and development tools
	go get -t ./...
	go get github.com/alecthomas/gometalinter
	go get golang.org/x/tools/cmd/stringer
	go install golang.org/x/tools/cmd/goyacc
	gometalinter --install

lint: ## lint the package
	gometalinter ./... --exclude expressions/scanner.go --exclude expressions/scanner.go --exclude '.*_string.go'

test: ## test the package
	go test ./...

# Source: https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
