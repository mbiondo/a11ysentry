package domain

import (
	"context"
	"testing"
)

func TestAccessibilityAnalyzer_Analyze(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()

	tests := []struct {
		name      string
		nodes     []USN
		wantCodes []string
	}{
		{
			name: "Missing Alt Text",
			nodes: []USN{
				{Role: RoleImage, Label: "", Source: Source{Platform: PlatformWebReact}},
			},
			wantCodes: []string{"WCAG_1_1_1", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Missing Interactive Name",
			nodes: []USN{
				{Role: RoleButton, Label: "", Source: Source{Platform: PlatformWebReact}},
			},
			wantCodes: []string{"WCAG_4_1_2", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Heading Jump",
			nodes: []USN{
				{UID: "h1", Role: RoleHeading, Label: "Main Title", Source: Source{Platform: PlatformWebReact}},
				{UID: "h3", Role: RoleHeading, Label: "Sub Section", Source: Source{Platform: PlatformWebReact}},
			},
			wantCodes: []string{"WCAG_1_3_1", "WCAG_3_1_1"}, // H1 is present, but Lang is missing
		},
		{
			name: "Low Contrast",
			nodes: []USN{
				{
					Role:   RoleHeading,
					Label:  "Section Title",
					Traits: map[string]any{"color": "#999999", "background-color": "#ffffff"},
					Source: Source{Platform: PlatformWebReact},
				},
			},
			wantCodes: []string{"WCAG_1_4_3", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Duplicate IDs",
			nodes: []USN{
				{Traits: map[string]any{"id": "test"}, Source: Source{Platform: PlatformWebReact}},
				{Traits: map[string]any{"id": "test"}, Source: Source{Platform: PlatformWebReact}},
			},
			wantCodes: []string{"WCAG_4_1_1", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Focus Visibility Hidden",
			nodes: []USN{
				{
					Role:   RoleButton,
					Label:  "Click me",
					Traits: map[string]any{"style": "outline: none;"},
					Source: Source{Platform: PlatformWebReact},
				},
			},
			wantCodes: []string{"WCAG_2_4_7", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Touch Target Too Small (Mobile)",
			nodes: []USN{
				{
					Role:   RoleButton,
					Label:  "Save",
					Traits: map[string]any{"width": 20.0, "height": 20.0},
					Source: Source{Platform: PlatformAndroidCompose},
				},
			},
			wantCodes: []string{"WCAG_2_5_5"},
		},
		{
			name: "Modal Missing Focus Trap",
			nodes: []USN{
				{
					Role:   RoleModal,
					Traits: map[string]any{"aria-modal": "false"},
					Source: Source{Platform: PlatformWebReact},
				},
			},
			wantCodes: []string{"ADV_FOCUS_TRAP", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Document Level - Web Missing Lang and H1",
			nodes: []USN{
				{UID: "html", Source: Source{Platform: PlatformWebReact}},
			},
			wantCodes: []string{"WCAG_3_1_1", "G141"},
		},
		{
			name: "Generic Element with Click but no Keyboard/Focus",
			nodes: []USN{
				{
					Role:   "generic",
					Traits: map[string]any{"onclick": "handleClick"},
					Source: Source{Platform: PlatformWebReact},
				},
			},
			wantCodes: []string{"WCAG_2_1_1", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Generic Element with Click and Keyboard/Focus",
			nodes: []USN{
				{
					Role:   "generic",
					Traits: map[string]any{"onclick": "handleClick", "onkeydown": "handleKey", "tabindex": "0"},
					Source: Source{Platform: PlatformWebReact},
				},
			},
			wantCodes: []string{"WCAG_3_1_1", "G141"},
		},
		{
			name: "Mobile - No H1/Lang required",
			nodes: []USN{
				{Role: RoleButton, Label: "Save", Source: Source{Platform: PlatformAndroidCompose}},
			},
			wantCodes: []string{},
		},
		{
			name: "Ambiguous Link Purpose",
			nodes: []USN{
				{Role: RoleLink, Label: "Read more", Traits: map[string]any{"href": "/a"}, Source: Source{Platform: PlatformWebReact, Line: 1, RawHTML: "<a href='/a'>Read more</a>"}},
				{Role: RoleLink, Label: "Read more", Traits: map[string]any{"href": "/b"}, Source: Source{Platform: PlatformWebReact, Line: 2, RawHTML: "<a href='/b'>Read more</a>"}},
			},
			wantCodes: []string{"WCAG_2_4_4", "WCAG_2_4_4", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Missing Autocomplete",
			nodes: []USN{
				{Role: RoleInput, Traits: map[string]any{"type": "email"}, Source: Source{Platform: PlatformWebReact, Line: 1, RawHTML: "<input type='email'>"}},
			},
			wantCodes: []string{"WCAG_3_3_2", "WCAG_1_3_5", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Focusable in Hidden Area",
			nodes: []USN{
				{Role: RoleButton, Label: "Submit", Traits: map[string]any{"aria-hidden-inherited": true}, Source: Source{Platform: PlatformWebReact, Line: 1, RawHTML: "<button>Submit</button>"}},
			},
			wantCodes: []string{"WCAG_2_4_3_HIDDEN", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Redundant Title",
			nodes: []USN{
				{Role: RoleLink, Label: "Home", Traits: map[string]any{"title": "Home"}, Source: Source{Platform: PlatformWebReact, Line: 1, RawHTML: "<a title='Home'>Home</a>"}},
			},
			wantCodes: []string{"REDUNDANT_TITLE", "WCAG_3_1_1", "G141"},
		},
		{
			name: "Inline a11y-ignore",
			nodes: []USN{
				{
					Role:  RoleImage,
					Label: "",
					Source: Source{
						Platform:     PlatformWebReact,
						IgnoredRules: []string{"WCAG_1_1_1"},
						RawHTML:      "<img>",
					},
				},
			},
			wantCodes: []string{"WCAG_3_1_1", "G141"}, // WCAG_1_1_1 should be filtered
		},
	}

	cfg := DefaultConfig()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations, err := analyzer.Analyze(ctx, tt.nodes, cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(violations) != len(tt.wantCodes) {
				gotCodes := []string{}
				for _, v := range violations {
					gotCodes = append(gotCodes, v.ErrorCode)
				}
				t.Errorf("expected %d violations %v, got %d %v", len(tt.wantCodes), tt.wantCodes, len(violations), gotCodes)
				return
			}
			
			gotCodes := make(map[string]int)
			for _, v := range violations {
				gotCodes[v.ErrorCode]++
			}
			
			wantCodes := make(map[string]int)
			for _, code := range tt.wantCodes {
				wantCodes[code]++
			}
			
			for code, count := range wantCodes {
				if gotCodes[code] != count {
					t.Errorf("expected code %s to appear %d times, got %d", code, count, gotCodes[code])
				}
			}
		})
	}
}
