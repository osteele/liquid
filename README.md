# Liquid Template Parser

 [![][travis-svg]][travis-url]
 [![][appveyor-svg]][appveyor-url]
 [![][coveralls-svg]][coveralls-url]
 [![][go-report-card-svg]][go-report-card-url]
 [![][godoc-svg]][godoc-url]
 [![][license-svg]][license-url]

`liquid` is a pure Go implementation of [Shopify Liquid templates](https://shopify.github.io/liquid).
It was developed for use in the [Gojekyll](https://github.com/osteele/gojekyll) port of the Jekyll static site generator.

<!-- TOC -->

- [Liquid Template Parser](#liquid-template-parser)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Command-Line tool](#command-line-tool)
  - [Documentation](#documentation)
    - [Status](#status)
    - [Drops](#drops)
    - [Value Types](#value-types)
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

See the [API documentation][godoc-url] for additional examples.

### Command-Line tool

`go install gopkg.in/osteele/liquid.v0/cmd/liquid` installs a command-line `liquid` executable.
This is intended to make it easier to create test cases for bug reports.

```bash
$ liquid --help
usage: liquid [FILE]
$ echo '{{ "Hello World" | downcase | split: " " | first | append: "!"}}' | liquid
hello!
```

## Documentation

### Status

These features of Shopify Liquid aren't implemented:

- Warn and lax [error modes](https://github.com/shopify/liquid#error-modes).
- Non-strict filters. An undefined filter is currently an error.
- Strict variables. An undefined variable is not an error.

### Drops

Drops have a different design from the Shopify (Ruby) implementation.
A Ruby drop sets `liquid_attributes` to a list of attributes that are exposed to Liquid.
A Go drop implements `ToLiquid() interface{}`, that returns a proxy object.
Conventionally, the proxy is a `map` or `struct` that defines the exposed properties.
See <http://godoc.org/github.com/osteele/liquid#Drop> for additional information.

### Value Types

`Render` and friends take a `Bindings` parameter. This is a map of `string` to `interface{}`, that associates template variable names with Go values.

Any Go value can be used as a variable value. These values have special meaning:

- `false` and `nil`
  - These, and no other values, are recognized as false by `and`, `or`, `{% if %}`, `{% elsif %}`, and `{% case %}`.
- Integers
  - (Only) integers can be used as array indices: `array[1]`; `array[n]`, where `array` has an array value and `n` has an integer value.
  - (Only) integers can be used as the endpoints of a range: `{% for item in (1..5) %}`, `{% for item in (start..end) %}` where `start` and `end` have integer values.
- Integers and floats
  - Integers and floats are converted to their join type for comparison: `1 == 1.0` evaluates to `true`.  Similarly, `int8(1)`, `int16(1)`, `uint8(1)` etc. are all `==`.
  - [There is currently no special treatment of complex numbers.]
- Integers, floats, and strings
  - Integers, floats, and strings can be used in comparisons `<`, `>`, `<=`, `>=`. Integers and floats can be usefully compared with each other. Strings can be usefully compared with each other, but not with other values. Any other comparison, e.g. `1 < "one"`, `1 > "one"`, is always false.
- Arrays (and slices)
  - An array can be indexed by integer value: `array[1]`; `array[n]` where `n` has an integer value.
  - Arrays have `first`, `last`, and `size` properties: `array.first == array[0]`, `array[array.size-1] == array.last` (where `array.size > 0`)
- Maps
  - A map can be indexed by a string: `hash["key"]`; `hash[s]` where `s` has a string value
  - A map can be accessed using property syntax `hash.key`
  - Maps have a special `size` property, that returns the size of the map.
- Drops
  - A value `value` of a type that implements the `Drop` interface acts as the value `value.ToLiquid()`. There is no guarantee about how many times `ToLiquid` will be called. [This is in contrast to Shopify Liquid, which both uses a different interface for drops, and makes stronger guarantees.]
- Structs
  - A public field of a struct can be accessed by its name: `value.FieldName`, `value["fieldName"]`.
    - A field tagged e.g. `liquid:”name”` is accessed as `value.name` instead.
    - If the value of the field is a function that takes no arguments and returns either one or two arguments, accessing it invokes the function, and the value of the property is its first return value.
    - If the second return value is non-nil, accessing the field panics instead.
  - A function defined on a struct can be accessed by function name e.g. `value.Func`, `value["Func"]`.
    - The same rules apply as to accessing a func-valued public field.
  - Note that despite being array- and map-like, structs do not have a special `value.size` property.
- `[]byte`
  - A value of type `[]byte` is rendered as the corresponding string, and presented as a string to filters that expect one. A `[]byte` is not (currently) equivalent to a `string` for all uses; for example, `a < b`, `a contains b`, `hash[b]` will not behave as expected where `a` or `b` is a `[]byte`.
- `MapSlice`
  - An instance of `yaml.MapSlice` acts as a map. It implements `m.key`, `m[key]`, and `m.size`.

### References

* [Shopify.github.io/liquid](https://shopify.github.io/liquid)
* [Liquid for Designers](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers)
* [Liquid for Programmers](https://github.com/Shopify/liquid/wiki/Liquid-for-Programmers)
* [Help.shopify.com](https://help.shopify.com/themes/liquid) goes into more detail, but includes features that aren't present in core Liquid as used by Jekyll.

## Contributing

Bug reports, test cases, and code contributions are more than welcome.
Please refer to the [contribution guidelines](./CONTRIBUTING.md).

### Contributors

Thanks goes to these wonderful people ([emoji key](https://github.com/kentcdodds/all-contributors#emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore -->
| [<img src="https://avatars2.githubusercontent.com/u/674?v=4" width="100px;"/><br /><sub><b>Oliver Steele</b></sub>](https://osteele.com/)<br />[💻](https://github.com/osteele/liquid/commits?author=osteele "Code") [📖](https://github.com/osteele/liquid/commits?author=osteele "Documentation") [🤔](#ideas-osteele "Ideas, Planning, & Feedback") [🚇](#infra-osteele "Infrastructure (Hosting, Build-Tools, etc)") [👀](#review-osteele "Reviewed Pull Requests") [⚠️](https://github.com/osteele/liquid/commits?author=osteele "Tests") | [<img src="https://avatars0.githubusercontent.com/u/973593?v=4" width="100px;"/><br /><sub><b>James Littlejohn</b></sub>](https://github.com/thessem)<br />[💻](https://github.com/osteele/liquid/commits?author=thessem "Code") [📖](https://github.com/osteele/liquid/commits?author=thessem "Documentation") [⚠️](https://github.com/osteele/liquid/commits?author=thessem "Tests") | [<img src="https://avatars2.githubusercontent.com/u/12567?v=4" width="100px;"/><br /><sub><b>nsf</b></sub>](http://nosmileface.ru)<br />[💻](https://github.com/osteele/liquid/commits?author=nsf "Code") [⚠️](https://github.com/osteele/liquid/commits?author=nsf "Tests") |
| :---: | :---: | :---: |
<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/kentcdodds/all-contributors) specification. Contributions of any kind welcome!

### Attribution

| Package                                             | Author          | Description          | License            |
|-----------------------------------------------------|-----------------|----------------------|--------------------|
| [Ragel](http://www.colm.net/open-source/ragel/)     | Adrian Thurston | scanning expressions | MIT                |
| [gopkg.in/yaml.v2](https://github.com/go-yaml/yaml) | Canonical       | MapSlice             | Apache License 2.0 |

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

[appveyor-url]: https://ci.appveyor.com/project/osteele/liquid
[appveyor-svg]: https://ci.appveyor.com/api/projects/status/76tnj36879n671jx?svg=true
