package domain

import (
	"fmt"
	"math"
	"strings"
)

// CalculateContrast returns the contrast ratio between two hex colors.
func CalculateContrast(fg, bg string) float64 {
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

func isLandmark(role SemanticRole) bool {
	switch role {
	case RoleMain, RoleNav, RoleAside, RoleHeader, RoleFooter, RoleSection, RoleForm, RoleSearch:
		return true
	}
	return false
}
