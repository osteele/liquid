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
    - [x] `if`/`else`/`elsif`
    - [x] `unless`
    - [ ] `case`/`when`
  - [ ] Iteration
      - [x] modifiers
          - [ ] `limit`
          - [ ] `offset`
          - [ ] `range`
          - [x] `reversed`
      - [ ] `break`, `continue`
      - [ ] loop variables
      - [ ] `tablerow`
      - [ ] `cycle`
  - [x] Raw
  - [ ] Variable
    - [x] Assign
    - [ ] Capture
- [ ] Filters
  - [x] some
  - [ ] all

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

## References

<https://shopify.github.io/liquid>

<https://help.shopify.com/themes/liquid>

<https://github.com/Shopify/liquid/wiki/Liquid-for-Designers>


## Attribution

Kyoung-chan Lee's <https://github.com/leekchan/timeutil> for formatting dates.

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f) was essential to understanding `go yacc`.

The [original Liquid engine](https://shopify.github.io/liquid), of course, for the design and documentation of the Liquid template language.

(That said, this is a clean-room implementation to make sure it just implements the documented design.)

## License

MIT License
