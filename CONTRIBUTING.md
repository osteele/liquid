# Contributing

Here's some ways to help:

* Select an item from the [issues list](https://github.com/osteele/liquid/issues)
* Search the sources for FIXME and TODO comments using `make list-todo`
* Improve the [code coverage](https://coveralls.io/github/osteele/liquid?branch=master) - run `make coverage` to see current coverage

Review the [pull request template](https://github.com/osteele/liquid/blob/master/.github/PULL_REQUEST_TEMPLATE.md) before you get too far along on coding.

A note on lint: `nolint: gocyclo` has been used to disable cyclomatic complexity checks on generated functions, hand-written parsers, and some of the generic interpreter functions. IMO this check isn't appropriate for those classes of functions. This isn't a license to disable cyclomatic complexity checks or lint in general.

## Cookbook

### Set up your machine

Fork and clone the repo.

[Install go](https://golang.org/doc/install#install). On macOS running Homebrew, `brew install go` is easier than the linked instructions.

Install package dependencies and development tools:

```bash
make tools  # Install code generation tools
make deps   # Download Go dependencies
```

[Install golangci-lint](https://golangci-lint.run/usage/install/#local-installation).
On macOS: `brew install golangci-lint`

#### Set up Git Hooks (Recommended)

This project uses pre-commit hooks to automatically run formatting, linting, and tests before commits and pushes:

```bash
make install-hooks  # Install pre-commit hooks
```

This will:
- Install pre-commit if not already installed
- Set up hooks to run automatically on `git commit` and `git push`
- Run formatting (`go fmt`)
- Run linting (`golangci-lint`)
- Run tests (`go test -short`)
- Check for common issues (trailing whitespace, large files, merge conflicts)

To test the hooks manually:
```bash
make run-hooks  # Run all hooks on all files
```

To update hooks to latest versions:
```bash
make update-hooks  # Update pre-commit hooks
```

### Development Workflow

Quick start for development:

```bash
make all         # Clean, lint, test, and build everything
make pre-commit  # Run formatter, linter, and tests before committing
```

### Testing

```bash
make test        # Run all tests
make test-short  # Run short tests only
make coverage    # Generate test coverage report (HTML)
make benchmark   # Run performance benchmarks
```

### Code Quality

```bash
make fmt         # Format code
make lint        # Run linter
make lint-fix    # Run linter with auto-fix
make vet         # Run go vet
```

### Building

```bash
make build       # Build the binary
make install     # Build and install to GOPATH/bin
make clean       # Remove build artifacts
```

### Dependencies

```bash
make deps        # Download dependencies
make deps-update # Update dependencies to latest versions
make deps-list   # List all dependencies
make mod-tidy    # Clean up go.mod and go.sum
make mod-verify  # Verify dependencies are correct
make check-mod   # Check if go.mod is up to date
```

### Utilities

```bash
make list-todo   # Find all TODO and FIXME comments
make list-imports # List all package imports
make ci          # Run full CI suite locally
make help        # Show all available commands
```

### Preview the Documentation

```bash
godoc -http=:6060
open http://localhost:6060/pkg/github.com/osteele/liquid/
```

### Work on the Expression Parser and Lexer

To work on the lexer, install Ragel. On macOS: `brew install ragel`.

The parser and lexer tools are installed via `make tools`, which installs:
- `goyacc` for parser generation
- `stringer` for string method generation

After editing `scanner.rl` or `expressions.y`:

```bash
make generate   # Re-generate lexers and parsers
```

Or directly:

```bash
go generate ./...
```

Test just the scanner:

```bash
cd expressions
ragel -Z scanner.rl && go test
```
