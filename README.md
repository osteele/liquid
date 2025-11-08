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
  - [Security](#security)
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

`go get github.com/osteele/liquid` # latest version

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

### Filters

Filters transform template values. The library includes [standard Shopify Liquid filters](https://shopify.github.io/liquid/filters/abs/), and you can also define custom filters.

#### Basic Filter

```go
engine := liquid.NewEngine()
engine.RegisterFilter("has_prefix", strings.HasPrefix)

out, _ := engine.ParseAndRenderString(`{{ title | has_prefix: "Intro" }}`,
    map[string]any{"title": "Introduction"})
// Output: true
```

#### Filter with Optional Arguments

Use a function parameter to provide default values:

```go
engine.RegisterFilter("inc", func(a int, b func(int) int) int {
    return a + b(1)  // b(1) provides default value
})

out, _ := engine.ParseAndRenderString(`{{ n | inc }}`, map[string]any{"n": 10})
// Output: 11

out, _ = engine.ParseAndRenderString(`{{ n | inc: 5 }}`, map[string]any{"n": 10})
// Output: 15
```

#### Filters with Named Arguments

Filters can accept named arguments by including a `map[string]any` parameter:

```go
engine.RegisterFilter("img_url", func(image string, size string, opts map[string]any) string {
    scale := 1
    if s, ok := opts["scale"].(int); ok {
        scale = s
    }
    return fmt.Sprintf("https://cdn.example.com/%s?size=%s&scale=%d", image, size, scale)
})

// Use with named arguments
out, _ := engine.ParseAndRenderString(
    `{{image | img_url: '580x', scale: 2}}`,
    map[string]any{"image": "product.jpg"})
// Output: https://cdn.example.com/product.jpg?size=580x&scale=2

// Named arguments are optional
out, _ = engine.ParseAndRenderString(
    `{{image | img_url: '300x'}}`,
    map[string]any{"image": "product.jpg"})
// Output: https://cdn.example.com/product.jpg?size=300x&scale=1
```

The named arguments syntax follows Shopify Liquid conventions:
- Named arguments use the format `name: value`
- Multiple arguments are comma-separated: `filter: pos_arg, name1: value1, name2: value2`
- Positional arguments come before named arguments
- If the filter function's last parameter is `map[string]any`, it receives all named arguments

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

`go install github.com/osteele/liquid/cmd/liquid@latest` installs a command-line
`liquid` executable. This is intended to make it easier to create test cases for
bug reports.

```bash
$ liquid --help
usage: liquid [FILE]
$ echo '{{ "Hello World" | downcase | split: " " | first | append: "!"}}' | liquid
hello!
```

## Security

**Important**: If you plan to process untrusted templates (templates authored by users you don't fully trust), please review the [Security Policy](SECURITY.md) documentation.

Key security considerations:

- **Sandboxed Execution**: Templates cannot execute arbitrary code or access filesystem/network resources (by default)
- **DoS Vulnerabilities**: The engine is vulnerable to denial-of-service attacks via infinite loops and memory exhaustion when processing untrusted templates
- **Resource Limiting via FRender**: Use the `FRender` method with custom writers to implement timeouts and output size limits for untrusted templates
- **Third-Party Extensions**: Custom filters and tags execute arbitrary Go code and should be carefully audited

For detailed information about security guarantees, limitations, and production deployment recommendations, see [SECURITY.md](SECURITY.md). For implementing resource limits, see the [FRender documentation](./docs/FRender.md).

## Documentation

This section provides a comprehensive guide to using and extending the Liquid template engine. Documentation is organized by topic:

### Getting Started

- **[Installation](#installation)** - Install the library and command-line tool
- **[Usage](#usage)** - Quick start guide with examples
- **[Command-Line Tool](#command-line-tool)** - Testing templates from the command line
- **[API Documentation][godoc-url]** - Complete API reference on pkg.go.dev

### Core Concepts

- **[Value Types](#value-types)** - How Go values map to Liquid types
- **[Drops](#drops)** - Custom types in templates
- **[Status](#status)** - Feature compatibility with Shopify Liquid

### Advanced Usage

- **[Template Store](#template-store)** - Custom template storage (filesystem, database, etc.)
  - See also: [Template Store Example](./docs/TemplateStoreExample.md)
- **[Advanced Rendering](#advanced-rendering)** - FRender for streaming, timeouts, and size limits
  - See also: [FRender Documentation](./docs/FRender.md)

### Security & Performance

- **[Security](#security)** - Resource limits and security considerations
  - See also: [SECURITY.md](SECURITY.md)
- **[FRender Documentation](./docs/FRender.md)** - Implementing resource limits in production

### Internals

- **[Loop Semantics](./docs/loop-semantics.md)** - Comparison with Ruby Liquid implementation
- **[References](#references)** - Shopify Liquid documentation and resources

### Contributing

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - How to contribute to the project
- **[Contributors](#contributors)** - List of project contributors

---

### Status

These features of Shopify Liquid aren't implemented:

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
    engine.RegisterTemplateStore(myTemplateStore)
    ```

`FileTemplateStore` is the default mechanism for backwards compatibility.

Refer to [example](./docs/TemplateStoreExample.md) for an example implementation.

### Advanced Rendering

#### Custom Writers (FRender)

For advanced use cases like streaming to files, implementing timeouts, or limiting output size, use the `FRender` method to render directly to any `io.Writer`:

```go
var buf bytes.Buffer
err := template.FRender(&buf, bindings)
```

This is particularly useful for:
- Rendering large templates without buffering in memory
- Implementing cancellation via context
- Limiting output size from untrusted templates
- Custom output transformation

See the [FRender documentation](./docs/FRender.md) for detailed examples and security best practices.

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
    <td align="center"><a href="https://utpal.io/"><img src="https://avatars.githubusercontent.com/u/19898129?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Utpal Sarkar</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=uksarkar" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=uksarkar" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/imiskolee"><img src="https://avatars.githubusercontent.com/u/1549948?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Misko Lee</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=imiskolee" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/aisbergg"><img src="https://avatars.githubusercontent.com/u/14318942?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Andre Lehmann</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=aisbergg" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/jamesog"><img src="https://avatars.githubusercontent.com/u/982184?v=4?s=100" width="100px;" alt=""/><br /><sub><b>James O'Gorman</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=jamesog" title="Code">💻</a> <a href="https://github.com/osteele/liquid/issues?q=author%3Ajamesog" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://github.com/ofavre"><img src="https://avatars.githubusercontent.com/u/95129?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Olivier Favre</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=ofavre" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/peteraba"><img src="https://avatars.githubusercontent.com/u/1675360?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Peter Aba</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=peteraba" title="Documentation">📖</a></td>
    <td align="center"><a href="https://github.com/chrisghill"><img src="https://avatars.githubusercontent.com/u/15616541?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Christopher Hill</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=chrisghill" title="Code">💻</a> <a href="https://github.com/osteele/liquid/issues?q=author%3Achrisghill" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://github.com/wttw"><img src="https://avatars.githubusercontent.com/u/389596?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Steve Atkins</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=wttw" title="Code">💻</a> <a href="https://github.com/osteele/liquid/issues?q=author%3Awttw" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://github.com/prestonprice57"><img src="https://avatars.githubusercontent.com/u/10774823?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Preston Price</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=prestonprice57" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/jamslinger"><img src="https://avatars.githubusercontent.com/u/80337165?v=4?s=100" width="100px;" alt=""/><br /><sub><b>jamslinger</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=jamslinger" title="Code">💻</a> <a href="https://github.com/osteele/liquid/issues?q=author%3Ajamslinger" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://github.com/deining"><img src="https://avatars.githubusercontent.com/u/18169566?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Andreas Deininger</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=deining" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/magiusdarrigo"><img src="https://avatars.githubusercontent.com/u/43056803?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Matteo Agius-D'Arrigo</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=magiusdarrigo" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/codykrieger"><img src="https://avatars.githubusercontent.com/u/1311179?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Cody Krieger</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=codykrieger" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/stephanejais"><img src="https://avatars.githubusercontent.com/u/822431?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Stéphane JAIS</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=stephanejais" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/jam3sn"><img src="https://avatars.githubusercontent.com/u/7646700?v=4?s=100" width="100px;" alt=""/><br /><sub><b>James Newman</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=jam3sn" title="Code">💻</a> <a href="https://github.com/osteele/liquid/issues?q=author%3Ajam3sn" title="Bug reports">🐛</a></td>
    <td align="center"><a href="https://github.com/chrisatbd"><img src="https://avatars.githubusercontent.com/u/180913248?v=4?s=100" width="100px;" alt=""/><br /><sub><b>chris</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=chrisatbd" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/dop251"><img src="https://avatars.githubusercontent.com/u/995021?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Dmitry Panov</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=dop251" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/GauthierHacout"><img src="https://avatars.githubusercontent.com/u/71611631?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Gauthier Hacout</b></sub></a><br /><a href="https://github.com/osteele/liquid/issues?q=author%3AGauthierHacout" title="Bug reports">🐛</a></td>
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
