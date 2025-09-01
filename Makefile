# Makefile for github.com/osteele/liquid
SHELL := /bin/bash

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOGENERATE := $(GOCMD) generate

# Binary names
BINARY_NAME := liquid

# Packages
PACKAGES := $(shell $(GOCMD) list ./... | grep -v /vendor/)

# Coverage
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Tools - installed via tools.go
TOOLS_DIR := $(shell $(GOCMD) env GOPATH)/bin
GOYACC := $(TOOLS_DIR)/goyacc
STRINGER := $(TOOLS_DIR)/stringer
GOLANGCI_LINT := golangci-lint  # Use system-installed version

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

.DEFAULT_GOAL: help

## Help
.PHONY: help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2 } /^##@/ { printf "\n${YELLOW}%s${NC}\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: all
all: clean lint test build ## Run all checks and build

.PHONY: build
build: ## Build the binary
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/liquid

.PHONY: clean
clean: ## Remove build artifacts and temporary files
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@find . -type f -name '*.test' -delete
	@find . -type f -name '*.out' -delete

##@ Code Generation

.PHONY: generate
generate: tools ## Generate code (parsers, string methods, etc.)
	@echo "Generating code..."
	$(GOGENERATE) ./...

##@ Testing

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -race -v ./...

.PHONY: test-short
test-short: ## Run short tests
	$(GOTEST) -short ./...

.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GOTEST) -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "Total coverage: " $$3}'

.PHONY: benchmark
benchmark: ## Run benchmarks
	$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

.PHONY: lint
lint: check-golangci-lint ## Run linter
	@echo "Running linter..."
	$(GOLANGCI_LINT) run
	@echo "${GREEN}✓ Lint passed${NC}"

.PHONY: lint-fix
lint-fix: check-golangci-lint ## Run linter with auto-fix
	$(GOLANGCI_LINT) run --fix

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@find . -name "*.go" -not -path "./vendor/*" -not -name "scanner.go" -not -name "y.go" | xargs gofmt -w
	@echo "${GREEN}✓ Formatting complete${NC}"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "${GREEN}✓ Vet passed${NC}"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies to latest versions
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

.PHONY: deps-list
deps-list: ## List all dependencies
	@$(GOCMD) list -m all

.PHONY: mod-tidy
mod-tidy: ## Run go mod tidy
	$(GOMOD) tidy

.PHONY: mod-verify
mod-verify: ## Verify dependencies
	$(GOMOD) verify

##@ Tools

.PHONY: tools
tools: ## Install development tools
	@echo "Installing development tools..."
	@$(GOCMD) install golang.org/x/tools/cmd/goyacc@latest
	@$(GOCMD) install golang.org/x/tools/cmd/stringer@latest
	@echo "${GREEN}✓ Tools installed${NC}"
	@echo ""
	@echo "${YELLOW}Note: golangci-lint should be installed separately:${NC}"
	@echo "  brew install golangci-lint"
	@echo "  or"
	@echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s"

.PHONY: check-golangci-lint
check-golangci-lint:
	@which $(GOLANGCI_LINT) > /dev/null 2>&1 || (echo "${RED}Error: golangci-lint is not installed${NC}" && echo "Run 'make tools' for installation instructions" && exit 1)

##@ CI/CD

.PHONY: ci
ci: deps lint vet test ## Run CI checks

.PHONY: pre-commit
pre-commit: fmt lint test ## Run pre-commit checks

##@ Utilities

.PHONY: list-imports
list-imports: ## List all imports
	@$(GOCMD) list -f '{{join .Imports "\n"}}' ./... | grep -v 'github.com/osteele/liquid' | sort | uniq

.PHONY: list-todo
list-todo: ## List all TODO and FIXME comments
	@grep -r --include="*.go" -E "TODO|FIXME" . | grep -v vendor || echo "No TODOs or FIXMEs found"

.PHONY: check-mod
check-mod: ## Check if go.mod is up to date
	@$(GOCMD) mod tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "${RED}go.mod or go.sum needs updating. Run 'make mod-tidy'${NC}"; \
		exit 1; \
	fi
	@echo "${GREEN}✓ go.mod is up to date${NC}"

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@$(GOCMD) install ./cmd/liquid
	@echo "${GREEN}✓ Installed to $$($(GOCMD) env GOPATH)/bin/$(BINARY_NAME)${NC}"