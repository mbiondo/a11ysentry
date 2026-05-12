package domain

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
)

type ruleWCAG311 struct{}

func (r *ruleWCAG311) Name() string             { return "Language of Page" }
func (r *ruleWCAG311) ErrorCode() string        { return "WCAG_3_1_1" }
func (r *ruleWCAG311) ACTID() string            { return "bf051a" }
func (r *ruleWCAG311) DocumentationURL() string { return "https://www.w3.org/WAI/WCAG22/Techniques/html/H57" }

// langCodeRegex matches basic ISO 639-1/2/3 codes, optionally with region (e.g., en, en-US, es-419)
var langCodeRegex = regexp.MustCompile(`^[a-z]{2,3}(-[a-zA-Z0-9]{2,4})?$`)

// documentRootFiles are the only file names that are expected to own the <html lang> attribute.
// page.tsx, loading.tsx, error.tsx, components, etc. are NOT document roots.
var documentRootFiles = map[string]bool{
	"layout.tsx":    true,
	"layout.jsx":    true,
	"layout.js":     true,
	"_document.tsx": true,
	"_document.jsx": true,
	"_document.js":  true,
	"_app.tsx":      true,
	"_app.jsx":      true,
	"_app.js":       true,
	"index.html":    true,
	"index.htm":     true,
}

// isDocumentRootFile returns true only for files that are expected to own the <html> tag.
func isDocumentRootFile(filePath string) bool {
	base := strings.ToLower(filepath.Base(filePath))
	return documentRootFiles[base]
}

func (r *ruleWCAG311) Execute(ctx context.Context, analysisCtx *AnalysisContext) ([]Violation, error) {
	var violations []Violation

	isWebProject := false
	var rootNode *USN

	for i := range analysisCtx.Nodes {
		node := &analysisCtx.Nodes[i]
		p := node.Source.Platform
		if isWebPlatform(p) {
			isWebProject = true
		}
		// Identify the root html tag
		if node.UID == "html" || node.UID == "html-tag" || strings.ToLower(node.Label) == "html" {
			rootNode = node
		}
	}

	if !isWebProject {
		return nil, nil
	}

	// Case 1: we found a real <html> node — validate it has a valid lang attribute.
	// Only enforce on known document-root files; the HTML parser adds implicit <html> nodes
	// in fragment files (components, pages) even when they don't actually contain one.
	if rootNode != nil && !rootNode.Source.IsComponent && isDocumentRootFile(rootNode.Source.FilePath) {
		lang, _ := rootNode.Traits["lang"].(string)
		if lang == "" {
			violations = append(violations, Violation{
				ErrorCode:        r.ErrorCode(),
				Severity:         SeverityError,
				Message:          "The <html> element must have a 'lang' attribute to specify the document's language.",
				SourceRef:        rootNode.Source,
				DocumentationURL: r.DocumentationURL(),
			})
		} else if !langCodeRegex.MatchString(strings.ToLower(lang)) {
			violations = append(violations, Violation{
				ErrorCode:        "WCAG_3_1_1_INVALID",
				Severity:         SeverityError,
				Message:          "The 'lang' attribute value '" + lang + "' is not a valid BCP 47 language tag.",
				SourceRef:        rootNode.Source,
				DocumentationURL: r.DocumentationURL(),
			})
		}
		return violations, nil
	}

	// Case 2: no <html> node found.
	// Only report if the analysis source is a known document-root file.
	// Components, pages (page.tsx), loading.tsx, error.tsx etc. are NOT responsible for <html lang>.
	if rootNode == nil && !analysisCtx.HasLang {
		firstSource := getFirstAvailableSource(analysisCtx.Nodes)
		if isDocumentRootFile(firstSource.FilePath) {
			violations = append(violations, Violation{
				ErrorCode:        r.ErrorCode(),
				Severity:         SeverityError,
				Message:          "Document is missing the <html> tag or the 'lang' attribute. Screen readers cannot identify the page language.",
				SourceRef:        firstSource,
				DocumentationURL: r.DocumentationURL(),
			})
		}
	}

	return violations, nil
}

func getFirstAvailableSource(nodes []USN) Source {
	if len(nodes) == 0 {
		return Source{Line: 1, Column: 1}
	}
	for _, n := range nodes {
		if n.Source.Line > 0 {
			return n.Source
		}
	}
	// Fallback to the first node but force valid coordinates
	s := nodes[0].Source
	if s.Line == 0 {
		s.Line = 1
	}
	if s.Column == 0 {
		s.Column = 1
	}
	return s
}
