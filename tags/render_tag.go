package tags

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

// renderArgs represents the parsed arguments for a render tag
type renderArgs struct {
	templateName expressions.Expression
	params       map[string]expressions.Expression
	withValue    expressions.Expression // for "with" syntax
	withAlias    string                 // for "with ... as alias" syntax
	forValue     expressions.Expression // for "for" syntax
	forAlias     string                 // for "for ... as alias" syntax
}

// parseRenderArgs parses the arguments of a {% render %} tag
// Supports the following syntaxes:
//
//	{% render 'template' %}
//	{% render 'template', key: value, key2: value2 %}
//	{% render 'template' with object %}
//	{% render 'template' with object as name %}
//	{% render 'template' for array %}
//	{% render 'template' for array as item %}
//	{% render 'template' for array as item, key: value %}
func parseRenderArgs(source string) (*renderArgs, error) {
	args := &renderArgs{
		params: make(map[string]expressions.Expression),
	}

	// Trim whitespace
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("render tag requires a template name")
	}

	// Parse template name (first argument)
	// Find the end of the template name (could be a string or variable)
	var templateNameStr string
	var rest string

	// Check if it starts with a quote (string literal)
	if strings.HasPrefix(source, "'") || strings.HasPrefix(source, "\"") {
		quote := source[0]
		endQuote := strings.IndexByte(source[1:], quote)
		if endQuote == -1 {
			return nil, fmt.Errorf("unclosed quote in template name")
		}
		templateNameStr = source[0 : endQuote+2] // include both quotes
		rest = strings.TrimSpace(source[endQuote+2:])
	} else {
		// Variable name (no quotes)
		parts := strings.Fields(source)
		if len(parts) == 0 {
			return nil, fmt.Errorf("render tag requires a template name")
		}
		// Find where the template name ends (before comma, 'with', 'for', or end)
		templateNameStr = parts[0]
		// Remove the template name from source
		rest = strings.TrimSpace(source[len(templateNameStr):])
	}

	// Parse the template name as an expression
	templateExpr, err := expressions.Parse(templateNameStr)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}
	args.templateName = templateExpr

	// Remove leading comma if present
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, ",") {
		rest = strings.TrimSpace(rest[1:])
	}

	// Parse the rest of the arguments
	if rest == "" {
		return args, nil
	}

	// Check for 'with' or 'for' keywords
	if strings.HasPrefix(rest, "with ") {
		// Parse "with" syntax: with object [as alias] [, params]
		rest = strings.TrimSpace(rest[5:]) // remove "with "

		// Find the end of the object expression (before 'as' or ',')
		withEnd := len(rest)
		asIndex := strings.Index(rest, " as ")
		commaIndex := strings.IndexByte(rest, ',')

		if asIndex != -1 && (commaIndex == -1 || asIndex < commaIndex) {
			withEnd = asIndex
		} else if commaIndex != -1 {
			withEnd = commaIndex
		}

		withValueStr := strings.TrimSpace(rest[:withEnd])
		withExpr, err := expressions.Parse(withValueStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'with' value: %w", err)
		}
		args.withValue = withExpr

		rest = strings.TrimSpace(rest[withEnd:])

		// Check for 'as alias'
		if strings.HasPrefix(rest, "as ") {
			rest = strings.TrimSpace(rest[3:]) // remove "as "
			// Get the alias name (before comma or end)
			aliasEnd := strings.IndexByte(rest, ',')
			if aliasEnd == -1 {
				args.withAlias = strings.TrimSpace(rest)
				rest = ""
			} else {
				args.withAlias = strings.TrimSpace(rest[:aliasEnd])
				rest = strings.TrimSpace(rest[aliasEnd+1:])
			}
		}
	} else if strings.HasPrefix(rest, "for ") {
		// Parse "for" syntax: for array [as item] [, params]
		rest = strings.TrimSpace(rest[4:]) // remove "for "

		// Find the end of the array expression (before 'as' or ',')
		forEnd := len(rest)
		asIndex := strings.Index(rest, " as ")
		commaIndex := strings.IndexByte(rest, ',')

		if asIndex != -1 && (commaIndex == -1 || asIndex < commaIndex) {
			forEnd = asIndex
		} else if commaIndex != -1 {
			forEnd = commaIndex
		}

		forValueStr := strings.TrimSpace(rest[:forEnd])
		forExpr, err := expressions.Parse(forValueStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'for' value: %w", err)
		}
		args.forValue = forExpr

		rest = strings.TrimSpace(rest[forEnd:])

		// Check for 'as alias'
		if strings.HasPrefix(rest, "as ") {
			rest = strings.TrimSpace(rest[3:]) // remove "as "
			// Get the alias name (before comma or end)
			aliasEnd := strings.IndexByte(rest, ',')
			if aliasEnd == -1 {
				args.forAlias = strings.TrimSpace(rest)
				rest = ""
			} else {
				args.forAlias = strings.TrimSpace(rest[:aliasEnd])
				rest = strings.TrimSpace(rest[aliasEnd+1:])
			}
		}
	}

	// Parse remaining parameters (key: value pairs)
	if rest != "" {
		// Remove leading comma if present
		rest = strings.TrimSpace(rest)
		if strings.HasPrefix(rest, ",") {
			rest = strings.TrimSpace(rest[1:])
		}

		// Parse key-value pairs
		if err := parseKeyValuePairs(rest, args.params); err != nil {
			return nil, err
		}
	}

	return args, nil
}

// parseKeyValuePairs parses comma-separated key: value pairs
func parseKeyValuePairs(source string, params map[string]expressions.Expression) error {
	if source == "" {
		return nil
	}

	// Simple parser for key: value pairs
	// This is a basic implementation - a more robust parser would handle nested structures
	parts := splitPreservingQuotes(source, ',')

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split on first colon
		colonIndex := strings.IndexByte(part, ':')
		if colonIndex == -1 {
			return fmt.Errorf("invalid parameter format (expected 'key: value'): %s", part)
		}

		key := strings.TrimSpace(part[:colonIndex])
		valueStr := strings.TrimSpace(part[colonIndex+1:])

		// Validate key (must be a valid identifier)
		if !isValidIdentifier(key) {
			return fmt.Errorf("invalid parameter name: %s", key)
		}

		// Parse value as expression
		valueExpr, err := expressions.Parse(valueStr)
		if err != nil {
			return fmt.Errorf("invalid parameter value for '%s': %w", key, err)
		}

		params[key] = valueExpr
	}

	return nil
}

// splitPreservingQuotes splits a string by delimiter, but preserves quoted strings
func splitPreservingQuotes(s string, delimiter byte) []string {
	var result []string
	var current strings.Builder
	inQuote := false
	var quoteChar byte

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if (ch == '"' || ch == '\'') && (i == 0 || s[i-1] != '\\') {
			if !inQuote {
				inQuote = true
				quoteChar = ch
			} else if ch == quoteChar {
				inQuote = false
			}
			current.WriteByte(ch)
		} else if ch == delimiter && !inQuote {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// isValidIdentifier checks if a string is a valid identifier (alphanumeric + underscore)
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || (i > 0 && ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

// renderTag implements the {% render %} tag with isolated scope
// Syntax:
//
//	{% render 'template' %}
//	{% render 'template', key: value %}
//	{% render 'template' with object %}
//	{% render 'template' with object as name %}
//	{% render 'template' for array %}
//	{% render 'template' for array as item %}
func renderTag(source string) (func(io.Writer, render.Context) error, error) {
	args, err := parseRenderArgs(source)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, ctx render.Context) error {
		// Evaluate template name
		templateNameValue, err := ctx.Evaluate(args.templateName)
		if err != nil {
			return err
		}

		templateName, ok := templateNameValue.(string)
		if !ok {
			return ctx.Errorf("render requires a string template name; got %T", templateNameValue)
		}

		// Build the file path
		filename := filepath.Join(filepath.Dir(ctx.SourceFile()), templateName)

		// Handle 'for' parameter (render for each item in array)
		if args.forValue != nil {
			return renderFor(w, ctx, filename, args)
		}

		// Build isolated scope with parameters
		isolatedScope, err := buildIsolatedScope(ctx, args)
		if err != nil {
			return err
		}

		// Render with isolated scope
		s, err := renderFileIsolated(ctx, filename, isolatedScope)
		if err != nil {
			return err
		}

		_, err = io.WriteString(w, s)
		return err
	}, nil
}

// buildIsolatedScope creates an isolated scope with only the passed parameters
func buildIsolatedScope(ctx render.Context, args *renderArgs) (map[string]any, error) {
	scope := make(map[string]any)

	// Add 'with' parameter if present
	if args.withValue != nil {
		value, err := ctx.Evaluate(args.withValue)
		if err != nil {
			return nil, err
		}

		// Use alias if provided, otherwise use template filename as key
		if args.withAlias != "" {
			scope[args.withAlias] = value
		} else {
			// Default behavior: make the object available by template name
			// This is simplified - Shopify uses the template filename
			scope["object"] = value
		}
	}

	// Add explicit parameters
	for key, valueExpr := range args.params {
		value, err := ctx.Evaluate(valueExpr)
		if err != nil {
			return nil, fmt.Errorf("error evaluating parameter '%s': %w", key, err)
		}
		scope[key] = value
	}

	return scope, nil
}

// renderFor renders the template for each item in an array
func renderFor(w io.Writer, ctx render.Context, filename string, args *renderArgs) error {
	// Evaluate the array
	arrayValue, err := ctx.Evaluate(args.forValue)
	if err != nil {
		return err
	}

	// Convert to slice
	items, ok := convertToSlice(arrayValue)
	if !ok {
		return ctx.Errorf("'for' parameter must be an array; got %T", arrayValue)
	}

	// Determine the variable name for each item
	itemName := args.forAlias
	if itemName == "" {
		// Default: use template filename without extension as variable name
		// This is simplified - Shopify's behavior is more complex
		itemName = "item"
	}

	// Render for each item
	for i, item := range items {
		// Build scope with forloop object
		scope := make(map[string]any)

		// Add the current item
		scope[itemName] = item

		// Add forloop object (Shopify-compatible)
		scope["forloop"] = map[string]any{
			"first":   i == 0,
			"last":    i == len(items)-1,
			"index":   i + 1, // 1-indexed
			"index0":  i,     // 0-indexed
			"length":  len(items),
			"rindex":  len(items) - i,     // reverse index (1-indexed)
			"rindex0": len(items) - i - 1, // reverse index (0-indexed)
		}

		// Add explicit parameters
		for key, valueExpr := range args.params {
			value, err := ctx.Evaluate(valueExpr)
			if err != nil {
				return fmt.Errorf("error evaluating parameter '%s': %w", key, err)
			}
			scope[key] = value
		}

		// Render with isolated scope
		s, err := renderFileIsolated(ctx, filename, scope)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
	}

	return nil
}

// convertToSlice attempts to convert a value to []any
func convertToSlice(v any) ([]any, bool) {
	if v == nil {
		return nil, false
	}

	switch arr := v.(type) {
	case []any:
		return arr, true
	default:
		// Try reflection for other slice types
		return nil, false
	}
}

// renderFileIsolated renders a file with an isolated scope (no parent variables)
// Uses the RenderFileIsolated method which provides true variable isolation
func renderFileIsolated(ctx render.Context, filename string, isolatedScope map[string]any) (string, error) {
	return ctx.RenderFileIsolated(filename, isolatedScope)
}
