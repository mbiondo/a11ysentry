package web

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
	"golang.org/x/net/html"
)

// AttributeMapper allows frameworks to map custom attributes to standard HTML attributes.
type AttributeMapper func(key, val string) (mappedKey string, mappedVal string)

type htmlAdapter struct {
	platform     domain.Platform
	cssMap       map[string]map[string]string // className -> property -> value
	darkCSSMap   map[string]map[string]string // dark-mode overrides: className -> property -> value
	customColors map[string]string            // Tailwind custom token name -> hex (resolved from CSS files)
	cssVars      map[string]string            // CSS custom property name (--foo) -> hex
	darkCSSVars  map[string]string            // CSS custom properties for dark mode
	attrMapper   AttributeMapper              // maps framework-specific attributes
}

// NewHTMLAdapterWithMapper creates a new HTML adapter with a custom attribute mapper.
func NewHTMLAdapterWithMapper(platform domain.Platform, mapper AttributeMapper) ports.Adapter {
	adapter := &htmlAdapter{
		platform:     platform,
		cssMap:       make(map[string]map[string]string),
		darkCSSMap:   make(map[string]map[string]string),
		customColors: make(map[string]string),
		cssVars:      make(map[string]string),
		darkCSSVars:  make(map[string]string),
		attrMapper:   mapper,
	}
	return adapter
}

func genericWebMapper(key, val string) (string, string) {
	// Svelte/Solid events: on:click -> onclick
	if strings.HasPrefix(key, "on:") {
		return "on" + key[3:], val
	}
	// Svelte bindings: bind:alt -> alt
	if strings.HasPrefix(key, "bind:") {
		return key[5:], val
	}
	// React/JSX properties: htmlfor -> for
	if key == "htmlfor" {
		return "for", val
	}
	return key, val
}

func NewHTMLAdapter() ports.Adapter {
	return NewHTMLAdapterWithMapper(domain.PlatformWebReact, genericWebMapper)
}

func NewElectronAdapter() ports.Adapter {
	return NewHTMLAdapterWithMapper(domain.PlatformElectron, nil)
}

func NewTauriAdapter() ports.Adapter {
	return NewHTMLAdapterWithMapper(domain.PlatformTauri, nil)
}

func (a *htmlAdapter) Ingest(ctx context.Context, root *domain.FileNode) ([]domain.USN, error) {
	if root == nil {
		return nil, nil
	}
	return a.ingestRecursive(ctx, root, "", "", false, nil)
}

func (a *htmlAdapter) ingestRecursive(ctx context.Context, node *domain.FileNode, inheritedBG, inheritedFG string, inheritedHidden bool, inheritedIgnored []string) ([]domain.USN, error) {
	if node.IsCycle {
		// Stop processing cycles to avoid redundant work and errors.
		return nil, nil
	}
	if node.IsOpaque {
		// Opaque nodes don't have source code to analyze.
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

	// Build a map of opaque children for quick lookup in traverse
	opaqueMap := make(map[string]string)
	for _, child := range node.Children {
		if child.IsOpaque {
			pkg := child.OpaqueSource
			parts := strings.Split(pkg, "/")
			name := parts[len(parts)-1]
			opaqueMap[name] = pkg
			opaqueMap[pkg] = pkg
		}
	}

	contentLines := strings.Split(cleanContent, "\n")
	nodes := a.traverse(doc, node.FilePath, contentLines, cleanContent, isComponent, inheritedBG, inheritedFG, inheritedHidden, inheritedIgnored, opaqueMap)

	childBG := inheritedBG
	childFG := inheritedFG
	childHidden := inheritedHidden
	childIgnored := inheritedIgnored
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
		childIgnored = n.Source.IgnoredRules
	}

	var allNodes []domain.USN
	allNodes = append(allNodes, nodes...)

	for _, child := range node.Children {
		childNodes, err := a.ingestRecursive(ctx, child, childBG, childFG, childHidden, childIgnored)
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
	inThemeBlock := false

	// State machine for manual token accumulation inside @theme blocks.
	// tdewolff emits TokenGrammar (not CustomPropertyGrammar) for custom
	// properties inside @theme {}, so we must accumulate them ourselves.
	var themeProp strings.Builder
	var themeVal strings.Builder
	themeState := 0 // 0=idle, 1=prop, 2=val

	flushThemeDecl := func() {
		prop := strings.TrimSpace(themeProp.String())
		val := strings.TrimSpace(themeVal.String())
		themeProp.Reset()
		themeVal.Reset()
		themeState = 0
		if prop == "" || val == "" {
			return
		}
		a.cssVars[prop] = val
		if hex := a.normalizeColor(val); hex != "" {
			if strings.HasPrefix(prop, "--color-") {
				a.customColors[strings.TrimPrefix(prop, "--color-")] = hex
			}
		}
	}

	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			break
		}

		// Inside @theme, tokens arrive one by one — accumulate manually.
		if inThemeBlock {
			switch gt {
			case css.EndAtRuleGrammar:
				flushThemeDecl()
				inThemeBlock = false
			case css.TokenGrammar:
				tok := string(data)
				switch {
				case themeState == 0 && strings.HasPrefix(tok, "--"):
					themeProp.WriteString(tok)
					themeState = 1
				case themeState == 1 && tok == ":":
					themeState = 2
				case themeState == 2 && tok == ";":
					flushThemeDecl()
				case themeState == 1:
					// still accumulating prop name (e.g. multi-token)
					themeProp.WriteString(tok)
				case themeState == 2:
					// accumulate value, skip bare whitespace tokens
					if strings.TrimSpace(tok) != "" {
						themeVal.WriteString(tok)
					}
				}
			}
			continue
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
			if name == "theme" {
				inThemeBlock = true
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

func (a *htmlAdapter) traverse(n *html.Node, filename string, lines []string, fullContent string, isComponent bool, inheritedBG, inheritedFG string, inheritedHidden bool, inheritedIgnored []string, opaqueMap map[string]string) []domain.USN {
	var nodes []domain.USN

	if n.Type == html.ElementNode {
		raw := a.renderNode(n)
		line, col := a.findPosition(n, fullContent)
		isOpaque, opaqueSource := a.getOpaqueComponentInfo(n, opaqueMap)
		
		// Merge inherited ignores with current ones
		currentIgnored := a.getA11yIgnoreRules(n)
		allIgnored := append([]string{}, inheritedIgnored...)
		for _, r := range currentIgnored {
			found := false
			for _, ir := range inheritedIgnored {
				if r == ir {
					found = true
					break
				}
			}
			if !found {
				allIgnored = append(allIgnored, r)
			}
		}

		usn := domain.USN{
			UID:      a.getAttribute(n, "id"),
			Role:     a.mapRole(n.Data),
			Label:    a.getLabel(n),
			IsOpaque: isOpaque,
			Traits:   make(map[string]any),
			Source: domain.Source{
				Platform:     a.platform,
				FilePath:     filename,
				Line:         line,
				Column:       col,
				RawHTML:      raw,
				IsComponent:  isComponent,
				IsOpaque:     isOpaque,
				OpaqueSource: opaqueSource,
				IgnoredRules: allIgnored,
			},
		}

		if inheritedHidden {
			usn.Traits["aria-hidden-inherited"] = true
		}

		if ariaRole := a.getAttribute(n, "role"); ariaRole != "" {
			usn.Role = a.mapAriaRole(ariaRole)
		}
		if usn.UID == "" {
			usn.UID = n.Data
		}

		a.mapAttributes(n, &usn)
		a.processCSSClasses(n, &usn, inheritedBG)
		a.processInlineStyles(n, &usn)

		// Detect JSX spread props ({...something}) which the HTML parser strips.
		// Scan up to 20 lines forward from the tag to handle multi-line elements.
		if line > 0 && line <= len(lines) {
			hasspread := false
			for scanLine := line - 1; scanLine < len(lines) && scanLine < line+20; scanLine++ {
				srcLine := lines[scanLine]
				if strings.Contains(srcLine, "{...") {
					hasspread = true
					break
				}
				// Stop only on self-closing /> or a line whose trimmed content is just ">"
				trimmed := strings.TrimSpace(srcLine)
				if scanLine > line-1 && (strings.HasSuffix(trimmed, "/>") || trimmed == ">") {
					break
				}
			}
			if hasspread {
				usn.Traits["has-spread-props"] = true
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
	childIgnored := inheritedIgnored
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
		childIgnored = last.Source.IgnoredRules
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, a.traverse(c, filename, lines, fullContent, isComponent, childBG, childFG, childHidden, childIgnored, opaqueMap)...)
	}

	return nodes
}

func (a *htmlAdapter) getOpaqueComponentInfo(n *html.Node, opaqueMap map[string]string) (bool, string) {
	isCustom := (len(n.Data) > 0 && n.Data[0] >= 'A' && n.Data[0] <= 'Z')
	for k, v := range opaqueMap {
		if strings.EqualFold(n.Data, k) || strings.Contains(strings.ToLower(n.Data), strings.ToLower(k)) {
			return true, v
		}
	}
	return isCustom, ""
}

func (a *htmlAdapter) getA11yIgnoreRules(n *html.Node) []string {
	var ignoredRules []string
	for prev := n.PrevSibling; prev != nil; prev = prev.PrevSibling {
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
		break
	}
	return ignoredRules
}

func (a *htmlAdapter) mapAriaRole(ariaRole string) domain.SemanticRole {
	switch ariaRole {
	case "button": return domain.RoleButton
	case "link": return domain.RoleLink
	case "heading": return domain.RoleHeading
	case "dialog", "alertdialog": return domain.RoleModal
	case "main": return domain.RoleMain
	case "navigation": return domain.RoleNav
	case "complementary": return domain.RoleAside
	case "banner": return domain.RoleHeader
	case "contentinfo": return domain.RoleFooter
	case "region": return domain.RoleSection
	case "form": return domain.RoleForm
	case "search": return domain.RoleSearch
	case "status", "alert", "log": return domain.RoleLiveRegion
	default: return domain.SemanticRole(ariaRole)
	}
}

func (a *htmlAdapter) mapAttributes(n *html.Node, usn *domain.USN) {
	for _, attr := range n.Attr {
		k, v := attr.Key, attr.Val
		if a.attrMapper != nil {
			k, v = a.attrMapper(k, v)
			if k == "" {
				continue
			}
		}
		
		switch k {
		case "id", "lang", "type", "class", "className", "for", "aria-pressed", "aria-expanded", "aria-checked", "role", "tabindex", "aria-live", "href", "title", "autocomplete", "aria-hidden", "onclick", "onkeydown", "onkeyup", "onkeypress":
			usn.Traits[k] = v
		case "htmlfor":
			if v != "" {
				usn.Traits["htmlFor"] = v
			}
		case "alt":
			if v != "" {
				usn.Label = v
			}
		}
	}
}

func (a *htmlAdapter) processCSSClasses(n *html.Node, usn *domain.USN, inheritedBG string) {
	classAttr := a.getAttribute(n, "class")
	if classAttr == "" {
		classAttr = a.getAttribute(n, "className")
	}
	if classAttr == "" {
		return
	}

	classes := strings.Fields(classAttr)
	for _, c := range classes {
		if props, ok := a.cssMap[c]; ok {
			for k, v := range props {
				usn.Traits[k] = a.resolveVar(v)
			}
		}
		
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
			if hex := a.mapTailwindColor(c, inheritedBG); hex != "" {
				usn.Traits["color"] = hex
			} else {
				// text-* class found but color could not be resolved (CSS var, custom token, etc.)
				colorKey := strings.TrimPrefix(c, "text-")
				// Skip non-color text-* classes (text-sm, text-center, text-left, etc.)
				isLikelySizeOrAlign := len(colorKey) <= 3 ||
					colorKey == "center" || colorKey == "left" || colorKey == "right" ||
					colorKey == "justify" || colorKey == "wrap" || colorKey == "nowrap" ||
					colorKey == "ellipsis" || colorKey == "clip" || colorKey == "truncate" ||
					colorKey == "uppercase" || colorKey == "lowercase" || colorKey == "capitalize" ||
					colorKey == "mono" || colorKey == "sans" || colorKey == "serif"
				if !isLikelySizeOrAlign {
					usn.Traits["has-unresolved-color"] = true
				}
			}
		}
		if strings.HasPrefix(c, "bg-") {
			if hex := a.mapTailwindColor(c, inheritedBG); hex != "" {
				usn.Traits["background-color"] = hex
			} else {
				// bg-* class found but color could not be resolved
				bgKey := strings.TrimPrefix(c, "bg-")
				isLikelyNonColor := bgKey == "transparent" || bgKey == "inherit" ||
					strings.HasPrefix(bgKey, "gradient") || bgKey == "none" ||
					bgKey == "fixed" || bgKey == "local" || bgKey == "scroll" ||
					bgKey == "clip" || bgKey == "origin"
				if !isLikelyNonColor {
					usn.Traits["has-unresolved-bg"] = true
				}
			}
		}
		if strings.HasPrefix(c, "border-") {
			if hex := a.mapTailwindColor(c, inheritedBG); hex != "" {
				usn.Traits["border-color"] = hex
			}
		}
		if c == "no-underline" {
			usn.Traits["no-underline"] = true
		}
		if strings.HasPrefix(c, "dark:text-") {
			if hex := a.mapTailwindColor("text-"+strings.TrimPrefix(c, "dark:text-"), inheritedBG); hex != "" {
				usn.Traits["dark:color"] = hex
			}
		}
		if strings.HasPrefix(c, "dark:bg-") {
			if hex := a.mapTailwindColor("bg-"+strings.TrimPrefix(c, "dark:bg-"), inheritedBG); hex != "" {
				usn.Traits["dark:background-color"] = hex
			}
		}
	}
}

func (a *htmlAdapter) processInlineStyles(n *html.Node, usn *domain.USN) {
	styleAttr := a.getAttribute(n, "style")
	if styleAttr == "" {
		return
	}
	
	parts := strings.Split(styleAttr, ";")
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
		k, v := attr.Key, attr.Val
		if a.attrMapper != nil {
			k, v = a.attrMapper(k, v)
		}
		if k == "aria-label" || k == "alt" {
			if strings.TrimSpace(v) != "" {
				return v
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
		return 1, 1
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
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "transparent" || val == "inherit" || val == "currentcolor" {
		return ""
	}
	if strings.HasPrefix(val, "#") {
		switch len(val) {
		case 7: // #rrggbb
			return val
		case 4: // #rgb → #rrggbb
			r, g, b := val[1:2], val[2:3], val[3:4]
			return "#" + r + r + g + g + b + b
		case 9: // #rrggbbaa
			return a.applyOpacity(val[:7], "#ffffff", 0.0) // fallback white
		case 5: // #rgba
			return val[:4]
		}
	}
	if strings.HasPrefix(val, "hsl(") || strings.HasPrefix(val, "hsla(") {
		inner := val
		inner = strings.TrimPrefix(inner, "hsla(")
		inner = strings.TrimPrefix(inner, "hsl(")
		inner = strings.TrimSuffix(inner, ")")
		inner = strings.ReplaceAll(inner, "/", ",")
		parts := strings.Split(inner, ",")
		if len(parts) >= 3 {
			h, errH := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			s, errS := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[1]), "%")), 64)
			l, errL := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[2]), "%")), 64)
			if errH == nil && errS == nil && errL == nil {
				s /= 100.0
				l /= 100.0
				r, g, b := hslToRGB(h, s, l)
				return fmt.Sprintf("#%02x%02x%02x", r, g, b)
			}
		}
	}
	if strings.HasPrefix(val, "rgb(") {
		inner := strings.TrimSuffix(strings.TrimPrefix(val, "rgb("), ")")
		inner = strings.ReplaceAll(inner, "/", ",")
		parts := strings.Split(inner, ",")
		if len(parts) >= 3 {
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
		inner = strings.ReplaceAll(inner, "/", ",")
		parts := strings.Split(inner, ",")
		if len(parts) >= 4 {
			var r, g, b int
			var alpha float64
			_, err1 := fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &r)
			_, err2 := fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &g)
			_, err3 := fmt.Sscanf(strings.TrimSpace(parts[2]), "%d", &b)
			_, err4 := fmt.Sscanf(strings.TrimSpace(parts[3]), "%f", &alpha)
			if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
				return a.applyOpacity(fmt.Sprintf("#%02x%02x%02x", r, g, b), "#ffffff", alpha)
			}
		}
	}
	return ""
}

func (a *htmlAdapter) mapTailwindColor(class string, inheritedBG string) string {
	colorKey := class
	for _, prefix := range []string{"text-", "bg-", "border-", "ring-", "from-", "to-", "via-"} {
		if strings.HasPrefix(class, prefix) {
			colorKey = strings.TrimPrefix(class, prefix)
			break
		}
	}

	opacity := 1.0
	if strings.Contains(colorKey, "/") {
		parts := strings.Split(colorKey, "/")
		colorKey = parts[0]
		if len(parts) > 1 {
			if val, err := strconv.ParseFloat(parts[1], 64); err == nil {
				opacity = val / 100.0
			}
		}
	}

	if strings.HasPrefix(colorKey, "[") && strings.HasSuffix(colorKey, "]") {
		inner := colorKey[1 : len(colorKey)-1]
		if hex := a.normalizeColor(inner); hex != "" {
			return hex
		}
		return a.resolveVar(inner)
	}
	if strings.HasPrefix(colorKey, "var(") {
		return a.resolveVar(colorKey)
	}
	
	hex := ""
	if h, ok := a.customColors[colorKey]; ok {
		hex = h
	} else if h, ok := tailwindV3Colors[colorKey]; ok {
		hex = h
	}

	if hex != "" && opacity < 1.0 {
		base := inheritedBG
		if base == "" {
			// Smarter fallback: if we see 'dark' in the classes or it's a dark theme, assume black
			isDark := strings.Contains(class, "dark:")
			if isDark {
				base = "#000000"
			} else {
				base = "#ffffff"
			}
		}
		return a.applyOpacity(hex, base, opacity)
	}

	return hex
}

// hslToRGB converts HSL (h in [0,360], s and l in [0,1]) to RGB uint8 components.
func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	hk := h / 360.0
	toRGB := func(t float64) uint8 {
		if t < 0 { t += 1 }
		if t > 1 { t -= 1 }
		switch {
		case t < 1.0/6.0:
			t = p + (q-p)*6*t
		case t < 0.5:
			t = q
		case t < 2.0/3.0:
			t = p + (q-p)*(2.0/3.0-t)*6
		default:
			t = p
		}
		return uint8(math.Round(t * 255))
	}
	return toRGB(hk + 1.0/3.0), toRGB(hk), toRGB(hk - 1.0/3.0)
}

func (a *htmlAdapter) applyOpacity(fgHex string, bgHex string, alpha float64) string {
	if !strings.HasPrefix(fgHex, "#") || len(fgHex) != 7 {
		return fgHex
	}
	if !strings.HasPrefix(bgHex, "#") || len(bgHex) != 7 {
		bgHex = "#ffffff"
	}
	
	var r1, g1, b1 int
	var r2, g2, b2 int
	if n1, _ := fmt.Sscanf(fgHex[1:], "%02x%02x%02x", &r1, &g1, &b1); n1 == 3 {
		if n2, _ := fmt.Sscanf(bgHex[1:], "%02x%02x%02x", &r2, &g2, &b2); n2 == 3 {
			r := int(float64(r1)*alpha + float64(r2)*(1-alpha))
			g := int(float64(g1)*alpha + float64(g2)*(1-alpha))
			b := int(float64(b1)*alpha + float64(b2)*(1-alpha))
			return fmt.Sprintf("#%02x%02x%02x", r, g, b)
		}
	}
	return fgHex
}
