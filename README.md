# Go Liquid Template Parser
[![Build Status](https://travis-ci.org/osteele/liquid.svg?branch=master)](https://travis-ci.org/osteele/liquid)
[![Go Report Card](https://goreportcard.com/badge/github.com/osteele/liquid)](https://goreportcard.com/report/github.com/osteele/liquid)
[![GoDoc](https://godoc.org/github.com/osteele/liquid?status.svg)](http://godoc.org/github.com/osteele/liquid)

> "Any sufficiently complicated C or Fortran program contains an ad-hoc, informally-specified, bug-ridden, slow implementation of half of Common Lisp." – Philip Greenspun

`liquid` ports [Shopify Liquid templates](https://shopify.github.io/liquid) to Go. It was developed for use in [gojekyll](https://github.com/osteele/gojekyll).

`liquid` provides a functional API for defining tags and filters. See examples [here](https://github.com/osteele/liquid/blob/master/filters/filters.go), [here](https://github.com/osteele/gojekyll/blob/master/filters/filters.go), and [here](https://github.com/osteele/gojekyll/blob/master/tags/tags.go). On the one hand, this isn't idiomatic Go. On the other hand, this made it possible to implement a boatload of Liquid and Jekyll filters without an onerous amount of boilerplate – in some cases, simply by passing a reference to a function from the standard library.

<!-- TOC -->

- [Go Liquid Template Parser](#go-liquid-template-parser)
    - [Status](#status)
    - [Install](#install)
    - [Contribute](#contribute)
        - [Setup](#setup)
        - [Workflow](#workflow)
        - [Style Guide](#style-guide)
        - [Working on the Parser and Lexer](#working-on-the-parser-and-lexer)
    - [References](#references)
    - [Attribution](#attribution)
    - [Other Implementations](#other-implementations)
        - [Go](#go)
        - [Other Languages](#other-languages)
    - [License](#license)

<!-- /TOC -->

## Status

This library is in its early days. IMO it's not sufficiently mature to be worth snapping off a [versioned URL](http://labix.org/gopkg.in). In particular, the tag and filter extension API is likely to change.

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
    - [x] `case`
      - [x] `when`
      - [ ] `else`
  - [ ] Iteration
      - [x] modifiers (`limit`, `reversed`, `offset`)
      - [ ] `range`
      - [x] `break`, `continue`
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

### Workflow

```bash
go test ./...
```

Test just the scanner:

```bash
cd expressions
ragel -Z scanner.rl && go test
```

Preview the documentation:

```bash
godoc -http=:6060&
open http://localhost:6060/pkg/github.com/osteele/liquid/
```

### Style Guide

`make test` and `make lint` should pass.

The cyclomatic complexity checks on generated functions, hand-written parsers, and some of the generic interpreter functions, have been disabled (via `nolint: gocyclo`). IMO this check isn't appropriate for those classes of functions. This isn't a license to disable cyclomatic complexity or lint in general willy nilly.

### Working on the Parser and Lexer

To work on the lexer, install Ragel. On macOS: `brew install ragel`.

Do this after editing `scanner.rl` or `expressions.y`:

```bash
go generate ./...
```

## References

* [Shopify.github.io/liquid](https://shopify.github.io/liquid) is the definitive reference.
* [Help.shopify.com](https://help.shopify.com/themes/liquid) goes into more detail, but includes features that aren't present in core Liquid as used by Jekyll.
* Shopify's [Liquid for Designers](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers) is another take.


## Attribution

| Package                                               | Author          | Description                                  | License            |
|-------------------------------------------------------|-----------------|----------------------------------------------|--------------------|
| [gopkg.in/yaml.v2](https://github.com/go-yaml/yaml)   | Canonical       | YAML support (for printing parse trees)      | Apache License 2.0 |
| [jeffjen/datefmt](https://github.com/jeffjen/datefmt) | Jeffrey Jen     | Go bindings to GNU `strftime` and `strptime` | MIT                |
| [Ragel](http://www.colm.net/open-source/ragel/)       | Adrian Thurston | scanning expressions                         | MIT                |

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f) was essential to understanding `go yacc`.

The [original Liquid engine](https://shopify.github.io/liquid), of course, for the design and documentation of the Liquid template language. Many of the tag and filter test cases are taken directly from the Liquid documentation.

## Other Implementations

### Go

* [karlseguin/liquid](https://github.com/karlseguin/liquid) is a dormant implementation that inspired a lot of forks.
* [acstech/liquid](https://github.com/acstech/liquid) is a more active fork of Karl Seguin's implementation. I submitted a couple of pull requests there.
* [hownowstephen/go-liquid](https://github.com/hownowstephen/go-liquid) is a more recent entry.

After trying each of these, and looking at how to extend them, I concluded that I wasn't going to get very far without a parser generator. I also wanted an easy API for writing filters.

### Other Languages

 See Shopify's [ports of Liquid to other environments](https://github.com/Shopify/liquid/wiki/Ports-of-Liquid-to-other-environments).

## License

MIT License
