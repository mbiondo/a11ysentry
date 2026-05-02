package domain

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// Severity represents whether a violation is definitive or context-dependent.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Violation represents an accessibility issue found by the engine.
type Violation struct {
	ErrorCode        string
	Severity         Severity
	Message          string
	SourceRef        Source
	FixSnippet       string
	DocumentationURL string
}

// ViolationReport aggregates all violations for a single analysis session.
type ViolationReport struct {
	ID          int64
	ProjectName string
	ProjectRoot string
	FilePath    string
	Platform    Platform
	Timestamp   int64 // Unix timestamp
	Violations  []Violation
}

// Analyzer is the core logic that takes USN nodes and returns violations.
type Analyzer interface {
	Analyze(ctx context.Context, nodes []USN, cfg ProjectConfig) ([]Violation, error)
}

type accessibilityAnalyzer struct{}

func NewAnalyzer() Analyzer {
	return &accessibilityAnalyzer{}
}

func (a *accessibilityAnalyzer) Analyze(ctx context.Context, nodes []USN, cfg ProjectConfig) ([]Violation, error) {
	var violations []Violation

	// Internal helper to add violations based on project config
	add := func(v Violation) {
		ruleCfg, found := cfg.Rules[v.ErrorCode]
		if found {
			if !ruleCfg.Enabled {
				return
			}
			if ruleCfg.Severity != "" {
				v.Severity = ruleCfg.Severity
			}
		}
		violations = append(violations, v)
	}

	usedIDs := make(map[string]bool)
	labelsByFor := make(map[string]string)
	landmarkLabels := make(map[SemanticRole]map[string]Source) // role -> label -> first_source
	mainCount := 0
	lastHeadingLevel := 0
	hasH1 := false
	hasLang := false

	// Pass 1: Collect metadata (Labels for inputs, doc info, landmarks)
	for _, node := range nodes {
		// Identify Web document-level traits
		if node.UID == "html" || node.UID == "html-tag" {
			if lang, ok := node.Traits["lang"].(string); ok && lang != "" {
				hasLang = true
			}
		}

		// Collect <label for="..."> — support both 'for' (HTML) and 'htmlFor' (JSX/React)
		forAttr := ""
		if f, ok := node.Traits["for"].(string); ok && f != "" {
			forAttr = f
		} else if f, ok := node.Traits["htmlFor"].(string); ok && f != "" {
			forAttr = f
		}
		if forAttr != "" {
			if node.Label != "" {
				labelsByFor[forAttr] = node.Label
			}
		}

		// Initialize landmark tracking maps
		if _, exists := landmarkLabels[node.Role]; !exists && isLandmark(node.Role) {
			landmarkLabels[node.Role] = make(map[string]Source)
		}
	}

	// Pass 2: Rule validation
	for i, node := range nodes {
		p := node.Source.Platform
		isMobile := p == PlatformAndroidCompose || p == PlatformAndroidView || p == PlatformIOSSwiftUI || p == PlatformFlutterDart || p == PlatformReactNative
		isGaming := p == PlatformUnity || p == PlatformGodot

		// Rule 1: Images (WCAG 1.1.1)
		if node.Role == RoleImage && node.Label == "" {
			msg := "Image missing alternative text."
			if isMobile {
				msg = "Mobile image missing content description."
			} else if isGaming {
				msg = "Gaming texture/sprite missing accessibility label."
			}

			add(Violation{
				ErrorCode:        "WCAG_1_1_1",
				Severity:         SeverityError,
				Message:          fmt.Sprintf("%s Every image must have an 'alt', 'aria-label', or platform-specific description attribute.", msg),
				SourceRef:        node.Source,
				FixSnippet:       "Add a descriptive label for users with screen readers.",
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G94",
			})
		}

		// Rule 2 & 8: Form Labels & Interactive Names (WCAG 4.1.2, 3.3.2)
		if node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput {
			effectiveLabel := node.Label
			if node.Role == RoleInput {
				if id, ok := node.Traits["id"].(string); ok {
					if l, found := labelsByFor[id]; found {
						effectiveLabel = l
					}
				}
			}

			if effectiveLabel == "" {
				code := "WCAG_4_1_2"
				if node.Role == RoleInput {
					code = "WCAG_3_3_2"
				}
				add(Violation{
					ErrorCode:        code,
					Severity:         SeverityError,
					Message:          fmt.Sprintf("%s missing accessible name or label.", node.Role),
					SourceRef:        node.Source,
					FixSnippet:       "Ensure the control has a visible label or an internal accessibility name.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}
		}

		// Rule 3: Heading order (WCAG 1.3.1)
		if node.Role == RoleHeading {
			level := 0
			_, _ = fmt.Sscanf(string(node.UID), "h%d", &level)
			if level == 1 {
				hasH1 = true
			}
			if level > lastHeadingLevel+1 && lastHeadingLevel != 0 {
				add(Violation{
					ErrorCode:        "WCAG_1_3_1",
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Heading levels should only increase by one. Jumped from H%d to H%d.", lastHeadingLevel, level),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Adjust level to H%d to maintain hierarchy.", lastHeadingLevel+1),
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G141",
				})
			}
			lastHeadingLevel = level
		}

		// Rule 4: Unique IDs (WCAG 4.1.1)
		if id, ok := node.Traits["id"].(string); ok && id != "" {
			if usedIDs[id] {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_4_1_1",
					Severity:         SeverityError,
					Message:          fmt.Sprintf("Duplicate ID found: '%s'. IDs must be unique for focus management.", id),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("id=\"%s-unique\"", id),
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H93",
				})
			}
			usedIDs[id] = true
		}

		// Advanced Rules: Landmarks (WCAG 2.4.1 / ARIA 1.1)
		if isLandmark(node.Role) {
			if node.Role == RoleMain {
				mainCount++
				if mainCount > 1 {
					add(Violation{
						ErrorCode:        "WCAG_2_4_1",
						Severity:         SeverityError,
						Message:          "Multiple <main> elements found. A document should only have one primary content landmark.",
						SourceRef:        node.Source,
						FixSnippet:       "Remove redundant <main> elements or convert them to <section>.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H64",
					})
				}
			}

			// Track unique labels for landmarks of same type
			if node.Label != "" {
				landmarkLabels[node.Role][node.Label] = node.Source
			}

			// Check if multiple landmarks of same type exist without labels
			sameRoleCount := 0
			for _, n2 := range nodes {
				if n2.Role == node.Role {
					sameRoleCount++
				}
			}
			if sameRoleCount > 1 && node.Label == "" && node.Role != RoleMain {
				add(Violation{
					ErrorCode:        "ARIA_1_1",
					Severity:         SeverityWarning,
					Message:          fmt.Sprintf("Multiple %s landmarks found without distinguishing labels. Screen reader users won't know the difference between them.", node.Role),
					SourceRef:        node.Source,
					FixSnippet:       fmt.Sprintf("Add aria-label=\"...\" to distinguish this %s from others.", node.Role),
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA6",
				})
			}
		}

		// Grouping: Fieldset Legend (WCAG 1.3.1)
		if node.Role == RoleFieldset {
			hasLegend := false
			// Check immediate children (simplified check in current USN structure)
			// In a more robust engine, we'd check node.Hierarchy.Children
			for _, child := range nodes[i+1:] {
				// This is a naive check assuming legend follows fieldset closely in the flat slice
				// A real tree walker would be better.
				if child.Role == RoleLegend {
					hasLegend = true
					break
				}
				// If we encounter another fieldset or a major structural element, stop searching
				if child.Role == RoleFieldset || child.Role == RoleMain || child.Role == RoleHeader {
					break
				}
			}
			if !hasLegend {
				add(Violation{
					ErrorCode:        "WCAG_1_3_1_LEGEND",
					Severity:         SeverityError,
					Message:          "<fieldset> is missing a <legend>. Grouped inputs need a descriptive legend for context.",
					SourceRef:        node.Source,
					FixSnippet:       "<legend>Group Description</legend>",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H71",
				})
			}
		}

		// Rule 5: Focus Visibility (WCAG 2.4.7) — expanded
		if node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput {
			if style, ok := node.Traits["style"].(string); ok {
				styleLower := strings.ReplaceAll(style, " ", "")
				if strings.Contains(styleLower, "outline:none") ||
					strings.Contains(styleLower, "outline:0") ||
					strings.Contains(styleLower, "outline:0px") {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_2_4_7",
						Severity:         SeverityError,
						Message:          "Focus indicator hidden via 'outline: none'. Keyboard users won't know where the focus is.",
						SourceRef:        node.Source,
						FixSnippet:       "Remove 'outline: none' or provide a high-contrast custom :focus style.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G149",
					})
				}
			}
			// Detect tabindex < 0 on interactive elements (removes from tab order)
			if tabindex, ok := node.Traits["tabindex"].(string); ok {
				if tabindex == "-1" {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_2_4_3",
						Severity:         SeverityWarning,
						Message:          "Interactive element has tabindex=\"-1\", removing it from the natural tab order. Ensure it is reachable via a custom focus management strategy.",
						SourceRef:        node.Source,
						FixSnippet:       "Remove tabindex=\"-1\" unless you are managing focus programmatically (e.g. modal, roving tabindex).",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
					})
				}
			}
		}

		// Rule 11: WCAG 2.4.3 — Positive tabindex breaks natural focus order
		if tabindex, ok := node.Traits["tabindex"].(string); ok {
			var tabVal int
			if n, err := fmt.Sscanf(tabindex, "%d", &tabVal); n == 1 && err == nil && tabVal > 0 {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_4_3",
					Severity:         SeverityWarning,
					Message:          fmt.Sprintf("tabindex=\"%d\" (positive value) forces a custom focus order that may confuse keyboard and AT users.", tabVal),
					SourceRef:        node.Source,
					FixSnippet:       "Use tabindex=\"0\" to include elements in the natural DOM order instead of forcing a custom sequence.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H4",
				})
			}
		}

		// Rule 12: WCAG 1.4.11 — Non-text contrast for UI components (3:1 minimum).
		// Applies to inputs: check border-color against background (or white if unknown).
		if node.Role == RoleInput {
			borderColor, hasBorder := node.Traits["border-color"].(string)
			bg, hasBg := node.Traits["background-color"].(string)
			if !hasBg {
				bg = "#ffffff" // assume white background when unknown
			}
			if hasBorder {
				ratio := a.calculateContrast(borderColor, bg)
				if ratio < 3.0 {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_11",
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Input border has insufficient contrast ratio (%.2f:1). Minimum for UI components is 3:1.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Increase the border color contrast so the input boundary is clearly visible against its background.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G195",
					})
				}
			}
		}

		// Rule 13: WCAG 4.1.3 — Status messages must use a live region so AT announces
		// them without moving focus. Elements with role="status", "alert", or "log" that
		// lack an aria-live attribute may not be announced to screen reader users.
		if roleAttr, ok := node.Traits["role"].(string); ok {
			switch roleAttr {
			case "status", "alert", "log":
				if _, hasLive := node.Traits["aria-live"].(string); !hasLive {
					implicit := map[string]string{
						"status": "polite",
						"alert":  "assertive",
						"log":    "polite",
					}
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_4_1_3",
						Severity:         SeverityWarning,
						Message:          fmt.Sprintf("Element with role=\"%s\" is missing an explicit aria-live attribute. Some AT may not announce it automatically.", roleAttr),
						SourceRef:        node.Source,
						FixSnippet:       fmt.Sprintf("Add aria-live=\"%s\" to ensure screen readers announce status changes without requiring focus.", implicit[roleAttr]),
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA22",
					})
				}
			}
		}

		// Rule 14: WCAG 2.4.6 — Headings must be descriptive (non-empty).
		if node.Role == RoleHeading && node.Label == "" {
			violations = append(violations, Violation{
				ErrorCode:        "WCAG_2_4_6",
				Severity:         SeverityWarning,
				Message:          "Heading element has no visible text content. Headings must be descriptive to aid navigation.",
				SourceRef:        node.Source,
				FixSnippet:       "Add meaningful text to the heading or remove it if it is decorative.",
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G130",
			})
		}

		// Rule 15: WCAG 1.4.1 — Color must not be the sole means of conveying information.
		// Detectable signal: a link with text-decoration removed may rely on color alone to
		// distinguish itself from surrounding text (no underline, no bold, no other visual cue).
		if node.Role == RoleLink {
			style, _ := node.Traits["style"].(string)
			noUnderline := strings.Contains(strings.ReplaceAll(style, " ", ""), "text-decoration:none") ||
				strings.Contains(strings.ReplaceAll(style, " ", ""), "text-decoration-line:none")
			if noUnderline {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_1_4_1",
					Severity:         SeverityWarning,
					Message:          "Link has 'text-decoration: none' — if the only visual distinction from surrounding text is color, this violates WCAG 1.4.1.",
					SourceRef:        node.Source,
					FixSnippet:       "Ensure links are visually distinguishable from body text by at least one non-color cue (underline, bold, outline on hover, etc.).",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G183",
				})
			}
			// Also flag Tailwind no-underline class.
			if noUnderlineClass, ok := node.Traits["no-underline"].(bool); ok && noUnderlineClass {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_1_4_1",
					Severity:         SeverityWarning,
					Message:          "Link has 'no-underline' class — if the only visual distinction from surrounding text is color, this violates WCAG 1.4.1.",
					SourceRef:        node.Source,
					FixSnippet:       "Ensure links are visually distinguishable from body text by at least one non-color cue (underline, bold, outline on hover, etc.).",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G183",
				})
			}
		}

		// Rule 16: Keyboard Navigation for clickable elements (WCAG 2.1.1)
		// Elements with click handlers must have keyboard handlers and be focusable.
		hasMouseClick := node.Traits["onclick"] != nil || node.Traits["@click"] != nil || node.Traits["v-on:click"] != nil || node.Traits["(click)"] != nil || node.Traits["on:click"] != nil
		hasKeyboard := node.Traits["onkeydown"] != nil || node.Traits["@keydown"] != nil || node.Traits["v-on:keydown"] != nil || node.Traits["(keydown)"] != nil || node.Traits["on:keydown"] != nil ||
			node.Traits["onkeyup"] != nil || node.Traits["@keyup"] != nil || node.Traits["v-on:keyup"] != nil || node.Traits["(keyup)"] != nil || node.Traits["on:keyup"] != nil ||
			node.Traits["onkeypress"] != nil || node.Traits["@keypress"] != nil || node.Traits["v-on:keypress"] != nil || node.Traits["(keypress)"] != nil || node.Traits["on:keypress"] != nil
		hasTabIndex := node.Traits["tabindex"] != nil

		isNativeInteractive := node.Role == RoleButton || node.Role == RoleLink || node.Role == RoleInput

		if hasMouseClick && !isNativeInteractive {
			if !hasKeyboard || !hasTabIndex {
				violations = append(violations, Violation{
					ErrorCode:        "WCAG_2_1_1",
					Severity:         SeverityError,
					Message:          "Non-interactive element with a click handler is missing keyboard support (keydown/keyup) or is not focusable (tabindex).",
					SourceRef:        node.Source,
					FixSnippet:       "Add a keydown handler and tabindex=\"0\", or change the element to a native <button>.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G90",
				})
			}
		}

		// Rule 6: Color Contrast (WCAG 1.4.3)
		if node.Role == RoleHeading || node.Role == RoleButton || node.Role == "generic" {
			fg, hasFg := node.Traits["color"].(string)
			bg, hasBg := node.Traits["background-color"].(string)
			if hasFg && hasBg {
				ratio := a.calculateContrast(fg, bg)
				if ratio < 4.5 {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3",
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Low contrast ratio (%.2f:1). Target is 4.5:1.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Adjust colors to meet WCAG AA standards.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G18",
					})
				}
			} else if hasFg || hasBg {
				// At least one color was declared but the other could not be resolved.
				// Only warn for text-bearing elements — headings, interactive elements,
				// and generic elements that have visible text (Label != "").
				// Pure structural containers (div, article, header) without their own
				// text content are excluded to avoid noise.
				isTextBearer := node.Role == RoleHeading ||
					node.Role == RoleButton ||
					node.Role == RoleLink ||
					(node.Role == "generic" && node.Label != "")
				if isTextBearer {
					missing := "color and background-color"
					if hasFg {
						missing = "background-color"
					} else if hasBg {
						missing = "color"
					}
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_UNRESOLVED",
						Severity:         SeverityWarning,
						Message:          fmt.Sprintf("Color contrast could not be validated: %s is not statically resolvable (CSS variable, custom token, or runtime value).", missing),
						SourceRef:        node.Source,
						FixSnippet:       "Provide the resolved color values via --css or ensure colors are defined in static CSS/Tailwind classes.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G18",
					})
				}
			}

			// Rule 6b: Dark mode contrast (WCAG 1.4.3) — check dark: Tailwind overrides.
			darkFg, hasDarkFg := node.Traits["dark:color"].(string)
			darkBg, hasDarkBg := node.Traits["dark:background-color"].(string)
			if hasDarkFg && hasDarkBg {
				ratio := a.calculateContrast(darkFg, darkBg)
				if ratio < 4.5 {
					violations = append(violations, Violation{
						ErrorCode:        "WCAG_1_4_3_DARK",
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Dark mode contrast ratio (%.2f:1) is below 4.5:1 minimum.", ratio),
						SourceRef:        node.Source,
						FixSnippet:       "Adjust the dark: color classes to ensure sufficient contrast in dark mode.",
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G18",
					})
				}
			}
			// Rule 7: Touch Target Size (WCAG 2.5.5 & SC 2.5.8)
			if node.Role == RoleButton || node.Role == RoleLink {
				w, _ := node.Traits["width"].(float64)
				h, _ := node.Traits["height"].(float64)

				targetSize := 44.0 // Mobile default
				errorCode := "WCAG_2_5_5"

				if !isMobile && !isGaming {
					targetSize = 24.0 // Web WCAG 2.2 SC 2.5.8
					errorCode = "WCAG_2_5_8"
				}

				if (w > 0 && w < targetSize) || (h > 0 && h < targetSize) {
					violations = append(violations, Violation{
						ErrorCode:        errorCode,
						Severity:         SeverityError,
						Message:          fmt.Sprintf("Touch target too small (%.0fx%.0f). Target is %.0fx%.0f px/DP.", w, h, targetSize, targetSize),
						SourceRef:        node.Source,
						FixSnippet:       fmt.Sprintf("Increase the size to at least %.0fx%.0f.", targetSize, targetSize),
						DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G202",
					})
				}
			}

			// Rule 10: Aria State for Interactive Elements
			if node.Role == RoleButton {
				role, _ := node.Traits["role"].(string)
				if role == "switch" || role == "checkbox" || role == "toggle" {
					_, hasPressed := node.Traits["aria-pressed"].(string)
					_, hasExpanded := node.Traits["aria-expanded"].(string)
					_, hasChecked := node.Traits["aria-checked"].(string)

					if !hasPressed && !hasExpanded && !hasChecked {
						violations = append(violations, Violation{
							ErrorCode:        "WCAG_4_1_2",
							Severity:         SeverityError,
							Message:          "Interactive button acting as toggle/switch missing state attribute (aria-pressed, aria-expanded, or aria-checked).",
							SourceRef:        node.Source,
							FixSnippet:       "Add the appropriate aria-state attribute to reflect the current toggle condition.",
							DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA5",
						})
					}
				}
			}
		}

		// Final Web-only doc checks
		// Rule 9: Focus Trap for Modals (Advanced)
		if node.Role == RoleModal {
			hasAriaModal := node.Traits["aria-modal"] == "true"
			if !hasAriaModal {
				add(Violation{
					ErrorCode:        "ADV_FOCUS_TRAP",
					Severity:         SeverityError,
					Message:          "Modal/Dialog missing 'aria-modal=\"true\"'. Without it, screen readers may allow users to navigate outside the modal while it is active.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-modal=\"true\" to the dialog element.",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA14",
				})
			}
		}

		// Rule 13: Live Regions (WCAG 4.1.3)
		if node.Role == RoleLiveRegion {
			_, hasLive := node.Traits["aria-live"].(string)
			if !hasLive {
				add(Violation{
					ErrorCode:        "WCAG_4_1_3",
					Severity:         SeverityWarning,
					Message:          "Live region missing explicit 'aria-live' attribute. Dynamic content updates may not be announced to screen reader users.",
					SourceRef:        node.Source,
					FixSnippet:       "Add aria-live=\"polite\" (for status updates) or aria-live=\"assertive\" (for critical alerts).",
					DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/aria/ARIA19",
				})
			}
		}
	}

	// Final Web-only doc checks
	isWebProject := false
	isComponent := false
	for _, n := range nodes {
		p := n.Source.Platform
		if p == PlatformWebReact || p == PlatformWebVue || p == PlatformWebSvelte ||
			p == PlatformWebAngular || p == PlatformWebAstro ||
			p == PlatformBlazor || p == PlatformElectron || p == PlatformTauri {
			isWebProject = true
		}
		if n.Source.IsComponent {
			isComponent = true
		}
	}

	if isWebProject {
		// Components (no <html> root) may legitimately omit lang and H1 — report as WARNING.
		severity := SeverityError
		contextNote := ""
		if isComponent {
			severity = SeverityWarning
			contextNote = " (component file - verify if this is a root document or a nested component)"
		}

		// Use source info from the first node if available
		var firstSource Source
		if len(nodes) > 0 {
			// Find the first node that has a real line number (not synthetic)
			for _, n := range nodes {
				if n.Source.Line > 0 {
					firstSource = n.Source
					break
				}
			}
			// Fallback to first node if none have lines (unlikely but safe)
			if firstSource.Line == 0 {
				firstSource = nodes[0].Source
			}
		}
		
		// If still 0, default to 1
		if firstSource.Line == 0 {
			firstSource.Line = 1
			firstSource.Column = 1
		}

		if !hasLang {
			violations = append(violations, Violation{
				ErrorCode:        "WCAG_3_1_1",
				Severity:         severity,
				Message:          "Document missing language attribute." + contextNote,
				SourceRef:        firstSource,
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/html/H57",
			})
		}
		if !hasH1 {
			violations = append(violations, Violation{
				ErrorCode:        "G141",
				Severity:         severity,
				Message:          "Page missing an H1 heading." + contextNote,
				SourceRef:        firstSource,
				DocumentationURL: "https://www.w3.org/WAI/WCAG22/Techniques/general/G141",
			})
		}
	}

	// Deduplicate violations with the same errorCode + rawHTML to avoid
	// reporting the same pattern N times (e.g. repeated components at the same source).
	seen := make(map[string]bool)
	unique := violations[:0]
	for _, v := range violations {
		key := v.ErrorCode + "|" + v.SourceRef.RawHTML
		if !seen[key] {
			seen[key] = true
			unique = append(unique, v)
		}
	}
	return unique, nil
}

func (a *accessibilityAnalyzer) calculateContrast(fg, bg string) float64 {
	l1 := a.getLuminance(fg)
	l2 := a.getLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func (a *accessibilityAnalyzer) getLuminance(hex string) float64 {
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
