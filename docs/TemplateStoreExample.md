# Template Store Example

This document describes the implementation of an `TemplateStore` that uses an embedded file system as its storage type.

Add a go file to your project with configuration properties and the ReadTemplate() implementation

```go
package your_package_name

import (
	"embed"
	"fmt"
)

type EmbeddedFileSystemTemplateStore struct {
	Folder  embed.FS
	RootDir string
}

// implementation of ITemplateProvider
func (tl *EmbeddedFileSystemTemplateStore) ReadTemplate(filename string) ([]byte, error) {

	fileName := fmt.Sprintf("%v/%v", tl.RootDir, filename)
	templateFile, _ := tl.Folder.ReadFile(fileName)

	return templateFile, nil
}

```
initialize your embedded folder.  for details on go embedded package see [embed](https://pkg.go.dev/embed)

```go

//go:embed all:templates
var folder embed.FS

```
create store and register with engine

```go
	// use the embedded file system loader for now.
	embedFileSystemTemplateStore := &your_package_name.EmbeddedFileSystemTemplateStore{
		Folder:  folder,
		RootDir: "templates",
	}

    //create engine
    engine := liquid.NewEngine()

    //register with the engine
	engine.RegisterTemplateStore(embedFileSystemTemplateStore)

    //ready to go
```