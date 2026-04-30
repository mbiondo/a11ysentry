package web

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// cssinjsColorRe matches color/background-color declarations inside CSS-in-JS
// template literals. Supports #rrggbb, #rgb, and rgb(r,g,b) forms.
var cssinjsColorRe = regexp.MustCompile(
	`(?i)(color|background-color)\s*:\s*(#[0-9a-fA-F]{3,6}|rgb\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*\))`)

// jsConstColorRe matches JS/TS constant declarations with hex color values.
//   const PRIMARY = '#1a2b3c'
//   const BG_DARK = "#0f172a"
var jsConstColorRe = regexp.MustCompile(
	`(?m)const\s+(\w+)\s*=\s*['"](\#[0-9a-fA-F]{3,6})['"]`)

// styledTagRe detects styled-components tagged template literal openers.
var styledTagRe = regexp.MustCompile(
	"(?:styled(?:\\.\\w+|\\(\\w+\\))|css|createGlobalStyle|keyframes)\\s*`")

// LoadCSSinJS scans a JS/TS/TSX/JSX file for CSS-in-JS color values and
// registers them in the adapter for later contrast resolution.
//
// Two-pass strategy:
//  1. JS const assignments: const FOO = '#hex' → stored in jsConsts + cssVars
//  2. Template literal blocks (styled-components, emotion css``) →
//     color/bg declarations extracted and stored in cssVars keyed by
//     a synthetic name so inline style resolution can pick them up.
//
// This is best-effort: only registers colors when property + parseable value
// are both present. Unresolvable interpolations are skipped (no false positives).
func (a *htmlAdapter) LoadCSSinJS(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	src := string(data)

	// Pass 1: collect JS color constants.
	jsConsts := make(map[string]string) // identifier → hex
	for _, m := range jsConstColorRe.FindAllStringSubmatch(src, -1) {
		name := m[1]
		hex := strings.ToLower(m[2])
		jsConsts[name] = hex
		a.cssVars["--js-"+name] = hex
	}

	// Pass 2: extract styled-components / css`` template literal blocks.
	if styledTagRe.MatchString(src) {
		blocks := extractTemplateLiterals(src)
		for i, block := range blocks {
			if !strings.Contains(block, ":") || !strings.Contains(block, ";") {
				continue
			}
			resolved := resolveInterpolations(block, jsConsts)
			for _, m := range cssinjsColorRe.FindAllStringSubmatch(resolved, -1) {
				prop := strings.TrimSpace(strings.ToLower(m[1]))
				val := strings.TrimSpace(m[2])
				hex := a.normalizeColor(val)
				if hex == "" {
					continue
				}
				// Key: synthetic name scoped to this file+block index.
				key := "--cssinjs-" + strconv.Itoa(i) + "-" + prop
				a.cssVars[key] = hex
			}
		}
	}

	return nil
}

// extractTemplateLiterals returns all backtick-delimited string contents from src.
// Handles ${} interpolation nesting (shallow — does not recurse into nested templates).
func extractTemplateLiterals(src string) []string {
	var blocks []string
	i := 0
	for i < len(src) {
		if src[i] != '`' {
			i++
			continue
		}
		// Start of a template literal.
		i++
		start := i
		depth := 0
		for i < len(src) {
			switch {
			case src[i] == '\\':
				i += 2 // skip escaped char
			case src[i] == '$' && i+1 < len(src) && src[i+1] == '{':
				depth++
				i += 2
			case src[i] == '}' && depth > 0:
				depth--
				i++
			case src[i] == '`' && depth == 0:
				blocks = append(blocks, src[start:i])
				i++
				goto nextBlock
			default:
				i++
			}
		}
	nextBlock:
	}
	return blocks
}

// resolveInterpolations replaces ${IDENT} in block with the matching jsConsts value.
func resolveInterpolations(block string, consts map[string]string) string {
	re := regexp.MustCompile(`\$\{(\w+)\}`)
	return re.ReplaceAllStringFunc(block, func(match string) string {
		m := re.FindStringSubmatch(match)
		if len(m) > 1 {
			if val, ok := consts[m[1]]; ok {
				return val
			}
		}
		return match
	})
}
