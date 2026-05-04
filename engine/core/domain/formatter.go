package domain

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ToESLintStyle converts a list of violations into a human-readable format similar to ESLint.
func ToESLintStyle(violations []Violation, projectRoot string) string {
	if len(violations) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, v := range violations {
		relPath := v.SourceRef.FilePath
		if projectRoot != "" {
			if rel, err := filepath.Rel(projectRoot, v.SourceRef.FilePath); err == nil {
				relPath = rel
			}
		}

		severity := "error  "
		if v.Severity == SeverityWarning {
			severity = "warning"
		}

		// Format: path:line:col  severity  message  [code]
		fmt.Fprintf(&sb, "  %s:%d:%d  %s  %s  [%s]\n",
			relPath,
			v.SourceRef.Line,
			v.SourceRef.Column,
			severity,
			v.Message,
			v.ErrorCode)
	}

	return sb.String()
}

// ToTOON converts a list of violations into a Token-Oriented Object Notation string.
// Format: violations[count]{code,sev,file,line,snippet,msg,fix}:
func ToTOON(violations []Violation) string {
	if len(violations) == 0 {
		return "violations[0]{code,sev,file,line,snippet,msg,fix}:"
	}

	header := fmt.Sprintf("violations[%d]{code,sev,file,line,snippet,msg,fix}:", len(violations))
	var rows []string
	rows = append(rows, header)

	for _, v := range violations {
		// Truncate RawHTML to keep it lean
		snippet := strings.TrimSpace(v.SourceRef.RawHTML)
		if len(snippet) > 60 {
			snippet = snippet[:57] + "..."
		}
		// Escape commas in values to prevent CSV-style breakage
		snippet = strings.ReplaceAll(snippet, ",", ";")
		msg := strings.ReplaceAll(v.Message, ",", ";")
		fix := strings.ReplaceAll(v.FixSnippet, ",", ";")
		
		severity := "E"
		if v.Severity == SeverityWarning {
			severity = "W"
		}

		row := fmt.Sprintf("  %s,%s,%s,%d,%s,%s,%s",
			v.ErrorCode,
			severity,
			v.SourceRef.FilePath,
			v.SourceRef.Line,
			snippet,
			msg,
			fix,
		)
		rows = append(rows, row)
	}

	return strings.Join(rows, "\n")
}

// HierarchyToTOON converts a list of related file nodes into a Token-Oriented Object Notation string.
// Format: hierarchy[count]{file,rel,depth,isCycle}:
func HierarchyToTOON(root *FileNode, targetPath string) string {
	if root == nil {
		return "hierarchy[0]{file,rel,depth,isCycle}:"
	}

	var items []string
	flattenHierarchy(root, 0, "", &items)

	header := fmt.Sprintf("hierarchy[%d]{file,rel,depth,isCycle}:", len(items))
	return header + "\n" + strings.Join(items, "\n")
}

func flattenHierarchy(n *FileNode, depth int, rel string, res *[]string) {
	if n == nil {
		return
	}

	isCycle := "0"
	if n.IsCycle {
		isCycle = "1"
	}

	row := fmt.Sprintf("  %s,%s,%d,%s",
		n.FilePath,
		rel,
		depth,
		isCycle,
	)
	*res = append(*res, row)

	for _, child := range n.Children {
		flattenHierarchy(child, depth+1, "child", res)
	}
}
