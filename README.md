# Go Liquid Template Parser

`goliquid` is a Go implementation of the [Shopify Liquid template language](https://shopify.github.io/liquid/tags/variable/), for use in [Gojekyll](https://github.com/osteele/gojekyll).

## Status

- [ ] Basics
  - [ ] Constants
  - [x] Variables
  - [ ] Operators
  - [ ] Arrays
  - [ ] Whitespace Control
- [ ] Tags
  - [ ] Comment
  - [ ] Control Flow
    - [x] if/else/elsif
    - [x] unless
    - [ ] case/when
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
go install golang.org/x/tools/cmd/goyacc
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

## Attribution

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f) was essential to understanding `go yacc`.
