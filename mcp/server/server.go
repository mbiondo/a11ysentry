package server

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"a11ysentry/adapters/android"
	"a11ysentry/adapters/blazor"
	"a11ysentry/adapters/dotnet"
	"a11ysentry/adapters/flutter"
	"a11ysentry/adapters/godot"
	"a11ysentry/adapters/ios"
	"a11ysentry/adapters/javadesktop"
	"a11ysentry/adapters/reactnative"
	"a11ysentry/adapters/unity"
	"a11ysentry/adapters/web"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"a11ysentry/scanner"
	androidfw "a11ysentry/scanner/android"
	astrofw "a11ysentry/scanner/astro"
	"a11ysentry/scanner/generic"
	iosfw "a11ysentry/scanner/ios"
	"a11ysentry/scanner/nextjs"
	"a11ysentry/scanner/nuxt"
	"a11ysentry/scanner/sveltekit"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func Start() {
	// Create MCP server
	s := mcpserver.NewMCPServer(
		"A11ySentry MCP Server",
		"1.0.0",
		mcpserver.WithLogging(),
	)

	// Register the accessibility analysis tool
	tool := mcp.NewTool("analyze_accessibility",
		mcp.WithDescription("Audit source files or full project directories for accessibility violations. If a directory is provided, it automatically discovers project roots, resolves component trees, and audits them in context. For single files, supports comma-separated paths for multi-file context."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Absolute or relative path to the source file(s) or directory to analyze. Supports comma-separated paths for multi-file context."),
		),
	)

	s.AddTool(tool, analyzeHandler)

	// Start stdio server
	log.Println("Starting A11ySentry MCP server on stdio...")
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func analyzeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathInput, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("argument 'path' is required: %v", err)), nil
	}

	paths := strings.Split(pathInput, ",")
	if len(paths) == 1 {
		p := strings.TrimSpace(paths[0])
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return analyzeDirectory(ctx, p)
		}
	}

	return analyzeFiles(ctx, paths, pathInput)
}

func analyzeDirectory(ctx context.Context, dir string) (*mcp.CallToolResult, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Could not resolve directory: %v", err)), nil
	}

	roots := scanner.FindProjectRoots(absDir)
	if len(roots) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("No supported project roots found in %s", absDir)), nil
	}

	var allViolations []domain.Violation
	for _, root := range roots {
		fw := scanner.Detect(root,
			nextjs.New(),
			astrofw.New(),
			nuxt.New(),
			sveltekit.New(),
			androidfw.New(),
			iosfw.New(),
			generic.New(),
		)

		uiFiles, cssFiles, err := fw.CollectFiles(root)
		if err != nil {
			log.Printf("Error scanning project root %s: %v", root, err)
			continue
		}

		importGraph := scanner.BuildImportGraph(uiFiles, fw, root)
		trees := fw.BuildPageTrees(uiFiles, importGraph, root)

		for _, tree := range trees {
			adapter, _ := getAdapterAndPlatform(tree.Files[0], fw.Name())
			if adapter == nil {
				continue
			}

			// Pre-load project CSS for color resolution (no-op if not a web adapter)
			web.LoadProjectCSS(adapter, append(cssFiles, tree.Files...))

			nodes, err := adapter.Ingest(ctx, tree.Files)
			if err != nil {
				log.Printf("Error ingesting tree %s: %v", tree.Label, err)
				continue
			}

			violations, _ := domain.NewAnalyzer().Analyze(ctx, nodes)
			allViolations = append(allViolations, violations...)
		}
	}

	if len(allViolations) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("✅ No accessibility violations found in directory %s.", dir)), nil
	}

	toonReport := domain.ToTOON(allViolations)
	return mcp.NewToolResultText(fmt.Sprintf("❌ Found %d accessibility violations in directory %s (TOON Format):\n\n%s", len(allViolations), dir, toonReport)), nil
}

func analyzeFiles(ctx context.Context, paths []string, originalInput string) (*mcp.CallToolResult, error) {
	var allNodes []domain.USN

	for _, p := range paths {
		p = strings.TrimSpace(p)
		adapter, _ := getAdapterAndPlatform(p, "")
		if adapter == nil {
			log.Printf("Skipping unsupported file: %s", p)
			continue
		}

		// 2. Ingest Source
		nodes, err := adapter.Ingest(ctx, []string{p})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error reading file %s: %v", p, err)), nil
		}
		allNodes = append(allNodes, nodes...)
	}

	if len(allNodes) == 0 {
		return mcp.NewToolResultError("No valid source files provided for analysis."), nil
	}

	analyzer := domain.NewAnalyzer()

	// 3. Perform Analysis on the combined tree
	violations, err := analyzer.Analyze(ctx, allNodes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Analysis failed: %v", err)), nil
	}

	// 4. Format Result (Optimized with TOON)
	if len(violations) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("✅ No accessibility violations found in %s.", originalInput)), nil
	}

	toonReport := domain.ToTOON(violations)
	return mcp.NewToolResultText(fmt.Sprintf("❌ Found %d accessibility violations in %s (TOON Format):\n\n%s", len(violations), originalInput, toonReport)), nil
}

func getAdapterAndPlatform(filePath, fwName string) (ports.Adapter, domain.Platform) {
	ext := strings.ToLower(filepath.Ext(filePath))
	fwName = strings.ToLower(fwName)

	if strings.Contains(fwName, "android") {
		return android.NewAndroidAdapter(), domain.PlatformAndroidCompose
	}
	if strings.Contains(fwName, "ios") {
		return ios.NewIOSAdapter(), domain.PlatformIOSSwiftUI
	}

	switch ext {
	case ".html", ".htm", ".astro":
		return web.NewHTMLAdapter(), domain.PlatformWebReact
	case ".kt", ".xml":
		return android.NewAndroidAdapter(), domain.PlatformAndroidCompose
	case ".swift":
		return ios.NewIOSAdapter(), domain.PlatformIOSSwiftUI
	case ".dart":
		return flutter.NewFlutterAdapter(), domain.PlatformFlutterDart
	case ".xaml", ".cs":
		return dotnet.NewDotNetAdapter(), domain.PlatformDotNetXAML
	case ".fxml":
		return javadesktop.NewJavaDesktopAdapter(), domain.PlatformJavaFX
	case ".js", ".jsx", ".ts", ".tsx":
		// Best effort for mobile vs web if no framework name provided
		if strings.Contains(filePath, "android") || strings.Contains(filePath, "ios") {
			return reactnative.NewReactNativeAdapter(), domain.PlatformReactNative
		}
		return web.NewHTMLAdapter(), domain.PlatformWebReact
	case ".razor":
		return blazor.NewBlazorAdapter(), domain.PlatformBlazor
	case ".prefab", ".unity":
		return unity.NewUnityAdapter(), domain.PlatformUnity
	case ".tscn":
		return godot.NewGodotAdapter(), domain.PlatformGodot
	case ".java":
		return android.NewAndroidAdapter(), domain.PlatformAndroidView
	default:
		return nil, ""
	}
}

