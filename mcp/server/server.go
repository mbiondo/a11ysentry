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
	django_adapter "a11ysentry/adapters/django"
	flask_adapter "a11ysentry/adapters/flask"
	angular_adapter "a11ysentry/adapters/angular"
	vue_adapter "a11ysentry/adapters/vue"
	pyqt_adapter "a11ysentry/adapters/pyqt"
	electron_adapter "a11ysentry/adapters/electron"
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
	"a11ysentry/scanner/django"
	"a11ysentry/scanner/flask"
	"a11ysentry/scanner/angular"
	"a11ysentry/scanner/vue"
	dotnetfw "a11ysentry/scanner/dotnet"
	pyqtfw "a11ysentry/scanner/pyqt"
	electronfw "a11ysentry/scanner/electron"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	repo ports.Repository
}

func Start(repo ports.Repository) {
	srv := &MCPServer{repo: repo}

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

	s.AddTool(tool, srv.analyzeHandler)

	// Register the component context tool
	contextTool := mcp.NewTool("get_component_context",
		mcp.WithDescription("Get the architectural context of a specific component file. It returns the component's ancestors (layouts/parents) and children, helping you understand how it fits into the rendering hierarchy and avoid redundant props."),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Absolute path to the component file to investigate."),
		),
	)

	s.AddTool(contextTool, srv.contextHandler)

	// Register the audit history tool
	historyTool := mcp.NewTool("get_audit_history",
		mcp.WithDescription("Retrieve the history of past accessibility audits from the local database."),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of reports to retrieve (default: 10)."),
		),
	)

	s.AddTool(historyTool, srv.historyHandler)

	// Start stdio server
	log.Println("Starting A11ySentry MCP server on stdio...")
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func (srv *MCPServer) analyzeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathInput, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("argument 'path' is required: %v", err)), nil
	}

	paths := strings.Split(pathInput, ",")
	if len(paths) == 1 {
		p := strings.TrimSpace(paths[0])
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return srv.analyzeDirectory(ctx, p)
		}
	}

	return srv.analyzeFiles(ctx, paths, pathInput)
}

func (srv *MCPServer) analyzeDirectory(ctx context.Context, dir string) (*mcp.CallToolResult, error) {
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
			django.New(),
			flask.New(),
			angular.New(),
			vue.New(),
			dotnetfw.New(),
			pyqtfw.New(),
			electronfw.New(),
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

		cfg, _ := domain.LoadConfig(filepath.Join(root, "a11ysentry.json"))

		for _, tree := range trees {
			adapter, platform := getAdapterAndPlatform(tree.Root.FilePath, fw.Name())
			if adapter == nil {
				continue
			}

			// Pre-load project CSS for color resolution (no-op if not a web adapter)
			allFiles := tree.Root.Flatten()
			web.LoadProjectCSS(adapter, append(cssFiles, allFiles...))

			nodes, err := adapter.Ingest(ctx, tree.Root)
			if err != nil {
				log.Printf("Error ingesting tree %s: %v", tree.Label, err)
				continue
			}

			violations, _ := domain.NewAnalyzer().Analyze(ctx, nodes, cfg)
			allViolations = append(allViolations, violations...)

			// Persist to repository
			report := domain.ViolationReport{
				ProjectName: filepath.Base(root),
				ProjectRoot: root,
				FilePath:    tree.Label,
				Platform:    platform,
				Timestamp:   time.Now().Unix(),
				Violations:  violations,
			}
			_ = srv.repo.SaveReport(ctx, report)
		}
	}

	if len(allViolations) == 0 {
		return mcp.NewToolResultText("✅ No violations found."), nil
	}

	return mcp.NewToolResultText(domain.ToTOON(allViolations)), nil
}

func (srv *MCPServer) analyzeFiles(ctx context.Context, paths []string, originalInput string) (*mcp.CallToolResult, error) {
	var allViolations []domain.Violation
	analyzer := domain.NewAnalyzer()

	for _, p := range paths {
		p = strings.TrimSpace(p)
		absPath, err := filepath.Abs(p)
		if err != nil {
			log.Printf("Could not resolve absolute path for %s: %v", p, err)
			continue
		}
		// 1. Detect framework to use its collection/tree building logic
		dir := filepath.Dir(absPath)
		fw := scanner.Detect(dir,
			nextjs.New(),
			astrofw.New(),
			nuxt.New(),
			sveltekit.New(),
			django.New(),
			flask.New(),
			angular.New(),
			vue.New(),
			dotnetfw.New(),
			pyqtfw.New(),
			electronfw.New(),
			androidfw.New(),
			iosfw.New(),
			generic.New(),
		)

		// 2. Build a localized tree for this specific file
		// We treat the single file as a potential root and let the scanner find its children.
		uiFiles := []string{absPath}
		importGraph := scanner.BuildImportGraph(uiFiles, fw, dir)
		
		// We manually create a root node and try to resolve its children if it's a supported framework
		rootNode := &domain.FileNode{FilePath: absPath}
		
		// If the framework supports building trees, we try to "expand" this file's context
		// Note: This is a simplified version of project-aware analysis but focused on a single entry point.
		adapter, platform := getAdapterAndPlatform(absPath, fw.Name())
		if adapter == nil {
			log.Printf("Skipping unsupported file: %s", absPath)
			continue
		}

		// Try to populate children via the import graph
		if deps, ok := importGraph[absPath]; ok {
			for _, dep := range deps {
				childNode := &domain.FileNode{FilePath: dep}
				rootNode.Children = append(rootNode.Children, childNode)
			}
		}

		// 3. Load project config if nearby
		cfg, _ := domain.LoadConfig(filepath.Join(dir, "a11ysentry.json"))

		// 4. Ingest the tree (adapter handles the recursion if we provided children)
		nodes, err := adapter.Ingest(ctx, rootNode)
		if err != nil {
			log.Printf("Error ingesting file %s: %v", p, err)
			continue
		}

		violations, _ := analyzer.Analyze(ctx, nodes, cfg)
		allViolations = append(allViolations, violations...)

		// Persist to repository
		report := domain.ViolationReport{
			ProjectName: filepath.Base(dir),
			ProjectRoot: dir,
			FilePath:    absPath,
			Platform:    platform,
			Timestamp:   time.Now().Unix(),
			Violations:  violations,
		}
		_ = srv.repo.SaveReport(ctx, report)
	}

	if len(allViolations) == 0 {
		return mcp.NewToolResultText("✅ No violations found."), nil
	}

	return mcp.NewToolResultText(domain.ToTOON(allViolations)), nil
}

func (srv *MCPServer) contextHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathInput, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("argument 'path' is required: %v", err)), nil
	}

	absPath, err := filepath.Abs(pathInput)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Could not resolve absolute path: %v", err)), nil
	}

	// 1. Find project root
	root := scanner.FindProjectRoot(absPath)
	if root == "" {
		root = filepath.Dir(absPath)
	}

	// 2. Detect framework
	fw := scanner.Detect(root,
		nextjs.New(), astrofw.New(), nuxt.New(), sveltekit.New(),
		django.New(), flask.New(), angular.New(), vue.New(),
		dotnetfw.New(), pyqtfw.New(), electronfw.New(), androidfw.New(),
		iosfw.New(), generic.New(),
	)

	// 3. Scan and build trees
	uiFiles, _, err := fw.CollectFiles(root)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error collecting files: %v", err)), nil
	}

	importGraph := scanner.BuildImportGraph(uiFiles, fw, root)
	trees := fw.BuildPageTrees(uiFiles, importGraph, root)

	// 4. Find trees containing the target path
	var relatedNodes []*domain.FileNode
	for _, tree := range trees {
		found := findNodeInTree(tree.Root, absPath)
		if found != nil {
			relatedNodes = append(relatedNodes, found)
		}
	}

	if len(relatedNodes) == 0 {
		// If not in a tree, try standalone context
		standalone := &domain.FileNode{FilePath: absPath}
		if deps, ok := importGraph[absPath]; ok {
			for _, dep := range deps {
				standalone.Children = append(standalone.Children, &domain.FileNode{FilePath: dep})
			}
		}
		return mcp.NewToolResultText(domain.HierarchyToTOON(standalone, absPath)), nil
	}

	// Merge all related trees into a virtual root for TOON output
	virtualRoot := &domain.FileNode{
		FilePath: "Context for " + filepath.Base(absPath),
		Children: relatedNodes,
	}

	return mcp.NewToolResultText(domain.HierarchyToTOON(virtualRoot, absPath)), nil
}

func (srv *MCPServer) historyHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := request.GetInt("limit", 10)

	history, err := srv.repo.GetHistory(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Error retrieving history: %v", err)), nil
	}

	if len(history) == 0 {
		return mcp.NewToolResultText("No audit history found."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Audit History (last %d reports):\n\n", len(history)))
	for _, report := range history {
		t := time.Unix(report.Timestamp, 0).Format("2006-01-02 15:04:05")
		status := "✅ PASS"
		if len(report.Violations) > 0 {
			status = fmt.Sprintf("❌ %d violations", len(report.Violations))
		}
		fmt.Fprintf(&sb, "[%s] %s (%s) - %s\n", t, report.FilePath, report.Platform, status)
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func findNodeInTree(root *domain.FileNode, target string) *domain.FileNode {
	if root == nil {
		return nil
	}
	if root.FilePath == target {
		// Return a copy to avoid mutating the original tree
		return &domain.FileNode{
			FilePath: root.FilePath,
			Children: root.Children,
			IsCycle:  root.IsCycle,
		}
	}
	for _, child := range root.Children {
		found := findNodeInTree(child, target)
		if found != nil {
			// We return a copy that preserves the structural context but focuses on the target
			return &domain.FileNode{
				FilePath: root.FilePath,
				Children: []*domain.FileNode{found},
			}
		}
	}
	return nil
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
	}

	if fwName == "django" {
		return django_adapter.NewDjangoAdapter(), domain.Platform("django")
	}
	if fwName == "flask" {
		return flask_adapter.NewFlaskAdapter(), domain.Platform("flask")
	}
	if fwName == "angular" {
		return angular_adapter.NewAngularAdapter(), domain.Platform("angular")
	}
	if fwName == "vue" {
		return vue_adapter.NewVueAdapter(), domain.Platform("vue")
	}
	if fwName == "pyqt" {
		return pyqt_adapter.NewPyQtAdapter(), domain.Platform("pyqt")
	}
	if fwName == "electron" {
		return electron_adapter.NewElectronAdapter(), domain.Platform("electron")
	}

	return nil, ""
}
