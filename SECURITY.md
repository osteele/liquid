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

### Currently, there are NO built-in mechanisms for:
- Execution timeouts
- Memory limits
- CPU usage limits
- Template complexity limits
- Iteration count limits

**Recommendation**: If you need to process untrusted templates, implement your own timeout and resource limiting mechanisms at the application level. See [Production Deployment Recommendations](#production-deployment-recommendations) below.

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

### 1. Timeout Protection

Wrap template execution with context timeouts:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/osteele/liquid"
)

func renderWithTimeout(engine *liquid.Engine, template string, bindings map[string]any, timeout time.Duration) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    resultChan := make(chan struct {
        output string
        err    error
    }, 1)

    go func() {
        output, err := engine.ParseAndRenderString(template, bindings)
        resultChan <- struct {
            output string
            err    error
        }{output, err}
    }()

    select {
    case result := <-resultChan:
        return result.output, result.err
    case <-ctx.Done():
        return "", fmt.Errorf("template rendering timed out after %v", timeout)
    }
}
```

**Note**: This approach has limitations - the goroutine will continue running until completion even after timeout.

### 2. Resource Limits

Run template execution in resource-limited environments:
- Use OS-level process isolation (containers, VMs)
- Set memory limits (cgroups, container limits)
- Use separate processes with ulimits

### 3. Input Validation

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

### 4. Minimal Bindings

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

### 5. Output Sanitization

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

### 6. Template Review and Approval

For sensitive applications:
- Implement a template review process
- Use version control for templates
- Audit template changes before deployment
- Consider static analysis of templates

### 7. Rate Limiting

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

- **Go**: **No built-in resource limits currently**
  - Must implement timeouts externally (as shown in [Production Deployment Recommendations](#production-deployment-recommendations))
  - No complexity scoring mechanism
  - No built-in iteration count limits
  - More challenging to implement proper resource constraints
  - Goroutine-based timeouts have limitations (goroutine continues running)

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

The **Ruby implementation is currently the more battle-tested and safer choice** due to:
- Built-in resource limiting (render_score_limit, render_length_limit)
- 15+ years of production hardening at Shopify
- Larger security-focused community
- Better-established security track record

**However**, both implementations:
- Share the same fundamental security model and core guarantees
- Require additional application-level safeguards for untrusted templates
- Are vulnerable to DoS without proper external controls
- Are equally safe for trusted templates (e.g., your own template files)

**The Go implementation is perfectly suitable when:**
- You control all templates (e.g., static site generator, internal tools)
- You can implement external resource limiting (timeouts, OS-level limits)
- You need Go's performance characteristics and deployment simplicity
- You're willing to implement more defensive programming practices

**Choose Ruby if:**
- You need built-in resource limiting without external infrastructure
- You're processing large volumes of untrusted templates
- You want the most battle-tested implementation
- You need established security guarantees

**If using the Go implementation with untrusted templates**, you **must** implement:
- External timeouts (goroutine-based as shown above)
- Resource limits (memory, CPU via OS-level controls)
- Template complexity validation before execution
- Rate limiting per user/source
- Comprehensive monitoring and alerting
- Regular security reviews of custom extensions

## Security Best Practices Summary

✅ **DO**:
- Implement timeouts for template execution
- Use resource limits (memory, CPU)
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

## License

This security documentation is provided under the same MIT license as the rest of the project.
