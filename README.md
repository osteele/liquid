# Go Liquid Template Parser

`goliquid` is a very early-stage Go implementation of the [Shopify Liquid template language](https://shopify.github.io/liquid), for use in [Gojekyll](https://github.com/osteele/gojekyll).

## Status
[![Build Status](https://travis-ci.org/osteele/liquid.svg?branch=master)](https://travis-ci.org/osteele/liquid)
[![Go Report Card](https://goreportcard.com/badge/github.com/osteele/liquid)](https://goreportcard.com/report/github.com/osteele/liquid)

- [ ] Basics
  - [x] Literals
    - [ ] String Escapes
  - [x] Variables
  - [ ] Operators
  - [x] Arrays
  - [ ] Whitespace Control
- [ ] Tags
  - [x] Comment
  - [ ] Control Flow
    - [x] if/else/elsif
    - [x] unless
    - [ ] case/when
  - [ ] Iteration
        - [x] for
            - [ ] limit, offset, range, reversed
        - [ ] break, continue
        - [ ] loop variables
        - [ ] tablerow
        - [ ] cycle
  - [ ] Raw
  - [ ] Variable
    - [x] Assign
    - [ ] Capture
- [ ] Filters

## Install

`go get -u github.com/osteele/goliquid`

## Contribute

### Setup

```bash
make setup
```

Install Ragel. On macOS: `brew install ragel`.

### Workflow

```bash
go generate ./...
go test ./...
```

Test just the scanner:

```bash
cd expressions
ragel -Z scanner.rl && go test
```

## Attribution

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f) was essential to understanding `go yacc`.

## License

MIT License
