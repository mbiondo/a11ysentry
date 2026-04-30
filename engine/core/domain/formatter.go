package domain

import (
	"fmt"
	"strings"
)

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
