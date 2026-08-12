# Render to an `io.Writer` with `FRender`

`FRender` sends template output to an `io.Writer`. Use it to write directly to
a file or connection, transform output, or stop rendering when a writer
reaches a limit.

## Basic use

```go
engine := liquid.NewEngine()
template, err := engine.ParseTemplate([]byte(`<h1>{{ page.title }}</h1>`))
if err != nil {
    log.Fatal(err)
}

bindings := liquid.Bindings{
    "page": map[string]string{"title": "Introduction"},
}

var output bytes.Buffer
if err := template.FRender(&output, bindings); err != nil {
    log.Fatal(err)
}
fmt.Println(output.String())
```

`Template.FRender` returns a `SourceError`. Writer failures wrap their original
error, so callers can inspect them with `errors.Is`.

`Engine.ParseAndFRender` combines parsing and rendering:

```go
err := engine.ParseAndFRender(
    os.Stdout,
    []byte(`Hello, {{ name }}!`),
    liquid.Bindings{"name": "Ada"},
)
```

## Write directly to a file

```go
file, err := os.Create("output.html")
if err != nil {
    return err
}

renderErr := template.FRender(file, bindings)
closeErr := file.Close()
return errors.Join(renderErr, closeErr)
```

This avoids creating the complete rendered result as a `[]byte` or `string`
before writing it.

## Limit output and check cancellation

The engine has no built-in output or execution limit. A writer can reject
output after a byte limit and check a context at each write:

```go
var ErrOutputLimit = errors.New("template output limit exceeded")

type BoundedWriter struct {
    Context  context.Context
    Writer   io.Writer
    MaxBytes int64
    written  int64
}

func (w *BoundedWriter) Write(p []byte) (int, error) {
    select {
    case <-w.Context.Done():
        return 0, w.Context.Err()
    default:
    }

    if int64(len(p)) > w.MaxBytes-w.written {
        return 0, ErrOutputLimit
    }

    n, err := w.Writer.Write(p)
    w.written += int64(n)
    if err == nil && n != len(p) {
        err = io.ErrShortWrite
    }
    return n, err
}
```

Use the writer with a timeout context:

```go
func renderBounded(
    template *liquid.Template,
    bindings liquid.Bindings,
    timeout time.Duration,
    maxBytes int64,
) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    var output bytes.Buffer
    writer := &BoundedWriter{
        Context:  ctx,
        Writer:   &output,
        MaxBytes: maxBytes,
    }

    if err := template.FRender(writer, bindings); err != nil {
        switch {
        case errors.Is(err, context.DeadlineExceeded):
            return "", fmt.Errorf("template render timed out: %w", err)
        case errors.Is(err, ErrOutputLimit):
            return "", fmt.Errorf(
                "template output exceeded %d bytes: %w",
                maxBytes,
                err,
            )
        default:
            return "", err
        }
    }
    return output.String(), nil
}
```

Cancellation is cooperative. The writer checks its context only when Liquid
writes output. A template that performs a long computation without writing may
not stop at the deadline. Run untrusted templates in a process or container
with enforceable CPU, memory, and wall-clock limits.

## Transform output

Writer wrappers can transform each output chunk:

```go
type UpperWriter struct {
    Writer io.Writer
}

func (w UpperWriter) Write(p []byte) (int, error) {
    upper := bytes.ToUpper(p)
    n, err := w.Writer.Write(upper)
    if err == nil && n != len(upper) {
        err = io.ErrShortWrite
    }
    return n, err
}
```

Transformations must preserve `io.Writer` semantics. Return the number of input
bytes consumed, propagate the wrapped writer's error, and return
`io.ErrShortWrite` when a writer accepts fewer bytes without an error.

## Choose a render method

| Method | Result | Use it when |
| --- | --- | --- |
| `Template.Render` | `[]byte` | The caller needs bytes. |
| `Template.RenderString` | `string` | The caller needs a string. |
| `Template.FRender` | Writes to `io.Writer` | Output should stream through a destination or wrapper. |
| `Engine.ParseAndFRender` | Parses, then writes | The caller has source bytes and does not need to retain a compiled template. |

For the full threat model, read the [security policy](../SECURITY.md).
