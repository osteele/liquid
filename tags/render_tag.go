package tags

import (
	"fmt"
	"io"
	"maps"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

type renderParam struct {
	name  string
	value expressions.Expression
}

type renderArgs struct {
	templateName expressions.Expression
	params       []renderParam
	withValue    expressions.Expression
	withAlias    string
	forValue     expressions.Expression
	forAlias     string
}

type isolatedFileRenderer interface {
	RenderFileIsolated(string, map[string]any) (string, error)
}

type isolatedFileWriter interface {
	RenderFileIsolatedTo(io.Writer, string, map[string]any) error
}

// parseRenderArgs parses the arguments of a render tag.
func parseRenderArgs(source string) (*renderArgs, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("render tag requires a template name")
	}

	templateSource, rest, err := takeRenderTemplateName(source)
	if err != nil {
		return nil, err
	}

	templateExpr, err := expressions.Parse(templateSource)
	if err != nil {
		return nil, fmt.Errorf("invalid template name: %w", err)
	}

	args := &renderArgs{templateName: templateExpr}
	rest = trimOptionalComma(rest)
	if rest == "" {
		return args, nil
	}

	var valueSource, alias, paramSource string
	if modifierSource, ok := consumeRenderKeyword(rest, "with"); ok {
		valueSource, alias, paramSource, err = parseRenderModifier(modifierSource)
		if err != nil {
			return nil, fmt.Errorf("invalid 'with' argument: %w", err)
		}
		args.withValue, err = expressions.Parse(valueSource)
		if err != nil {
			return nil, fmt.Errorf("invalid 'with' value: %w", err)
		}
		args.withAlias = alias
		rest = paramSource
	} else if modifierSource, ok := consumeRenderKeyword(rest, "for"); ok {
		valueSource, alias, paramSource, err = parseRenderModifier(modifierSource)
		if err != nil {
			return nil, fmt.Errorf("invalid 'for' argument: %w", err)
		}
		args.forValue, err = expressions.Parse(valueSource)
		if err != nil {
			return nil, fmt.Errorf("invalid 'for' value: %w", err)
		}
		args.forAlias = alias
		rest = paramSource
	}

	args.params, err = parseRenderParams(rest)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func takeRenderTemplateName(source string) (string, string, error) {
	if source[0] == '\'' || source[0] == '"' {
		end := findClosingQuote(source, 0)
		if end < 0 {
			return "", "", fmt.Errorf("unclosed quote in template name")
		}

		return source[:end+1], strings.TrimSpace(source[end+1:]), nil
	}

	end := strings.IndexFunc(source, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	if end < 0 {
		return source, "", nil
	}

	return source[:end], strings.TrimSpace(source[end:]), nil
}

func trimOptionalComma(source string) string {
	source = strings.TrimSpace(source)
	if strings.HasPrefix(source, ",") {
		return strings.TrimSpace(source[1:])
	}

	return source
}

func consumeRenderKeyword(source, keyword string) (string, bool) {
	if !strings.HasPrefix(source, keyword) {
		return source, false
	}
	rest := source[len(keyword):]
	if rest == "" {
		return "", true
	}
	r, _ := utf8.DecodeRuneInString(rest)
	if !unicode.IsSpace(r) {
		return source, false
	}

	return strings.TrimSpace(rest), true
}

func parseRenderModifier(source string) (value, alias, params string, err error) {
	if source == "" {
		return "", "", "", fmt.Errorf("missing value")
	}

	asIndex, aliasIndex := findTopLevelAs(source)
	commaIndex := findTopLevelParamComma(source)
	if asIndex >= 0 && (commaIndex < 0 || asIndex < commaIndex) {
		value = strings.TrimSpace(source[:asIndex])
		aliasSource := strings.TrimSpace(source[aliasIndex:])
		var aliasEnd int
		alias, aliasEnd = readRenderIdentifier(aliasSource)
		if alias == "" {
			return "", "", "", fmt.Errorf("missing alias after 'as'")
		}

		rest := strings.TrimSpace(aliasSource[aliasEnd:])
		if rest != "" && !strings.HasPrefix(rest, ",") {
			return "", "", "", fmt.Errorf("unexpected text after alias: %s", rest)
		}
		params = trimOptionalComma(rest)
	} else if commaIndex >= 0 {
		value = strings.TrimSpace(source[:commaIndex])
		params = strings.TrimSpace(source[commaIndex+1:])
	} else {
		value = strings.TrimSpace(source)
	}

	if value == "" {
		return "", "", "", fmt.Errorf("missing value")
	}

	return value, alias, params, nil
}

func findTopLevelAs(source string) (int, int) {
	quote := byte(0)
	depth := 0
	for i := 0; i < len(source); i++ {
		ch := source[i]
		if quote != 0 {
			if ch == quote && !isEscaped(source, i) {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		switch ch {
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		}
		if depth != 0 || !unicode.IsSpace(rune(ch)) {
			continue
		}

		keywordStart := i
		for keywordStart < len(source) && unicode.IsSpace(rune(source[keywordStart])) {
			keywordStart++
		}
		if !strings.HasPrefix(source[keywordStart:], "as") {
			continue
		}
		afterKeyword := keywordStart + len("as")
		if afterKeyword < len(source) && unicode.IsSpace(rune(source[afterKeyword])) {
			return i, afterKeyword
		}
	}

	return -1, -1
}

func findTopLevelParamComma(source string) int {
	quote := byte(0)
	depth := 0
	for i := 0; i < len(source); i++ {
		ch := source[i]
		if quote != 0 {
			if ch == quote && !isEscaped(source, i) {
				quote = 0
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			continue
		}
		switch ch {
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 && hasRenderParamPrefix(source[i+1:]) {
				return i
			}
		}
	}

	return -1
}

func hasRenderParamPrefix(source string) bool {
	source = strings.TrimSpace(source)
	name, end := readRenderIdentifier(source)
	if name == "" {
		return false
	}

	return strings.HasPrefix(strings.TrimSpace(source[end:]), ":")
}

func parseRenderParams(source string) ([]renderParam, error) {
	var params []renderParam
	source = strings.TrimSpace(source)
	for source != "" {
		name, nameEnd := readRenderIdentifier(source)
		if name == "" {
			return nil, fmt.Errorf("invalid parameter format (expected 'key: value'): %s", source)
		}

		rest := strings.TrimSpace(source[nameEnd:])
		if !strings.HasPrefix(rest, ":") {
			return nil, fmt.Errorf("invalid parameter format (expected 'key: value'): %s", source)
		}
		rest = strings.TrimSpace(rest[1:])
		if rest == "" {
			return nil, fmt.Errorf("missing parameter value for '%s'", name)
		}

		commaIndex := findTopLevelParamComma(rest)
		valueSource := rest
		if commaIndex >= 0 {
			valueSource = strings.TrimSpace(rest[:commaIndex])
			source = strings.TrimSpace(rest[commaIndex+1:])
		} else {
			source = ""
		}

		value, err := expressions.Parse(valueSource)
		if err != nil {
			return nil, fmt.Errorf("invalid parameter value for '%s': %w", name, err)
		}

		replaced := false
		for i := range params {
			if params[i].name == name {
				params[i].value = value
				replaced = true
				break
			}
		}
		if !replaced {
			params = append(params, renderParam{name: name, value: value})
		}
	}

	return params, nil
}

func readRenderIdentifier(source string) (string, int) {
	end := 0
	for index, r := range source {
		valid := unicode.IsLetter(r) || r == '_' || (index > 0 && (unicode.IsDigit(r) || r == '-'))
		if !valid {
			break
		}
		end = index + utf8.RuneLen(r)
	}
	if end == 0 {
		return "", 0
	}

	return source[:end], end
}

func findClosingQuote(source string, start int) int {
	quote := source[start]
	for i := start + 1; i < len(source); i++ {
		if source[i] == quote && !isEscaped(source, i) {
			return i
		}
	}

	return -1
}

func isEscaped(source string, index int) bool {
	backslashes := 0
	for i := index - 1; i >= 0 && source[i] == '\\'; i-- {
		backslashes++
	}

	return backslashes%2 == 1
}

func renderTag(source string) (func(io.Writer, render.Context) error, error) {
	args, err := parseRenderArgs(source)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, ctx render.Context) error {
		templateNameValue, err := ctx.Evaluate(args.templateName)
		if err != nil {
			return err
		}
		templateName, ok := templateNameValue.(string)
		if !ok {
			return ctx.Errorf("render requires a string template name; got %T", templateNameValue)
		}

		fileRenderer, ok := ctx.(isolatedFileRenderer)
		if !ok {
			return ctx.Errorf("render requires isolated file rendering support")
		}

		filename, err := resolveTemplatePath(ctx.SourceFile(), templateName)
		if err != nil {
			return ctx.WrapError(err)
		}
		alias := args.withAlias
		if alias == "" {
			alias = args.forAlias
		}
		if alias == "" {
			alias = defaultRenderAlias(templateName)
		}

		params, err := evaluateRenderParams(ctx, args.params)
		if err != nil {
			return err
		}
		if args.forValue != nil {
			return renderFor(w, ctx, fileRenderer, filename, alias, params, args.forValue)
		}

		scope := maps.Clone(params)
		if args.withValue != nil {
			var value any
			value, err = ctx.Evaluate(args.withValue)
			if err != nil {
				return err
			}
			scope[alias] = value
		}

		if fw, ok := fileRenderer.(isolatedFileWriter); ok {
			return fw.RenderFileIsolatedTo(w, filename, scope)
		}

		s, err := fileRenderer.RenderFileIsolated(filename, scope)
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, s)

		return err
	}, nil
}

func defaultRenderAlias(templateName string) string {
	base := path.Base(filepath.ToSlash(templateName))
	extension := path.Ext(base)
	if extension != "" && extension != base {
		return strings.TrimSuffix(base, extension)
	}

	return base
}

func evaluateRenderParams(ctx render.Context, params []renderParam) (map[string]any, error) {
	values := make(map[string]any, len(params))
	for _, param := range params {
		value, err := ctx.Evaluate(param.value)
		if err != nil {
			return nil, fmt.Errorf("error evaluating parameter '%s': %w", param.name, err)
		}
		values[param.name] = value
	}

	return values, nil
}

func renderFor(
	w io.Writer,
	ctx render.Context,
	fileRenderer isolatedFileRenderer,
	filename string,
	alias string,
	params map[string]any,
	collectionExpr expressions.Expression,
) error {
	collection, err := ctx.Evaluate(collectionExpr)
	if err != nil {
		return err
	}

	items, ok := renderItems(collection)
	if !ok {
		return ctx.Errorf("'for' parameter must be an array; got %T", collection)
	}

	for i := 0; i < items.Len(); i++ {
		scope := maps.Clone(params)
		scope[alias] = items.Index(i)
		scope["forloop"] = map[string]any{
			"first":   i == 0,
			"last":    i == items.Len()-1,
			"index":   i + 1,
			"index0":  i,
			"length":  items.Len(),
			"rindex":  items.Len() - i,
			"rindex0": items.Len() - i - 1,
		}

		if fw, ok := fileRenderer.(isolatedFileWriter); ok {
			if err := fw.RenderFileIsolatedTo(w, filename, scope); err != nil {
				return err
			}
			continue
		}

		s, err := fileRenderer.RenderFileIsolated(filename, scope)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
	}

	return nil
}

func renderItems(value any) (iterable, bool) {
	if value == nil {
		return nil, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return nil, false
	}

	return sliceWrapper(rv), true
}
