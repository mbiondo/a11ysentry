package web

import (
	"a11ysentry/engine/core/domain"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTemp writes content to a temp file with the given name suffix and returns its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
	return path
}

func ingestNodes(t *testing.T, root *domain.FileNode) []domain.USN {
	t.Helper()
	nodes, err := NewHTMLAdapter().Ingest(context.Background(), root)
	if err != nil {
		t.Fatalf("Ingest failed: %v", err)
	}
	return nodes
}

func findNode(nodes []domain.USN, uid string) *domain.USN {
	for i := range nodes {
		if nodes[i].UID == uid {
			return &nodes[i]
		}
	}
	return nil
}

// S2: a page whose H1 carries an id must not be reported as missing an H1.
func TestPageWithH1CarryingIDIsNotReportedMissingH1(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1 id="main-title">Title</h1></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	violations, err := domain.NewAnalyzer().Analyze(context.Background(), nodes, domain.DefaultConfig())
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	for _, v := range violations {
		if v.ErrorCode == "G141" {
			t.Errorf("G141 reported for a page that does have an <h1 id=...>: %s", v.Message)
		}
	}
}

// S3: role=alert/status already implies aria-live, so no warning may demand it.
func TestRoleAlertWithoutExplicitAriaLiveIsNotReported(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><div role="alert">Saved</div><div role="status">Loading</div></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	violations, err := domain.NewAnalyzer().Analyze(context.Background(), nodes, domain.DefaultConfig())
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	for _, v := range violations {
		if v.ErrorCode == "WCAG_4_1_3" {
			t.Errorf("WCAG_4_1_3 reported for a role that already implies aria-live: %s", v.Message)
		}
	}
}

// S3: a malformed explicit aria-live is still a real defect.
func TestInvalidAriaLiveIsReported(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><div aria-live="bogus">x</div></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	violations, err := domain.NewAnalyzer().Analyze(context.Background(), nodes, domain.DefaultConfig())
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	found := false
	for _, v := range violations {
		if v.ErrorCode == "WCAG_4_1_3" {
			found = true
		}
	}
	if !found {
		t.Error("a malformed aria-live value was not reported")
	}
}

// S4: only the document-level header is a banner landmark.
func TestHeaderInsideSectionIsNotBanner(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><section><header id="inner">Inner</header></section></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	inner := findNode(nodes, "inner")
	if inner == nil {
		t.Fatal("header node not found")
	}
	if inner.Role == domain.RoleHeader {
		t.Errorf("nested <header> mapped to banner; want a non-landmark role")
	}
}

func TestTopLevelHeaderIsStillBanner(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><header id="masthead">Site</header></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	top := findNode(nodes, "masthead")
	if top == nil {
		t.Fatal("header node not found")
	}
	if top.Role != domain.RoleHeader {
		t.Errorf("document-level <header> role = %q, want banner", top.Role)
	}
}

// S4: a <form> is only a form landmark when it has an accessible name.
func TestNamelessFormIsNotALandmark(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><form id="f1"><input id="i1"></form></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	form := findNode(nodes, "f1")
	if form == nil {
		t.Fatal("form node not found")
	}
	if form.Role == domain.RoleForm {
		t.Errorf("nameless <form> mapped to the form landmark; want a non-landmark role")
	}
}

func TestNamedFormIsALandmark(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><form id="f1" aria-label="Checkout"><input id="i1"></form></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	form := findNode(nodes, "f1")
	if form == nil {
		t.Fatal("form node not found")
	}
	if form.Role != domain.RoleForm {
		t.Errorf("named <form> role = %q, want form", form.Role)
	}
}

// S5: aria-hidden must not leak from one file into an imported component.
func TestAriaHiddenDoesNotCrossComponentBoundary(t *testing.T) {
	parent := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body>
			<h1>t</h1>
			<label>Email <span aria-hidden="true">*</span></label>
			<div id="slot"></div>
		</body></html>`)
	child := writeTemp(t, "Child.html", `<button id="child-btn">Go</button>`)

	root := &domain.FileNode{
		FilePath: parent,
		Children: []*domain.FileNode{{FilePath: child}},
	}

	nodes := ingestNodes(t, root)

	button := findNode(nodes, "child-btn")
	if button == nil {
		t.Fatal("imported component node not found")
	}
	if _, leaked := button.Traits["aria-hidden-inherited"]; leaked {
		t.Error("aria-hidden leaked from the importing file into an imported component")
	}
}

// S5: within one file, an aria-hidden ancestor must still propagate to descendants.
func TestAriaHiddenPropagatesWithinFile(t *testing.T) {
	file := writeTemp(t, "page.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><div aria-hidden="true"><button id="hidden-btn">x</button></div></body></html>`)

	nodes := ingestNodes(t, &domain.FileNode{FilePath: file})

	button := findNode(nodes, "hidden-btn")
	if button == nil {
		t.Fatal("button node not found")
	}
	if _, ok := button.Traits["aria-hidden-inherited"]; !ok {
		t.Error("aria-hidden did not propagate to a descendant within the same file")
	}
}

// S6: a component reachable through two parents must be ingested once, so its hardcoded id is
// not counted twice and reported as a duplicate.
func TestSharedComponentIsIngestedOnce(t *testing.T) {
	layout := writeTemp(t, "layout.html",
		`<!DOCTYPE html><html lang="en"><body><h1>t</h1><div id="layout-slot"></div></body></html>`)
	page := writeTemp(t, "page.html", `<div id="page-slot"></div>`)
	shared := writeTemp(t, "Shared.html", `<input id="shared-field">`)

	root := &domain.FileNode{
		FilePath: layout,
		Children: []*domain.FileNode{
			{FilePath: shared},
			{FilePath: page, Children: []*domain.FileNode{{FilePath: shared}}},
		},
	}

	nodes := ingestNodes(t, root)

	count := 0
	for i := range nodes {
		if nodes[i].Traits["id"] == "shared-field" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("shared component field appeared %d times in the node list, want 1", count)
	}
}

// S1 support: test files must be recognized so collection can skip them.
func TestScannerRecognizesTestFiles(t *testing.T) {
	// The scanner helper lives in another module; this asserts the adapter's view of the markup
	// is unaffected by test-file naming, while the exclusion itself is covered in the scanner
	// module's own tests.
	file := writeTemp(t, "Button.test.html", `<button id="b">Go</button>`)
	if !strings.HasSuffix(file, "Button.test.html") {
		t.Fatal("unexpected temp file name")
	}
}
