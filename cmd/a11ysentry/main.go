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
	"time"

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
	watchFlag := flag.Bool("watch", false, "Watch input files for changes and re-analyze automatically")
	flag.Parse()

	homeDir, _ := os.UserHomeDir()
	dbDir := filepath.Join(homeDir, ".a11ysentry")
	_ = os.MkdirAll(dbDir, 0755)
	repo, err := sqlite.NewSQLiteRepository(filepath.Join(dbDir, "history.db"))
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	var excludes []string
	if *excludeFlag != "" {
		for _, e := range strings.Split(*excludeFlag, ",") {
			if e = strings.TrimSpace(e); e != "" {
				excludes = append(excludes, e)
			}
		}
	}

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
		runProjectAnalysis(*dirFlag, *formatFlag, excludes, repo)
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
			runProjectAnalysis(args[0], *formatFlag, excludes, repo)
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
		runWatch(args, *formatFlag, *platformFlag, extraCSS, repo)
		return
	}

	allReports, hasErrors, hasWarnings := analyzeFiles(args, *formatFlag, *platformFlag, extraCSS, repo)
	printReports(allReports, *formatFlag)

	if hasErrors {
		os.Exit(1)
	} else if hasWarnings {
		os.Exit(2)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Project-aware analysis (--dir flag) — delegates to scanner package
// ─────────────────────────────────────────────────────────────────────────────

func runProjectAnalysis(dir, format string, excludes []string, repo ports.Repository) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not resolve directory: %v\n", err)
		os.Exit(1)
	}

	// Discover project roots (handles monorepos / multi-project directories).
	roots := scanner.FindProjectRoots(absDir, excludes...)
	if len(roots) == 0 {
		fmt.Fprintf(os.Stderr, "No supported project roots found in %s\n", absDir)
		os.Exit(1)
	}
	if len(roots) > 1 && format == "text" {
		fmt.Printf("A11ySentry -- Found %d project(s) in %s\n\n", len(roots), absDir)
	}

	var allReports []domain.ViolationReport
	hasErrors, hasWarnings := false, false

	for _, root := range roots {
		errs, warns, reports := analyzeProject(root, format, repo)
		allReports = append(allReports, reports...)
		if errs {
			hasErrors = true
		}
		if warns {
			hasWarnings = true
		}
	}

	if format != "text" {
		printReports(allReports, format)
	}

	if hasErrors {
		os.Exit(1)
	} else if hasWarnings {
		os.Exit(2)
	}
}

func analyzeProject(absDir, format string, repo ports.Repository) (hasErrors, hasWarnings bool, reports []domain.ViolationReport) {
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

	if format == "text" {
		fmt.Printf("A11ySentry -- Project: %s [%s]\n\n", absDir, fw.Name())
	}

	uiFiles, _, err := fw.CollectFiles(absDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning directory: %v\n", err)
		return
	}
	if len(uiFiles) == 0 {
		if format == "text" {
			fmt.Printf("  No supported UI files found — skipping.\n\n")
		}
		return
	}

	// 2. Build the import graph.
	importGraph := scanner.BuildImportGraph(uiFiles, fw, absDir)

	// 3. Build page trees (framework-specific).
	trees := fw.BuildPageTrees(uiFiles, importGraph, absDir)

	if format == "text" {
		fmt.Printf("  Found %d files, %d page tree(s)\n\n", len(uiFiles), len(trees))
	}

	// 4. Analyze each tree as a unit.
	for _, tree := range trees {
		if format == "text" {
			fmt.Printf("  Page: %s\n", tree.Label)
			for _, f := range tree.Files[1:] {
				fmt.Printf("     |-- %s\n", shortPath(f, absDir))
			}
		}

		// Choose the right adapter and platform based on the framework.
		var adapter ports.Adapter
		platform := domain.PlatformWebReact

		switch fw.Name() {
		case "Android":
			adapter = android.NewAndroidAdapter()
			platform = domain.PlatformAndroidCompose
		case "iOS":
			adapter = ios.NewIOSAdapter()
			platform = domain.PlatformIOSSwiftUI
		default:
			adapter = web.NewHTMLAdapter()
		}

		analyzer := domain.NewAnalyzer()
		for _, f := range tree.Files {
			usns, err := adapter.Ingest(context.Background(), []string{f})
			if err != nil {
				continue
			}
			violations, err := analyzer.Analyze(context.Background(), usns)
			if err != nil {
				continue
			}
			report := domain.ViolationReport{
				FilePath:   f,
				Platform:   platform,
				Violations: violations,
			}

			// Persistence.
			_ = repo.SaveReport(context.Background(), report)

			reports = append(reports, report)
			if reportHasErrors(report) {
				hasErrors = true
			}
			if reportHasWarnings(report) {
				hasWarnings = true
			}
		}
	}
	return
}

func analyzeFiles(paths []string, format, platformName string, extraCSS []string, repo ports.Repository) (reports []domain.ViolationReport, hasErrors, hasWarnings bool) {
	// Standard analysis (not project-aware) — falls back to Generic adapter.
	var adapter ports.Adapter
	platform := domain.PlatformWebReact

	// Map platform flag to internal domain types.
	if platformName != "" {
		switch strings.ToLower(platformName) {
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
		usns, err := adapter.Ingest(context.Background(), []string{absPath})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", p, err)
			continue
		}
		violations, err := analyzer.Analyze(context.Background(), usns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", p, err)
			continue
		}
		report := domain.ViolationReport{
			FilePath:   absPath,
			Platform:   platform,
			Violations: violations,
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

func printReports(reports []domain.ViolationReport, format string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(reports, "", "  ")
		fmt.Println(string(data))
	case "sarif":
		s := sarif.FromReports(reports)
		data, _ := json.MarshalIndent(s, "", "  ")
		fmt.Println(string(data))
	default:
		for _, r := range reports {
			fmt.Printf("%s\n", domain.ToTOON(r.Violations))
		}
	}
}

func runWatch(paths []string, format, platform string, extraCSS []string, repo ports.Repository) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, p := range paths {
		_ = watcher.Add(p)
	}

	fmt.Printf(titleStyle.Render("A11ySentry")+" -- Watching %d file(s) for changes...\n\n", len(paths))

	// Initial run.
	reports, _, _ := analyzeFiles(paths, format, platform, extraCSS, repo)
	printReports(reports, format)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Printf("\n[%s] File changed: %s\n", time.Now().Format("15:04:05"), event.Name)
				reports, _, _ := analyzeFiles([]string{event.Name}, format, platform, extraCSS, repo)
				printReports(reports, format)
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
	mcpFlags.Parse(args)

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
	cfg := map[string]any{
		"version":       "1.0",
		"paths":         []string{"src", "app", "pages", "components"},
		"exclude":       []string{"node_modules", ".git", "examples", "landing"},
		"format":        "sarif",
		"exitOnWarning": false,
		"platform":      "",
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	fmt.Println("Created a11ysentry.json")
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
        run: a11ysentry --format sarif --dir . > results.sarif
        continue-on-error: true
      - name: Upload SARIF to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
          category: a11ysentry
      - name: Run Accessibility Audit (Text / Exit Codes)
        run: a11ysentry --dir .
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
