package domain

import (
	"fmt"
	"math"
	"strings"
)

// CalculateContrast returns the contrast ratio between two hex colors.
// Returns -1 if either color is empty or invalid (cannot be parsed as 6-digit hex).
func CalculateContrast(fg, bg string) float64 {
	fgHex := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(fg)), "#")
	bgHex := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(bg)), "#")
	if len(fgHex) != 6 || len(bgHex) != 6 {
		return -1
	}
	l1 := GetLuminance(fg)
	l2 := GetLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// GetLuminance returns the relative luminance of a hex color.
func GetLuminance(hex string) float64 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0.5
	}
	var r, g, b uint8
	_, _ = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)

	rs := float64(r) / 255.0
	gs := float64(g) / 255.0
	bs := float64(b) / 255.0

	f := func(c float64) float64 {
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*f(rs) + 0.7152*f(gs) + 0.0722*f(bs)
}

// CalculateOpacity returns the percentage of opaque nodes in a slice of USNs.
func CalculateOpacity(nodes []USN) float64 {
	if len(nodes) == 0 {
		return 0.0
	}
	opaqueCount := 0
	for _, n := range nodes {
		if n.IsOpaque {
			opaqueCount++
		}
	}
	return (float64(opaqueCount) / float64(len(nodes))) * 100.0
}

func isLandmark(role SemanticRole) bool {
	switch role {
	case RoleMain, RoleNav, RoleAside, RoleHeader, RoleFooter, RoleSection, RoleForm, RoleSearch:
		return true
	}
	return false
}

// nodeHasAccessibleName reports whether a node carries a name an assistive technology can use.
//
// Label holds aria-label/alt/text-derived names. A node named through aria-labelledby or title is
// equally named, so rules that only look at Label would wrongly treat it as unnamed and demand a
// label it already has.
func nodeHasAccessibleName(node USN) bool {
	if node.Label != "" {
		return true
	}
	if v, ok := node.Traits["aria-labelledby"].(string); ok && v != "" {
		return true
	}
	if v, ok := node.Traits["title"].(string); ok && v != "" {
		return true
	}
	return false
}

// headingLevel returns the heading level (1-6) of a node, or 0 when it is not a heading.
//
// The level comes from the element's tag name rather than from UID: UID holds the id attribute
// when the element has one, so <h1 id="main-title"> would otherwise parse as level 0 and the page
// would look like it had no H1 and no heading hierarchy.
func headingLevel(node USN) int {
	tag, ok := node.Traits["tag"].(string)
	if !ok || tag == "" {
		tag = string(node.UID)
	}

	level := 0
	_, _ = fmt.Sscanf(tag, "h%d", &level)
	return level
}
