package web

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type htmlAdapter struct {
	platform     domain.Platform
	cssMap       map[string]map[string]string // className -> property -> value
	darkCSSMap   map[string]map[string]string // dark-mode overrides: className -> property -> value
	customColors map[string]string            // Tailwind custom token name -> hex (resolved from CSS files)
	cssVars      map[string]string            // CSS custom property name (--foo) -> hex
}

func NewHTMLAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformWebReact,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
	}
}

func NewElectronAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformElectron,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
	}
}

func NewTauriAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformTauri,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
	}
}

func (a *htmlAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	if root == nil {
		return nil, nil
	}
	// Initial ingestion starts with no inherited context.
	return a.ingestRecursive(ctx, root, "", "")
}

func (a *htmlAdapter) ingestRecursive(ctx context.Context, node *domain.FileNode, inheritedBG, inheritedFG string) ([]domain.USN, error) {
	data, err := os.ReadFile(node.FilePath)
	if err != nil {
		return nil, err
	}
	content := string(data)

	// Strip Astro/Markdown frontmatter (--- ... ---)
	// We replace it with newlines to preserve line numbers.
	cleanContent := content
	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "---")
		if endIdx != -1 {
			frontmatter := content[:endIdx+6]
			newlines := strings.Repeat("\n", strings.Count(frontmatter, "\n"))
			cleanContent = newlines + content[endIdx+6:]
		}
	}

	// Detect if this file is a full HTML document or a partial component.
	// If it's a Page Tree root, we should NOT treat it as a component.
	isComponent := !strings.Contains(strings.ToLower(cleanContent), "<html")
	if inheritedBG == "" && inheritedFG == "" {
		// Heuristic: if no context inherited, it's likely a root document.
		isComponent = false
	}

	doc, err := html.Parse(strings.NewReader(cleanContent))
	if err != nil {
		return nil, err
	}

	// Pass 1: Extract CSS from this file
	a.extractCSS(doc)

	// Pass 2: Traverse this file with inherited context from parent file
	nodes := a.traverse(doc, node.FilePath, nil, cleanContent, isComponent, inheritedBG, inheritedFG)

	// Determine the effective background and foreground to propagate to children files.
	// We look for the most specific colors defined in the current file.
	childBG := inheritedBG
	childFG := inheritedFG
	for _, n := range nodes {
		if bg, ok := n.Traits["background-color"].(string); ok && bg != "" {
			childBG = bg
		}
		if fg, ok := n.Traits["color"].(string); ok && fg != "" {
			childFG = fg
		}
	}

	var allNodes []domain.USN
	allNodes = append(allNodes, nodes...)

	// Recursively ingest children (imported components or nested layouts)
	for _, child := range node.Children {
		childNodes, err := a.ingestRecursive(ctx, child, childBG, childFG)
		if err != nil {
			return nil, err
		}
		allNodes = append(allNodes, childNodes...)
	}

	return allNodes, nil
}

func (a *htmlAdapter) extractCSS(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "style" {
		var content strings.Builder
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				content.WriteString(c.Data)
			}
		}
		a.parseCSS(content.String())
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		a.extractCSS(c)
	}
}

func (a *htmlAdapter) parseCSS(css string) {
	a.parseCSSInto(css, a.cssMap)
}

func (a *htmlAdapter) parseDarkCSS(css string) {
	a.parseCSSInto(css, a.darkCSSMap)
}

func (a *htmlAdapter) parseCSSInto(css string, target map[string]map[string]string) {
	// Simple regex parser for .class { prop: val; }
	re := regexp.MustCompile(`\.([a-zA-Z0-9_-]+)\s*\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(css, -1)

	for _, m := range matches {
		className := m[1]
		rules := m[2]

		if _, ok := target[className]; !ok {
			target[className] = make(map[string]string)
		}

		propRe := regexp.MustCompile(`(color|background-color|border-color)\s*:\s*([^;]+)`)
		pMatches := propRe.FindAllStringSubmatch(rules, -1)
		for _, pm := range pMatches {
			target[className][strings.TrimSpace(pm[1])] = strings.TrimSpace(pm[2])
		}
	}
}

func (a *htmlAdapter) traverse(n *html.Node, filename string, lines []string, fullContent string, isComponent bool, inheritedBG string, inheritedFG string) []domain.USN {
	var nodes []domain.USN

	if n.Type == html.ElementNode {
		raw := a.renderNode(n)
		line, col := a.findPosition(n, fullContent)

		usn := domain.USN{
			UID:    a.getAttribute(n, "id"),
			Role:   a.mapRole(n.Data),
			Label:  a.getLabel(n),
			Traits: make(map[string]any),
			Source: domain.Source{
				Platform:    a.platform,
				FilePath:    filename,
				Line:        line,
				Column:      col,
				RawHTML:     raw,
				IsComponent: isComponent,
			},
		}
		if usn.UID == "" {
			usn.UID = n.Data
		}

		// 1. Merge CSS Class Traits (Lower priority)
		if classAttr := a.getAttribute(n, "class"); classAttr != "" {
			classes := strings.Fields(classAttr)
			for _, c := range classes {
				if props, ok := a.cssMap[c]; ok {
					for k, v := range props {
						usn.Traits[k] = v
					}
				}
			}
		}

		// 2. Map direct attributes and inline styles (Higher priority override)
		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "lang" || attr.Key == "type" || attr.Key == "class" || attr.Key == "className" || attr.Key == "for" ||
				attr.Key == "aria-pressed" || attr.Key == "aria-expanded" || attr.Key == "aria-checked" || attr.Key == "role" ||
				attr.Key == "tabindex" || attr.Key == "aria-live" ||
				attr.Key == "onclick" || attr.Key == "onkeydown" || attr.Key == "onkeyup" || attr.Key == "onkeypress" ||
				attr.Key == "@click" || attr.Key == "v-on:click" || attr.Key == "(click)" || attr.Key == "on:click" ||
				attr.Key == "@keydown" || attr.Key == "v-on:keydown" || attr.Key == "(keydown)" || attr.Key == "on:keydown" ||
				attr.Key == "@keyup" || attr.Key == "v-on:keyup" || attr.Key == "(keyup)" || attr.Key == "on:keyup" ||
				attr.Key == "@keypress" || attr.Key == "v-on:keypress" || attr.Key == "(keypress)" || attr.Key == "on:keypress" {
				usn.Traits[attr.Key] = attr.Val
			}

			// React JSX: htmlFor → for (for label association).
			// Note: HTML parsers lowercase all attributes, so htmlFor becomes htmlfor.
			if attr.Key == "htmlfor" && attr.Val != "" {
				usn.Traits["htmlFor"] = attr.Val
			}
			// Angular/Vue template bindings: [attr]="expr" means the value is dynamic (bound).
			// The parser sees [alt] as a literal key with a binding expression as value.
			// Treat any [alt] or [attr.alt] as "label is bound dynamically" — skip empty-alt violation.
			if (attr.Key == "[alt]" || attr.Key == "v-bind:alt" || attr.Key == ":alt" || attr.Key == "[attr.alt]") && attr.Val != "" {
				usn.Label = "{{" + attr.Val + "}}" // mark as dynamically bound
			}

			// Tailwind / Utility-first heuristics
			if attr.Key == "class" || attr.Key == "className" {
				classes := strings.Fields(attr.Val)
				for _, c := range classes {
					// Spacing (w-12, h-4, etc) -> 1 unit = 4px
					if strings.HasPrefix(c, "w-") {
						var val float64
						if _, err := fmt.Sscanf(c, "w-%f", &val); err == nil {
							usn.Traits["width"] = val * 4
						}
					}
					if strings.HasPrefix(c, "h-") {
						var val float64
						if _, err := fmt.Sscanf(c, "h-%f", &val); err == nil {
							usn.Traits["height"] = val * 4
						}
					}
					// Colors (text-red-500, bg-slate-900) — only set if resolved.
					// If mapTailwindColor returns "" the color is unknown; we must NOT
					// set the trait so the contrast rule skips the element instead of
					// producing a false-positive 1.00:1 ratio.
					if strings.HasPrefix(c, "text-") {
						if hex := a.mapTailwindColor(c); hex != "" {
							usn.Traits["color"] = hex
						}
					}
					if strings.HasPrefix(c, "bg-") {
						if hex := a.mapTailwindColor(c); hex != "" {
							usn.Traits["background-color"] = hex
						}
					}
					if strings.HasPrefix(c, "border-") {
						if hex := a.mapTailwindColor(c); hex != "" {
							usn.Traits["border-color"] = hex
						}
					}
					// WCAG 1.4.1: no-underline class on links removes text-decoration.
					if c == "no-underline" {
						usn.Traits["no-underline"] = true
					}
					// Dark mode Tailwind classes: dark:bg-*, dark:text-*
					if strings.HasPrefix(c, "dark:text-") {
						if hex := a.mapTailwindColor("text-" + strings.TrimPrefix(c, "dark:text-")); hex != "" {
							usn.Traits["dark:color"] = hex
						}
					}
					if strings.HasPrefix(c, "dark:bg-") {
						if hex := a.mapTailwindColor("bg-" + strings.TrimPrefix(c, "dark:bg-")); hex != "" {
							usn.Traits["dark:background-color"] = hex
						}
					}
				}
			}

			// Map framework-specific labels
			if attr.Key == "aria-label" || attr.Key == "alt" ||
				attr.Key == ":alt" || attr.Key == "v-bind:alt" ||
				attr.Key == "bind:alt" {
				if strings.TrimSpace(attr.Val) != "" {
					usn.Label = attr.Val
				}
			}

			if attr.Key == "style" {
				parts := strings.Split(attr.Val, ";")
				for _, p := range parts {
					kv := strings.Split(p, ":")
					if len(kv) < 2 {
						continue
					}
					key := strings.TrimSpace(kv[0])
					val := strings.TrimSpace(strings.Join(kv[1:], ":"))
					if key == "color" || key == "background-color" || key == "border-color" {
						if strings.HasPrefix(val, "#") {
							usn.Traits[key] = val
						} else if strings.HasPrefix(val, "rgb(") {
							if hex := a.normalizeColor(val); hex != "" {
								usn.Traits[key] = hex
							}
						} else if strings.HasPrefix(val, "var(") {
							// Resolve CSS custom property if we've loaded external CSS.
							inner := strings.TrimSuffix(strings.TrimPrefix(val, "var("), ")")
							if hex, ok := a.cssVars[strings.TrimSpace(inner)]; ok {
								usn.Traits[key] = hex
							}
						}
					}
				}
			}
		}

		// Apply inherited background-color if this node doesn't have its own.
		// CSS background-color is not inherited by default, but for contrast analysis
		// we need the effective background a child is rendered on.
		if _, hasBG := usn.Traits["background-color"]; !hasBG && inheritedBG != "" {
			usn.Traits["background-color"] = inheritedBG
		}
		// CSS `color` IS inherited. Propagate the parent's foreground color when the
		// element doesn't declare one explicitly, so contrast checks can fire correctly
		// for children of e.g. <body class="text-white">.
		if _, hasFG := usn.Traits["color"]; !hasFG && inheritedFG != "" {
			usn.Traits["color"] = inheritedFG
		}

		nodes = append(nodes, usn)
	}

	// Determine the background-color and color to propagate to children.
	childBG := inheritedBG
	childFG := inheritedFG
	if n.Type == html.ElementNode && len(nodes) > 0 {
		last := nodes[len(nodes)-1]
		if bg, ok := last.Traits["background-color"].(string); ok && bg != "" {
			childBG = bg
		}
		if fg, ok := last.Traits["color"].(string); ok && fg != "" {
			childFG = fg
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, a.traverse(c, filename, lines, fullContent, isComponent, childBG, childFG)...)
	}

	return nodes
}

func (a *htmlAdapter) mapRole(tag string) domain.SemanticRole {
	switch tag {
	case "button":
		return domain.RoleButton
	case "a":
		return domain.RoleLink
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return domain.RoleHeading
	case "img":
		return domain.RoleImage
	case "input":
		return domain.RoleInput
	default:
		return "generic"
	}
}

func (a *htmlAdapter) getLabel(n *html.Node) string {
	// 1. Check aria-label or alt (highest priority — explicit accessible name)
	for _, attr := range n.Attr {
		if attr.Key == "aria-label" || attr.Key == "alt" || attr.Key == ":alt" || attr.Key == "bind:alt" {
			if strings.TrimSpace(attr.Val) != "" {
				return attr.Val
			}
		}
	}

	// 2. For text-bearing elements, capture their direct text content.
	// This includes headings, interactive elements, and inline text elements
	// (p, span, label, legend, caption) so that contrast checks are triggered.
	switch n.Data {
	case "button", "a",
		"h1", "h2", "h3", "h4", "h5", "h6",
		"label", "legend", "caption",
		"p", "span", "li", "td", "th", "dt", "dd", "figcaption":
		t := strings.TrimSpace(a.getText(n))
		if t != "" {
			return t
		}
	}

	return ""
}

func (a *htmlAdapter) getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += a.getText(c)
	}
	return text
}

func (a *htmlAdapter) getAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (a *htmlAdapter) renderNode(n *html.Node) string {
	var buf bytes.Buffer
	// We only want the start tag for better UX in reports
	buf.WriteString("<" + n.Data)
	for _, attr := range n.Attr {
		fmt.Fprintf(&buf, " %s=\"%s\"", attr.Key, attr.Val)
	}
	buf.WriteString(">")
	return buf.String()
}

func (a *htmlAdapter) findPosition(n *html.Node, fullContent string) (int, int) {
	// Attempt to find the tag in the source more robustly.
	// 1. Try exact match of a reconstructed start tag (best effort).
	raw := a.renderNode(n)
	idx := strings.Index(fullContent, raw)
	
	// 2. If exact match fails, try a fuzzy match based on tag name and key attributes.
	if idx == -1 {
		tagName := n.Data
		var patterns []string
		patterns = append(patterns, "<"+tagName)
		
		// Add some attributes to the search pattern if available
		for _, attr := range n.Attr {
			// Skip attributes that might be transformed or are too common
			if attr.Key == "class" || attr.Key == "id" || attr.Key == "src" || attr.Key == "href" {
				val := regexp.QuoteMeta(attr.Val)
				patterns = append(patterns, fmt.Sprintf("%s\\s*=\\s*['\"]%s['\"]", attr.Key, val))
			}
		}
		
		// Build a regex that matches the tag name and at least one identifying attribute
		if len(patterns) > 1 {
			for i := 1; i < len(patterns); i++ {
				rePattern := "(?i)" + patterns[0] + "[^>]*" + patterns[i]
				re, err := regexp.Compile(rePattern)
				if err == nil {
					loc := re.FindStringIndex(fullContent)
					if loc != nil {
						idx = loc[0]
						break
					}
				}
			}
		}
		
		// 3. Fallback to just the first occurrence of the tag name if still not found
		if idx == -1 {
			idx = strings.Index(fullContent, "<"+tagName)
		}
	}

	if idx == -1 {
		return 0, 0
	}

	line := 1
	col := 1
	for i := 0; i < idx; i++ {
		if fullContent[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

// LoadProjectCSS feeds external CSS/SCSS files into the adapter so that
// custom color tokens and CSS custom properties are available when Ingest
// is called. It is a no-op if the adapter is not an *htmlAdapter.
// Intended for --dir mode in the CLI where standalone CSS files exist.
//
// It also scans JS/TS/TSX/JSX files in the same list for CSS-in-JS patterns
// (styled-components, emotion css“, etc.) using a best-effort heuristic.
func LoadProjectCSS(adapter ports.Adapter, cssFiles []string) {
	ha, ok := adapter.(*htmlAdapter)
	if !ok {
		return
	}
	for _, f := range cssFiles {
		lower := strings.ToLower(f)
		switch {
		case strings.HasSuffix(lower, ".css") || strings.HasSuffix(lower, ".scss"):
			_ = ha.LoadExternalCSS(f)
		case strings.HasSuffix(lower, ".js") || strings.HasSuffix(lower, ".ts") ||
			strings.HasSuffix(lower, ".jsx") || strings.HasSuffix(lower, ".tsx"):
			// Check for tailwind.config.js / tailwind.config.ts first.
			base := strings.ToLower(filepath.Base(f))
			if base == "tailwind.config.js" || base == "tailwind.config.ts" || base == "tailwind.config.mjs" {
				_ = ha.LoadTailwindConfig(f)
			} else {
				_ = ha.LoadCSSinJS(f)
			}
		}
	}
}

// LoadExternalCSS parses an external CSS or SCSS file and extracts:
//  1. CSS custom properties (--foo: #hex) → cssVars
//  2. Class-level color declarations → cssMap
//  3. Tailwind v4 @theme blocks (--color-*: #hex) → customColors
//
// This should be called before Ingest() when using --dir mode so that
// contrast checks can resolve tokens defined in standalone CSS files.
func (a *htmlAdapter) LoadExternalCSS(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	css := string(data)

	// Parse CSS custom properties at root / :root / @theme scope.
	// Matches: --color-primary: #1a2b3c;  or  --brand: rgb(10,20,30);
	varRe := regexp.MustCompile(`--([\w-]+)\s*:\s*([^;}\n]+)`)
	for _, m := range varRe.FindAllStringSubmatch(css, -1) {
		name := strings.TrimSpace(m[1])
		val := strings.TrimSpace(m[2])
		hex := a.normalizeColor(val)
		if hex == "" {
			continue
		}
		a.cssVars["--"+name] = hex

		// Tailwind v4 convention: --color-<name> maps to the Tailwind token <name>.
		// e.g. --color-primary → customColors["primary"] = hex
		// e.g. --color-brand-500 → customColors["brand-500"] = hex
		if strings.HasPrefix(name, "color-") {
			token := strings.TrimPrefix(name, "color-")
			a.customColors[token] = hex
		}
	}

	// Resolve any CSS-var references now that we have the var map.
	for k, v := range a.cssVars {
		if strings.HasPrefix(v, "var(") {
			inner := strings.TrimSuffix(strings.TrimPrefix(v, "var("), ")")
			if resolved, ok := a.cssVars[strings.TrimSpace(inner)]; ok {
				a.cssVars[k] = resolved
			}
		}
	}

	// Also parse regular class rules (same as inline <style> parsing).
	a.parseCSS(css)

	// Extract @media (prefers-color-scheme: dark) { ... } blocks and parse them
	// as dark-mode overrides stored in darkCSSMap.
	darkRe := regexp.MustCompile(`(?s)@media\s*\(\s*prefers-color-scheme\s*:\s*dark\s*\)\s*\{(.+?)\}(?:\s*\})`)
	for _, dm := range darkRe.FindAllStringSubmatch(css, -1) {
		a.parseDarkCSS(dm[1])
	}

	return nil
}

// LoadTailwindConfig parses a tailwind.config.js / .ts file and extracts custom
// color tokens defined in theme.colors or theme.extend.colors.
// Uses best-effort regex-based JS parsing (no AST).
//
// Supports the patterns:
//
//	colors: { brand: '#1a2b3c', primary: { DEFAULT: '#fff', 500: '#abc' } }
func (a *htmlAdapter) LoadTailwindConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// Match  'tokenName': '#hexvalue'  or  "tokenName": "#hexvalue"
	// Covers both string and numeric sub-keys (e.g. 500: '#...').
	tokenRe := regexp.MustCompile(`['"]?([\w-]+)['"]?\s*:\s*['"](#[0-9a-fA-F]{3,8})['"]`)
	matches := tokenRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		name := strings.ToLower(m[1])
		hex := a.normalizeColor(m[2])
		if hex == "" {
			continue
		}
		// Skip common non-color keys.
		switch name {
		case "default", "content", "screens", "spacing", "fontsize", "borderradius":
			continue
		}
		a.customColors[name] = hex
	}
	return nil
}

// normalizeColor converts common CSS color formats to lowercase hex.
// Supports: #rrggbb, #rgb, #rrggbbaa, #rgba, rgb(r,g,b), rgba(r,g,b,a).
// Alpha channels are composited over white (255,255,255) for contrast analysis.
// Returns "" if unrecognized.
func (a *htmlAdapter) normalizeColor(val string) string {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "#") {
		switch len(val) {
		case 7: // #rrggbb
			return strings.ToLower(val)
		case 4: // #rgb → #rrggbb
			r, g, b := val[1:2], val[2:3], val[3:4]
			return "#" + r + r + g + g + b + b
		case 9: // #rrggbbaa → composite over white
			var r, g, b, aa int
			if n, _ := fmt.Sscanf(val[1:], "%02x%02x%02x%02x", &r, &g, &b, &aa); n == 4 {
				alpha := float64(aa) / 255.0
				r = int(float64(r)*alpha + 255*(1-alpha))
				g = int(float64(g)*alpha + 255*(1-alpha))
				b = int(float64(b)*alpha + 255*(1-alpha))
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		case 5: // #rgba → composite over white
			rs, gs, bs, as_ := val[1:2], val[2:3], val[3:4], val[4:5]
			var r, g, b, aa int
			if n, _ := fmt.Sscanf(rs+rs+gs+gs+bs+bs+as_+as_, "%02x%02x%02x%02x", &r, &g, &b, &aa); n == 4 {
				alpha := float64(aa) / 255.0
				r = int(float64(r)*alpha + 255*(1-alpha))
				g = int(float64(g)*alpha + 255*(1-alpha))
				b = int(float64(b)*alpha + 255*(1-alpha))
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		}
	}
	if strings.HasPrefix(val, "rgb(") {
		// rgb(r, g, b)
		inner := strings.TrimSuffix(strings.TrimPrefix(val, "rgb("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			var r, g, b int
			_, err1 := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &r)
			_, err2 := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &g)
			_, err3 := fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &b)
			if err1 == nil && err2 == nil && err3 == nil {
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		}
	}
	if strings.HasPrefix(val, "rgba(") {
		// rgba(r, g, b, a)
		inner := strings.TrimSuffix(strings.TrimPrefix(val, "rgba("), ")")
		parts := strings.Split(inner, ",")
		if len(parts) == 4 {
			var r, g, b int
			var alpha float64
			_, err1 := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &r)
			_, err2 := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &g)
			_, err3 := fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &b)
			_, err4 := fmt.Sscanf(strings.TrimSpace(parts[3]), "%f", &alpha)
			if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
				r = int(float64(r)*alpha + 255*(1-alpha))
				g = int(float64(g)*alpha + 255*(1-alpha))
				b = int(float64(b)*alpha + 255*(1-alpha))
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		}
	}
	return ""
}

// mapTailwindColor resolves a Tailwind utility class to a hex color string.
// It accepts classes like "text-blue-600", "bg-slate-900", "text-white".
// Returns "" if the color cannot be resolved (unknown custom token, CSS var, etc.).
func (a *htmlAdapter) mapTailwindColor(class string) string {
	// Strip the prefix to get the color key (e.g. "text-blue-600" → "blue-600").
	colorKey := class
	for _, prefix := range []string{"text-", "bg-", "border-", "ring-", "from-", "to-", "via-"} {
		if strings.HasPrefix(class, prefix) {
			colorKey = strings.TrimPrefix(class, prefix)
			break
		}
	}

	// Check CSS variables resolved at runtime — try cssVars first, then skip.
	if strings.HasPrefix(colorKey, "[") || strings.HasPrefix(colorKey, "var(") {
		inner := strings.TrimSuffix(strings.TrimPrefix(colorKey, "var("), ")")
		if hex, ok := a.cssVars[strings.TrimSpace(inner)]; ok {
			return hex
		}
		return ""
	}

	// Check custom color tokens already resolved from project CSS (cssMap).
	if hex, ok := a.customColors[colorKey]; ok {
		return hex
	}

	// Look up the standard Tailwind v3 palette.
	if hex, ok := tailwindV3Colors[colorKey]; ok {
		return hex
	}

	return ""
}
