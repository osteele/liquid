# Security policy

This document describes the security boundary of `osteele/liquid`. Read it
before you render templates from users you do not trust.

## Security model

Liquid templates cannot import Go packages or call arbitrary Go code directly.
They can read values from the bindings supplied by the application and call
the filters and tags registered on the engine.

That boundary depends on application configuration:

- The built-in `include` and `render` tags read templates through the
  configured `TemplateStore`.
- The default `FileTemplateStore` confines reads to its `Root`. An empty root
  uses the current working directory. Include and render paths must stay
  beneath the source template's directory.
- A custom `TemplateStore`, filter, tag, `Drop`, zero-argument struct method,
  or function-valued struct field runs application Go code. Treat each one as
  part of the trusted computing base.
- Templates can read every public value exposed through their bindings. Do not
  bind secrets or infrastructure objects.
- The engine does not make network requests unless registered application code
  does so.

The engine does not enable HTML escaping by default. Call
`Engine.SetAutoEscapeReplacer(render.HtmlEscaper)` when output enters an HTML
context, or escape the output in the application.

## Resource exhaustion

The engine does not impose limits on iterations, assignments, output size, or
execution time. A template can consume excessive CPU or memory through large
ranges, nested loops, repeated string growth, or deeply nested input.

`FRender` can enforce an output limit and check a cancellation context whenever
the template writes output. It does not provide a hard execution deadline. A
long computation that does not write may continue until its next write or
until it finishes.

Use process or container isolation when you need enforceable CPU, memory, and
wall-clock limits. Apply all of the following controls to untrusted templates:

- Set OS-level CPU and memory limits.
- Render in a worker process that the caller can terminate.
- Limit template source size and nesting.
- Limit output with an `FRender` writer.
- Expose a minimal binding map.
- Register only audited stores, filters, tags, drops, and callable fields or
  methods.
- Rate-limit render requests.

See [Rendering to custom writers](docs/FRender.md) for cooperative cancellation
and output-limit examples.

## Template and data injection

Keep template source separate from untrusted data. Pass data through bindings
instead of concatenating it into a template:

```go
// Unsafe: input can change the template program.
source := "Hello " + userInput

// Safer: input remains data.
source := "Hello {{ name }}"
bindings := liquid.Bindings{"name": userInput}
```

A template can reveal any value reachable from its bindings. Prefer maps or
narrow presentation structs over domain objects with sensitive public fields
or callable methods.

Escape output for its destination. HTML escaping does not make output safe for
JavaScript, CSS, SQL, shell commands, URLs, or other contexts.

## Template stores

The default file store uses `os.OpenRoot` to prevent `..` traversal and
symbolic-link escapes. Set an explicit root when templates live outside the
process working directory:

```go
engine.RegisterTemplateStore(&render.FileTemplateStore{
    Root: "/srv/myapp/templates",
})
```

Template paths are relative. The `include` and `render` tags reject absolute
paths and paths that escape the source template's directory.

A custom store must enforce its own authorization rules. If users can choose
template names, allow only known templates or tenant-scoped prefixes. Do not
pass names directly to an unrestricted filesystem, database, object store, or
network client.

## Extension safety

Validate all arguments received by custom filters and tags. Use least-privilege
clients, bounded operations, and explicit allowlists. Do not assume a template
will call an extension only in the way shown in its documentation.

Drops and callable struct members deserve the same review. Their Go code can
access process state even though the Liquid language cannot.

## Reporting a vulnerability

Please report sensitive vulnerabilities through
[GitHub's private vulnerability reporting form](https://github.com/osteele/liquid/security/advisories/new).
Include the affected versions, impact, reproduction steps, and any suggested
remediation.

Use a public issue only when disclosure does not put users at risk.
