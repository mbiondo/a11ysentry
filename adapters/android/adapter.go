package android

import (
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type androidAdapter struct{}

func NewAndroidAdapter() ports.Adapter {
	return &androidAdapter{}
}

func (a *androidAdapter) Ingest(ctx context.Context, files []string) ([]domain.USN, error) {
	var allNodes []domain.USN
	nodeChan := make(chan []domain.USN, len(files))
	errChan := make(chan error, len(files))

	for _, file := range files {
		go func(f string) {
			content, err := os.ReadFile(f)
			if err != nil {
				errChan <- err
				return
			}

			ext := filepath.Ext(f)
			var nodes []domain.USN

			switch ext {
			case ".kt":
				nodes = a.parseCompose(string(content), f)
			case ".xml":
				nodes = a.parseXML(string(content), f)
			case ".java":
				nodes = a.parseJava(string(content), f)
			}
			nodeChan <- nodes
		}(file)
	}

	for i := 0; i < len(files); i++ {
		select {
		case nodes := <-nodeChan:
			allNodes = append(allNodes, nodes...)
		case err := <-errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return allNodes, nil
}

func (a *androidAdapter) parseCompose(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	imageRegex := regexp.MustCompile(`Image\s*\(`)
	contentDescRegex := regexp.MustCompile(`contentDescription\s*=\s*\"([^\"]*)\"`)
	nullContentDescRegex := regexp.MustCompile(`contentDescription\s*=\s*null`)
	textRegex := regexp.MustCompile(`Text\s*\(\s*\"([^\"]*)\"`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		if imageRegex.MatchString(trimmed) {
			label := ""
			lookAhead := 5
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				if m := contentDescRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
				if nullContentDescRegex.MatchString(lines[j]) {
					label = "" // Explicitly null
					break
				}
			}
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-compose-img-%d", filename, i),
				Role:  domain.RoleImage,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformAndroidCompose,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "Image"),
					RawHTML:  trimmed,
				},
			})
		}

		if strings.Contains(trimmed, "Button") {
			label := ""
			lookAhead := 5
			for j := i; j < i+lookAhead && j < len(lines); j++ {
				if m := textRegex.FindStringSubmatch(lines[j]); m != nil {
					label = m[1]
					break
				}
			}

			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-compose-btn-%d", filename, i),
				Role:  domain.RoleButton,
				Label: label,
				Source: domain.Source{
					Platform: domain.PlatformAndroidCompose,
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

func (a *androidAdapter) parseXML(content string, filename string) []domain.USN {
	var nodes []domain.USN
	decoder := xml.NewDecoder(strings.NewReader(content))

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		switch se := token.(type) {
		case xml.StartElement:
			role := domain.SemanticRole("")
			label := ""

			// Map common Android View tags to roles
			switch se.Name.Local {
			case "Button", "ImageButton":
				role = domain.RoleButton
			case "ImageView":
				role = domain.RoleImage
			case "EditText":
				role = domain.RoleInput
			case "TextView":
				// If it's a title or large text, it might be a heading,
				// but for now, we just ingest it to check labels.
			}

			if role != "" {
				traits := make(map[string]any)
				for _, attr := range se.Attr {
					if attr.Name.Local == "contentDescription" || attr.Name.Local == "text" || attr.Name.Local == "hint" {
						label = attr.Value
					}
					// Capture layout_width/height as hints for touch target
					if attr.Name.Local == "layout_width" || attr.Name.Local == "layout_height" {
						val := strings.TrimSuffix(attr.Value, "dp")
						var size float64
						fmt.Sscanf(val, "%f", &size)
						if size > 0 {
							key := "width"
							if attr.Name.Local == "layout_height" {
								key = "height"
							}
							traits[key] = size
						}
					}
				}

				nodes = append(nodes, domain.USN{
					UID:    fmt.Sprintf("%s-xml-%s-%d", filename, se.Name.Local, decoder.InputOffset()),
					Role:   role,
					Label:  label,
					Traits: traits,
					Source: domain.Source{
						Platform: domain.PlatformAndroidView,
						FilePath: filename,
						Line:     1,
						RawHTML:  fmt.Sprintf("<%s ...>", se.Name.Local),
					},
				})
			}
		}
	}

	return nodes
}

func (a *androidAdapter) parseJava(content string, filename string) []domain.USN {
	var nodes []domain.USN
	lines := strings.Split(content, "\n")

	// Look for setContentDescription("...")
	descRegex := regexp.MustCompile(`\.setContentDescription\s*\(\s*\"([^\"]*)\"`)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip comment lines
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		if matches := descRegex.FindStringSubmatch(trimmed); matches != nil {
			nodes = append(nodes, domain.USN{
				UID:   fmt.Sprintf("%s-java-desc-%d", filename, i),
				Role:  domain.RoleImage,
				Label: matches[1],
				Source: domain.Source{
					Platform: domain.PlatformAndroidView,
					FilePath: filename,
					Line:     i + 1,
					Column:   strings.Index(line, "setContentDescription"),
					RawHTML:  trimmed,
				},
			})
		}
	}

	return nodes
}
