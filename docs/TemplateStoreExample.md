# Load templates from an embedded filesystem

A `TemplateStore` supplies files to the `include` and `render` tags. This
example embeds a `templates` directory in the application binary.

```go
package templates

import (
    "embed"
    "io/fs"
)

//go:embed all:templates
var files embed.FS

type Store struct {
    FS fs.FS
}

func (s *Store) ReadTemplate(name string) ([]byte, error) {
    return fs.ReadFile(s.FS, name)
}

func NewStore() (*Store, error) {
    templateFS, err := fs.Sub(files, "templates")
    if err != nil {
        return nil, err
    }
    return &Store{FS: templateFS}, nil
}
```

Register the store before parsing or rendering templates:

```go
store, err := templates.NewStore()
if err != nil {
    log.Fatal(err)
}

engine := liquid.NewEngine()
engine.RegisterTemplateStore(store)
```

Give the source template a path relative to the embedded `templates` directory
when includes must resolve relative to it:

```go
template, err := engine.ParseTemplateLocation(source, "pages/index.liquid", 1)
```

An include such as `{% include "../shared/header.liquid" %}` is rejected
because include and render paths cannot escape the source template's directory.
Organize reusable partials beneath that directory, or implement an application
tag with an explicit allowlist for cross-directory lookup.

Custom stores must enforce any additional authorization rules your application
needs. See the [security policy](../SECURITY.md).
