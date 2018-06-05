# Contributing

Here's some ways to help:

* Select an item from the [issues list](https://github.com/osteele/liquid/issues)
* Search the sources for FIXME and TODO comments.
* Improve the [code coverage](https://coveralls.io/github/osteele/liquid?branch=master).

Review the [pull request template](https://github.com/osteele/liquid/blob/master/.github/PULL_REQUEST_TEMPLATE.md) before you get too far along on coding.

A note on lint: `nolint: gocyclo` has been used to disable cyclomatic complexity checks on generated functions, hand-written parsers, and some of the generic interpreter functions. IMO this check isn't appropriate for those classes of functions. This isn't a license to disable cyclomatic complexity checks or lint in general.

## Cookbook

### Set up your machine

Fork and clone the repo.

[Install go](https://golang.org/doc/install#install). On macOS running Homebrew, `brew install go` is easier than the linked instructions.

Install package dependencies and development tools:

* `make setup`
* `go get -t ./...`

### Test and Lint

```bash
make pre-commit
```

You can also do these individually:

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
