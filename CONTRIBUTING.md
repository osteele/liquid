# Contributing

Refer to the [(original) Liquid contribution guidelines](https://github.com/Shopify/liquid/blob/master/CONTRIBUTING.md).

In addition to those checklists, I also won't merge:

- [ ] Performance improvements that don't include a benchmark.
- [ ] Meager (<3%) performance improvements that increase code verbosity or complexity.

A caveat: The cyclomatic complexity checks on generated functions, hand-written parsers, and some of the generic interpreter functions, have been disabled (via `nolint: gocyclo`). IMO this check isn't appropriate for those classes of functions. This isn't a license to disable cyclomatic complexity or lint in general.

## Cookbook

### Set up your machine

Fork and clone the repo.

[Install go](https://golang.org/doc/install#install). On macOS running Homebrew, `brew install go` is easier than the linked instructions.

Install package dependencies and development tools:

* `make install-dev-tools`
* `go get -t ./...`

### Test and Lint

```bash
go test ./...
make lint
```

### Preview the Documentation

```bash
godoc -http=:6060
open http://localhost:6060/pkg/github.com/osteele/liquid/
```

### Work on the Expression Parser and Lexer

To work on the lexer, install Ragel. On macOS: `brew install ragel`.

Do this after editing `scanner.rl` or `expressions.y`:

```bash
go generate ./...
```

Test just the scanner:

```bash
cd expression
ragel -Z scanner.rl && go test
```
