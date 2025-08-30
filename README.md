# Liquid Template Parser

[![go badge][go-svg]][go-url]
[![Golangci-lint badge][golangci-lint-svg]][golangci-lint-url]
[![Go Report Card badge][go-report-card-svg]][go-report-card-url]
[![Go Doc][godoc-svg]][godoc-url]
[![MIT License][license-svg]][license-url]

`liquid` is a pure Go implementation of [Shopify Liquid
templates](https://shopify.github.io/liquid). It was developed for use in the
[Gojekyll](https://github.com/osteele/gojekyll) port of the Jekyll static site
generator.

<!-- TOC -->

- [Liquid Template Parser](#liquid-template-parser)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Command-Line tool](#command-line-tool)
  - [Documentation](#documentation)
    - [Status](#status)
    - [Drops](#drops)
    - [Value Types](#value-types)
    - [Template Store](#template-store)
    - [References](#references)
  - [Contributing](#contributing)
    - [Contributors](#contributors)
    - [Attribution](#attribution)
  - [Other Implementations](#other-implementations)
    - [Go](#go)
    - [Other Languages](#other-languages)
  - [License](#license)

<!-- /TOC -->

## Installation

`go get gopkg.in/osteele/liquid.v1` # latest snapshot

`go get -u github.com/osteele/liquid` # development version

## Usage

```go
engine := liquid.NewEngine()
template := `<h1>{{ page.title }}</h1>`
bindings := map[string]any{
    "page": map[string]string{
        "title": "Introduction",
    },
}
out, err := engine.ParseAndRenderString(template, bindings)
if err != nil { log.Fatalln(err) }
fmt.Println(out)
// Output: <h1>Introduction</h1>
```

See the [API documentation][godoc-url] for additional examples.

### Jekyll Compatibility

This library was originally developed for [Gojekyll](https://github.com/osteele/gojekyll), a Go port of Jekyll. 
As such, it includes optional Jekyll-specific extensions that are not part of the Shopify Liquid specification.

To enable Jekyll compatibility mode:

```go
engine := liquid.NewEngine()
engine.EnableJekyllExtensions()
```

Jekyll extensions include:

- **Dot notation in assign tags**: `{% assign page.canonical_url = "/about/" %}`
  - In standard Liquid, this would be a syntax error
  - With Jekyll extensions enabled, this creates or updates nested object properties
  - Intermediate objects are created automatically if they don't exist

Example:

```go
engine := liquid.NewEngine()
engine.EnableJekyllExtensions()  // Enable Jekyll-specific features

template := `{% assign page.meta.author = "John Doe" %}{{ page.meta.author }}`
bindings := map[string]any{
    "page": map[string]any{
        "title": "Home",
    },
}
out, _ := engine.ParseAndRenderString(template, bindings)
// Output: John Doe
```

**Note**: Jekyll extensions are disabled by default to maintain compatibility with standard Shopify Liquid.

### Command-Line tool

`go install gopkg.in/osteele/liquid.v0/cmd/liquid` installs a command-line
`liquid` executable. This is intended to make it easier to create test cases for
bug reports.

```bash
$ liquid --help
usage: liquid [FILE]
$ echo '{{ "Hello World" | downcase | split: " " | first | append: "!"}}' | liquid
hello!
```

## Documentation

### Status

These features of Shopify Liquid aren't implemented:

- Filter keyword parameters, for example `{{ image | img_url: '580x', scale: 2
  }}`. [[Issue #42](https://github.com/osteele/liquid/issues/42)]
- Warn and lax [error modes](https://github.com/shopify/liquid#error-modes).
- Non-strict filters. An undefined filter is currently an error.

### Drops

Drops have a different design from the Shopify (Ruby) implementation. A Ruby
drop sets `liquid_attributes` to a list of attributes that are exposed to
Liquid. A Go drop implements `ToLiquid() any`, that returns a proxy
object. Conventionally, the proxy is a `map` or `struct` that defines the
exposed properties. See <http://godoc.org/github.com/osteele/liquid#Drop> for
additional information.

### Value Types

`Render` and friends take a `Bindings` parameter. This is a map of `string` to
`any`, that associates template variable names with Go values.

Any Go value can be used as a variable value. These values have special meaning:

- `false` and `nil`
  - These, and no other values, are recognized as false by `and`, `or`, `{% if
    %}`, `{% elsif %}`, and `{% case %}`.
- Integers
  - (Only) integers can be used as array indices: `array[1]`; `array[n]`, where
    `array` has an array value and `n` has an integer value.
  - (Only) integers can be used as the endpoints of a range: `{% for item in
    (1..5) %}`, `{% for item in (start..end) %}` where `start` and `end` have
    integer values.
- Integers and floats
  - Integers and floats are converted to their join type for comparison: `1 ==
    1.0` evaluates to `true`.  Similarly, `int8(1)`, `int16(1)`, `uint8(1)` etc.
    are all `==`.
  - [There is currently no special treatment of complex numbers.]
- Integers, floats, and strings
  - Integers, floats, and strings can be used in comparisons `<`, `>`, `<=`,
    `>=`. Integers and floats can be usefully compared with each other. Strings
    can be usefully compared with each other, but not with other values. Any
    other comparison, e.g. `1 < "one"`, `1 > "one"`, is always false.
- Arrays (and slices)
  - An array can be indexed by integer value: `array[1]`; `array[n]` where `n`
    has an integer value.
  - Arrays have `first`, `last`, and `size` properties: `array.first ==
    array[0]`, `array[array.size-1] == array.last` (where `array.size > 0`)
- Maps
  - A map can be indexed by a string: `hash["key"]`; `hash[s]` where `s` has a
    string value
  - A map can be accessed using property syntax `hash.key`
  - Maps have a special `size` property, that returns the size of the map.
- Drops
  - A value `value` of a type that implements the `Drop` interface acts as the
    value `value.ToLiquid()`. There is no guarantee about how many times
    `ToLiquid` will be called. [This is in contrast to Shopify Liquid, which
    both uses a different interface for drops, and makes stronger guarantees.]
- Structs
  - A public field of a struct can be accessed by its name: `value.FieldName`, `value["fieldName"]`.
    - A field tagged e.g. `liquid:”name”` is accessed as `value.name` instead.
    - If the value of the field is a function that takes no arguments and
      returns either one or two arguments, accessing it invokes the function,
      and the value of the property is its first return value.
    - If the second return value is non-nil, accessing the field panics instead.
  - A function defined on a struct can be accessed by function name e.g.
    `value.Func`, `value["Func"]`.
    - The same rules apply as to accessing a func-valued public field.
  - Note that despite being array- and map-like, structs do not have a special
    `value.size` property.
- `[]byte`
  - A value of type `[]byte` is rendered as the corresponding string, and
    presented as a string to filters that expect one. A `[]byte` is not
    (currently) equivalent to a `string` for all uses; for example, `a < b`, `a
    contains b`, `hash[b]` will not behave as expected where `a` or `b` is a
    `[]byte`.
- `MapSlice`
  - An instance of `yaml.MapSlice` acts as a map. It implements `m.key`,
    `m[key]`, and `m.size`.

### Template Store

The template store allows for usage of varying template storage implementations (embedded file system, database, service, etc).  In order to use:

1. Create a struct that implements TemplateStore
    ```go
    type TemplateStore interface {
	      ReadTemplate(templatename string) ([]byte, error)
    }
    ```
1. Register with the engine
    ```go
    engine.RegisterTemplateStore()
    ```

`FileTemplateStore` is the default mechanism for backwards compatibility.

Refer to [example](./docs/TemplateStoreExample.md) for an example implementation.

### References

- [Shopify.github.io/liquid](https://shopify.github.io/liquid)
- [Liquid for Designers](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers)
- [Liquid for Programmers](https://github.com/Shopify/liquid/wiki/Liquid-for-Programmers)
- [Help.shopify.com](https://help.shopify.com/themes/liquid)

## Contributing

Bug reports, test cases, and code contributions are more than welcome.
Please refer to the [contribution guidelines](./CONTRIBUTING.md).

### Contributors

Thanks goes to these wonderful people ([emoji key](https://github.com/kentcdodds/all-contributors#emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://osteele.com/"><img src="https://avatars2.githubusercontent.com/u/674?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Oliver Steele</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=osteele" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=osteele" title="Documentation">📖</a> <a href="#ideas-osteele" title="Ideas, Planning, & Feedback">🤔</a> <a href="#infra-osteele" title="Infrastructure (Hosting, Build-Tools, etc)">🚇</a> <a href="https://github.com/osteele/liquid/pulls?q=is%3Apr+reviewed-by%3Aosteele" title="Reviewed Pull Requests">👀</a> <a href="https://github.com/osteele/liquid/commits?author=osteele" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/thessem"><img src="https://avatars0.githubusercontent.com/u/973593?v=4?s=100" width="100px;" alt=""/><br /><sub><b>James Littlejohn</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=thessem" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=thessem" title="Documentation">📖</a> <a href="https://github.com/osteele/liquid/commits?author=thessem" title="Tests">⚠️</a></td>
    <td align="center"><a href="http://nosmileface.ru"><img src="https://avatars2.githubusercontent.com/u/12567?v=4?s=100" width="100px;" alt=""/><br /><sub><b>nsf</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=nsf" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=nsf" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://tobias.salzmann.berlin/"><img src="https://avatars.githubusercontent.com/u/796084?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Tobias Salzmann</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=Eun" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/bendoerr"><img src="https://avatars.githubusercontent.com/u/253068?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Ben Doerr</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=bendoerr" title="Code">💻</a></td>
    <td align="center"><a href="https://daniil.it/"><img src="https://avatars.githubusercontent.com/u/7339644?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Daniil Gentili</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=danog" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/carolynvs"><img src="https://avatars.githubusercontent.com/u/1368985?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Carolyn Van Slyck</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=carolynvs" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/kke"><img src="https://avatars.githubusercontent.com/u/224971?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Kimmo Lehto</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=kke" title="Code">💻</a></td>
    <td align="center"><a href="https://vito.io/"><img src="https://avatars.githubusercontent.com/u/77198?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Victor "Vito" Gama</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=heyvito" title="Code">💻</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the
[all-contributors](https://github.com/kentcdodds/all-contributors)
specification. Contributions of any kind welcome!

### Attribution

| Package                                             | Author          | Description          | License            |
|-----------------------------------------------------|-----------------|----------------------|--------------------|
| [Ragel](http://www.colm.net/open-source/ragel/)     | Adrian Thurston | scanning expressions | MIT                |
| [gopkg.in/yaml.v2](https://github.com/go-yaml/yaml) | Canonical       | MapSlice             | Apache License 2.0 |

Michael Hamrah's [Lexing with Ragel and Parsing with Yacc using
Go](https://medium.com/@mhamrah/lexing-with-ragel-and-parsing-with-yacc-using-go-81e50475f88f)
was essential to understanding `go yacc`.

The [original Liquid engine](https://shopify.github.io/liquid), of course, for
the design and documentation of the Liquid template language. Many of the tag
and filter test cases are taken directly from the Liquid documentation.

## Other Implementations

### Go

- [karlseguin/liquid](https://github.com/karlseguin/liquid) is a dormant
  implementation that inspired a lot of forks.
- [acstech/liquid](https://github.com/acstech/liquid) is a more active fork of
  Karl Seguin's implementation.
- [hownowstephen/go-liquid](https://github.com/hownowstephen/go-liquid)

### Other Languages

 See Shopify's [ports of Liquid to other environments](https://github.com/Shopify/liquid/wiki/Ports-of-Liquid-to-other-environments).

## License

MIT License

[coveralls-url]: https://coveralls.io/r/osteele/liquid?branch=master
[coveralls-svg]: https://img.shields.io/coveralls/osteele/liquid.svg?branch=master

[go-url]: https://github.com/osteele/liquid/actions?query=workflow%3A%22Build+Status%22
[go-svg]: https://github.com/osteele/liquid/actions/workflows/go.yml/badge.svg

[golangci-lint-url]: https://github.com/osteele/liquid/actions?query=workflow%3Lint
[golangci-lint-svg]: https://github.com/osteele/liquid/actions/workflows/golangci-lint.yml/badge.svg

[godoc-url]: https://godoc.org/github.com/osteele/liquid
[godoc-svg]: https://godoc.org/github.com/osteele/liquid?status.svg

[license-url]: https://github.com/osteele/liquid/blob/master/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg

[go-report-card-url]: https://goreportcard.com/report/github.com/osteele/liquid
[go-report-card-svg]: https://goreportcard.com/badge/github.com/osteele/liquid
