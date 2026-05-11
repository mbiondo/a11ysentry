package main

import (
	"os"
	"strings"
)

func main() {
	filePath := "E:/repositories/semantix/adapters/web/adapter.go"
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	
	// We want to replace traverse and all its helper methods up to LoadProjectCSS
	startIdx := strings.Index(string(content), "func (a *htmlAdapter) traverse(")
	if startIdx == -1 {
		panic("traverse not found")
	}
	
	endIdx := strings.Index(string(content), "func LoadProjectCSS(")
	if endIdx == -1 {
		panic("LoadProjectCSS not found")
	}

	newTraverse := `func (a *htmlAdapter) traverse(n *html.Node, filename string, lines []string, fullContent string, isComponent bool, inheritedBG string, inheritedFG string, inheritedHidden bool, opaqueMap map[string]string) []domain.USN {
	var nodes []domain.USN

	if n.Type == html.ElementNode {
		raw := a.renderNode(n)
		line, col := a.findPosition(n, fullContent)

		isOpaque, opaqueSource := a.getOpaqueComponentInfo(n, opaqueMap)
		ignoredRules := a.getA11yIgnoreRules(n)

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
				IgnoredRules: ignoredRules,
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
		a.processCSSClasses(n, &usn)
		a.processInlineStyles(n, &usn)

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
		nodes = append(nodes, a.traverse(c, filename, lines, fullContent, isComponent, childBG, childFG, childHidden, opaqueMap)...)
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
	default: return domain.Role(ariaRole)
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

func (a *htmlAdapter) processCSSClasses(n *html.Node, usn *domain.USN) {
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

`
	
	newContent := string(content[:startIdx]) + newTraverse + string(content[endIdx:])
	
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		panic(err)
	}
}
