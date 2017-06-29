# Go Liquid Template Parser
[![GoDoc](https://godoc.org/github.com/osteele/liquid?status.svg)](http://godoc.org/github.com/osteele/liquid)

`goliquid` is a pure Go implementation of [Shopify Liquid templates](https://shopify.github.io/liquid), for use in [gojekyll](https://github.com/osteele/gojekyll).

## Status
[![Build Status](https://travis-ci.org/osteele/liquid.svg?branch=master)](https://travis-ci.org/osteele/liquid)
[![Go Report Card](https://goreportcard.com/badge/github.com/osteele/liquid)](https://goreportcard.com/report/github.com/osteele/liquid)

- [ ] Basics
  - [x] Literals
    - [ ] String Escapes
  - [x] Variables
  - [ ] Operators (partial)
  - [x] Arrays
  - [ ] Whitespace Control
- [ ] Tags
  - [x] Comment
  - [ ] Control Flow
    - [x] `if`/`else`/`elsif`
    - [x] `unless`
    - [ ] `case`
      - [x] `when`
      - [ ] `else`
  - [ ] Iteration
      - [x] modifiers (`limit`, `reversed`, `offset`)
      - [ ] `range`
      - [ ] `break`, `continue`
      - [x] loop variables
      - [ ] `tablerow`
      - [ ] `cycle`
  - [x] Raw
  - [x] Variables
    - [x] Assign
    - [x] Capture
- [ ] Filters
  - [ ] `sort_natural`, `uniq`, `escape`, `truncatewords`, `url_decode`, `url_encode`
  - [x] everything else

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

The [original Liquid engine](https://shopify.github.io/liquid), of course, for the design and documentation of the Liquid template language. Many of the tag and filter test cases are taken directly from the Liquid documentation.

(That said, this is a clean-room implementation to make sure it just implements the documented design.)

## Other Implementations

Go:

* <https://godoc.org/github.com/karlseguin/liquid> is a partial implementation.
* <https://godoc.org/github.com/acstech/liquid> is a more active fork of Karl Seguin's implementation.
* <https://godoc.org/github.com/hownowstephen/go-liquid> is a more recent entry.

<https://github.com/Shopify/liquid/wiki/Ports-of-Liquid-to-other-environments> lists ports to other languages.

## License

MIT License
