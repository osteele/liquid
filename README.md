# Liquid templates for Go

[![go badge][go-svg]][go-url]
[![Golangci-lint badge][golangci-lint-svg]][golangci-lint-url]
[![Go Report Card badge][go-report-card-svg]][go-report-card-url]
[![Go Doc][godoc-svg]][godoc-url]
[![MIT License][license-svg]][license-url]

`liquid` is a pure Go implementation of
[Shopify Liquid](https://shopify.github.io/liquid). It was developed for
[Gojekyll](https://github.com/osteele/gojekyll), a Go port of the Jekyll
static-site generator.

## Installation

```bash
go get github.com/osteele/liquid@latest
```

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
if err != nil {
    log.Fatal(err)
}
fmt.Println(out)
// Output: <h1>Introduction</h1>
```

See the [API documentation][godoc-url] for additional examples.

### Jekyll compatibility

Optional Jekyll extensions support syntax that is not part of Shopify Liquid.

To enable Jekyll compatibility mode:

```go
engine := liquid.NewEngine()
engine.EnableJekyllExtensions()
```

Jekyll mode allows dot notation in assignment targets, such as
`{% assign page.canonical_url = "/about/" %}`. It creates missing intermediate
maps. Nested assignments use copy-on-write and do not modify maps supplied by
the caller.

Example:

```go
engine := liquid.NewEngine()
engine.EnableJekyllExtensions()

template := `{% assign page.meta.author = "John Doe" %}{{ page.meta.author }}`
bindings := map[string]any{
    "page": map[string]any{
        "title": "Home",
    },
}
out, err := engine.ParseAndRenderString(template, bindings)
if err != nil {
    log.Fatal(err)
}
fmt.Println(out)
// Output: John Doe
```

Jekyll extensions are disabled by default.

### Command-line tool

`go install github.com/osteele/liquid/cmd/liquid@latest` installs a command-line
`liquid` executable for testing templates and preparing bug reports.

```bash
$ liquid --help
usage: liquid [FILE]
$ echo '{{ "Hello World" | downcase | split: " " | first | append: "!"}}' | liquid
hello!
```

## Security

Read the [security policy](SECURITY.md) before rendering untrusted templates.
The engine has no built-in CPU, memory, iteration, or output limits.
Auto-escaping is opt-in. The `include` and `render` tags can read through the
configured template store. Registered extensions and callable bound values
execute application Go code.

Use [`FRender`](./docs/FRender.md) to limit output and check cooperative
cancellation. Use process or container isolation when you need enforceable
resource limits.

## Documentation

The [API reference][godoc-url] documents exported Go types and methods. The
guides cover [custom template stores](./docs/TemplateStoreExample.md),
[`FRender`](./docs/FRender.md), [security](SECURITY.md), and
[loop-modifier differences](./docs/loop-semantics.md).

### Status

The following Shopify Liquid feature is not implemented:

- Warn and lax [error modes](https://github.com/shopify/liquid#error-modes).
  - `Engine.LaxFilters()` does provide Shopify-compatible pass-through behavior
    for undefined filters.

### Drops

Drops have a different design from the Shopify (Ruby) implementation. A Ruby
drop sets `liquid_attributes` to a list of attributes that are exposed to
Liquid. A Go drop implements `ToLiquid() any`, that returns a proxy
object. The proxy is usually a map or struct that defines the exposed
properties. See the
[`Drop` API documentation](https://pkg.go.dev/github.com/osteele/liquid#Drop)
for details.

### Value Types

`Render` and friends take a `Bindings` parameter. This is a map of `string` to
`any` that associates template variable names with Go values.

Any Go value can be used as a variable value. These values have special
meaning:

- `false` and `nil`
  - These, and no other values, are recognized as false by `and`, `or`, `{% if
    %}`, `{% elsif %}`, and `{% case %}`.
- Integers
  - Integers can be used as array indices: `array[1]`; `array[n]`, where
    `array` has an array value and `n` has an integer value.
  - (Only) integers can be used as the endpoints of a range: `{% for item in
    (1..5) %}`, `{% for item in (start..end) %}` where `start` and `end` have
    integer values.
- Integers and floats
  - Integers and floats are converted to their join type for comparison: `1 ==
    1.0` evaluates to `true`. Similarly, `int8(1)`, `int16(1)`, and `uint8(1)`
    are all equal.
  - Complex numbers receive no special treatment.
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
    string value.
  - A map can be accessed using property syntax: `hash.key`.
  - Maps have a special `size` property, that returns the size of the map.
- Drops
  - A value `value` of a type that implements the `Drop` interface acts as the
    value `value.ToLiquid()`. There is no guarantee about how many times
    `ToLiquid` will be called. [This is in contrast to Shopify Liquid, which
    both uses a different interface for drops, and makes stronger guarantees.]
- Structs
  - A public field of a struct can be accessed by its name: `value.FieldName`,
    `value["FieldName"]`.
    - A field tagged `liquid:"name"` is accessed as `value.name` instead.
    - If the value of the field is a function that takes no arguments and
      returns either one or two values, accessing it invokes the function,
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

`TemplateStore` loads files for the `include` and `render` tags. Implement it
to load templates from an embedded filesystem, database, or service:

```go
type TemplateStore interface {
    ReadTemplate(templateName string) ([]byte, error)
}

engine.RegisterTemplateStore(myTemplateStore)
```

`FileTemplateStore` is the default. It confines reads to `Root`; an empty root
uses the current working directory. Include and render paths are relative to
the source template and cannot escape its directory.

See the [embedded template-store example](./docs/TemplateStoreExample.md).

### Advanced Rendering

#### Custom Writers (FRender)

Use `FRender` to write directly to an `io.Writer`:

```go
var buf bytes.Buffer
err := template.FRender(&buf, bindings)
```

Writer wrappers can limit output, check a cancellation context when output is
written, or transform output. Writer errors are returned from `FRender` and
support `errors.Is`.

See the [`FRender` guide](./docs/FRender.md) for examples and limitations.

### References

- [Shopify.github.io/liquid](https://shopify.github.io/liquid)
- [Liquid for Designers](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers)
- [Liquid for Programmers](https://github.com/Shopify/liquid/wiki/Liquid-for-Programmers)
- [Help.shopify.com](https://help.shopify.com/themes/liquid)

## Contributing

Bug reports, test cases, documentation, and code contributions are welcome.
Read the [contribution guide](./CONTRIBUTING.md) before opening a pull request.

### Contributors

Thanks to these contributors
([emoji key](https://github.com/kentcdodds/all-contributors#emoji-key)):

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
    <td align="center"><a href="https://github.com/jaime-amate"><img src="https://avatars.githubusercontent.com/u/16927375?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Jaime Amate</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=jaime-amate" title="Code">💻</a></td>
  </tr>
  <tr>
    <td align="center"><a href="https://github.com/michaelhvisser"><img src="https://avatars.githubusercontent.com/u/40672164?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Michael Visser</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=michaelhvisser" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=michaelhvisser" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/pierre-b"><img src="https://avatars.githubusercontent.com/u/1058531?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Pierre</b></sub></a><br /><a href="https://github.com/osteele/liquid/commits?author=pierre-b" title="Code">💻</a> <a href="https://github.com/osteele/liquid/commits?author=pierre-b" title="Documentation">📖</a> <a href="#ideas-pierre-b" title="Ideas, Planning, & Feedback">🤔</a> <a href="https://github.com/osteele/liquid/commits?author=pierre-b" title="Tests">⚠️</a></td>
    <td align="center"><a href="https://github.com/ns-jjiang"><img src="https://avatars.githubusercontent.com/u/117057470?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Jush Jiang</b></sub></a><br /><a href="#ideas-ns-jjiang" title="Ideas, Planning, & Feedback">🤔</a></td>
    <td align="center"><a href="https://github.com/tuchida"><img src="https://avatars.githubusercontent.com/u/201790?v=4?s=100" width="100px;" alt=""/><br /><sub><b>tuchida</b></sub></a><br /><a href="#ideas-tuchida" title="Ideas, Planning, & Feedback">🤔</a></td>
    <td align="center"><a href="https://github.com/tanema"><img src="https://avatars.githubusercontent.com/u/463193?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Tim Anema</b></sub></a><br /><a href="#ideas-tanema" title="Ideas, Planning, & Feedback">🤔</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the
[all-contributors](https://github.com/kentcdodds/all-contributors)
specification. Contributions of all kinds are welcome.

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

[go-url]: https://github.com/osteele/liquid/actions/workflows/test.yml
[go-svg]: https://github.com/osteele/liquid/actions/workflows/test.yml/badge.svg

[golangci-lint-url]: https://github.com/osteele/liquid/actions/workflows/lint.yml
[golangci-lint-svg]: https://github.com/osteele/liquid/actions/workflows/lint.yml/badge.svg

[godoc-url]: https://pkg.go.dev/github.com/osteele/liquid
[godoc-svg]: https://pkg.go.dev/badge/github.com/osteele/liquid.svg

[license-url]: https://github.com/osteele/liquid/blob/main/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg

[go-report-card-url]: https://goreportcard.com/report/github.com/osteele/liquid
[go-report-card-svg]: https://goreportcard.com/badge/github.com/osteele/liquid
