# Rendering to Custom Writers with FRender

The `FRender` method enables rendering Liquid templates directly to any `io.Writer` implementation, providing fine-grained control over output handling. This is particularly useful for performance optimization, resource limiting, and security constraints.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Use Cases](#use-cases)
  - [Direct File Writing](#direct-file-writing)
  - [Context-Based Cancellation](#context-based-cancellation)
  - [Limiting Output Size](#limiting-output-size)
  - [Custom Output Transformation](#custom-output-transformation)
- [API Reference](#api-reference)

## Basic Usage

The simplest use of `FRender` writes template output to any `io.Writer`:

```go
engine := liquid.NewEngine()
template, err := engine.ParseTemplate([]byte(`<h1>{{ page.title }}</h1>`))
if err != nil {
    log.Fatal(err)
}

bindings := map[string]any{
    "page": map[string]string{"title": "Introduction"},
}

var buf bytes.Buffer
err = template.FRender(&buf, bindings)
if err != nil {
    log.Fatal(err)
}

fmt.Println(buf.String())
// Output: <h1>Introduction</h1>
```

## Use Cases

### Direct File Writing

Avoid unnecessary memory allocation by rendering large templates directly to files:

```go
engine := liquid.NewEngine()
template, err := engine.ParseTemplate(sourceBytes)
if err != nil {
    log.Fatal(err)
}

file, err := os.Create("output.html")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Stream directly to file without intermediate buffers
err = template.FRender(file, bindings)
if err != nil {
    log.Fatal(err)
}
```

### Context-Based Cancellation

Prevent runaway template rendering by implementing cancellation via context:

```go
// CancelWriter wraps an io.Writer with context cancellation support
type CancelWriter struct {
    ctx context.Context
    w   io.Writer
}

func (cw *CancelWriter) Write(p []byte) (n int, err error) {
    select {
    case <-cw.ctx.Done():
        return 0, cw.ctx.Err()
    default:
        return cw.w.Write(p)
    }
}

func renderWithTimeout(template *liquid.Template, bindings liquid.Bindings, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    var buf bytes.Buffer
    cw := &CancelWriter{ctx: ctx, w: &buf}

    err := template.FRender(cw, bindings)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return "", fmt.Errorf("template rendering exceeded %v timeout", timeout)
        }
        return "", err
    }

    return buf.String(), nil
}

// Usage
engine := liquid.NewEngine()
template, _ := engine.ParseTemplate([]byte(`{% for i in (1..1000000) %}{{ i }}{% endfor %}`))

result, err := renderWithTimeout(template, liquid.Bindings{}, 100*time.Millisecond)
if err != nil {
    log.Printf("Rendering stopped: %v", err)
}
```

This is crucial when rendering untrusted templates that might contain deeply nested loops or expensive operations.

### Limiting Output Size

Protect against excessive output from untrusted templates:

```go
// LimitWriter enforces a maximum output size
type LimitWriter struct {
    w        io.Writer
    written  int64
    maxBytes int64
}

var ErrOutputLimitExceeded = errors.New("output size limit exceeded")

func NewLimitWriter(w io.Writer, maxBytes int64) *LimitWriter {
    return &LimitWriter{w: w, maxBytes: maxBytes}
}

func (lw *LimitWriter) Write(p []byte) (n int, err error) {
    if lw.written+int64(len(p)) > lw.maxBytes {
        return 0, ErrOutputLimitExceeded
    }

    n, err = lw.w.Write(p)
    lw.written += int64(n)
    return n, err
}

func renderWithSizeLimit(template *liquid.Template, bindings liquid.Bindings, maxBytes int64) (string, error) {
    var buf bytes.Buffer
    lw := NewLimitWriter(&buf, maxBytes)

    err := template.FRender(lw, bindings)
    if err != nil {
        if errors.Is(err, ErrOutputLimitExceeded) {
            return "", fmt.Errorf("template output exceeded %d bytes", maxBytes)
        }
        return "", err
    }

    return buf.String(), nil
}

// Usage - limit untrusted template output to 1MB
result, err := renderWithSizeLimit(template, bindings, 1024*1024)
if err != nil {
    log.Printf("Rendering failed: %v", err)
}
```

### Custom Output Transformation

Transform output on-the-fly without post-processing:

```go
// UpperCaseWriter converts all output to uppercase
type UpperCaseWriter struct {
    w io.Writer
}

func (uc *UpperCaseWriter) Write(p []byte) (n int, err error) {
    upper := bytes.ToUpper(p)
    return uc.w.Write(upper)
}

// MinifyWriter could strip whitespace, compress, etc.
type MinifyWriter struct {
    w io.Writer
}

func (mw *MinifyWriter) Write(p []byte) (n int, err error) {
    // Remove extra whitespace
    compressed := regexp.MustCompile(`\s+`).ReplaceAll(p, []byte(" "))
    _, err = mw.w.Write(compressed)
    return len(p), err // Return original length for proper accounting
}

// Usage
var buf bytes.Buffer
upperWriter := &UpperCaseWriter{w: &buf}
template.FRender(upperWriter, bindings)
```

## API Reference

### Template.FRender

```go
func (t *Template) FRender(w io.Writer, vars Bindings) SourceError
```

Executes the template with the specified variable bindings and writes output to `w`.

**Parameters:**
- `w`: Any type implementing `io.Writer` interface
- `vars`: Variable bindings (typically `map[string]any`)

**Returns:**
- `SourceError`: Error with source location information, or `nil` on success

**Error Handling:**

`FRender` returns errors from:
1. Template execution errors (undefined variables, filter errors, etc.)
2. Writer errors (disk full, context cancellation, custom limits, etc.)

Both error types are returned as `SourceError` when possible, providing line number information for template-related issues.

### Engine.ParseAndFRender

```go
func (e *Engine) ParseAndFRender(w io.Writer, source []byte, b Bindings) SourceError
```

Convenience method that parses a template and immediately renders it to a writer.

**Example:**

```go
engine := liquid.NewEngine()
var buf bytes.Buffer

err := engine.ParseAndFRender(&buf, []byte(`{{ greeting }}`), liquid.Bindings{
    "greeting": "Hello, World!",
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(buf.String())
```

## Comparison with Render Methods

| Method | Return Type | Use Case |
|--------|-------------|----------|
| `Render(vars)` | `([]byte, error)` | Small templates, need byte slice |
| `RenderString(vars)` | `(string, error)` | Small templates, need string |
| `FRender(w, vars)` | `error` | Large output, streaming, custom handling |

**When to use FRender:**
- Template output > 1MB (avoid memory allocation)
- Writing to files or network connections
- Need cancellation or resource limits
- Want custom output transformation
- Rendering untrusted templates

**When to use Render/RenderString:**
- Small templates with predictable output
- Need the result as a value for further processing
- Simpler code for straightforward use cases

## Performance Considerations

`FRender` can significantly improve performance for large templates:

```go
// Memory-inefficient for large output
data, _ := template.Render(bindings)
file.Write(data)  // Entire output buffered in memory

// Memory-efficient streaming
file, _ := os.Create("output.html")
template.FRender(file, bindings)  // Streams directly to disk
```

For a 100MB template output:
- `Render()` approach: ~100MB memory usage
- `FRender()` approach: ~4KB memory usage (typical buffer size)

## Security Best Practices

When rendering untrusted templates, always use FRender with protective wrappers:

```go
type SafeWriter struct {
    ctx      context.Context
    w        io.Writer
    written  int64
    maxBytes int64
}

func (sw *SafeWriter) Write(p []byte) (n int, err error) {
    // Check context cancellation
    select {
    case <-sw.ctx.Done():
        return 0, sw.ctx.Err()
    default:
    }

    // Check size limit
    if sw.written+int64(len(p)) > sw.maxBytes {
        return 0, ErrOutputLimitExceeded
    }

    n, err = sw.w.Write(p)
    sw.written += int64(n)
    return n, err
}

func renderUntrusted(template *liquid.Template, bindings liquid.Bindings) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var buf bytes.Buffer
    safeWriter := &SafeWriter{
        ctx:      ctx,
        w:        &buf,
        maxBytes: 10 * 1024 * 1024, // 10MB limit
    }

    err := template.FRender(safeWriter, bindings)
    return buf.String(), err
}
```

This approach protects against:
- Infinite loops or deeply nested iterations
- Excessive memory consumption
- DoS attacks via template complexity
