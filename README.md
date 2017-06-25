# Go Liquid Template Parser

`goliquid` is a Go implementation of the [Shopify Liquid template language](https://shopify.github.io/liquid/tags/variable/), for use in [Gojekyll](https://github.com/osteele/gojekyll).

## Status

- [ ] Basics
  - [ ] Constants
  - [ ] Variables
  - [ ] Operators
  - [ ] Arrays
  - [ ] Whitespace Control
- [ ] Tags
  - [ ] Comment
  - [ ] Control Flow
  - [ ] Iteration
        - [ ] for
            - [ ] limit, offset, range, reversed
        - [ ] break, continue
        - [ ] loop variables
        - [ ] tablerow
        - [ ] cycle
  - [ ] Raw
  - [ ] Variable
    - [ ] Assign
    - [ ] Capture
- [ ] Filters

## Install

`go get -u github.com/osteele/goliquid`

## Contribute

### Setup

```bash
go get golang.org/x/tools/cmd/stringer
```

Install Ragel. On macOS: `brew install ragel`.

### Workflow

```bash
go generate
go test
```

Test just the scanner:

```bash
ragel -Z scanner.rl && go test -run TestExpressionParser
```