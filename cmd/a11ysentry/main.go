package main

import (
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
	"a11ysentry/cmd/a11ysentry/internal/sarif"
	"a11ysentry/cmd/a11ysentry/internal/tui"
	"a11ysentry/engine/core/domain"
	"a11ysentry/engine/core/ports"
	"a11ysentry/engine/persistence/sqlite"
	"a11ysentry/mcp/registration"
	"a11ysentry/mcp/server"
	"a11ysentry/scanner"
	androidfw "a11ysentry/scanner/android"
	astrofw "a11ysentry/scanner/astro"
	"a11ysentry/scanner/generic"
	iosfw "a11ysentry/scanner/ios"
	"a11ysentry/scanner/nextjs"
	"a11ysentry/scanner/nuxt"
	sveltekit "a11ysentry/scanner/sveltekit"

	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fsnotify/fsnotify"
)

var (
	Version    = "0.0.1"
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true)
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("A11ySentry version %s\n", Version)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		handleMCPSubcommand(os.Args[2:])
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		handleInitSubcommand(os.Args[2:])
		return
	}

	tuiFlag := flag.Bool("tui", false, "Start the interactive TUI dashboard")
	formatFlag := flag.String("format", "text", "Output format: text, json, json-ld, sarif")
	platformFlag := flag.String("platform", "", "Force platform: react, vue, svelte, angular, astro, android, ios, flutter, dotnet, reactnative, blazor, unity, godot, electron, tauri")
	dirFlag := flag.String("dir", "", "Analyze a full project directory, resolving component trees automatically")
	excludeFlag := flag.String("exclude", "", "Comma-separated list of directories to exclude from analysis")
	cssFlag := flag.String("css", "", "Comma-separated list of external CSS/SCSS files to pre-load (for color resolution in single-file mode)")
	outputFlag := flag.String("output", "", "Save results to a specific file (e.g. results.txt). If empty in project mode, defaults to date_project.txt")
	watchFlag := flag.Bool("watch", false, "Watch input files for changes and re-analyze automatically")
	flag.Parse()

	// Load configuration
	configPath := "a11ysentry.json"
	cfg, err := domain.LoadConfig(configPath)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading config file: %v", err)
	}

	// Merge flags with config (Flags take precedence if explicitly set)
	if *formatFlag != "text" {
		cfg.Format = *formatFlag
	}
	if *platformFlag != "" {
		cfg.Platform = *platformFlag
	}
	if *excludeFlag != "" {
		cfg.Exclude = append(cfg.Exclude, strings.Split(*excludeFlag, ",")...)
	}

	homeDir, _ := os.UserHomeDir()
	dbDir := filepath.Join(homeDir, ".a11ysentry")
	_ = os.MkdirAll(dbDir, 0755)
	repo, err := sqlite.NewSQLiteRepository(filepath.Join(dbDir, "history.db"))
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	projectRoot, _ := detectProjectRoot()
	startTime := time.Now()

	if *tuiFlag {
		m := tui.NewMainModel(repo)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("TUI Error: %v", err)
		}
		return
	}

	// --dir: project-aware analysis delegated entirely to the scanner.
	if *dirFlag != "" {
		runProjectAnalysis(*dirFlag, cfg, repo, startTime, *outputFlag)
		return
	}

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	// Auto-detect: if the single positional arg is a directory, delegate to project analysis.
	if len(args) == 1 {
		if info, err := os.Stat(args[0]); err == nil && info.IsDir() {
			runProjectAnalysis(args[0], cfg, repo, startTime, *outputFlag)
			return
		}
	}

	var extraCSS []string
	if *cssFlag != "" {
		for _, f := range strings.Split(*cssFlag, ",") {
			if f = strings.TrimSpace(f); f != "" {
				extraCSS = append(extraCSS, f)
			}
		}
	}

	if *watchFlag {
		runWatch(args, cfg, extraCSS, repo, projectRoot)
		return
	}

	allReports, hasErrors, hasWarnings := analyzeFiles(args, cfg, extraCSS, repo)
	printReports(allReports, cfg.Format, projectRoot, *outputFlag)

	if cfg.Format == "text" {
		elapsed := time.Since(startTime)
		fmt.Printf("\nAnalysis completed in %v. Total files analyzed: %d.\n", elapsed.Round(time.Millisecond), len(args))
	}

	if hasErrors {
		os.Exit(1)
	} else if hasWarnings {
		os.Exit(2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Project-aware analysis (--dir flag) — delegates to scanner package
// ─────────────────────────────────────────────────────────────────────────────

func runProjectAnalysis(dir string, cfg domain.ProjectConfig, repo ports.Repository, startTime time.Time, outputFilePath string) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve directory: %v\n", err)
		os.Exit(1)
	}

	projectName := filepath.Base(absDir)

	// Discover project roots (handles monorepos / multi-project directories).
	roots := scanner.FindProjectRoots(absDir, cfg.Exclude...)
	if len(roots) == 0 {
		fmt.Fprintf(os.Stderr, "No supported project roots found in %s\n", absDir)
		os.Exit(1)
	}
	if cfg.Format == "text" {
		if len(roots) > 1 {
			fmt.Printf("A11ySentry -- Found %d project(s) in %s\n\n", len(roots), absDir)
		} else {
			// For single project, we still want the label if we are in text mode
			fmt.Printf("A11ySentry -- Project: %s\n", absDir)
		}
	}

	var allReports []domain.ViolationReport
	hasErrors, hasWarnings := false, false
	totalFiles := 0

	for _, root := range roots {
		errs, warns, reports, fileCount := analyzeProject(root, cfg, repo)
		allReports = append(allReports, reports...)
		totalFiles += fileCount
		if errs {
			hasErrors = true
		}
		if warns {
			hasWarnings = true
		}
	}

	// Determine output file path if not provided
	if cfg.Format == "text" && outputFilePath == "" {
		date := time.Now().Format("2006-01-02")
		outputFilePath = fmt.Sprintf("%s_%s.txt", date, projectName)
	}

	printReports(allReports, cfg.Format, absDir, outputFilePath)

	if cfg.Format == "text" && outputFilePath != "" {
		elapsed := time.Since(startTime)
		fmt.Printf("\nAnalysis completed in %v. Total files analyzed: %d.\n", elapsed.Round(time.Millisecond), totalFiles)
		fmt.Printf("Results saved to: %s\n", outputFilePath)
	}

	if hasErrors {
		os.Exit(1)
	} else if hasWarnings {
		os.Exit(2)
	}
}

func analyzeProject(absDir string, cfg domain.ProjectConfig, repo ports.Repository) (hasErrors, hasWarnings bool, reports []domain.ViolationReport, fileCount int) {
	// 1. Detect framework and collect files.
	fw := scanner.Detect(absDir,
		nextjs.New(),
		astrofw.New(),
		nuxt.New(),
		sveltekit.New(),
		androidfw.New(),
		iosfw.New(),
		generic.New(),
	)

	uiFiles, cssFiles, err := fw.CollectFiles(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		return
	}
	if len(uiFiles) == 0 {
		if cfg.Format == "text" {
			fmt.Printf("  No supported UI files found — skipping.\n\n")
		}
		return
	}
	fileCount = len(uiFiles)

	// 2. Build the import graph.
	importGraph := scanner.BuildImportGraph(uiFiles, fw, absDir)

	// 3. Build page trees (framework-specific).
	trees := fw.BuildPageTrees(uiFiles, importGraph, absDir)

	if cfg.Format == "text" {
		fmt.Printf("  Found %d files, %d page tree(s)\n\n", len(uiFiles), len(trees))
	}

	// 4. Analyze each tree as a unit.
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	totalTrees := len(trees)
	analyzedTrees := 0
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 60

	for _, t := range trees {
		wg.Add(1)
		go func(tree scanner.PageTree) {
			defer wg.Done()

			// Choose the right adapter and platform based on the framework.
			var adapter ports.Adapter
			platform := domain.PlatformWebReact

			switch fw.Name() {
			case "Android (Kotlin/Java)":
				adapter = android.NewAndroidAdapter()
				platform = domain.PlatformAndroidCompose
			case "iOS (Swift/SwiftUI)":
				adapter = ios.NewIOSAdapter()
				platform = domain.PlatformIOSSwiftUI
			default:
				adapter = web.NewHTMLAdapter()
			}

			// Pre-load CSS from the whole tree for web frameworks.
			if fw.Name() != "Android (Kotlin/Java)" && fw.Name() != "iOS (Swift/SwiftUI)" {
				allFiles := tree.Root.Flatten()
				web.LoadProjectCSS(adapter, append(cssFiles, allFiles...))
			}

			// Ingest the entire tree as a single analysis unit.
			usns, err := adapter.Ingest(context.Background(), tree.Root)
			if err != nil {
				return
			}

			analyzer := domain.NewAnalyzer()
			violations, err := analyzer.Analyze(context.Background(), usns, cfg)
			if err != nil {
				return
			}

			report := domain.ViolationReport{
				ProjectName: filepath.Base(absDir),
				ProjectRoot: absDir,
				FilePath:    tree.Label,
				Platform:    platform,
				Violations:  violations,
			}

			// Persistence.
			_ = repo.SaveReport(context.Background(), report)

			mu.Lock()
			analyzedTrees++
			if cfg.Format == "text" {
				// \r moves to start of line, \033[K clears the rest of the line
				fmt.Printf("\r\033[K  Analyzing: %s %s (%d/%d)", 
					tree.Label,
					prog.ViewAs(float64(analyzedTrees)/float64(totalTrees)),
					analyzedTrees, totalTrees)
			}

			reports = append(reports, report)
			if reportHasErrors(report) {
				hasErrors = true
			}
			if reportHasWarnings(report) {
				hasWarnings = true
			}
			mu.Unlock()
		}(t)
	}
	wg.Wait()
	if cfg.Format == "text" {
		fmt.Println() // New line after progress finishes
	}
	return
}

func analyzeFiles(paths []string, cfg domain.ProjectConfig, extraCSS []string, repo ports.Repository) (reports []domain.ViolationReport, hasErrors, hasWarnings bool) {
	// Standard analysis (not project-aware) — falls back to Generic adapter.
	var adapter ports.Adapter
	platform := domain.PlatformWebReact

	// Map platform flag to internal domain types.
	if cfg.Platform != "" {
		switch strings.ToLower(cfg.Platform) {
		case "android":
			adapter = android.NewAndroidAdapter()
			platform = domain.PlatformAndroidCompose
		case "ios":
			adapter = ios.NewIOSAdapter()
			platform = domain.PlatformIOSSwiftUI
		case "flutter":
			adapter = flutter.NewFlutterAdapter()
			platform = domain.PlatformFlutterDart
		case "reactnative":
			adapter = reactnative.NewReactNativeAdapter()
			platform = domain.PlatformReactNative
		case "dotnet":
			adapter = dotnet.NewDotNetAdapter()
			platform = domain.PlatformDotNetXAML
		case "blazor":
			adapter = blazor.NewBlazorAdapter()
			platform = domain.PlatformBlazor
		case "unity":
			adapter = unity.NewUnityAdapter()
			platform = domain.PlatformUnity
		case "godot":
			adapter = godot.NewGodotAdapter()
			platform = domain.PlatformGodot
		case "javadesktop":
			adapter = javadesktop.NewJavaDesktopAdapter()
			platform = domain.PlatformJavaFX
		default:
			adapter = web.NewHTMLAdapter()
		}
	} else {
		adapter = web.NewHTMLAdapter()
	}

	analyzer := domain.NewAnalyzer()
	for _, p := range paths {
		absPath, _ := filepath.Abs(p)
		rootNode := &domain.FileNode{FilePath: absPath}

		// Try to find a project root for this file to provide better TUI grouping
		pRoot := filepath.Dir(absPath)
		pName := filepath.Base(pRoot)
		if roots := scanner.FindProjectRoots(pRoot); len(roots) > 0 {
			pRoot = roots[0]
			pName = filepath.Base(pRoot)
		}

		usns, err := adapter.Ingest(context.Background(), rootNode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", p, err)
			continue
		}
		violations, err := analyzer.Analyze(context.Background(), usns, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", p, err)
			continue
		}
		report := domain.ViolationReport{
			ProjectName: pName,
			ProjectRoot: pRoot,
			FilePath:    absPath,
			Platform:    platform,
			Violations:  violations,
		}

		_ = repo.SaveReport(context.Background(), report)
		reports = append(reports, report)
		if reportHasErrors(report) {
			hasErrors = true
		}
		if reportHasWarnings(report) {
			hasWarnings = true
		}
	}
	return
}

func reportHasErrors(r domain.ViolationReport) bool {
	for _, v := range r.Violations {
		if v.Severity == domain.SeverityError {
			return true
		}
	}
	return false
}

func reportHasWarnings(r domain.ViolationReport) bool {
	for _, v := range r.Violations {
		if v.Severity == domain.SeverityWarning {
			return true
		}
	}
	return false
}

func printReports(reports []domain.ViolationReport, format, projectRoot, outputFilePath string) {
	var out []byte
	var err error

	switch format {
	case "json":
		out, err = json.MarshalIndent(reports, "", "  ")
	case "sarif":
		s := sarif.FromReports(reports)
		out, err = json.MarshalIndent(s, "", "  ")
	default:
		// Text mode handles its own output logic below
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling results: %v\n", err)
		return
	}

	if format != "text" {
		if outputFilePath != "" {
			if err := os.WriteFile(outputFilePath, out, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing results to %s: %v\n", outputFilePath, err)
			}
		} else {
			fmt.Println(string(out))
		}
		return
	}

	// Text mode logic
	var b strings.Builder
	for _, r := range reports {
		if len(r.Violations) > 0 {
			fmt.Fprintf(&b, "\nPage: %s\n", shortPath(r.FilePath, projectRoot))
			b.WriteString(domain.ToESLintStyle(r.Violations, projectRoot))
		}
	}

	if outputFilePath != "" {
		header := fmt.Sprintf("A11ySentry Analysis Report - %s\n", time.Now().Format(time.RFC1123))
		separator := strings.Repeat("=", 60) + "\n\n"
		final := header + separator + b.String()
		if err := os.WriteFile(outputFilePath, []byte(final), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing results to %s: %v\n", outputFilePath, err)
		}
	} else {
		fmt.Print(b.String())
	}
}

func runWatch(paths []string, cfg domain.ProjectConfig, extraCSS []string, repo ports.Repository, projectRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close() //nolint:errcheck

	for _, p := range paths {
		_ = watcher.Add(p)
	}

	fmt.Printf(titleStyle.Render("A11ySentry")+" -- Watching %d file(s) for changes...\n\n", len(paths))

	// Initial run.
	reports, _, _ := analyzeFiles(paths, cfg, extraCSS, repo)
	printReports(reports, cfg.Format, projectRoot, "")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Printf("\n[%s] File changed: %s\n", time.Now().Format("15:04:05"), event.Name)
				reports, _, _ := analyzeFiles([]string{event.Name}, cfg, extraCSS, repo)
				printReports(reports, cfg.Format, projectRoot, "")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func shortPath(path, root string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

// ─────────────────────────────────────────────────────────────────────────────
// MCP subcommands
// ─────────────────────────────────────────────────────────────────────────────

func handleMCPSubcommand(args []string) {
	mcpFlags := flag.NewFlagSet("mcp", flag.ExitOnError)
	registerFlag := mcpFlags.Bool("register", false, "Register A11ySentry in AI Agents (Claude, Cursor, etc.)")
	binaryPath := mcpFlags.String("path", "", "Path to the a11ysentry binary (defaults to current executable)")
	if err := mcpFlags.Parse(args); err != nil {
		log.Fatal(err)
	}

	if *registerFlag {
		if *binaryPath == "" {
			exe, _ := os.Executable()
			*binaryPath = exe
		}

		p := tea.NewProgram(tui.MultiSelectModel{
			Title: "Register A11ySentry in AI Agents",
			Choices: []tui.Choice{
				{Label: "Claude Desktop"},
				{Label: "Cursor IDE"},
				{Label: "VSCode (Cline/Roo-Code)"},
				{Label: "Gemini CLI MCP"},
				{Label: "Gemini CLI Skill"},
				{Label: "Qwen"},
				{Label: "OpenCode"},
			},
		})
		m, err := p.Run()
		if err != nil {
			log.Fatalf("Error running registration TUI: %v", err)
		}

		if m, ok := m.(tui.MultiSelectModel); ok && m.Finished {
			if len(m.Choices) == 0 {
				fmt.Println("No agents selected for registration.")
				return
			}

			// Try to detect repo root for skill installation
			wd, _ := os.Getwd()

			fmt.Printf("\nRegistering A11ySentry MCP from %s...\n", *binaryPath)
			for _, choice := range m.Choices {
				if !choice.Selected {
					continue
				}

				var err error
				switch choice.Label {
				case "Claude Desktop":
					err = registration.RegisterClaude(*binaryPath)
				case "Cursor IDE":
					err = registration.RegisterCursor(*binaryPath)
				case "VSCode (Cline/Roo-Code)":
					err = registration.RegisterVSCode(*binaryPath)
				case "Gemini CLI MCP":
					err = registration.RegisterGemini(*binaryPath)
				case "Gemini CLI Skill":
					err = registration.RegisterSkill(wd)
				case "Qwen":
					err = registration.RegisterQwen(*binaryPath)
				case "OpenCode":
					err = registration.RegisterOpenCode(*binaryPath)
				}

				if err != nil {
					fmt.Printf("  [!] Error registering in %s: %v\n", choice.Label, err)
				} else {
					fmt.Printf("  [✓] Registered in %s\n", choice.Label)
				}
			}
			fmt.Println("\nRegistration complete.")
			return
		}
	}

	server.Start()
}

// ─────────────────────────────────────────────────────────────────────────────
// Init subcommand
// ─────────────────────────────────────────────────────────────────────────────

func handleInitSubcommand(args []string) {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	forceFlag := initFlags.Bool("force", false, "Overwrite existing files without prompting")
	skipHooksFlag := initFlags.Bool("skip-hooks", false, "Skip git pre-commit hook setup")
	skipActionsFlag := initFlags.Bool("skip-actions", false, "Skip GitHub Actions workflow creation")
	_ = initFlags.Parse(args)

	// Interactive mode if no flags provided
	useActions := !*skipActionsFlag
	useHooks := !*skipHooksFlag
	force := *forceFlag

	if len(args) == 0 {
		fmt.Println(titleStyle.Render("A11ySentry Project Initialization") + "\n")

		// Ask about GitHub Actions
		p1 := tea.NewProgram(tui.PromptYesNo{Question: "Setup GitHub Actions workflow?"})
		m1, _ := p1.Run()
		if res, ok := m1.(tui.PromptYesNo); ok && res.Finished {
			useActions = res.Result
		} else {
			return
		}

		// Ask about Git Hooks
		p2 := tea.NewProgram(tui.PromptYesNo{Question: "Setup Git pre-commit hook?"})
		m2, _ := p2.Run()
		if res, ok := m2.(tui.PromptYesNo); ok && res.Finished {
			useHooks = res.Result
		} else {
			return
		}
	}

	projectRoot, err := detectProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nProject root: %s\n", projectRoot)

	if err := createConfigFile(projectRoot, force); err != nil {
		fmt.Fprintf(os.Stderr, "Config: %v\n", err)
	}
	if useActions {
		if err := createGitHubActions(projectRoot, force); err != nil {
			fmt.Fprintf(os.Stderr, "GitHub Actions: %v\n", err)
		}
	}
	if useHooks {
		if err := createPreCommitHook(projectRoot, force); err != nil {
			fmt.Fprintf(os.Stderr, "Pre-commit hook: %v\n", err)
		}
	}

	fmt.Println("\nA11ySentry initialized! Your project is ready for accessibility CI/CD.")
	fmt.Println("  Run `a11ysentry --dir .` to test locally")
	fmt.Println("  Push to GitHub to trigger the Actions workflow")
}

func detectProjectRoot() (string, error) {
	out, err := execCommand("git", "rev-parse", "--show-toplevel").CombinedOutput()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine project root: %w", err)
	}
	fmt.Println("Not a git repository. Using current directory.")
	return wd, nil
}

var execCommand = exec.Command

func createConfigFile(root string, force bool) error {
	path := filepath.Join(root, "a11ysentry.json")
	if !force && fileExists(path) {
		fmt.Println("a11ysentry.json already exists, skipping (use --force to overwrite)")
		return nil
	}
	
	// Use the domain's default config to ensure consistency
	cfg := domain.DefaultConfig()
	cfg.Format = "sarif" // Overriding default for new projects as discussed
	
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	fmt.Println("Created a11ysentry.json with default rules.")
	return nil
}

func createGitHubActions(root string, force bool) error {
	workflowDir := filepath.Join(root, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0755); err != nil {
		return fmt.Errorf("creating workflow dir: %w", err)
	}
	path := filepath.Join(workflowDir, "a11y.yml")
	if !force && fileExists(path) {
		fmt.Println(".github/workflows/a11y.yml already exists, skipping (use --force to overwrite)")
		return nil
	}
	content := `name: Accessibility Audit
on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]
permissions:
  security-events: write
  contents: read
jobs:
  a11y:
    name: A11ySentry Analysis
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install A11ySentry
        run: curl -sSL https://raw.githubusercontent.com/mbiondo/a11ysentry/main/install.sh | bash
      - name: Run Accessibility Audit (SARIF)
        run: a11ysentry --format sarif . > results.sarif
        continue-on-error: true
      - name: Upload SARIF to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
          category: a11ysentry
      - name: Run Accessibility Audit (Text / Exit Codes)
        run: a11ysentry .
        continue-on-error: true
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing workflow: %w", err)
	}
	fmt.Println("Created .github/workflows/a11y.yml")
	return nil
}

func createPreCommitHook(root string, force bool) error {
	hookDir := filepath.Join(root, ".git", "hooks")
	if !dirExists(hookDir) {
		fmt.Println("Not a git repository — skipping pre-commit hook.")
		return nil
	}
	path := filepath.Join(hookDir, "pre-commit")
	if !force && fileExists(path) {
		fmt.Println(".git/hooks/pre-commit already exists, skipping (use --force to overwrite)")
		return nil
	}
	content := `#!/bin/sh
# A11ySentry pre-commit hook
echo "A11ySentry: checking staged files..."
files=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(html|htm|astro|vue|svelte|tsx|jsx|ts|js|razor|kt|xml|dart|swift|xaml|cs|fxml|java)$' || true)
if [ -z "$files" ]; then
  echo "  No relevant files staged. Skipping."
  exit 0
fi
# Note: users can customize exclusions here if needed
echo "$files" | xargs a11ysentry 2>&1
exit_code=$?
if [ $exit_code -eq 1 ]; then
  echo "ERROR: Accessibility errors found. Commit aborted."
  exit 1
elif [ $exit_code -eq 2 ]; then
  echo "WARNING: Accessibility warnings found. Commit allowed."
fi
exit 0
`
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return fmt.Errorf("writing hook: %w", err)
	}
	fmt.Println("Created .git/hooks/pre-commit hook")
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func printUsage() {
	fmt.Printf("A11ySentry version %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  a11ysentry init                  Scaffold CI/CD: GitHub Actions, pre-commit hook, config")
	fmt.Println("  a11ysentry <file1> <file2>        Analyze files directly")
	fmt.Println("  a11ysentry --dir ./src            Analyze a full project directory")
	fmt.Println("  a11ysentry --exclude node_modules Comma-separated list of directories to skip")
	fmt.Println("  a11ysentry --format json          Output results in JSON")
	fmt.Println("  a11ysentry --format sarif         Output SARIF (GitHub Code Scanning)")
	fmt.Println("  a11ysentry --css style.css        Pre-load external CSS for color resolution")
	fmt.Println("  a11ysentry --watch                Re-analyze on file changes")
	fmt.Println("  a11ysentry --tui                  Open the TUI dashboard")
	fmt.Println("  a11ysentry --platform vue         Force platform (react, vue, svelte, angular, astro, android, ios, ...)")
	fmt.Println("  a11ysentry mcp                   Start MCP server")
	fmt.Println("  a11ysentry --version, -v          Show version")
}
