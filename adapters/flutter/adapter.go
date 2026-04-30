package flutter

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type flutterAdapter struct{}

func NewFlutterAdapter() ports.Adapter {
	return &flutterAdapter{}
}

func (a *flutterAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}

		nodes := a.parseDart(string(content), file)
		allNodes = append(allNodes, nodes...)
	}

	return allNodes, nil
}

func (a *flutterAdapter) parseDart(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Regex for Semantics(label: "...")
	semanticsRegex := regexp.MustCompile(`Semantics\s*\(\s*[^)]*label\s*:\s*\"([^\"]*)\"`)
	// Regex for Text("...")
	textRegex := regexp.MustCompile(`Text\s*\(\s*\"([^\"]*)\"`)
	// Regex for Image.asset(..., semanticLabel: "...")
	imageRegex := regexp.MustCompile(`Image\s*\.[^(\n]*\(\s*[^)]*semanticLabel\s*:\s*\"([^\"]*)\"`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comment lines — single-line (//) and block comment markers (/* * */)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Match Semantics
		if matches := semanticsRegex.FindStringSubmatch(trimmed); matches != nil {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-flutter-semantics-%d", filename, i),
				Role:  domain.RoleLiveRegion, // Placeholder for semantic wrapper
				Label: matches[1],
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Semantics"),
					RawHTML:  trimmed,
				},
			})
		}

		// Match Images
		if matches := imageRegex.FindStringSubmatch(trimmed); matches != nil {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-flutter-img-%d", filename, i),
				Role:  domain.RoleImage,
				Label: matches[1],
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Image"),
					RawHTML:  trimmed,
				},
			})
		}

		// Match Buttons (ElevatedButton, TextButton, IconButton)
		if strings.Contains(trimmed, "Button") && !strings.Contains(trimmed, "ButtonStyle") {
			label := ""
			// Look ahead for Text() or Semantics label
			lookAhead := 5
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				if m := textRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
				if m := semanticsRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
			}
			// If no label found ahead, also look backward for a Semantics wrapper
			// (e.g. Semantics(label: "...") wrapping a child: IconButton(...))
			// Join lines into a block to handle multi-line Semantics declarations.
			if label == "" {
				lookBack := 5
				start := i - lookBack
				if start < 0 {
					start = 0
				}
				backBlock := strings.Join(lines[start:i], " ")
				if m := semanticsRegex.FindStringSubmatch(backBlock); m != nil {
					label = m[1]
				}
			}

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-flutter-btn-%d", filename, i),
				Role:  domain.RoleButton,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformFlutterDart,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Button"),
					RawHTML:  trimmed,
				},
			})
		}
	}

	return nodes
}
