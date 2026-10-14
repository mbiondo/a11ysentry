package scanner

import "testing"

// S1: test, spec, mock and snapshot files must be recognized so collection can skip them. They
// hold fixtures and assertions, not rendered UI, so auditing them yields false positives such as
// "page missing an H1".
func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Test and spec files.
		{"src/components/Button.test.tsx", true},
		{"src/components/Button.spec.tsx", true},
		{"src/lib/api.test.ts", true},
		{"src/lib/api.spec.js", true},
		{"src/components/Card.test.vue", true},

		// Test directories.
		{"src/__tests__/Button.tsx", true},
		{"src/components/__tests__/Header.tsx", true},
		{"src/__mocks__/api.ts", true},
		{"src/__snapshots__/Button.tsx.snap", true},

		// Other languages.
		{"api/test_views.py", true},
		{"api/views_test.py", true},
		{"lib/widget_test.dart", true},
		{"Sources/ButtonTests.swift", true},
		{"android/ButtonTest.kt", true},
		{"dotnet/ButtonTests.cs", true},

		// Ordinary UI files must NOT be excluded.
		{"src/components/Button.tsx", false},
		{"src/pages/index.astro", false},
		{"app/layout.tsx", false},
		{"src/lib/latest.ts", false},
		{"src/components/Testimonial.tsx", false},
		{"src/components/Contest.tsx", false},
		{"src/components/Spec.md", false},
		// A file merely named with "test" as a substring is not a test file.
		{"src/components/ButtonTestCase.tsx", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsTestFile(tt.path); got != tt.want {
				t.Errorf("IsTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
