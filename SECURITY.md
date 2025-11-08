# Security Policy

## Overview

This document outlines the security model, guarantees, and limitations of the `osteele/liquid` Go implementation of the Liquid template language. This information is particularly important if you plan to run end-user-supplied templates in production environments.

## Security Model

Liquid was designed by Shopify to allow end-user modification of templates while preventing malicious code execution. This Go implementation follows that security model with the following characteristics:

### Built-in Security Guarantees

1. **No Disk Access**: The core engine and built-in filters/tags do not access the filesystem
   - Exception: When using `{% include %}` tags with templates from disk (controlled by your `TemplateStore` implementation)

2. **No Network Access**: The core engine and built-in filters/tags do not make network requests

3. **Sandboxed Execution**: Templates cannot execute arbitrary code or access Go language features
   - No access to Go functions outside of registered filters and tags
   - No ability to import packages or define functions
   - No access to reflection or unsafe operations (from the template itself)

4. **Controlled Data Access**: Templates can only access data explicitly provided via bindings
   - No access to environment variables
   - No access to command-line arguments
   - No access to global state (unless explicitly exposed)

## Known Security Limitations

### ⚠️ Denial of Service (DoS) Vulnerabilities

**This implementation is vulnerable to DoS attacks when processing untrusted templates.** Common attack vectors include:

1. **Infinite Loops**
   ```liquid
   {% for i in (1..999999999) %}
     {% for j in (1..999999999) %}
       {{ i }} {{ j }}
     {% endfor %}
   {% endfor %}
   ```

2. **Memory Exhaustion**
   ```liquid
   {% assign huge = "x" %}
   {% for i in (1..30) %}
     {% assign huge = huge | append: huge %}
   {% endfor %}
   ```

3. **Regex Complexity** (via filters that use regular expressions)
   - Certain patterns can cause catastrophic backtracking

4. **Deep Nesting**
   - Deeply nested data structures or template constructs may cause stack overflow

### Mitigation via FRender

While there are **no automatic built-in limits** like Ruby's `resource_limits`, the `FRender` method (available since v1.4.0) enables implementing these protections:

✅ **Available via FRender**:
- Execution timeouts (via context cancellation)
- Output size limits (via custom writers)
- Memory protection (via streaming output)

❌ **Not currently available**:
- CPU usage limits
- Template complexity scoring
- Iteration count limits

**Recommendation**: When processing untrusted templates, use `FRender` with custom writer implementations for timeout and size limiting. See [Production Deployment Recommendations](#production-deployment-recommendations) below for detailed examples.

### Data Injection Risks

1. **Template Injection**: If you construct templates from untrusted data, attackers can inject malicious template code
   ```go
   // ❌ DANGEROUS - Never do this with untrusted input
   template := "Hello {{ user_input }}"  // user_input could be "}} {% for i in (1..999999999) %}..."
   ```

2. **XSS via Unescaped Output**: The `raw` filter bypasses HTML escaping
   ```liquid
   {{ user_input | raw }}  <!-- Could inject malicious HTML/JavaScript -->
   ```

3. **Data Exfiltration**: Templates can expose any data in the bindings context
   ```go
   // If you include sensitive data in bindings, templates can access it
   bindings := map[string]any{
       "user": userData,
       "secrets": apiKeys,  // ❌ Templates can access this!
   }
   ```

### Third-Party Extension Risks

**⚠️ CRITICAL: Custom tags and filters execute arbitrary Go code**

When you register custom filters or tags, you are giving template authors the ability to invoke that code:

```go
// This filter will execute with whatever arguments the template provides
engine.RegisterFilter("custom_filter", func(input any) any {
    // This code runs in your application's context
    // It has full access to filesystem, network, etc.
    return input
})
```

**Recommendations**:
- Carefully audit all custom filters and tags before deploying
- Assume template authors will call your extensions with malicious inputs
- Validate and sanitize all inputs in custom extensions
- Apply principle of least privilege - don't register extensions you don't need

## Production Deployment Recommendations

If you plan to execute untrusted templates (templates authored by users you don't fully trust), consider implementing these safeguards:

### 1. Timeout Protection with FRender

Use `FRender` with a context-aware writer to implement **proper cancellation** that actually stops rendering:

```go
package main

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "io"
    "time"

    "github.com/osteele/liquid"
)

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

func renderWithTimeout(template *liquid.Template, bindings map[string]any, timeout time.Duration) (string, error) {
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
```

**Advantages over goroutine approach**:
- ✅ Actually stops rendering when timeout occurs (not just detection)
- ✅ No resource leaks from continuing goroutines
- ✅ Clean error handling via context
- ✅ Can be combined with other writer wrappers

### 2. Output Size Limits with FRender

Protect against memory exhaustion from excessive output:

```go
import (
    "errors"
    "io"
)

var ErrOutputLimitExceeded = errors.New("output size limit exceeded")

// LimitWriter enforces a maximum output size
type LimitWriter struct {
    w        io.Writer
    written  int64
    maxBytes int64
}

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

func renderWithSizeLimit(template *liquid.Template, bindings map[string]any, maxBytes int64) (string, error) {
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
```

This is equivalent to Ruby's `render_length_limit` option.

### 3. Combined Protection (Recommended)

For production use with untrusted templates, combine timeout and size limits:

```go
// SafeWriter combines context cancellation and size limiting
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

func renderUntrusted(template *liquid.Template, bindings map[string]any) (string, error) {
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

See [docs/FRender.md](./docs/FRender.md) for more examples and patterns.

### 4. Additional Resource Limits

For defense in depth, also consider:
- OS-level process isolation (containers, VMs)
- Memory limits (cgroups, container limits)
- CPU limits via containerization

### 5. Input Validation

```go
// Validate template complexity before execution
func validateTemplate(template string) error {
    if len(template) > 100000 {
        return fmt.Errorf("template too large")
    }

    // Check for suspicious patterns
    loopCount := strings.Count(template, "{% for ")
    if loopCount > 10 {
        return fmt.Errorf("too many loops")
    }

    return nil
}
```

### 6. Minimal Bindings

Only expose data that templates absolutely need:

```go
// ✅ GOOD - minimal exposure
bindings := map[string]any{
    "product_name": product.Name,
    "product_price": product.Price,
}

// ❌ BAD - exposing entire objects
bindings := map[string]any{
    "product": product,  // May expose unintended fields
    "database": db,      // Never expose infrastructure
}
```

### 7. Output Sanitization

Always sanitize template output before displaying in web contexts:

```go
import "html"

output, err := engine.ParseAndRenderString(template, bindings)
if err != nil {
    return err
}

// Sanitize output if displaying in HTML
safeOutput := html.EscapeString(output)
```

### 8. Template Review and Approval

For sensitive applications:
- Implement a template review process
- Use version control for templates
- Audit template changes before deployment
- Consider static analysis of templates

### 9. Rate Limiting

Limit how often users can render templates:
- Prevent abuse and DoS attacks
- Implement per-user rate limits
- Monitor for suspicious patterns

## Security Audit Status

**⚠️ This codebase has not undergone an independent security audit.**

While the core engine aims to be secure by design, users should be aware that:
- There may be undiscovered vulnerabilities
- Security issues may exist in dependencies
- New attack vectors may be discovered

If you discover a security vulnerability, please see the [Reporting Vulnerabilities](#reporting-vulnerabilities) section below.

## Comparison with Other Implementations

### Shopify Liquid (Ruby)

The original Ruby implementation and this Go implementation share the **same fundamental security design** from Shopify Liquid, but there are important differences to consider when choosing between them for security-sensitive applications.

#### Core Security Model: Very Similar

Both implementations provide:
- Sandboxed execution (no arbitrary code execution from templates)
- No filesystem access from built-in tags/filters
- No network access from built-in tags/filters
- Templates can only access explicitly provided data bindings
- Same vulnerability to DoS attacks (infinite loops, memory exhaustion)

#### Key Differences

**1. Production Battle-Testing**

- **Ruby**:
  - Used by Shopify to process millions of templates daily since 2006
  - Extensive real-world security testing through actual attack attempts
  - Edge cases discovered and hardened over 15+ years
  - Large community continuously identifying and reporting issues
  - Well-documented CVEs and security patches

- **Go**:
  - Newer implementation (since 2017) with less production usage at scale
  - Fewer real-world attack scenarios encountered
  - Smaller community, less security scrutiny
  - **No independent security audit** (as documented above)
  - Security issues may remain undiscovered

**2. Built-in Resource Limiting**

- **Ruby**: Has built-in resource limiting capabilities:
  ```ruby
  Liquid::Template.parse(template, resource_limits: {
    render_length_limit: 100000,      # Maximum output size
    render_score_limit: 1000,         # Complexity scoring to prevent expensive operations
    assign_score_limit: 500           # Limits on variable assignments
  })
  ```
  - Complexity scoring tracks "render score" to prevent expensive operations
  - Better timeout support through Ruby's threading model
  - Can abort rendering when limits are exceeded

- **Go**: Provides resource limiting via `FRender` (custom writer pattern):
  - ✅ **Timeout support**: Context-based cancellation via custom writers (since v1.4.0)
  - ✅ **Output size limits**: `render_length_limit` equivalent via `LimitWriter`
  - ✅ **Proper cancellation**: Actually stops rendering (not just detection)
  - ❌ **No complexity scoring**: No `render_score_limit` equivalent
  - ❌ **No iteration limits**: No `assign_score_limit` equivalent
  - **Requires custom writer implementation** (see [Production Deployment Recommendations](#production-deployment-recommendations))

**3. Memory Safety**

- **Ruby**: Memory-safe language with garbage collection
  - No buffer overflows or memory corruption from language itself
  - Dynamic typing provides flexibility but less compile-time safety

- **Go**: Memory-safe language with garbage collection
  - **Similar memory safety guarantees**
  - Static typing provides additional compile-time safety
  - Both can still exhaust memory through malicious template logic

**4. Type System & Attack Surface**

- **Ruby**: Dynamic typing with powerful reflection
  - Larger attack surface if custom filters/tags use reflection carelessly
  - Method calls can be intercepted and controlled via metaprogramming
  - More runtime flexibility but requires careful security review

- **Go**: Static typing with limited reflection
  - Smaller attack surface in custom extensions
  - Type safety provides compile-time guardrails
  - Less flexibility but inherently more restrictive

**5. Community Security Review**

- **Ruby**:
  - Original implementation by Shopify's security-conscious team
  - 15+ years of community scrutiny and security research
  - Established vulnerability disclosure and patching process
  - Known security properties well-documented

- **Go**:
  - Smaller community, less extensive security review
  - Newer implementation with shorter security track record
  - Security properties less thoroughly tested in production
  - Potential vulnerabilities may be undiscovered

#### Bottom Line: Which to Choose?

**For processing untrusted templates at scale in production:**

**Ruby** remains the more battle-tested choice due to:
- 15+ years of production hardening at Shopify
- Automatic complexity scoring (`render_score_limit`, `assign_score_limit`)
- Larger security-focused community
- Better-established security track record

**Go** now provides comparable timeout and output limiting via FRender:
- ✅ Proper timeout support with cancellation (via `FRender` + context)
- ✅ Output size limiting equivalent to `render_length_limit`
- ❌ No automatic complexity scoring for CPU/iteration limits
- Requires custom writer implementation (more code, but flexible)

**Both implementations**:
- Share the same fundamental security model and core guarantees
- Are equally safe for trusted templates (e.g., your own template files)
- Require template complexity validation for full DoS protection

**Choose Go when:**
- You control all templates (e.g., static site generator, internal tools)
- You're willing to implement `FRender` writers for timeout/size limits
- You need Go's performance characteristics and deployment simplicity
- You want fine-grained control over resource limiting logic

**Choose Ruby when:**
- You need automatic complexity scoring without custom code
- You want a single `resource_limits` configuration instead of custom writers
- You're processing large volumes of untrusted templates
- You prefer the most battle-tested implementation

**If using the Go implementation with untrusted templates**, you **must** use:
- ✅ `FRender` with context-aware writer for timeouts ([example above](#1-timeout-protection-with-frender))
- ✅ `FRender` with size-limiting writer for output limits ([example above](#2-output-size-limits-with-frender))
- ✅ Template complexity validation before execution
- ✅ Rate limiting per user/source
- ✅ Comprehensive monitoring and alerting
- ✅ Regular security reviews of custom extensions

## Security Best Practices Summary

✅ **DO**:
- **Use `FRender` for untrusted templates** with timeout and size-limiting writers
- Implement timeouts via context-aware writers
- Limit output size to prevent memory exhaustion
- Validate template complexity before execution
- Minimize data exposed in bindings
- Sanitize template output
- Audit custom filters and tags
- Keep the library updated
- Monitor for suspicious template patterns

❌ **DON'T**:
- Trust user-provided templates without limits
- Construct templates from untrusted data
- Expose sensitive data in bindings
- Register unsafe custom filters/tags
- Allow unbounded template execution
- Disable output escaping without careful consideration

## Reporting Vulnerabilities

If you discover a security vulnerability in this library, please report it by:

1. **Opening a GitHub Issue**: [Create an issue](https://github.com/osteele/liquid/issues/new) with the "security" label
   - Provide a detailed description of the vulnerability
   - Include steps to reproduce
   - If possible, provide a proof of concept

2. **For sensitive vulnerabilities**: If you believe public disclosure would be harmful, please contact the maintainer directly through GitHub before creating a public issue.

Please include:
- Description of the vulnerability
- Affected versions
- Steps to reproduce
- Potential impact
- Suggested remediation (if any)

## Additional Resources

- [Shopify Liquid Documentation](https://shopify.github.io/liquid/)
- [OWASP Template Injection](https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/18-Testing_for_Server-side_Template_Injection)
- [OWASP XSS Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)

## Version History

- **2025-01-08**: Initial security documentation created (addresses [#35](https://github.com/osteele/liquid/issues/35))
  - Added comprehensive comparison with Ruby implementation
  - Documented security guarantees, limitations, and DoS vulnerabilities
  - Provided production deployment recommendations with code examples
  - Updated to reflect FRender capabilities for timeout and output size limiting
  - Co-authored by Claude Code

## License

This security documentation is provided under the same MIT license as the rest of the project.
