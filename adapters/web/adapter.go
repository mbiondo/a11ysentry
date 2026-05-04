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

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
)

type htmlAdapter struct {
	platform     domain.Platform
	cssMap       map[string]map[string]string // className -> property -> value
	darkCSSMap   map[string]map[string]string // dark-mode overrides: className -> property -> value
	customColors map[string]string            // Tailwind custom token name -> hex (resolved from CSS files)
	cssVars      map[string]string            // CSS custom property name (--foo) -> hex
	darkCSSVars  map[string]string            // CSS custom properties for dark mode
}

func NewHTMLAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformWebReact,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
		darkCSSVars:  make(map[string]string),
	}
}

func NewElectronAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformElectron,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
		darkCSSVars:  make(map[string]string),
	}
}

func NewTauriAdapter() ports.Adapter {
	return &htmlAdapter{
		platform:     domain.PlatformTauri,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
		darkCSSVars:  make(map[string]string),
	}
}

func (a *htmlAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	if root == nil {
		return nil, nil
	}
	return a.ingestRecursive(ctx, root, "", "", false)
}

func (a *htmlAdapter) ingestRecursive(ctx context.Context, node *domain.FileNode, inheritedBG, inheritedFG string, inheritedHidden bool) ([]domain.USN, error) {
	if node.IsCycle {
		// Stop processing cycles to avoid redundant work and errors.
		return nil, nil
	}
	data, err := os.ReadFile(node.FilePath)
	if err != nil {
		return nil, err
	}
	content := string(data)

	cleanContent := content
	if strings.HasPrefix(content, "---") {
		endIdx := strings.Index(content[3:], "---")
		if endIdx != -1 {
			frontmatter := content[:endIdx+6]
			newlines := strings.Repeat("\n", strings.Count(frontmatter, "\n"))
			cleanContent = newlines + content[endIdx+6:]
		}
	}

	isComponent := !strings.Contains(strings.ToLower(cleanContent), "<html")
	if inheritedBG == "" && inheritedFG == "" {
		isComponent = false
	}

	doc, err := html.Parse(strings.NewReader(cleanContent))
	if err != nil {
		return nil, err
	}

	a.extractCSS(doc)

	nodes := a.traverse(doc, node.FilePath, nil, cleanContent, isComponent, inheritedBG, inheritedFG, inheritedHidden)

	childBG := inheritedBG
	childFG := inheritedFG
	childHidden := inheritedHidden
	for _, n := range nodes {
		if bg, ok := n.Traits["background-color"].(string); ok && bg != "" {
			childBG = bg
		}
		if fg, ok := n.Traits["color"].(string); ok && fg != "" {
			childFG = fg
		}
		if h, ok := n.Traits["aria-hidden"].(string); ok && h == "true" {
			childHidden = true
		}
	}

	var allNodes []domain.USN
	allNodes = append(allNodes, nodes...)

	for _, child := range node.Children {
		childNodes, err := a.ingestRecursive(ctx, child, childBG, childFG, childHidden)
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

func (a *htmlAdapter) parseCSS(cssStr string) {
	a.processCSS(cssStr, a.cssMap, false)
}

func (a *htmlAdapter) processCSS(cssStr string, target map[string]map[string]string, isDark bool) {
	p := css.NewParser(parse.NewInputString(cssStr), false)
	var currentSelectors []string
	inDarkMedia := isDark

	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			break
		}

		switch gt {
		case css.AtRuleGrammar, css.BeginAtRuleGrammar:
			name := strings.TrimPrefix(strings.ToLower(string(data)), "@")
			if name == "media" {
				var b strings.Builder
				for _, v := range p.Values() {
					b.Write(v.Data)
				}
				prelude := b.String()
				if strings.Contains(prelude, "prefers-color-scheme") && strings.Contains(prelude, "dark") {
					inDarkMedia = true
				}
			}
		case css.EndAtRuleGrammar:
			inDarkMedia = isDark
		case css.BeginRulesetGrammar:
			currentSelectors = nil
			var b strings.Builder
			for _, v := range p.Values() {
				b.Write(v.Data)
			}
			rawSelectors := b.String()
			for _, s := range strings.Split(rawSelectors, ",") {
				sel := strings.TrimSpace(s)
				if sel == ":root" || sel == "@theme" || strings.HasPrefix(sel, ":root") {
					currentSelectors = append(currentSelectors, ":root")
				} else if strings.HasPrefix(sel, ".") {
					currentSelectors = append(currentSelectors, strings.TrimPrefix(sel, "."))
				}
			}
		case css.DeclarationGrammar, css.CustomPropertyGrammar:
			prop := strings.TrimSpace(string(data))
			var b strings.Builder
			for _, v := range p.Values() {
				b.Write(v.Data)
			}
			val := strings.TrimSpace(b.String())

			if strings.HasPrefix(prop, "--") || gt == css.CustomPropertyGrammar {
				vMap := a.cssVars
				if inDarkMedia {
					vMap = a.darkCSSVars
				}
				vMap[prop] = val
				if hex := a.normalizeColor(val); hex != "" {
					if strings.HasPrefix(prop, "--color-") {
						a.customColors[strings.TrimPrefix(prop, "--color-")] = hex
					}
				}
				continue
			}

			if len(currentSelectors) == 0 {
				continue
			}
			if prop != "color" && prop != "background-color" && prop != "border-color" {
				continue
			}

			currentTarget := target
			if inDarkMedia {
				currentTarget = a.darkCSSMap
			}

			for _, sel := range currentSelectors {
				if sel == ":root" {
					continue
				}
				if _, ok := currentTarget[sel]; !ok {
					currentTarget[sel] = make(map[string]string)
				}
				currentTarget[sel][prop] = val
			}
		}
	}
}

func (a *htmlAdapter) resolveVar(val string) string {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "var(") {
		inner := strings.TrimSuffix(strings.TrimPrefix(val, "var("), ")")
		key := strings.TrimSpace(inner)
		if strings.Contains(key, ",") {
			parts := strings.Split(key, ",")
			key = strings.TrimSpace(parts[0])
		}
		if res, ok := a.cssVars[key]; ok {
			return a.resolveVar(res)
		}
		if strings.Contains(inner, ",") {
			parts := strings.Split(inner, ",")
			return a.resolveVar(strings.TrimSpace(parts[1]))
		}
	}
	return val
}

func (a *htmlAdapter) traverse(n *html.Node, filename string, lines []string, fullContent string, isComponent bool, inheritedBG string, inheritedFG string, inheritedHidden bool) []domain.USN {
	var nodes []domain.USN

	if n.Type == html.ElementNode {
		raw := a.renderNode(n)
		line, col := a.findPosition(n, fullContent)

		// Check for preceding comments to support a11y-ignore
		var ignoredRules []string
		for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
			// Skip whitespace text nodes to find adjacent comments
			if prev.Type == html.TextNode && strings.TrimSpace(prev.Data) == "" {
				continue
			}
			if prev.Type == html.CommentNode {
				comment := strings.TrimSpace(prev.Data)
				if strings.HasPrefix(comment, "a11y-ignore:") {
					rulesPart := strings.TrimPrefix(comment, "a11y-ignore:")
					for _, r := range strings.Split(rulesPart, ",") {
						ignoredRules = append(ignoredRules, strings.TrimSpace(r))
					}
				}
			}
			// Only look at the immediate preceding non-whitespace node
			break
		}

		usn := domain.USN{
			UID:    a.getAttribute(n, "id"),
			Role:   a.mapRole(n.Data),
			Label:  a.getLabel(n),
			Traits: make(map[string]any),
			Source: domain.Source{
				Platform:     a.platform,
				FilePath:     filename,
				Line:         line,
				Column:       col,
				RawHTML:      raw,
				IsComponent:  isComponent,
				IgnoredRules: ignoredRules,
			},
		}

		if inheritedHidden {
			usn.Traits["aria-hidden-inherited"] = true
		}

		if ariaRole := a.getAttribute(n, "role"); ariaRole != "" {
			switch ariaRole {
			case "button":
				usn.Role = domain.RoleButton
			case "link":
				usn.Role = domain.RoleLink
			case "heading":
				usn.Role = domain.RoleHeading
			case "dialog", "alertdialog":
				usn.Role = domain.RoleModal
			case "main":
				usn.Role = domain.RoleMain
			case "navigation":
				usn.Role = domain.RoleNav
			case "complementary":
				usn.Role = domain.RoleAside
			case "banner":
				usn.Role = domain.RoleHeader
			case "contentinfo":
				usn.Role = domain.RoleFooter
			case "region":
				usn.Role = domain.RoleSection
			case "form":
				usn.Role = domain.RoleForm
			case "search":
				usn.Role = domain.RoleSearch
			case "status", "alert", "log":
				usn.Role = domain.RoleLiveRegion
			}
		}
		if usn.UID == "" {
			usn.UID = n.Data
		}

		if classAttr := a.getAttribute(n, "class"); classAttr != "" {
			classes := strings.Fields(classAttr)
			for _, c := range classes {
				if props, ok := a.cssMap[c]; ok {
					for k, v := range props {
						usn.Traits[k] = a.resolveVar(v)
					}
				}
			}
		}

		for _, attr := range n.Attr {
			if attr.Key == "id" || attr.Key == "lang" || attr.Key == "type" || attr.Key == "class" || attr.Key == "className" || attr.Key == "for" ||
				attr.Key == "aria-pressed" || attr.Key == "aria-expanded" || attr.Key == "aria-checked" || attr.Key == "role" ||
				attr.Key == "tabindex" || attr.Key == "aria-live" || attr.Key == "href" || attr.Key == "title" || attr.Key == "autocomplete" || attr.Key == "aria-hidden" ||
				attr.Key == "onclick" || attr.Key == "onkeydown" || attr.Key == "onkeyup" || attr.Key == "onkeypress" ||
				attr.Key == "@click" || attr.Key == "v-on:click" || attr.Key == "(click)" || attr.Key == "on:click" ||
				attr.Key == "@keydown" || attr.Key == "v-on:keydown" || attr.Key == "(keydown)" || attr.Key == "on:keydown" ||
				attr.Key == "@keyup" || attr.Key == "v-on:keyup" || attr.Key == "(keyup)" || attr.Key == "on:keyup" ||
				attr.Key == "@keypress" || attr.Key == "v-on:keypress" || attr.Key == "(keypress)" || attr.Key == "on:keypress" {
				usn.Traits[attr.Key] = attr.Val
			}

			if attr.Key == "htmlfor" && attr.Val != "" {
				usn.Traits["htmlFor"] = attr.Val
			}
			if (attr.Key == "[alt]" || attr.Key == "v-bind:alt" || attr.Key == ":alt" || attr.Key == "[attr.alt]") && attr.Val != "" {
				usn.Label = "{{" + attr.Val + "}}"
			}

			if attr.Key == "class" || attr.Key == "className" {
				classes := strings.Fields(attr.Val)
				for _, c := range classes {
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
					if c == "no-underline" {
						usn.Traits["no-underline"] = true
					}
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
							if res := a.resolveVar(val); res != val {
								usn.Traits[key] = res
							}
						}
					}
				}
			}
		}

		if _, hasBG := usn.Traits["background-color"]; !hasBG && inheritedBG != "" {
			usn.Traits["background-color"] = inheritedBG
		}
		if _, hasFG := usn.Traits["color"]; !hasFG && inheritedFG != "" {
			usn.Traits["color"] = inheritedFG
		}

		nodes = append(nodes, usn)
	}

	childBG := inheritedBG
	childFG := inheritedFG
	childHidden := inheritedHidden
	if n.Type == html.ElementNode && len(nodes) > 0 {
		last := nodes[len(nodes)-1]
		if bg, ok := last.Traits["background-color"].(string); ok && bg != "" {
			childBG = bg
		}
		if fg, ok := last.Traits["color"].(string); ok && fg != "" {
			childFG = fg
		}
		if h, ok := last.Traits["aria-hidden"].(string); ok && h == "true" {
			childHidden = true
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, a.traverse(c, filename, lines, fullContent, isComponent, childBG, childFG, childHidden)...)
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
	case "input", "select", "textarea":
		return domain.RoleInput
	case "main":
		return domain.RoleMain
	case "nav":
		return domain.RoleNav
	case "aside":
		return domain.RoleAside
	case "header":
		return domain.RoleHeader
	case "footer":
		return domain.RoleFooter
	case "section":
		return domain.RoleSection
	case "form":
		return domain.RoleForm
	case "search":
		return domain.RoleSearch
	case "fieldset":
		return domain.RoleFieldset
	case "legend":
		return domain.RoleLegend
	default:
		return "generic"
	}
}

func (a *htmlAdapter) getLabel(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "aria-label" || attr.Key == "alt" || attr.Key == ":alt" || attr.Key == "bind:alt" || attr.Key == "v-bind:alt" || attr.Key == "[alt]" || attr.Key == "[attr.alt]" || attr.Key == "[attr.aria-label]" {
			if strings.TrimSpace(attr.Val) != "" {
				if attr.Key == ":alt" || attr.Key == "bind:alt" || attr.Key == "v-bind:alt" || attr.Key == "[alt]" || attr.Key == "[attr.alt]" || attr.Key == "[attr.aria-label]" {
					return "{{" + attr.Val + "}}"
				}
				return attr.Val
			}
		}
	}

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
	buf.WriteString("<" + n.Data)
	for _, attr := range n.Attr {
		fmt.Fprintf(&buf, " %s=\"%s\"", attr.Key, attr.Val)
	}
	buf.WriteString(">")
	return buf.String()
}

func (a *htmlAdapter) findPosition(n *html.Node, fullContent string) (int, int) {
	raw := a.renderNode(n)
	idx := strings.Index(fullContent, raw)
	if idx == -1 {
		tagName := n.Data
		idx = strings.Index(fullContent, "<"+tagName)
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
			base := strings.ToLower(filepath.Base(f))
			if base == "tailwind.config.js" || base == "tailwind.config.ts" || base == "tailwind.config.mjs" {
				_ = ha.LoadTailwindConfig(f)
			} else {
				_ = ha.LoadCSSinJS(f)
			}
		}
	}
}

func (a *htmlAdapter) LoadExternalCSS(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	cssStr := string(data)
	a.processCSS(cssStr, a.cssMap, false)
	return nil
}

func (a *htmlAdapter) LoadTailwindConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	tokenRe := regexp.MustCompile(`['"]?([\w-]+)['"]?\s*:\s*['"](#[0-9a-fA-F]{3,8})['"]`)
	matches := tokenRe.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		name := strings.ToLower(m[1])
		hex := a.normalizeColor(m[2])
		if hex == "" {
			continue
		}
		switch name {
		case "default", "content", "screens", "spacing", "fontsize", "borderradius":
			continue
		}
		a.customColors[name] = hex
	}
	return nil
}

func (a *htmlAdapter) normalizeColor(val string) string {
	val = strings.TrimSpace(val)
	if strings.HasPrefix(val, "#") {
		switch len(val) {
		case 7: // #rrggbb
			return strings.ToLower(val)
		case 4: // #rgb → #rrggbb
			r, g, b := val[1:2], val[2:3], val[3:4]
			return "#" + r + r + g + g + b + b
		case 9: // #rrggbbaa
			var r, g, b, aa int
			if n, _ := fmt.Sscanf(val[1:], "%02x%02x%02x%02x", &r, &g, &b, &aa); n == 4 {
				alpha := float64(aa) / 255.0
				r = int(float64(r)*alpha + 255*(1-alpha))
				g = int(float64(g)*alpha + 255*(1-alpha))
				b = int(float64(b)*alpha + 255*(1-alpha))
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		case 5: // #rgba
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

func (a *htmlAdapter) mapTailwindColor(class string) string {
	colorKey := class
	for _, prefix := range []string{"text-", "bg-", "border-", "ring-", "from-", "to-", "via-"} {
		if strings.HasPrefix(class, prefix) {
			colorKey = strings.TrimPrefix(class, prefix)
			break
		}
	}
	if strings.HasPrefix(colorKey, "[") || strings.HasPrefix(colorKey, "var(") {
		return a.resolveVar(colorKey)
	}
	if hex, ok := a.customColors[colorKey]; ok {
		return hex
	}
	if hex, ok := tailwindV3Colors[colorKey]; ok {
		return hex
	}
	return ""
}
