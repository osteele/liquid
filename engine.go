package liquid

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
	"github.com/osteele/liquid/tags"
)

// An Engine parses template source into renderable text.
//
// An engine can be configured with additional filters and tags.
//
// Configuration methods (RegisterTag, RegisterFilter, SetGlobals, etc.) must be
// called before the engine is first used for parsing or rendering. Calling them
// after the first parse panics with a clear message, preventing data races on
// the shared grammar and filter maps.
type Engine struct {
	cfg    render.Config
	cache  *sync.Map   // nil = disabled; key: string (source), value: *Template
	frozen atomic.Bool // set to true on first parse; guards config mutations
}

// freeze marks the engine as in-use. Called at the start of every parse entry
// point. Subsequent configuration mutations will panic.
func (e *Engine) freeze() {
	e.frozen.Store(true)
}

// checkNotFrozen panics if the engine has already been used for parsing.
// Call this at the top of every configuration-mutation method.
func (e *Engine) checkNotFrozen(method string) {
	if e.frozen.Load() {
		panic("liquid: " + method + "() called after the engine has been used for parsing; configure the engine before first use")
	}
}

// NewEngine returns a new Engine.
func NewEngine() *Engine {
	e := Engine{cfg: render.NewConfig()}
	filters.AddStandardFilters(&e.cfg)
	tags.AddStandardTags(&e.cfg)

	return &e
}

// NewBasicEngine returns a new Engine without the standard filters or tags.
func NewBasicEngine() *Engine {
	return &Engine{cfg: render.NewConfig()}
}

// RegisterBlock defines a block e.g. {% tag %}…{% endtag %}.
func (e *Engine) RegisterBlock(name string, td Renderer) {
	e.checkNotFrozen("RegisterBlock")
	e.cfg.AddBlock(name).Renderer(func(w io.Writer, ctx render.Context) error {
		s, err := td(ctx)
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, s)

		return err
	})
}

// RegisterFilter defines a Liquid filter, for use as `{{ value | my_filter }}` or `{{ value | my_filter: arg }}`.
//
// A filter is a function that takes at least one input, and returns one or two outputs.
// If it returns two outputs, the second must have type error.
//
// Examples:
//
// * https://github.com/osteele/liquid/blob/main/filters/standard_filters.go
//
// * https://github.com/osteele/gojekyll/blob/master/filters/filters.go
func (e *Engine) RegisterFilter(name string, fn any) {
	e.checkNotFrozen("RegisterFilter")
	e.cfg.AddFilter(name, fn)
}

// RegisterTag defines a tag e.g. {% tag %}.
//
// Further examples are in https://github.com/osteele/gojekyll/blob/master/tags/tags.go
func (e *Engine) RegisterTag(name string, td Renderer) {
	e.checkNotFrozen("RegisterTag")
	// For simplicity, don't expose the two stage parsing/rendering process to clients.
	// Client tags do everything at runtime.
	e.cfg.AddTag(name, func(_ string) (func(io.Writer, render.Context) error, error) {
		return func(w io.Writer, ctx render.Context) error {
			s, err := td(ctx)
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, s)

			return err
		}, nil
	})
}

func (e *Engine) RegisterTemplateStore(templateStore render.TemplateStore) {
	e.checkNotFrozen("RegisterTemplateStore")
	e.cfg.TemplateStore = templateStore
}

// SetGlobals sets variables that are accessible in every rendering context,
// including isolated sub-contexts created by the {% render %} tag.
// Scope bindings passed to Render take priority over globals when keys conflict.
func (e *Engine) SetGlobals(globals map[string]any) {
	e.checkNotFrozen("SetGlobals")
	e.cfg.Globals = globals
}

// GetGlobals returns the engine-level global variables.
func (e *Engine) GetGlobals() map[string]any {
	return e.cfg.Globals
}

// StrictVariables causes the renderer to error when the template contains an undefined variable.
func (e *Engine) StrictVariables() {
	e.checkNotFrozen("StrictVariables")
	e.cfg.StrictVariables = true
}

// LaxFilters causes the renderer to silently pass through the input value
// when the template contains an undefined filter, matching Shopify Liquid behavior.
// By default, undefined filters cause an error.
func (e *Engine) LaxFilters() {
	e.checkNotFrozen("LaxFilters")
	e.cfg.LaxFilters = true
}

// EnableJekyllExtensions enables Jekyll-specific extensions to Liquid.
// This includes support for dot notation in assign tags (e.g., {% assign page.canonical_url = value %}).
// Note: This is not part of the Shopify Liquid standard but is used in Jekyll and Gojekyll.
func (e *Engine) EnableJekyllExtensions() {
	e.checkNotFrozen("EnableJekyllExtensions")
	e.cfg.JekyllExtensions = true
}

// SetTrimTagLeft controls whether whitespace to the left of every {% tag %} is
// automatically trimmed, equivalent to adding {%- to every tag.
func (e *Engine) SetTrimTagLeft(v bool) {
	e.checkNotFrozen("SetTrimTagLeft")
	e.cfg.TrimTagLeft = v
}

// SetTrimTagRight controls whether whitespace to the right of every {% tag %} is
// automatically trimmed, equivalent to adding -%} to every tag.
func (e *Engine) SetTrimTagRight(v bool) {
	e.checkNotFrozen("SetTrimTagRight")
	e.cfg.TrimTagRight = v
}

// SetTrimOutputLeft controls whether whitespace to the left of every {{ output }}
// is automatically trimmed, equivalent to adding {{- to every output expression.
func (e *Engine) SetTrimOutputLeft(v bool) {
	e.checkNotFrozen("SetTrimOutputLeft")
	e.cfg.TrimOutputLeft = v
}

// SetTrimOutputRight controls whether whitespace to the right of every {{ output }}
// is automatically trimmed, equivalent to adding -}} to every output expression.
func (e *Engine) SetTrimOutputRight(v bool) {
	e.checkNotFrozen("SetTrimOutputRight")
	e.cfg.TrimOutputRight = v
}

// SetGreedy controls whether whitespace trimming removes all consecutive blank
// characters including newlines (true, the default), or only trims inline
// blanks (space/tab) plus at most one newline (false).
func (e *Engine) SetGreedy(v bool) {
	e.checkNotFrozen("SetGreedy")
	e.cfg.Greedy = v
}

// ParseTemplate creates a new Template using the engine configuration.
func (e *Engine) ParseTemplate(source []byte) (*Template, SourceError) {
	e.freeze()
	return newTemplate(&e.cfg, source, "", 1)
}

// ParseString creates a new Template using the engine configuration.
// If the template cache is enabled (EnableCache), previously parsed templates
// are returned from the cache without re-parsing.
func (e *Engine) ParseString(source string) (*Template, SourceError) {
	if e.cache != nil {
		if cached, ok := e.cache.Load(source); ok {
			return cached.(*Template), nil
		}
	}
	tpl, err := e.ParseTemplate([]byte(source))
	if err != nil {
		return nil, err
	}
	if e.cache != nil {
		e.cache.Store(source, tpl)
	}
	return tpl, nil
}

// ParseTemplateLocation is the same as ParseTemplate, except that the source location is used
// for error reporting and for the {% include %} tag.
//
// The path and line number are used for error reporting.
// The path is also the reference for relative pathnames in the {% include %} tag.
func (e *Engine) ParseTemplateLocation(source []byte, path string, line int) (*Template, SourceError) {
	e.freeze()
	return newTemplate(&e.cfg, source, path, line)
}

// ParseAndRender parses and then renders the template.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. For example, adding WithStrictVariables() enables strict variable
// checking even if StrictVariables was not called on the engine.
func (e *Engine) ParseAndRender(source []byte, b Bindings, opts ...RenderOption) ([]byte, SourceError) {
	tpl, err := e.ParseString(string(source))
	if err != nil {
		return nil, err
	}

	return tpl.Render(b, opts...)
}

// ParseAndFRender parses and then renders the template into w.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. See ParseAndRender for details.
func (e *Engine) ParseAndFRender(w io.Writer, source []byte, b Bindings, opts ...RenderOption) SourceError {
	tpl, err := e.ParseString(string(source))
	if err != nil {
		return err
	}

	return tpl.FRender(w, b, opts...)
}

// ParseAndRenderString is a convenience wrapper for ParseAndRender, that takes string input and returns a string.
//
// RenderOptions can be passed to override engine-level settings for this
// call only. See ParseAndRender for details.
func (e *Engine) ParseAndRenderString(source string, b Bindings, opts ...RenderOption) (string, SourceError) {
	bs, err := e.ParseAndRender([]byte(source), b, opts...)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

// Delims sets the action delimiters to the specified strings, to be used in subsequent calls to
// ParseTemplate, ParseTemplateLocation, ParseAndRender, or ParseAndRenderString. An empty delimiter
// stands for the corresponding default: objectLeft = {{, objectRight = }}, tagLeft = {% , tagRight = %}
func (e *Engine) Delims(objectLeft, objectRight, tagLeft, tagRight string) *Engine {
	e.checkNotFrozen("Delims")
	// Empty strings fall back to the standard defaults.
	defaults := []string{"{{", "}}", "{%", "%}"}
	delims := []string{objectLeft, objectRight, tagLeft, tagRight}
	for i, d := range delims {
		if d == "" {
			delims[i] = defaults[i]
		}
	}
	e.cfg.Delims = delims
	return e
}

// ParseTemplateAndCache is the same as ParseTemplateLocation, except that the
// source location is used for error reporting and for the {% include %} tag.
// If parsing is successful, provided source is then cached, and can be retrieved
// by {% include %} tags, as long as there is not a real file in the provided path.
//
// The path and line number are used for error reporting.
// The path is also the reference for relative pathnames in the {% include %} tag.
func (e *Engine) ParseTemplateAndCache(source []byte, path string, line int) (*Template, SourceError) {
	t, err := e.ParseTemplateLocation(source, path, line)
	if err != nil {
		return t, err
	}

	e.cfg.Cache.Store(path, source)

	return t, err
}

// SetAutoEscapeReplacer enables auto-escape functionality where the output of expression blocks ({{ ... }}) is
// passed though a render.Replacer during rendering, unless it's been marked as safe by applying the 'safe' filter.
// This filter is automatically registered when this method is called. The filter must be applied last.
// A replacer is provided for escaping HTML (see render.HtmlEscaper).
func (e *Engine) SetAutoEscapeReplacer(replacer render.Replacer) {
	e.checkNotFrozen("SetAutoEscapeReplacer")
	e.cfg.SetAutoEscapeReplacer(replacer)
}

// SetGlobalFilter sets a function that is applied to the evaluated value of every
// {{ expression }} before it is written to the output. This is analogous to Ruby
// Liquid's global_filter option.
//
// The function receives the evaluated Liquid value (string, int, float64, bool, nil, etc.)
// and returns a transformed value or an error.
//
// Example:
//
//	engine.SetGlobalFilter(func(v any) (any, error) {
//	    if s, ok := v.(string); ok {
//	        return strings.ToLower(s), nil
//	    }
//	    return v, nil
//	})
func (e *Engine) SetGlobalFilter(fn func(any) (any, error)) {
	e.checkNotFrozen("SetGlobalFilter")
	e.cfg.SetGlobalFilter(fn)
}

// RegisterTagAnalyzer registers a static analysis function for a simple tag previously
// registered with RegisterTag. The analyzer is invoked during static analysis to
// determine which variables the tag reads and which it defines in scope.
//
// Use render.NodeAnalysis.Arguments to declare variable expressions the tag reads, and
// render.NodeAnalysis.LocalScope to declare variable names the tag defines.
func (e *Engine) RegisterTagAnalyzer(name string, a render.TagAnalyzer) {
	e.checkNotFrozen("RegisterTagAnalyzer")
	e.cfg.AddTagAnalyzer(name, a)
}

// RegisterBlockAnalyzer registers a static analysis function for a block tag previously
// registered with RegisterBlock. The analyzer is invoked during static analysis.
func (e *Engine) RegisterBlockAnalyzer(name string, a render.BlockAnalyzer) {
	e.checkNotFrozen("RegisterBlockAnalyzer")
	e.cfg.AddBlockAnalyzer(name, a)
}

// UnregisterTag removes the named tag definition from the engine's configuration.
// After calling UnregisterTag the tag will no longer be recognized by subsequent
// parsing or rendering operations. The call is idempotent — unregistering a tag
// that is not registered is a no-op.
//
// Note: UnregisterTag is intentionally excluded from the frozen-engine guard.
// It is designed to be called after testing or hot-reload scenarios where the
// engine may have already been used. Callers are responsible for ensuring no
// concurrent renders are in progress when calling this method.
func (e *Engine) UnregisterTag(name string) {
	e.cfg.UnregisterTag(name)
}

// LaxTags causes unknown {% tag %} names to be silently compiled as no-ops
// (empty output) instead of producing a parse/compile error.
// Analogous to Ruby Liquid's error_mode: :lax for tag names.
// Must be called before ParseTemplate / ParseString.
func (e *Engine) LaxTags() {
	e.checkNotFrozen("LaxTags")
	e.cfg.LaxTags = true
}

// SetExceptionHandler registers a function that is called when a render-time
// error occurs instead of stopping the render. The handler receives the error
// and returns a string that is written to the output in place of the failing
// node. Rendering continues with the next node after the handler returns.
//
// This sets an engine-level default; individual render calls can override it
// with WithErrorHandler.
//
// Analogous to Ruby Liquid's exception_renderer option.
func (e *Engine) SetExceptionHandler(fn func(error) string) {
	e.checkNotFrozen("SetExceptionHandler")
	e.cfg.ExceptionHandler = fn
}

// EnableCache enables a simple in-memory template cache keyed by source string.
// When enabled, ParseString (and the convenience methods ParseAndRenderString,
// ParseAndRender, ParseAndFRender) will return cached *Template values for
// source strings that have been parsed before, avoiding redundant parsing.
//
// The cache is unbounded; call ClearCache to release cached templates.
// Useful for hot-path rendering where the same template source is used many times.
func (e *Engine) EnableCache() {
	e.checkNotFrozen("EnableCache")
	e.cache = &sync.Map{}
}

// ClearCache evicts all entries from the template cache.
// Has no effect if the cache is not enabled.
func (e *Engine) ClearCache() {
	if e.cache != nil {
		e.cache.Range(func(k, _ any) bool {
			e.cache.Delete(k)
			return true
		})
	}
}

// ParseTemplateAudit parses source in error-recovering mode and returns a
// *ParseResult containing the compiled template and all parse-time diagnostics.
//
// Unlike ParseTemplate, ParseTemplateAudit never returns a SourceError.
// All problems are captured as Diagnostic entries in ParseResult.Diagnostics,
// using the same Diagnostic type used by (*Template).RenderAudit.
//
// ParseResult.Template is non-nil when parsing produced a usable compiled
// template. Callers should check Template before rendering:
//
//	result := eng.ParseTemplateAudit(source)
//	for _, d := range result.Diagnostics {
//	    log.Printf("%s at line %d: %s", d.Severity, d.Range.Start.Line, d.Message)
//	}
//	if result.Template != nil {
//	    output, err := result.Template.RenderString(binds)
//	    _ = output; _ = err
//	}
//
// Diagnostics that may appear:
//
//   - "unclosed-tag" (error): a block tag was opened but never closed;
//     ParseResult.Template is nil when this occurs.
//   - "unexpected-tag" (error): a closing or clause tag appeared without a
//     matching open block; ParseResult.Template is nil when this occurs.
//   - "syntax-error" (error): invalid expression inside {{ }} or tag args.
//   - "undefined-filter" (error): a filter name used is not registered.
//   - "empty-block" (info): a block tag has no content in any branch.
func (e *Engine) ParseTemplateAudit(source []byte) *ParseResult {
	e.freeze()

	loc := parser.SourceLoc{Pathname: "", LineNo: 1}
	cr := e.cfg.CompileAudit(string(source), loc)

	// Convert internal ParseDiags to public Diagnostics.
	diags := make([]Diagnostic, 0, len(cr.Diags))
	for _, d := range cr.Diags {
		diags = append(diags, parseDiagToPublic(d))
	}

	if cr.FatalError != nil {
		return &ParseResult{Template: nil, Diagnostics: diags}
	}

	// Build a usable *Template from the compiled node.
	tpl := &Template{root: cr.Node, cfg: &e.cfg}

	// Static analysis: reuse the same walk that Validate() uses.
	staticDiags := tpl.collectValidationDiags()
	diags = append(diags, staticDiags...)

	return &ParseResult{Template: tpl, Diagnostics: diags}
}

// ParseStringAudit is the string-input convenience variant of ParseTemplateAudit.
func (e *Engine) ParseStringAudit(source string) *ParseResult {
	return e.ParseTemplateAudit([]byte(source))
}
