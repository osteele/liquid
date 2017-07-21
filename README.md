# Liquid Template Parser

 [![][travis-svg]][travis-url] [![][coveralls-svg]][coveralls-url] [![][go-report-card-svg]][go-report-card-url] [![][godoc-svg]][godoc-url] [![][license-svg]][license-url]

`liquid` is a Go implementation of [Shopify Liquid templates](https://shopify.github.io/liquid). It was developed for use in the [Gojekyll](https://github.com/osteele/gojekyll) static site generator.

<!-- TOC -->

- [Liquid Template Parser](#liquid-template-parser)
    - [Differences from Liquid](#differences-from-liquid)
    - [Stability](#stability)
    - [Install](#install)
    - [Usage](#usage)
        - [Command-Line tool](#command-line-tool)
    - [Contributing](#contributing)
    - [References](#references)
    - [Attribution](#attribution)
    - [Other Implementations](#other-implementations)
        - [Go](#go)
        - [Other Languages](#other-languages)
    - [License](#license)

<!-- /TOC -->

## Differences from Liquid

The [feature parity board](https://github.com/osteele/liquid/projects/1) lists differences from Liquid.

In brief, these aren't implemented:

- Warn and lax [error modes](https://github.com/shopify/liquid#error-modes).
- Non-strict filters. An undefined filter is currently an error.
- Strict variables. An undefined variable is not an error.

## Stability

This library is at an early stage of development.
It has been mostly used by its author.

Only the liquid package itself, and the sub-package types that are used in that top-level package, are guaranteed stable. For example, `render.Context` is documented as the parameter type for tag definitions; it therefore won't change incompatibly with minor versions.

## Install

`go get gopkg.in/osteele/liquid.v1` # latest snapshot

`go get -u github.com/osteele/liquid` # development version

## Usage

```go
engine := liquid.NewEngine()
template := `<h1>{{ page.title }}</h1>`
bindings := map[string]interface{}{
    "page": map[string]string{
        "title": "Introduction",
    },
}
out, err := engine.ParseAndRenderString(template, bindings)
if err != nil { log.Fatalln(err) }
fmt.Println(out)
// Output: <h1>Introduction</h1>
```

### Command-Line tool

`go install gopkg.in/osteele/liquid.v0/cmd/liquid` installs a command-line `liquid` executable.
This is intended to make it easier to create test cases for bug reports.

```bash
$ liquid --help
usage: liquid [FILE]
$ echo '{{ "Hello World" | downcase | split: " " | first | append: "!"}}' | liquid
hello!
```

## Contributing

Bug reports, test cases, and code contributions are more than welcome.
Please refer to the [contribution guidelines](./CONTRIBUTING.md).

## References

* [Shopify.github.io/liquid](https://shopify.github.io/liquid)
* [Liquid for Designers](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers)
* [Liquid for Programmers](https://github.com/Shopify/liquid/wiki/Liquid-for-Programmers)
* [Help.shopify.com](https://help.shopify.com/themes/liquid) goes into more detail, but includes features that aren't present in core Liquid as used by Jekyll.

## Attribution

| Package                                             | Author          | Description                             | License            |
|-----------------------------------------------------|-----------------|-----------------------------------------|--------------------|
| [Ragel](http://www.colm.net/open-source/ragel/)     | Adrian Thurston | scanning expressions                    | MIT                |
| [gopkg.in/yaml.v2](https://github.com/go-yaml/yaml) | Canonical       | YAML support (for printing parse trees) | Apache License 2.0 |

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f) was essential to understanding `go yacc`.

The [original Liquid engine](https://shopify.github.io/liquid), of course, for the design and documentation of the Liquid template language. Many of the tag and filter test cases are taken directly from the Liquid documentation.

## Other Implementations

### Go

* [karlseguin/liquid](https://github.com/karlseguin/liquid) is a dormant implementation that inspired a lot of forks.
* [acstech/liquid](https://github.com/acstech/liquid) is a more active fork of Karl Seguin's implementation.
* [hownowstephen/go-liquid](https://github.com/hownowstephen/go-liquid)

### Other Languages

 See Shopify's [ports of Liquid to other environments](https://github.com/Shopify/liquid/wiki/Ports-of-Liquid-to-other-environments).

## License

MIT License

[coveralls-url]: https://coveralls.io/r/osteele/liquid?branch=master
[coveralls-svg]: https://img.shields.io/coveralls/osteele/liquid.svg?branch=master

[godoc-url]: https://godoc.org/github.com/osteele/liquid
[godoc-svg]: https://godoc.org/github.com/osteele/liquid?status.svg

[license-url]: https://github.com/osteele/liquid/blob/master/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg

[go-report-card-url]: https://goreportcard.com/report/github.com/osteele/liquid
[go-report-card-svg]: https://goreportcard.com/badge/github.com/osteele/liquid

[travis-url]: https://travis-ci.org/osteele/liquid
[travis-svg]: https://img.shields.io/travis/osteele/liquid.svg?branch=master
