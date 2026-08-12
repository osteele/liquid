# Contributing

Bug reports, focused test cases, documentation fixes, and code contributions
are welcome. Before starting a large change, open an issue to confirm the
approach.

Review the
[pull request template](https://github.com/osteele/liquid/blob/main/.github/PULL_REQUEST_TEMPLATE.md)
before writing the final description.

## Set up the repository

Install [Go](https://go.dev/doc/install), then fork and clone the repository.
Download the pinned tools and module dependencies:

```bash
make tools
make deps
```

The lint target uses the `golangci-lint` version pinned in `go.mod`. You do not
need a global installation.

Optional pre-commit hooks run formatting, lint, tests, and basic repository
checks:

```bash
make install-hooks
```

Run all hooks manually with `make run-hooks`. Update the hook definitions with
`make update-hooks`.

## Develop and test

Run the standard pre-commit checks before opening a pull request:

```bash
make pre-commit
```

Useful targets include:

```bash
make test         # Run all tests
make test-short   # Run short tests
make coverage     # Write coverage.out and print package coverage
make benchmark    # Run benchmarks
make fmt          # Format Go source
make lint         # Run golangci-lint
make lint-fix     # Apply supported lint fixes
make vet          # Run go vet
make build        # Build the command
make ci           # Run the local CI sequence
```

Use `make help` to list every target.

Do not suppress a lint finding without a specific reason. Existing
`nolint:gocyclo` directives cover generated functions, hand-written parsers,
and generic interpreter functions where the complexity metric is not useful.

## Manage dependencies

```bash
make deps          # Download dependencies
make deps-update   # Update dependencies
make deps-list     # List dependencies
make mod-tidy      # Run go mod tidy
make mod-verify    # Verify downloaded modules
make check-mod     # Check whether go.mod and go.sum are current
```

## Generate the parser

The expression lexer uses Ragel, and the parser uses `goyacc`. Install Ragel
before editing `expressions/scanner.rl`. On macOS, run `brew install ragel`.

After changing `expressions/scanner.rl` or `expressions/expressions.y`,
regenerate the checked-in Go files:

```bash
make generate
```

The generation target downloads the pinned Go tools and checks for Ragel.

## Preview API documentation

Run a local documentation server:

```bash
go install golang.org/x/tools/cmd/godoc@latest
godoc -http=:6060
```

Then open <http://localhost:6060/pkg/github.com/osteele/liquid/>.
